package feed

import (
	"fmt"
	"path"
)

// FeedManager is the main interface tu feeds and contains the path to ybFeed
// data folder. It is the proper way to get a Feed with proper websocket and
// notifications settings based on current deployment configuration.
type FeedManager struct {
	NotificationSettings *NotificationSettings

	path             string
	websocketManager *WebSocketManager
}

// NewFeedManager returns a FeedManager initialized with the mandatory
// path and websocket manager w.
func NewFeedManager(path string, w *WebSocketManager) *FeedManager {
	result := &FeedManager{
		path:             path,
		websocketManager: w,
	}
	return result
}

// GetFeed returns the Feed with name feedName. Authentication is not vaidates
// Be careful when using GetFeed that the result isn't returned to the browser
// directly. It should ony be used for internal methods
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

// GetFeedWithAuth returns the Feed feedName if the secret is valid,
// otherwise it returns an error. GetFeedWithAuth should always be user
// when fetching a Feed for end user consumption
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
