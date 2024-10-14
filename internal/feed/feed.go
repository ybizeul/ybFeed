// Package feed implements functions for managing feed, like create a new feed,
// add or remove items, send notifications, etc.
package feed

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
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

type FileTypeInfo struct {
	FileExtension    string
	FileNameTemplate string
}

// GetItemType returns FeedItemType for filenamebase on file extension
func GetItemType(fileName string) FeedItemType {
	fL.Logger.Debug("Finding item type", slog.String("filename", fileName), slog.String("ext", path.Ext(fileName)))
	var itemType FeedItemType
	switch {

	case path.Ext(fileName) == ".txt":
		itemType = Text
	case path.Ext(fileName) == ".png":
		itemType = Image
	case path.Ext(fileName) == ".jpg":
		itemType = Image
	default:
		itemType = Binary
	}
	fL.Logger.Debug("Result", slog.Int("type", int(itemType)))
	return itemType
}

// fL is a logger for feed activity
var fL = yblog.NewYBLogger("feed", []string{"DEBUG", "DEBUG_FEED"})

// Errors related to feed actions
var (
	FeedErrorNotFound             = errors.New("feed not found")
	FeedErrorInvalidSecret        = errors.New("invalid Secret")
	FeedErrorIncorrectSecret      = errors.New("incorrect Secret")
	FeedErrorAlreadyExists        = errors.New("feed already exists")
	FeedErrorUnableToReadContent  = errors.New("unable to read feed content")
	FeedErrorUnableToReadItemInfo = errors.New("unable to read item info")
	FeedErrorItemNotFound         = errors.New("feed item not found")
	FeedErrorInvalidContentType   = errors.New("invalid content-type")
	FeedErrorMaxBodySizeExceeded  = errors.New("max body size exceeded")
	FeedErrorItemEmpty            = errors.New("feed item is empty")
	FeedErrorErrorReading         = errors.New("error while reading new item")
	FeedErrorErrorWriting         = errors.New("error while reading new item")
	FeedErrorInvalidFeedItem      = errors.New("invalid feed item, cannot get internal files")
)

// Feed is the internal representation of a Feed and contains all the
// informations needed to perform its tasks
type Feed struct {
	Path                 string
	Config               FeedConfig
	NotificationSettings *NotificationSettings
	WebSocketManager     *WebSocketManager
}

// NotificationSettings contains the necessary key pair to send web push
// notifications. A key pair is generated on the first start of ybFeed and
// stored in ybFeed configuration. After that, every feed created is
// configured the same key pair.
type NotificationSettings struct {
	VAPIDPublicKey  string
	VAPIDPrivateKey string
}

// NewFeed creates a new feed at feedPath. feedPath is an absolute or relative
// (to ybFeed binary) path to the directory where feed items will be stored.
// The last directory of the path if the feed name.
func NewFeed(feedPath string) (*Feed, error) {
	fL.Logger.Info("Creating new feed", slog.String("feed", feedPath))

	// Return error if feed already exists
	_, err := os.Stat(feedPath)
	if err == nil {
		return nil, FeedErrorAlreadyExists
	}

	// Create feed directory
	err = os.Mkdir(feedPath, 0700)
	if err != nil {
		fL.Logger.Error("Error creating feed directory", slog.String("directory", feedPath))
		return nil, err
	}

	// Prepare Feed struct and assign a random secret in Config
	feed := Feed{
		Path: feedPath,
		Config: FeedConfig{
			Secret: uuid.NewString(),
		},
	}

	feed.Config.feed = &feed

	// Write feed configuration
	err = feed.Config.Write()
	if err != nil {
		return nil, err
	}

	return &feed, nil
}

// GetFeed returns the feed at feedPath, which is an absolute or relative (to
// ybFeed binary) path to the directory where feed items are stored.
func GetFeed(feedPath string) (*Feed, error) {
	// Check that the feed exists
	if _, err := os.Stat(feedPath); os.IsNotExist(err) {
		return nil, FeedErrorNotFound
	}

	// Create the Feed struct for the required feed
	result := &Feed{
		Path: feedPath,
	}

	// Retrieve configuration for the feed
	c, err := FeedConfigForFeed(result)
	if err != nil {
		return nil, fmt.Errorf("cannot get config for feed '%s': %w", feedPath, err)
	}

	c.feed = result
	result.Config = *c

	return result, nil
}

