package feed

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slog"
)

type FeedError struct {
	Message string
	Code    int
}

func (m *FeedError) Error() string {
	return m.Message
}

type FeedItemType int

const (
	Text = iota
	Image
	Binary
)

type Feed struct {
	Name   string     `json:"name"`
	Secret string     `json:"secret"`
	Items  []FeedItem `json:"items"`

	path string
}
type FeedItem struct {
	Name string       `json:"name"`
	Date time.Time    `json:"date"`
	Type FeedItemType `json:"type"`
	Feed *Feed        `json:"-"`
}

func NewFeed(basePath string, feedName string) (*Feed, error) {
	slog.Info("Creating new feed", slog.String("feed", feedName))
	feedPath := path.Join(basePath, feedName)
	_, err := os.Stat(feedPath)
	if err == nil {
		return nil, &FeedError{
			Code:    400,
			Message: "Feed already exists",
		}
	}
	os.Mkdir(feedPath, 0700)
	secret := uuid.NewString()
	err = os.WriteFile(path.Join(basePath, feedName, "secret"), []byte(secret), 0600)
	if err != nil {
		log.Errorf("Unable to write file %s", err.Error())
		return nil, err
	}
	return &Feed{
		Name:   feedName,
		Secret: secret,
		Items:  []FeedItem{},
		path:   path.Join(basePath, feedName),
	}, nil
}

func GetFeed(basePath string, feedName string, secret string) (*Feed, error) {
	feedPath := path.Join(basePath, feedName)

	feedLog := slog.Default().With(slog.String("feed", feedName))

	feedLog.Debug("Getting feed", slog.Int("secret_len", len(secret)))

	if _, err := os.Stat(feedPath); os.IsNotExist(err) {
		return nil, &FeedError{
			Code:    404,
			Message: "Feed does not exists",
		}
	}

	if secret == "" {
		code := 401
		feedLog.Error("No secret was provided", slog.Int("return", code))
		return nil, &FeedError{
			Code:    code,
			Message: "Unauthorized",
		}
	}

	if len(secret) != 4 {
		feedSecret, err := os.ReadFile(path.Join(feedPath, "secret"))
		code := 500
		if err != nil {
			feedLog.Error("Unable to read secret", slog.Int("return", code))
			return nil, &FeedError{
				Code:    code,
				Message: err.Error(),
			}
		}

		if string(feedSecret) != secret {
			code := 401
			feedLog.Error("Invalid secret", slog.Int("return", code))
			return nil, &FeedError{
				Code:    code,
				Message: "Authentication failed",
			}
		}
		secret = string(feedSecret)
	} else {
		stat, err := os.Stat(path.Join(feedPath, "pin"))
		if err != nil {
			if os.IsNotExist(err) {
				code := 401
				feedLog.Error("No PIN configured", slog.Int("return", code))
				return nil, &FeedError{
					Code:    code,
					Message: "No PIN configured",
				}
			} else {
				code := 500
				feedLog.Error("Unable to read PIN", slog.Int("return", code))
				return nil, &FeedError{
					Code:    code,
					Message: err.Error(),
				}
			}
		}

		maxTime := stat.ModTime().Add(2 * time.Minute)
		if maxTime.Before(time.Now()) {
			code := 401
			feedLog.Warn("PIN expired", slog.Int("return", code))
			os.Remove(path.Join(feedPath, "pin"))
			return nil, &FeedError{
				Code:    code,
				Message: "Authentication failed",
			}
		}

		s, err := os.ReadFile(path.Join(feedPath, "pin"))

		if err != nil {
			code := 500
			feedLog.Error("Unable to read secret", slog.Int("return", code))
			return nil, &FeedError{
				Code:    code,
				Message: err.Error(),
			}
		}
		if string(s) != secret {
			code := 401
			feedLog.Warn("PIN Incorrect", slog.Int("return", code))
			return nil, &FeedError{
				Code:    code,
				Message: "Authentication failed",
			}
		}
	}

	var d []fs.DirEntry
	var err error
	if d, err = os.ReadDir(feedPath); err != nil {
		code := 500
		feedLog.Error("Unable to feed content", slog.Int("return", code))

		return nil, &FeedError{
			Code:    code,
			Message: "Unable to open directory for read",
		}
	}

	result := Feed{
		Name:   feedName,
		Secret: secret,
		path:   path.Join(basePath, feedName),
	}

	items := []FeedItem{}
	for _, f := range d {
		if f.Name() == "secret" || f.Name() == "pin" {
			continue
		}
		info, err := f.Info()
		if err != nil {
			code := 500
			e := "Unable to read file info"

			feedLog.Error(e, slog.Int("return", code))

			return nil, &FeedError{
				Code:    code,
				Message: e,
			}
		}

		var itemType FeedItemType
		if strings.HasSuffix(f.Name(), ".txt") {
			itemType = Text
		} else if strings.HasSuffix(f.Name(), ".png") || strings.HasSuffix(f.Name(), ".jpg") {
			itemType = Image
		} else {
			itemType = Binary
		}
		items = append(items, FeedItem{
			Name: f.Name(),
			Date: info.ModTime(),
			Type: itemType,
			Feed: &result,
		})
	}
	sort.Slice(items, func(i, j2 int) bool {
		return items[i].Date.After(items[j2].Date)
	})

	result.Items = items

	return &result, nil
}

