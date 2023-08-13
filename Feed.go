package main

import (
	"fmt"
	"io"
	"io/fs"
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
	Name   string
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

func NewFeed(feedName string) (*Feed, error) {
	log.Infof("Creating new feed %s", feedName)
	os.Mkdir(path.Join(dataDir, feedName), 0700)
	secret := uuid.NewString()
	err := os.WriteFile(path.Join(dataDir, feedName, "secret"), []byte(secret), 0600)
	if err != nil {
		log.Errorf("Unable towrite file %s", err.Error())
		return nil, err
	}
	return &Feed{
		Name:   feedName,
		Secret: secret,
		Items:  []FeedItem{},
		path:   path.Join(dataDir, feedName),
	}, nil
}

func GetFeed(feedName string, secret string) (*Feed, error) {
	feedPath := path.Join(dataDir, feedName)

	log.Debugf("Getting feed %s with secret[%d]", feedName, len(secret))

	if _, err := os.Stat(feedPath); os.IsNotExist(err) {
		return nil, &FeedError{
			Code:    404,
			Message: "Feed does not exists",
		}
	}

	if secret == "" {
		log.Debugf("No secret was provided, returning 401")
		return nil, &FeedError{
			Code:    401,
			Message: "Unauthorized",
		}
	}

	if len(secret) != 4 {
		feedSecret, err := os.ReadFile(path.Join(feedPath, "secret"))

		if err != nil {
			return nil, &FeedError{
				Code:    500,
				Message: err.Error(),
			}
		}

		if string(feedSecret) != secret {
			return nil, &FeedError{
				Code:    401,
				Message: "Authentication failed",
			}
		}
	} else {
		stat, err := os.Stat(path.Join(feedPath, "pin"))
		if err != nil {
			return nil, &FeedError{
				Code:    500,
				Message: err.Error(),
			}
		}

		maxTime := stat.ModTime().Add(2 * time.Minute)
		if maxTime.Before(time.Now()) {
			os.Remove(path.Join(feedPath, "pin"))
			return nil, &FeedError{
				Code:    401,
				Message: "Authentication failed",
			}
		} else {

		}

		s, err := os.ReadFile(path.Join(feedPath, "secret"))

		secret = string(s)

		if err != nil {
			return nil, &FeedError{
				Code:    500,
				Message: err.Error(),
			}
		}
	}

	var d []fs.DirEntry
	var err error
	if d, err = os.ReadDir(feedPath); err != nil {
		return nil, &FeedError{
			Code:    500,
			Message: "Unable to open directory for read",
		}
	}

	feed := Feed{
		Name:   feedName,
		Secret: secret,
		path:   path.Join(dataDir, feedName),
	}

	items := []FeedItem{}
	for _, f := range d {
		if f.Name() == "secret" || f.Name() == "pin" {
			continue
		}
		info, err := f.Info()
		if err != nil {
			return nil, &FeedError{
				Code:    500,
				Message: "Unable to read file info",
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
			Feed: &feed,
		})
	}
	sort.Slice(items, func(i, j2 int) bool {
		return items[i].Date.After(items[j2].Date)
	})

	feed.Items = items

	return &feed, nil
}

func (feed *Feed) GetItem(item string) ([]byte, error) {
	// Read item content
	var content []byte
	content, err := os.ReadFile(path.Join(feed.path, item))
	if err != nil {
		return nil, &FeedError{
			Code:    500,
			Message: "Unable to open file for read",
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
			Code:    500,
			Message: "Content-type not accepted",
		}
	}

	content, err := io.ReadAll(f)
	if err != nil {
		return &FeedError{
			Code:    500,
			Message: "Unable to read stream",
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
		return &FeedError{
			Code:    500,
			Message: "Unable to delete file",
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