// Name returns the feed name based on it's path. The name of a feed is always
// the base directory name
func (feed *Feed) Name() string {
	return path.Base(feed.Path)
}

// Public returns a marshalable representation of a feed, that can be returned
// to the client as a result to an API call or in a websocket.
func (feed *Feed) Public() (*PublicFeed, error) {
	// Get all public items for the feed
	items, err := feed.publicItems()
	if err != nil {
		return nil, err
	}

	// Prepare the PublicFeed struct that will be returned
	result := &PublicFeed{
		Name:   feed.Name(),
		Items:  items,
		Secret: feed.Config.Secret,
	}

	// Add the necessary web push notification public key for the browser
	// to be able to subscribe
	if feed.NotificationSettings != nil {
		result.VAPIDPublicKey = feed.NotificationSettings.VAPIDPublicKey
	}

	return result, nil
}

// publicItems returns a slice of marshalable structs representing feed items
func (feed *Feed) publicItems() ([]PublicFeedItem, error) {
	items := []PublicFeedItem{}

	var d []fs.DirEntry
	var err error

	// Prepare to read feed directory
	if d, err = os.ReadDir(feed.Path); err != nil {
		fL.Logger.Error("Unable to read feed content")
		return nil, FeedErrorUnableToReadContent
	}

	// Parse feed directory content, ignoring internal files
	for _, f := range d {
		if f.Name() == "secret" || f.Name() == "pin" || f.Name() == "config.json" {
			continue
		}
		info, err := f.Info()
		if err != nil {
			fL.Logger.Error("error while reading feed content", slog.String("feed", feed.Path), slog.String("item", f.Name()))
			return nil, fmt.Errorf("%w: %s", FeedErrorUnableToReadItemInfo, path.Join(feed.Path, f.Name()))
		}

		// Append feed item to result
		items = append(items, PublicFeedItem{
			Name: f.Name(),
			Date: info.ModTime(),
			Type: GetItemType(f.Name()),
			Feed: &PublicFeed{Name: feed.Name()},
		})
	}

	// Sort result based on date
	sort.Slice(items, func(i, j2 int) bool {
		return items[i].Date.After(items[j2].Date)
	})

	return items, nil
}

func (feed *Feed) Empty() error {
	items, err := feed.publicItems()
	if err != nil {
		return err
	}
	for _, item := range items {
		err := feed.RemoveItem(item.Name, false)
		if err != nil {
			return err
		}
	}

	// Notify all connected websockets
	if feed.WebSocketManager != nil {
		if err = feed.WebSocketManager.NotifyEmpty(feed); err != nil {
			return err
		}
	}

	return nil
}

// GetPublicItem returns a marshable struct for a specific feed item
func (feed *Feed) GetPublicItem(i string) (*PublicFeedItem, error) {
	if i == "secret" || i == "pin" || i == "config.json" {
		return nil, FeedErrorInvalidFeedItem
	}

	s, err := os.Stat(path.Join(feed.Path, path.Join("/", i)))

	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", FeedErrorItemNotFound, i)
		}
		return nil, err
	}

	return &PublicFeedItem{
		Name: i,
		Date: s.ModTime(),
		Type: GetItemType(i),
		Feed: &PublicFeed{
			Name:   feed.Name(),
			Secret: feed.Config.Secret,
		},
	}, nil
}

// GetItemData returns the content of a specific feed item
func (feed *Feed) GetItemData(item string) ([]byte, error) {
	fL.Logger.Debug("Getting Item", slog.String("feed", feed.Path), slog.String("name", item))
	var content []byte

	// Get path to feed item
	filePath := path.Join(feed.Path, path.Join("/"+item))

	if path.Base(filePath) == "secret" || path.Base(filePath) == "pin" || path.Base(filePath) == "config.json" {
		return nil, fmt.Errorf("%w: %s", FeedErrorItemNotFound, item)
	}
	// Read feed item content
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", FeedErrorItemNotFound, item)
		}
		return nil, err
	}

	return content, nil
}

// IsSecretValid returns an error if the provided secret doesn't allow access
// to the feed. secret can be a full secret or a PIN
func (feed *Feed) IsSecretValid(secret string) error {
	if secret == "" {
		return FeedErrorInvalidSecret
	}

	if len(secret) == 4 { // Secret is a PIN
		err := feed.Config.PIN.IsValid(secret)
		if err != nil {
			fL.Logger.Error(err.Error())
			return err
		}
	} else {
		if feed.Config.Secret != secret {
			fL.Logger.Error(FeedErrorIncorrectSecret.Error())
			return FeedErrorIncorrectSecret
		}
	}

	return nil
}

