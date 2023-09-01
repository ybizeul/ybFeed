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

// PublicFeed is a version of a feed that does not expose private informations
// In this context, the feed secret is not a private information as it needs to
// be transmitted as a cookie to the browser
type PublicFeed struct {
	Name   string           `json:"name"`
	Items  []PublicFeedItem `json:"items"`
	Secret string           `json:"secret"`
}

// Feed is the internal representation of a Feed and contains all the
// informations needed to perform its tasks
type Feed struct {
	Path   string
	Config FeedConfig
}

type PublicFeedItem struct {
	Name string       `json:"name"`
	Date time.Time    `json:"date"`
	Type FeedItemType `json:"type"`
	Feed *PublicFeed  `json:"-"`
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

	err = os.Mkdir(feedPath, 0700)
	if err != nil {
		slog.Error("Error creating feed directory", slog.String("feed", feedName), slog.String("directory", feedPath))
		return nil, err
	}

	feed := Feed{
		Path: feedPath,
		Config: FeedConfig{
			Secret: uuid.NewString(),
		},
	}

	feed.Config.feed = &feed

	err = feed.Config.Write()

	if err != nil {
		slog.Error("Unable to write config %s", err.Error())
		return nil, err
	}

	return &feed, nil
}
func GetFeed(feedPath string) (*Feed, error) {
	if _, err := os.Stat(feedPath); os.IsNotExist(err) {
		return nil, &FeedError{
			Code:    404,
			Message: "Feed does not exists",
		}
	}

	result := &Feed{Path: feedPath}

	c, err := FeedConfigForFeed(result)
	if err != nil {
		return nil, &FeedError{
			Code:    404,
			Message: "Feed does not exists",
		}
	}
	c.feed = result

	result.Config = *c

	return result, nil
}

func GetPublicFeed(basePath string, feedName string, secret string) (*PublicFeed, error) {
	feedPath := path.Join(basePath, feedName)

	feedLog := slog.Default().With(slog.String("feed", feedName))

	feedLog.Debug("Getting feed", slog.Int("secret_len", len(secret)))

	f, err := GetFeed(feedPath)

	err = f.IsSecretValid(secret)

	if err != nil {
		return nil, err
	}

	publicFeed := &PublicFeed{
		Name:   feedName,
		Secret: f.Config.Secret,
	}
	var d []fs.DirEntry
	if d, err = os.ReadDir(feedPath); err != nil {
		code := 500
		feedLog.Error("Unable to feed content", slog.Int("return", code))

		return nil, &FeedError{
			Code:    code,
			Message: "Unable to open directory for read",
		}
	}

	items := []PublicFeedItem{}
	for _, f := range d {
		if f.Name() == "secret" || f.Name() == "pin" || f.Name() == "config.json" {
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
		items = append(items, PublicFeedItem{
			Name: f.Name(),
			Date: info.ModTime(),
			Type: itemType,
			Feed: publicFeed,
		})
	}
	sort.Slice(items, func(i, j2 int) bool {
		return items[i].Date.After(items[j2].Date)
	})

	publicFeed.Items = items

	return publicFeed, nil
}
func (feed *Feed) Name() string {
	return path.Base(feed.Path)
}
func (feed *Feed) GetItem(item string) ([]byte, error) {
	// Read item content
	slog.Info("Getting Item", slog.String("feed", feed.Name()), slog.String("name", item))
	var content []byte
	filePath := path.Join(feed.Path, item)
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

func (feed *Feed) IsSecretValid(secret string) error {
	if secret == "" {
		code := 401
		slog.Error("No secret was provided", slog.Int("return", code))
		return &FeedError{
			Code:    code,
			Message: "Unauthorized",
		}
	}

	if len(secret) != 4 {
		if feed.Config.Secret != secret {
			code := 401
			slog.Error("Invalid secret", slog.Int("return", code))
			return &FeedError{
				Code:    code,
				Message: "Authentication failed",
			}
		}
	} else {
		err := feed.Config.PIN.IsValid(secret)
		if err != nil {
			return err
		}
	}

	return nil
}

func (feed *Feed) AddItem(contentType string, f io.ReadCloser) error {
	slog.Debug("Adding Item", slog.String("feed", feed.Name()), slog.String("content-type", contentType))
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
		matches, err := filepath.Glob(path.Join(feed.Path, filename) + ".*")
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

	err = os.WriteFile(path.Join(feed.Path, filename+"."+ext), content, 0600)
	if err != nil {
		return &FeedError{
			Code:    500,
			Message: "Unable to write file",
		}
	}

	slog.Info("Added Item", slog.String("name", filename+"."+ext), slog.String("feed", feed.Path), slog.String("content-type", contentType))

	return nil
}

func (feed *Feed) RemoveItem(item string) error {
	slog.Debug("Remove Item", slog.String("name", item), slog.String("feed", feed.Path))
	itemPath := path.Join(feed.Path, item)

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
	slog.Info("Removed Item", slog.String("name", item), slog.String("feed", feed.Name()))
	return nil
}

func (feed *Feed) SetPIN(pin string) error {
	err := feed.Config.SetPIN(pin)
	if err != nil {
		return err
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
