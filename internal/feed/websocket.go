package feed

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	ws "github.com/gorilla/websocket"
	"github.com/ybizeul/ybfeed/internal/utils"
)

var logLevel = new(slog.LevelVar)
var logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))

func init() {
	if os.Getenv("DEBUG_WEBSOCKET") != "" {
		logLevel.Set(slog.LevelDebug)
	}
}

type FeedSockets struct {
	feedName   string
	websockets []*ws.Conn
}

type FeedNotification struct {
	Action string         `json:"action"`
	Item   PublicFeedItem `json:"item"`
}

func (fs *FeedSockets) RemoveConn(c *ws.Conn) {
	logger.Debug("Removing connection",
		slog.Int("count", len(fs.websockets)),
		slog.Any("connections", fs.websockets),
		slog.String("connection", fmt.Sprintf("%p", c)))
	for i, conn := range fs.websockets {
		logger.Debug("Current connection", slog.String("connection", fmt.Sprintf("%p", conn)))
		if conn == c {
			logger.Debug("Found connection", slog.String("connection", fmt.Sprintf("%p", conn)))
			fs.websockets[i] = fs.websockets[len(fs.websockets)-1]
			fs.websockets = fs.websockets[:len(fs.websockets)-1]
		}
	}
}

type WebSocketManager struct {
	FeedSockets []*FeedSockets
	FeedManager *FeedManager
}

func NewWebSocketManager(fm *FeedManager) *WebSocketManager {
	return &WebSocketManager{
		FeedManager: fm,
	}
}

var upgrader = websocket.Upgrader{} // use default options

func (m *WebSocketManager) FeedSocketsForFeed(feedName string) *FeedSockets {
	logger.Debug("Searching FeedSockets", slog.Int("count", len(m.FeedSockets)), slog.String("feedName", feedName))

	for _, fs := range m.FeedSockets {
		if fs.feedName == feedName {
			return fs
		}
	}
	return nil
}

func (m *WebSocketManager) RunSocketForFeed(feedName string, w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)

	feedSockets := m.FeedSocketsForFeed(feedName)

	if feedSockets == nil {
		logger.Debug("Adding FeedSockets", slog.Int("count_before", len(m.FeedSockets)), slog.String("feedName", feedName))
		feedSockets = &FeedSockets{
			feedName: feedName,
		}
		m.FeedSockets = append(m.FeedSockets, feedSockets)
	}

	logger.Debug("Adding connection", slog.Int("count", len(feedSockets.websockets)))
	feedSockets.websockets = append(feedSockets.websockets, c)
	logger.Debug("Added connection", slog.Int("count", len(feedSockets.websockets)))

	logger.Info("WebSocket added", slog.Int("array size", len(feedSockets.websockets)))

	if err != nil {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to upgrade WebSocket")
	}
	secret, _ := utils.GetSecret(r)

	defer func() {
		feedSockets.RemoveConn(c)
		c.Close()
	}()

	for {
		mt, message, err := c.ReadMessage()
		logger.Debug("Message Received", slog.String("message", string(message)), slog.Int("messageType", mt))
		if err != nil {
			slog.Info("Error reading message", slog.String("error", err.Error()), slog.Int("messageType", mt))
			break
		}
		switch strings.TrimSpace(string(message)) {
		case "feed":
			pf, err := m.FeedManager.GetPublicFeed(feedName, secret)
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

func (m *WebSocketManager) NotifyAdd(item *PublicFeedItem) error {
	logger.Debug("Notify websocket", slog.Any("item", item), slog.Int("ws count", len(m.FeedSockets)))
	for _, f := range m.FeedSockets {
		slog.Info("checking feed", slog.String("feedName", f.feedName))
		if f.feedName == item.Feed.Name {
			slog.Info("Found feed", slog.String("feedName", f.feedName))
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

func (m *WebSocketManager) NotifyRemove(item *PublicFeedItem) error {
	logger.Debug("Notify websocket", slog.Any("item", item), slog.Int("ws count", len(m.FeedSockets)))
	for _, f := range m.FeedSockets {
		slog.Info("checking feed", slog.String("feedName", f.feedName))
		if f.feedName == item.Feed.Name {
			slog.Info("Found feed", slog.String("feedName", f.feedName))
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