// AddItem reads content from r and creates a new file in the feed directory
// with a name and file extension based on contentType, then notifies clients
func (f *Feed) AddItem(contentType string, filename string, r io.Reader) error {
	fL.Logger.Debug("Adding Item", slog.String("feed", f.Name()), slog.String("content-type", contentType))

	var err error

	// Build a map providing informations for each type
	mimeInfos := map[string]FileTypeInfo{
		"image/png":  {FileExtension: "png", FileNameTemplate: "Pasted Image"},
		"image/jpeg": {FileExtension: "jpg", FileNameTemplate: "Pasted Image"},
		"text/plain": {FileExtension: "txt", FileNameTemplate: "Pasted Text"},
	}

	// If the content-type isn't found, return an error
	info, ok := mimeInfos[contentType]
	if !ok {
		info = FileTypeInfo{
			FileExtension:    path.Ext(filename)[1:],
			FileNameTemplate: filename[:len(filename)-len(path.Ext(filename))],
		}
		//return fmt.Errorf("%w: %s", FeedErrorInvalidContentType, contentType)
	}

	// Obtain file extension and template for file name
	ext := info.FileExtension
	template := info.FileNameTemplate

	// Read content
	content, err := io.ReadAll(r)
	if err != nil {
		e, ok := err.(*http.MaxBytesError)
		if ok {
			return fmt.Errorf("%w: %d", FeedErrorMaxBodySizeExceeded, e.Limit)
		}
		return err
	}

	// Check the content is not empty
	if len(content) == 0 {
		return fmt.Errorf("%w: %s %s", FeedErrorItemEmpty, f.Path, template)
	}

	// Search for existing content with identical file type to increment
	// the index in file name
	fileIndex := 0
	for {
		fileIndexStr := ""
		if fileIndex > 0 {
			fileIndexStr = fmt.Sprintf(" %d", fileIndex)
		}
		filename = fmt.Sprintf("%s%s", template, fileIndexStr)
		matches, err := filepath.Glob(path.Join(f.Path, filename) + ".*")
		if err != nil {
			return fmt.Errorf("%w: %s", FeedErrorErrorReading, filename)
		}
		if len(matches) == 0 {
			break
		}
		fileIndex++
	}

	// Assign full file path
	filePath := path.Join(f.Path, filename+"."+ext)

	// Write content to file
	err = os.WriteFile(filePath, content, 0600)
	if err != nil {
		return fmt.Errorf("%w: %s", FeedErrorErrorWriting, filePath)
	}

	// Get PublicItem for the added content
	publicItem, err := f.GetPublicItem(filename + "." + ext)

	if err != nil {
		return err
	}

	// Notify additon to all connected browsers
	if f.WebSocketManager != nil {
		if err = f.WebSocketManager.NotifyAdd(publicItem); err != nil {
			return err
		}
	}
	// Send push notification to subscribed browsers
	err = f.sendPushNotification()
	if err != nil {
		fL.Logger.Error("Error sending push notification", slog.String("feed", f.Path), slog.String("error", err.Error()))
		return err
	}

	fL.Logger.Debug("Added Item", slog.String("name", filename+"."+ext), slog.String("feed", f.Path), slog.String("content-type", contentType))

	return nil
}

// RemoveItem deletes item from the feed directory and notifies clients
func (f *Feed) RemoveItem(item string, notify bool) error {
	fL.Logger.Debug("Remove Item", slog.String("name", item), slog.String("feed", f.Path))

	itemPath := path.Join(f.Path, path.Join("/", item))

	// Save public item before deletion for notification later
	publicItem, err := f.GetPublicItem(item)
	if err != nil {
		return err
	}

	// Delete item from directory
	err = os.Remove(itemPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", FeedErrorItemNotFound, itemPath)
		}
		return err
	}

	// Notify all connected websockets
	if f.WebSocketManager != nil && notify {
		if err = f.WebSocketManager.NotifyRemove(publicItem); err != nil {
			return err
		}
	}

	fL.Logger.Debug("Removed Item", slog.String("name", item), slog.String("feed", f.Path))
	return nil
}

// SetPIN configures the provided pin on the feed
func (feed *Feed) SetPIN(pin string) error {
	err := feed.Config.SetPIN(pin)
	if err != nil {
		return err
	}
	return nil
}
