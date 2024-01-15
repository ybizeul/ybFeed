package feed

import (
	"fmt"
	"path"
)

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

func (m *FeedManager) GetFeedWithAuth(feedName string, secret string) (*Feed, error) {
	result, err := m.GetFeed(feedName)

	if err != nil {
		return nil, err
	}

	err = result.IsSecretValid(secret)

	if err != nil {
		return nil, err
	}

	return result, nil
}
