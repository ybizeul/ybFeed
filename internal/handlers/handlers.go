package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"golang.org/x/exp/slog"

	"github.com/Appboy/webpush-go"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/ybizeul/ybfeed/internal/feed"
	"github.com/ybizeul/ybfeed/internal/utils"

	"github.com/ybizeul/ybfeed/web/ui"
)

var logLevel = new(slog.LevelVar)
var logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})).WithGroup("http")

func init() {
	if os.Getenv("DEBUG") != "" || os.Getenv("DEBUG_HTTP") != "" {
		logLevel.Set(slog.LevelDebug)
	}
}

var webUiHandler = http.FileServer(http.FS(ui.GetUiFs()))

// RootHandlerFunc figures out how to handle incoming HTTP requests.
// If the requests points to an existing file in web UI (CSS, JS, etc)
// then it serves this file from webUiHandler, otherwise it returns
// index.html for proper react routing
func RootHandlerFunc(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Root request", slog.Any("request", r))

	p := r.URL.Path

	ui := ui.GetUiFs()

	//
	// Serve path from web UI if file exists
	//

	// Strip "/" at the beginning of path
	p = p[1:]

	matches, err := fs.Glob(ui, p)

	if err != nil {
		logger.Error("Unable to get web ui fs", slog.String("error", err.Error()))
	}

	if len(matches) == 1 {
		webUiHandler.ServeHTTP(w, r)
		return
	}

	//
	// Everything else goes to index.html
	//

	content, err := fs.ReadFile(ui, "index.html")
	if err != nil {
		logger.Error("Unable to read index.html from web ui", slog.String("error", err.Error()))
	}

	_, err = w.Write(content)
	if err != nil {
		logger.Error("Error while writing HTTP response", slog.String("error", err.Error()))
	}
}

// Handle requests to /api
type ApiHandler struct {
	BasePath         string
	Version          string
	MaxBodySize      int
	Config           APIConfig
	HttpPort         int
	WebSocketManager *feed.WebSocketManager
	FeedManager      *feed.FeedManager
}

type APIConfig struct {
	NotificationSettings *NotificationSettings `json:"notification,omitempty"`
}

func APIConfigFromFile(p string) (*APIConfig, error) {
	var config = &APIConfig{}
	d, err := os.ReadFile(path.Join(p))
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		} else {
			return nil, err
		}
	}
	err = json.Unmarshal(d, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

type NotificationSettings struct {
	VAPIDPublicKey  string
	VAPIDPrivateKey string
}

func NewApiHandler(basePath string) (*ApiHandler, error) {
	if err := os.MkdirAll(basePath, 0700); err != nil {
		return nil, err
	}

	// Check configuration
	var config, err = APIConfigFromFile(path.Join(basePath, "config.json"))
	if err != nil {
		return nil, err
	}
	if config.NotificationSettings == nil {
		privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
		if err != nil {
			return nil, err
		}
		config.NotificationSettings = &NotificationSettings{
			VAPIDPublicKey:  publicKey,
			VAPIDPrivateKey: privateKey,
		}
	}

	ws := feed.WebSocketManager{}
	result := &ApiHandler{
		BasePath:         basePath,
		Config:           *config,
		FeedManager:      feed.NewFeedManager(basePath, &ws),
		WebSocketManager: &ws,
	}

	ws.FeedManager = result.FeedManager

	if err = result.WriteConfig(); err != nil {
		return nil, err
	}

	return result, nil
}
func (api *ApiHandler) WriteConfig() error {
	b, err := json.Marshal(api.Config)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(api.BasePath, "config.json"), b, 0600)
	if err != nil {
		return err
	}
	return nil
}
func (api *ApiHandler) StartServer() {
	r := api.GetServer()
	err := http.ListenAndServe(fmt.Sprintf(":%d", api.HttpPort), r)
	if err != nil {
		logger.Error("Unable to start HTTP server",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
}
func (api *ApiHandler) GetServer() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			w.Header().Add("ybFeed-Version", api.Version)
			w.Header().Add("ybFeed-VAPIDPublicKey", api.Config.NotificationSettings.VAPIDPublicKey)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	r.Mount("/ws/{feedName}", http.HandlerFunc(api.feedWSHandler))

	r.Get("/api", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("OK")); err != nil {
			logger.Error("Cannot write Ping response")
		}
	})

	r.Route("/api/feed", func(r chi.Router) {
		r.Get("/{feedName}", api.feedHandlerFunc)
		r.Post("/{feedName}", api.feedPostHandlerFunc)
		r.Patch("/{feedName}", api.feedPatchHandlerFunc)

		r.Get("/{feedName}/{itemName}", api.feedItemHandlerFunc)
		r.Delete("/{feedName}/{itemName}", api.feedItemDeleteHandlerFunc)

		r.Post("/{feedName}/subscription", api.feedSubscriptionHandlerFunc)
		r.Delete("/{feedName}/subscription", api.feedUnsubscribeHandlerFunc)
	})
	r.Get("/*", RootHandlerFunc)

	slog.Info("ybFeed starting",
		slog.String("version", api.Version),
		slog.String("data_dir", api.BasePath),
		slog.Int("port", api.HttpPort),
		slog.Int("max-upload-size", api.MaxBodySize))

	return r
}

