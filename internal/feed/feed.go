package feed

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/exp/slog"
)

type NotificationSettings struct {
	VAPIDPublicKey  string
	VAPIDPrivateKey string
}

var fLogLevel = new(slog.LevelVar)
var fLogger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: fLogLevel})).WithGroup("feedManager")

func init() {
	if os.Getenv("DEBUG") != "" || os.Getenv("DEBUG_FEED") != "" {
		fLogLevel.Set(slog.LevelDebug)
	}
}

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
	Path                 string
	Config               FeedConfig
	NotificationSettings *NotificationSettings
	WebSocketManager     *WebSocketManager
}

type PublicFeedItem struct {
	Name string       `json:"name"`
	Date time.Time    `json:"date"`
	Type FeedItemType `json:"type"`
	Feed *PublicFeed  `json:"feed"`
}

func (feed *Feed) Name() string {
	return path.Base(feed.Path)
}
func (feed *Feed) GetPublicItem(i string) (*PublicFeedItem, error) {

	s, err := os.Stat(path.Join(feed.Path, i))

	if err != nil {
		if os.IsNotExist(err) {
			return nil, &FeedError{
				Code:    404,
				Message: fmt.Sprintf("Feed does not exists (%s)", feed.Name()),
			}
		}
		return nil, &FeedError{
			Code:    500,
			Message: fmt.Sprintf("Unable to open feed (%s)", feed.Name()),
		}
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
	fLogger.Debug("Getting Item", slog.String("feed", feed.Name()), slog.String("name", item))
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
		fLogger.Error("No secret was provided", slog.Int("return", code))
		return &FeedError{
			Code:    code,
			Message: "Unauthorized",
		}
	}

	if len(secret) != 4 {
		if feed.Config.Secret != secret {
			code := 401
			fLogger.Error("Invalid secret", slog.Int("return", code))
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

func (f *Feed) AddItem(contentType string, r io.ReadCloser) error {
	fLogger.Debug("Adding Item", slog.String("feed", f.Name()), slog.String("content-type", contentType))
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

	content, err := io.ReadAll(r)
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
		matches, err := filepath.Glob(path.Join(f.Path, filename) + ".*")
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

	err = os.WriteFile(path.Join(f.Path, filename+"."+ext), content, 0600)
	if err != nil {
		return &FeedError{
			Code:    500,
			Message: "Unable to write file",
		}
	}

	f.sendPushNotification()

	publicItem, err := f.GetPublicItem(filename + "." + ext)

	if err != nil {
		return err
	}

	if err = f.WebSocketManager.NotifyAdd(publicItem); err != nil {
		return &FeedError{
			Code:    500,
			Message: err.Error(),
		}
	}

	fLogger.Debug("Added Item", slog.String("name", filename+"."+ext), slog.String("feed", f.Path), slog.String("content-type", contentType))

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

	if err = f.WebSocketManager.NotifyRemove(publicItem); err != nil {
		return &FeedError{
			Code:    500,
			Message: err.Error(),
		}
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