func (feed *Feed) GetItem(item string) ([]byte, error) {
	// Read item content
	slog.Info("Getting Item", slog.String("feed", feed.Name), slog.String("name", item))
	var content []byte
	filePath := path.Join(feed.path, item)
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &FeedError{
				Code:    404,
				Message: fmt.Sprintf("File does not exists (%s)", filePath),
			}
		}
		return nil, &FeedError{
			Code:    500,
			Message: fmt.Sprintf("Unable to open file '%s' for read", filePath),
		}
	}
	return content, nil
}
func (feed *Feed) AddItem(contentType string, f io.ReadCloser) error {
	slog.Debug("Adding Item", slog.String("feed", feed.Name), slog.String("content-type", contentType))
	fileExtensions := map[string]string{
		"image/png":  "png",
		"image/jpeg": "jpg",
		"text/plain": "txt",
	}

	filenameTemplate := map[string]string{
		"image/png":  "Pasted Image",
		"image/jpeg": "Pasted Image",
		"text/plain": "Pasted Text",
	}

	ext := fileExtensions[contentType]
	template := filenameTemplate[contentType]

	if len(ext) == 0 {
		return &FeedError{
			Code:    400,
			Message: "Content-type not accepted",
		}
	}

	content, err := io.ReadAll(f)
	if err != nil {
		_, ok := err.(*http.MaxBytesError)
		if ok {
			return &FeedError{
				Code:    413,
				Message: "Max body size exceeded",
			}
		}
		return &FeedError{
			Code:    500,
			Message: "Unable to read stream",
		}
	}

	if len(content) == 0 {
		return &FeedError{
			Code:    500,
			Message: "Content is empty",
		}
	}

	fileIndex := 1
	var filename string
	for {
		filename = fmt.Sprintf("%s %d", template, fileIndex)
		matches, err := filepath.Glob(path.Join(feed.path, filename) + ".*")
		if err != nil {
			return &FeedError{
				Code:    500,
				Message: "Unable to read feed content",
			}
		}
		if len(matches) == 0 {
			break
		}
		fileIndex++
	}

	err = os.WriteFile(path.Join(feed.path, filename+"."+ext), content, 0600)
	if err != nil {
		return &FeedError{
			Code:    500,
			Message: "Unable to write file",
		}
	}

	slog.Info("Added Item", slog.String("name", filename+"."+ext), slog.String("feed", feed.Name), slog.String("content-type", contentType))

	return nil
}

func (feed *Feed) RemoveItem(item string) error {
	slog.Debug("Remove Item", slog.String("name", item), slog.String("feed", feed.Name))
	itemPath := path.Join(feed.path, item)

	err := os.Remove(itemPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &FeedError{
				Code:    404,
				Message: err.Error(),
			}
		}
		return &FeedError{
			Code:    500,
			Message: err.Error(),
		}
	}
	slog.Info("Removed Item", slog.String("name", item), slog.String("feed", feed.Name))
	return nil
}
func (feed *Feed) SetPIN(pin string) error {
	pinPath := path.Join(feed.path, "pin")

	err := os.WriteFile(pinPath, []byte(pin), 0600)
	if err != nil {
		return &FeedError{
			Code:    500,
			Message: "Unable to write PIN",
		}
	}
	return nil
}

func FeedNameFromPath(path string) (*string, error) {
	s := strings.Split(path, "/")
	if len(s) < 4 {
		return nil, fmt.Errorf("No feed in URL (not enough components)")
	}
	result, err := url.QueryUnescape(s[3])
	if err != nil {
		return nil, err
	}
	return &result, nil
}