func (api *ApiHandler) feedWSHandler(w http.ResponseWriter, r *http.Request) {

	secret, _ := utils.GetSecret(r)

	feedName, _ := url.QueryUnescape(chi.URLParam(r, "feedName"))

	if feedName == "" {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to obtain feed name")
	}

	p, err := api.FeedManager.GetFeed(feedName)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	err = p.IsSecretValid(secret)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	api.WebSocketManager.RunSocketForFeed(feedName, w, r)
}

func (api *ApiHandler) feedHandlerFunc(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Feed API request", slog.Any("request", r))

	secret, fromURL := utils.GetSecret(r)

	feedName, _ := url.QueryUnescape(chi.URLParam(r, "feedName"))

	if feedName == "" {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to obtain feed name")
	}

	p, err := api.FeedManager.GetFeed(feedName)

	if err != nil {
		yberr := err.(*feed.FeedError)
		if yberr.Code == 404 {
			p, err = api.FeedManager.NewFeed(feedName)
			secret = p.Config.Secret
			if err != nil {
				yberr := err.(*feed.FeedError)
				utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
				return
			}
		} else {
			utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
			return
		}
	}

	result, err := api.FeedManager.GetPublicFeed(feedName, p.Config.Secret)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	err = p.IsSecretValid(secret)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "Secret",
		Value:   result.Secret,
		Path:    fmt.Sprintf("/api/feed/%s", feedName),
		Expires: time.Now().Add(time.Hour * 24 * 365 * 10),
	})

	if fromURL {
		http.SetCookie(w, &http.Cookie{
			Name:    "Secret",
			Value:   result.Secret,
			Path:    fmt.Sprintf("/api/feed/%s", feedName),
			Expires: time.Now().Add(time.Hour * 24 * 365 * 10),
		})
	}

	j, err := json.Marshal(result)
	if err != nil {
		utils.CloseWithCodeAndMessage(w, 500, err.Error())
		return
	}
	if _, err = w.Write(j); err != nil {
		logger.Error("Error while writing HTTP response", slog.String("error", err.Error()))
	}
}

func (api *ApiHandler) feedPatchHandlerFunc(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Feed API Set PIN request", slog.Any("request", r))
	secret, _ := utils.GetSecret(r)

	feedName, _ := url.QueryUnescape(chi.URLParam(r, "feedName"))
	if feedName == "" {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to obtain feed name")
	}
	f, err := api.FeedManager.GetFeed(feedName)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	err = f.IsSecretValid(secret)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}
	pin, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		if _, err = w.Write([]byte(err.Error())); err != nil {
			logger.Error("Error while writing HTTP response", slog.String("error", err.Error()))
		}
		return
	}

	if len(pin) != 4 {
		utils.CloseWithCodeAndMessage(w, 400, "Malformed PIN")
		return
	}
	if err = f.SetPIN(string(pin)); err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}
}

func (api *ApiHandler) feedItemHandlerFunc(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Item API GET request", slog.Any("request", r))

	secret, _ := utils.GetSecret(r)

	feedName, _ := url.QueryUnescape(chi.URLParam(r, "feedName"))
	if feedName == "" {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to obtain feed name")
		return
	}

	f, err := api.FeedManager.GetFeed(feedName)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	err = f.IsSecretValid(secret)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	feedItem, _ := url.QueryUnescape(chi.URLParam(r, "itemName"))

	if feedItem == "" {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to obtain feed item")
	}
	content, err := f.GetItemData(feedItem)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}
	if _, err = w.Write(content); err != nil {
		logger.Error("Error while writing HTTP response", slog.String("error", err.Error()))
	}
}

