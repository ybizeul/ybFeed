package feed

import (
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

var fmLogLevel = new(slog.LevelVar)
var fmLogger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: fmLogLevel})).WithGroup("feedManager")

func init() {
	if os.Getenv("DEBUG") != "" || os.Getenv("DEBUG_FEEDMANAGER") != "" {
		fmLogLevel.Set(slog.LevelDebug)
	}
}

type FeedManager struct {
	path             string
	websocketManager *WebSocketManager
}

func NewFeedManager(path string, w *WebSocketManager) *FeedManager {

	result := &FeedManager{
		path:             path,
		websocketManager: w,
	}
	return result
}
func (m *FeedManager) NewFeed(feedName string) (*Feed, error) {
	fmLogger.Info("Creating new feed", slog.String("feed", feedName))
	feedPath := path.Join(m.path, feedName)

	_, err := os.Stat(feedPath)
	if err == nil {
		return nil, &FeedError{
			Code:    400,
			Message: "Feed already exists",
		}
	}

	err = os.Mkdir(feedPath, 0700)
	if err != nil {
		fmLogger.Error("Error creating feed directory", slog.String("feed", feedName), slog.String("directory", feedPath))
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
		fmLogger.Error("Unable to write config %s", err.Error())
		return nil, err
	}

	return &feed, nil
}

func (m *FeedManager) GetFeed(feedName string) (*Feed, error) {
	feedPath := path.Join(m.path, feedName)
	if _, err := os.Stat(feedPath); os.IsNotExist(err) {
		return nil, &FeedError{
			Code:    404,
			Message: "Feed does not exists",
		}
	}

	result := &Feed{
		Path:             feedPath,
		WebSocketManager: m.websocketManager,
	}

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

func (m *FeedManager) GetPublicFeed(feedName string, secret string) (*PublicFeed, error) {
	fmLogger.Debug("Getting Public Feed", slog.Any("feedManager", m))
	feedPath := path.Join(m.path, feedName)

	feedLog := fmLogger.With(slog.String("feed", feedName))

	feedLog.Debug("Getting feed", slog.Int("secret_len", len(secret)))

	f, err := m.GetFeed(feedName)

	if err != nil {
		return nil, err
	}

	err = f.IsSecretValid(secret)

	if err != nil {
		return nil, err
	}

	publicFeed := &PublicFeed{
		Name:   feedName,
		Secret: f.Config.Secret,
	}
	publicFeedNoItems := &PublicFeed{
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
			Feed: publicFeedNoItems,
		})
	}
	sort.Slice(items, func(i, j2 int) bool {
		return items[i].Date.After(items[j2].Date)
	})

	publicFeed.Items = items

	return publicFeed, nil
}
