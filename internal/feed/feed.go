// Package feed implements
package feed

import (
	"errors"
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
	"github.com/ybizeul/ybfeed/pkg/yblog"
	"golang.org/x/exp/slog"
)

// PublicFeed is a version of a feed meant to provide a json representation of
// a feed that does not expose private informations
// In this context, the feed secret is not a private information as it needs to
// be transmitted as a cookie to the browser
type PublicFeed struct {
	Name           string           `json:"name"`
	Items          []PublicFeedItem `json:"items"`
	Secret         string           `json:"secret"`
	VAPIDPublicKey string           `json:"vapidpublickey"`
}

// PublicFeedItem is used to provide a json representation of a feed item.
type PublicFeedItem struct {
	Name string       `json:"name"`
	Date time.Time    `json:"date"`
	Type FeedItemType `json:"type"`
	Feed *PublicFeed  `json:"feed"`
}

// FeedItemType defines the type of an item in the feed
type FeedItemType int

const (
	Text = iota
	Image
	Binary
)

// var fLogLevel = new(slog.LevelVar)
// var fLogger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: fLogLevel})).WithGroup("feedManager")

var fL = yblog.NewYBLogger("feed", []string{"DEBUG", "DEBUG_FEED"})

var FeedErrorNotFound = errors.New("feed not found")
var FeedErrorInvalidSecret = errors.New("invalid Secret")
var FeedErrorAlreadyExists = errors.New("feed already exists")
var FeedErrorUnableToReadContent = errors.New("unable to read feed content")
var FeedErrorUnableToReadItemInfo = errors.New("unable to read item info")
var FeedErrorItemNotFound = errors.New("feed item not found")
var FeedErrorInvalidContentType = errors.New("invalid content-type")
var FeedErrorMaxBodySizeExceeded = errors.New("max body size exceeded")
var FeedErrorItemEmpty = errors.New("feed item is empty")
var FeedErrorErrorReading = errors.New("error while reading new item")
var FeedErrorErrorWriting = errors.New("error while reading new item")

// Feed is the internal representation of a Feed and contains all the
// informations needed to perform its tasks
type Feed struct {
	Path                 string
	Config               FeedConfig
	NotificationSettings *NotificationSettings
	WebSocketManager     *WebSocketManager
}

type NotificationSettings struct {
	VAPIDPublicKey  string
	VAPIDPrivateKey string
}

func NewFeed(feedPath string) (*Feed, error) {
	fL.Logger.Info("Creating new feed", slog.String("feed", feedPath))

	_, err := os.Stat(feedPath)
	if err == nil {
		return nil, FeedErrorAlreadyExists
	}

	err = os.Mkdir(feedPath, 0700)
	if err != nil {
		fL.Logger.Error("Error creating feed directory", slog.String("directory", feedPath))
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
		return nil, err
	}

	return &feed, nil
}

func GetFeed(feedPath string) (*Feed, error) {
	if _, err := os.Stat(feedPath); os.IsNotExist(err) {
		return nil, FeedErrorNotFound
	}
	result := &Feed{
		Path: feedPath,
	}

	c, err := FeedConfigForFeed(result)
	if err != nil {
		return nil, fmt.Errorf("cannot get config for feed '%s': %w", feedPath, err)
	}
	c.feed = result

	result.Config = *c

	return result, nil
}

func (feed *Feed) Name() string {
	return path.Base(feed.Path)
}

func (feed *Feed) Public() (*PublicFeed, error) {
	items, err := feed.publicItems()
	if err != nil {
		return nil, err
	}

	result := &PublicFeed{
		Name:   feed.Name(),
		Items:  items,
		Secret: feed.Config.Secret,
	}

	if feed.NotificationSettings != nil {
		result.VAPIDPublicKey = feed.NotificationSettings.VAPIDPublicKey
	}

	return result, nil
}