func (api *ApiHandler) feedPostHandlerFunc(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Item API POST request", slog.Any("request", r))

	secret, _ := utils.GetSecret(r)

	feedName, _ := url.QueryUnescape(chi.URLParam(r, "feedName"))
	if feedName == "" {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to obtain feed name")
	}

	f, err := api.FeedManager.GetFeed(feedName)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	err = f.IsSecretValid(secret)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	contentType := r.Header.Get("Content-type")

	err = f.AddItem(contentType, http.MaxBytesReader(w, r.Body, int64(api.MaxBodySize)))

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	// Send push notifications
	for _, subscription := range f.Config.Subscriptions {
		slog.Debug("Sending push notification", slog.String("endpoint", subscription.Endpoint))
		resp, _ := webpush.SendNotification([]byte(fmt.Sprintf("New item posted to feed %s", f.Name())), &subscription, &webpush.Options{
			Subscriber:      "example@example.com", // Do not include "mailto:"
			VAPIDPublicKey:  api.Config.NotificationSettings.VAPIDPublicKey,
			VAPIDPrivateKey: api.Config.NotificationSettings.VAPIDPrivateKey,
			TTL:             30,
		})
		slog.Debug("Response", slog.Any("resp", resp))
		defer resp.Body.Close()
	}

	if _, err = w.Write([]byte("OK")); err != nil {
		slog.Error("Error while writing HTTP response", slog.String("error", err.Error()))
	}
}

func (api *ApiHandler) feedItemDeleteHandlerFunc(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Item API DELETE request", slog.Any("request", r))

	secret, _ := utils.GetSecret(r)

	feedName, _ := url.QueryUnescape(chi.URLParam(r, "feedName"))
	if feedName == "" {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to obtain feed name")
	}

	f, err := api.FeedManager.GetFeed(feedName)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	err = f.IsSecretValid(secret)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	feedItem, _ := url.QueryUnescape(chi.URLParam(r, "itemName"))
	if feedItem == "" {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to obtain feed item")
	}

	err = f.RemoveItem(feedItem)
	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	if _, err = w.Write([]byte("Item Removed")); err != nil {
		logger.Error("Error while writing HTTP response", slog.String("error", err.Error()))
	}
}

func (api *ApiHandler) feedSubscriptionHandlerFunc(w http.ResponseWriter, r *http.Request) {

	logger.Debug("Feed subscription request", slog.Any("request", r))

	secret, _ := utils.GetSecret(r)

	feedName, _ := url.QueryUnescape(chi.URLParam(r, "feedName"))
	if feedName == "" {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to obtain feed name")
	}

	f, err := api.FeedManager.GetFeed(feedName)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	err = f.IsSecretValid(secret)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	body, err := io.ReadAll(r.Body)

	defer r.Body.Close()

	if err != nil {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to read subscription")
		return
	}

	var s webpush.Subscription

	err = json.Unmarshal(body, &s)

	if err != nil {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to parse subscription")
		return
	}

	err = f.Config.AddSubscription(s)

	if err != nil {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to add subscription")
		return
	}
}

func (api *ApiHandler) feedUnsubscribeHandlerFunc(w http.ResponseWriter, r *http.Request) {

	logger.Debug("Feed subscription request", slog.Any("request", r))

	secret, _ := utils.GetSecret(r)

	feedName, _ := url.QueryUnescape(chi.URLParam(r, "feedName"))
	if feedName == "" {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to obtain feed name")
	}

	f, err := api.FeedManager.GetFeed(feedName)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	err = f.IsSecretValid(secret)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	body, err := io.ReadAll(r.Body)

	defer r.Body.Close()

	if err != nil {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to read subscription")
		return
	}

	var s webpush.Subscription

	err = json.Unmarshal(body, &s)

	if err != nil {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to parse subscription")
		return
	}

	err = f.Config.DeleteSubscription(s)

	if err != nil {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to add subscription")
		return
	}
}
