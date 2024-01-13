package feed

import (
	"fmt"
	"os"
	"path"

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
	path                 string
	websocketManager     *WebSocketManager
	NotificationSettings *NotificationSettings
}

func NewFeedManager(path string, w *WebSocketManager) *FeedManager {

	result := &FeedManager{
		path:             path,
		websocketManager: w,
	}
	return result
}

func (m *FeedManager) GetFeed(feedName string) (*Feed, error) {
	feedPath := path.Join(m.path, feedName)

	result, err := GetFeed(feedPath)
	if err != nil {
		return nil, fmt.Errorf("cannot get feed '%s': %w", feedName, err)
	}
	result.WebSocketManager = m.websocketManager
	result.NotificationSettings = m.NotificationSettings

	return result, nil
}