func (feed *Feed) publicItems() ([]PublicFeedItem, error) {
	items := []PublicFeedItem{}

	var d []fs.DirEntry
	var err error

	if d, err = os.ReadDir(feed.Path); err != nil {
		fL.Logger.Error("Unable to read feed content")

		return nil, FeedErrorUnableToReadContent
	}

	for _, f := range d {
		if f.Name() == "secret" || f.Name() == "pin" || f.Name() == "config.json" {
			continue
		}
		info, err := f.Info()
		if err != nil {
			code := 500
			e := "Unable to read file info"

			fL.Logger.Error(e, slog.Int("return", code))

			return nil, fmt.Errorf("%w: %s", FeedErrorUnableToReadItemInfo, f.Name())
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
			Feed: &PublicFeed{Name: feed.Name()},
		})
	}
	sort.Slice(items, func(i, j2 int) bool {
		return items[i].Date.After(items[j2].Date)
	})

	return items, nil
}

func (feed *Feed) GetPublicItem(i string) (*PublicFeedItem, error) {

	s, err := os.Stat(path.Join(feed.Path, i))

	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", FeedErrorItemNotFound, i)
		}
		return nil, err
	}

	var itemType FeedItemType
	if strings.HasSuffix(i, ".txt") {
		itemType = Text
	} else if strings.HasSuffix(i, ".png") || strings.HasSuffix(i, ".jpg") {
		itemType = Image
	} else {
		itemType = Binary
	}
	return &PublicFeedItem{
		Name: i,
		Date: s.ModTime(),
		Type: itemType,
		Feed: &PublicFeed{
			Name:   feed.Name(),
			Secret: feed.Config.Secret,
		},
	}, nil
}

func (feed *Feed) GetItemData(item string) ([]byte, error) {
	// Read item content
	fL.Logger.Debug("Getting Item", slog.String("feed", feed.Name()), slog.String("name", item))
	var content []byte
	filePath := path.Join(feed.Path, item)
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", FeedErrorItemNotFound, item)
		}
		return nil, err
	}
	return content, nil
}

func (feed *Feed) IsSecretValid(secret string) error {
	if secret == "" {
		return FeedErrorInvalidSecret
	}

	if len(secret) != 4 {
		if feed.Config.Secret != secret {
			fL.Logger.Error("Invalid secret")
			return FeedErrorInvalidSecret
		}
	} else {
		err := feed.Config.PIN.IsValid(secret)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *Feed) AddItem(contentType string, r io.ReadCloser) error {
	fL.Logger.Debug("Adding Item", slog.String("feed", f.Name()), slog.String("content-type", contentType))
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
		return fmt.Errorf("%w: %s", FeedErrorInvalidContentType, contentType)
	}

	content, err := io.ReadAll(r)
	if err != nil {
		e, ok := err.(*http.MaxBytesError)
		if ok {
			return fmt.Errorf("%w: %d", FeedErrorMaxBodySizeExceeded, e.Limit)
		}
		return err
	}

	if len(content) == 0 {
		return fmt.Errorf("%w: %s %s", FeedErrorItemEmpty, f.Path, template)
	}

	fileIndex := 1
	var filename string
	for {
		filename = fmt.Sprintf("%s %d", template, fileIndex)
		matches, err := filepath.Glob(path.Join(f.Path, filename) + ".*")
		if err != nil {
			return fmt.Errorf("%w: %s", FeedErrorErrorReading, filename)
		}
		if len(matches) == 0 {
			break
		}
		fileIndex++
	}

	filePath := path.Join(f.Path, filename+"."+ext)
	err = os.WriteFile(filePath, content, 0600)
	if err != nil {
		return fmt.Errorf("%w: %s", FeedErrorErrorWriting, filePath)
	}

	f.sendPushNotification()

	publicItem, err := f.GetPublicItem(filename + "." + ext)

	if err != nil {
		return err
	}

	if err = f.WebSocketManager.NotifyAdd(publicItem); err != nil {
		return err
	}

	fL.Logger.Debug("Added Item", slog.String("name", filename+"."+ext), slog.String("feed", f.Path), slog.String("content-type", contentType))

	return nil
}

func (f *Feed) RemoveItem(item string) error {
	slog.Debug("Remove Item", slog.String("name", item), slog.String("feed", f.Path))
	itemPath := path.Join(f.Path, item)

	publicItem, err := f.GetPublicItem(item)
	if err != nil {
		return err
	}

	err = os.Remove(itemPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", FeedErrorItemNotFound, itemPath)
		}
		return err
	}

	if err = f.WebSocketManager.NotifyRemove(publicItem); err != nil {
		return err
	}

	slog.Debug("Removed Item", slog.String("name", item), slog.String("feed", f.Name()))
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
