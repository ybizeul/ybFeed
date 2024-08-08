package feed

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	ws "github.com/gorilla/websocket"
	"github.com/ybizeul/ybfeed/internal/utils"
	"github.com/ybizeul/ybfeed/pkg/yblog"
	"golang.org/x/exp/slog"
)

var wsL = yblog.NewYBLogger("push", []string{"DEBUG", "DEBUG_WEBSOCKET"})

// upgrader is used to upgrade a connection to a websocket
var upgrader = ws.Upgrader{} // use default options

// FeedSockets maintains a list of active websockets for a specific feed
// designated by feedName
type FeedSockets struct {
	feedName   string
	websockets []*ws.Conn
}

// RemoveConn removes the websocket c from the list of active websockets
func (fs *FeedSockets) RemoveConn(c *ws.Conn) {
	wsL.Logger.Debug("Removing connection",
		slog.Int("count", len(fs.websockets)),
		slog.Any("connections", fs.websockets),
		slog.String("connection", fmt.Sprintf("%p", c)))
	for i, conn := range fs.websockets {
		wsL.Logger.Debug("Current connection", slog.String("connection", fmt.Sprintf("%p", conn)))
		if conn == c {
			wsL.Logger.Debug("Found connection", slog.String("connection", fmt.Sprintf("%p", conn)))
			fs.websockets[i] = fs.websockets[len(fs.websockets)-1]
			fs.websockets = fs.websockets[:len(fs.websockets)-1]
		}
	}
}

// FeedNotification is used to marshall notification information message
// to the push service
type FeedNotification struct {
	Action string         `json:"action"`
	Item   PublicFeedItem `json:"item"`
}

// WebSocketManager bridges a FeedManager with a FeedSockets struct
type WebSocketManager struct {
	FeedSockets []*FeedSockets
	FeedManager *FeedManager
}

// NewWebSocketManager creates a new WebSocketManager. There is typically one
// WebSocketManager per ybFeed deployment.
func NewWebSocketManager(fm *FeedManager) *WebSocketManager {
	return &WebSocketManager{
		FeedManager: fm,
	}
}

// FeedSocketsForFeed returns the FeedSockets for feed feedName
func (m *WebSocketManager) FeedSocketsForFeed(feedName string) *FeedSockets {
	wsL.Logger.Debug("Searching FeedSockets", slog.Int("count", len(m.FeedSockets)), slog.String("feedName", feedName))

	// Loop through all FeedSockets to find the one for this feed
	for _, fs := range m.FeedSockets {
		if fs.feedName == feedName {
			return fs
		}
	}
	return nil
}

// RunSocketForFeed promotes an HTTP connection to a websocket and starts
// waiting for data. This function is blocking and typically runs from
// a http handler.
func (m *WebSocketManager) RunSocketForFeed(feedName string, w http.ResponseWriter, r *http.Request) {
	// Check if we already have websockets for this feed
	feedSockets := m.FeedSocketsForFeed(feedName)

	if feedSockets == nil { // No, then we create a new FeedSockets
		wsL.Logger.Debug("Adding FeedSockets", slog.Int("count_before", len(m.FeedSockets)), slog.String("feedName", feedName))
		feedSockets = &FeedSockets{
			feedName: feedName,
		}
		m.FeedSockets = append(m.FeedSockets, feedSockets)
	}

	// Upgrade http connection to websocket
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to upgrade WebSocket")
	}

	feedSockets.websockets = append(feedSockets.websockets, c)

	// Get provided secret and validate feed access
	secret, _ := utils.GetSecret(r)

	f, err := m.FeedManager.GetFeedWithAuth(feedName, secret)

	if err != nil {
		switch {
		case errors.Is(err, FeedErrorNotFound):
			utils.CloseWithCodeAndMessage(w, 404, "feed not found")
		case errors.Is(err, FeedErrorInvalidSecret):
			utils.CloseWithCodeAndMessage(w, 401, "invalid secret")
		case errors.Is(err, FeedErrorIncorrectSecret):
			utils.CloseWithCodeAndMessage(w, 401, "incorrect secret")
		default:
			utils.CloseWithCodeAndMessage(w, 500, err.Error())
		}
	}

	// Cleanup
	defer func() {
		feedSockets.RemoveConn(c)
		c.Close()
	}()

	// Start waiting for messages
	for {
		mt, message, err := c.ReadMessage()
		wsL.Logger.Debug("Message Received",
			slog.String("message", string(message)),
			slog.Int("messageType", mt))
		if err != nil {
			slog.Error("Error reading message",
				slog.String("error", err.Error()),
				slog.Int("messageType", mt))
			break
		}
		switch strings.TrimSpace(string(message)) {
		// Return pubic feed content
		case "feed":
			pf, err := f.Public()
			if err != nil {
				utils.CloseWithCodeAndMessage(w, 500, err.Error())
			}
			err = c.WriteJSON(pf)
			if err != nil {
				utils.CloseWithCodeAndMessage(w, 500, err.Error())
			}
		}
	}
}

// NotifyAdd notifies all connected websockets that an item has been added
func (m *WebSocketManager) NotifyAdd(item *PublicFeedItem) error {
	wsL.Logger.Debug("Notify websocket",
		slog.Any("item", item),
		slog.Int("ws count", len(m.FeedSockets)))
	for _, f := range m.FeedSockets {
		wsL.Logger.Debug("checking feed", slog.String("feedName", f.feedName))
		if f.feedName == item.Feed.Name {
			wsL.Logger.Debug("Found feed", slog.String("feedName", f.feedName))
			for _, w := range f.websockets {
				if err := w.WriteJSON(FeedNotification{
					Action: "add",
					Item:   *item,
				}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// NotifyRemove notify all connected websockets that an item has been removed
func (m *WebSocketManager) NotifyRemove(item *PublicFeedItem) error {
	wsL.Logger.Debug("Notify websocket",
		slog.Any("item", item),
		slog.Int("ws count", len(m.FeedSockets)))
	for _, f := range m.FeedSockets {
		wsL.Logger.Debug("checking feed", slog.String("feedName", f.feedName))
		if f.feedName == item.Feed.Name {
			wsL.Logger.Debug("found feed", slog.String("feedName", f.feedName))
			for _, w := range f.websockets {
				if err := w.WriteJSON(FeedNotification{
					Action: "remove",
					Item:   *item,
				}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
