package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	ws "github.com/gorilla/websocket"
	"golang.org/x/exp/slog"

	"github.com/Appboy/webpush-go"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/ybizeul/ybfeed/internal/feed"
	"github.com/ybizeul/ybfeed/internal/utils"
	"github.com/ybizeul/ybfeed/pkg/yblog"

	"github.com/ybizeul/ybfeed/web/ui"
)

var hL = yblog.NewYBLogger("http", []string{"DEBUG", "DEBUG_HTTP"})

var webUiHandler = http.FileServer(http.FS(ui.GetUiFs()))

// RootHandlerFunc figures out how to handle incoming HTTP requests.
// If the requests points to an existing file in web UI (CSS, JS, etc)
// then it serves this file from webUiHandler, otherwise it returns
// index.html for proper react routing
func RootHandlerFunc(w http.ResponseWriter, r *http.Request) {
	hL.Logger.Debug("Root request", slog.String("request_uri", r.RequestURI))

	p := r.URL.Path

	ui := ui.GetUiFs()

	//
	// Serve path from web UI if file exists
	//

	// Strip "/" at the beginning of path
	p = p[1:]

	matches, err := fs.Glob(ui, p)

	if err != nil {
		hL.Logger.Error("Unable to get web ui fs", slog.String("error", err.Error()))
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
		hL.Logger.Error("Unable to read index.html from web ui", slog.String("error", err.Error()))
	}

	_, err = w.Write(content)
	if err != nil {
		hL.Logger.Error("Error while writing HTTP response", slog.String("error", err.Error()))
	}
}

// Handle requests to /api
type ApiHandler struct {
	BasePath         string
	Version          string
	MaxBodySize      int
	Config           APIConfig
	HttpPort         int
	ListenAddr       string
	WebSocketManager *feed.WebSocketManager
	FeedManager      *feed.FeedManager
}

type APIConfig struct {
	NotificationSettings *feed.NotificationSettings `json:"notification,omitempty"`
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
		config.NotificationSettings = &feed.NotificationSettings{
			VAPIDPublicKey:  publicKey,
			VAPIDPrivateKey: privateKey,
		}
	}

	ws := feed.WebSocketManager{}

	fm := feed.NewFeedManager(basePath, &ws)
	fm.NotificationSettings = config.NotificationSettings
	result := &ApiHandler{
		BasePath:         basePath,
		Config:           *config,
		FeedManager:      fm,
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
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", api.ListenAddr, api.HttpPort), r)
	if err != nil {
		hL.Logger.Error("Unable to start HTTP server",
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
			hL.Logger.Error("Cannot write Ping response")
		}
	})

	r.Post("/api/secrets", api.postSecretsHandler)
	r.Route("/api/feeds", func(r chi.Router) {
		r.Get("/{feedName}", api.feedGetFunc)
		r.Post("/{feedName}", api.feedPostFunc)
		r.Patch("/{feedName}", api.feedPatchFunc)
		r.Post("/{feedName}/subscription", api.subscriptionPostFunc)
		r.Delete("/{feedName}/subscription", api.subscriptionDeleteFunc)
		r.Delete("/{feedName}/items", api.itemsDeleteFunc)
		r.Get("/{feedName}/items/{itemName}", api.itemGetFunc)
		r.Delete("/{feedName}/items/{itemName}", api.itemDeleteFunc)
	})
	r.Get("/*", RootHandlerFunc)

	slog.Info("ybFeed starting",
		slog.String("version", api.Version),
		slog.String("data_dir", api.BasePath),
		slog.Int("port", api.HttpPort),
		slog.String("address", api.ListenAddr),
		slog.Int("max-upload-size", api.MaxBodySize))

	return r
}

func (api *ApiHandler) feedWSHandler(w http.ResponseWriter, r *http.Request) {

	secret, _ := utils.GetSecret(r)

	feedName, _ := url.QueryUnescape(chi.URLParam(r, "feedName"))

	if feedName == "" {
		WriteError(w, http.StatusBadRequest, "Unable to obtain feed name")
		return
	}

	_, err := api.FeedManager.GetFeedWithAuth(feedName, secret)

	if err != nil {
		// A web socket doesn't have a standard http status code, so we need
		// to open it and close it with a relevant code
		var upgrader = ws.Upgrader{}
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		switch {
		case errors.Is(err, feed.FeedErrorNotFound):
			_ = c.WriteControl(ws.CloseMessage, ws.FormatCloseMessage(http.StatusNotFound+4000, ""), time.Now().Add(time.Second))
		case errors.Is(err, feed.FeedErrorInvalidSecret):
			_ = c.WriteControl(ws.CloseMessage, ws.FormatCloseMessage(http.StatusUnauthorized+4000, ""), time.Now().Add(time.Second))
		default:
			_ = c.WriteControl(ws.CloseMessage, ws.FormatCloseMessage(http.StatusInternalServerError+4000, ""), time.Now().Add(time.Second))
		}
		c.Close()
		return
	}

	api.WebSocketManager.RunSocketForFeed(feedName, w, r)
}

func (api *ApiHandler) feedGetFunc(w http.ResponseWriter, r *http.Request) {
	hL.Logger.Debug("Feed API request", slog.String("request_uri", r.RequestURI))

	feedName, _ := url.QueryUnescape(chi.URLParam(r, "feedName"))

	if feedName == "" {
		WriteError(w, http.StatusBadRequest, "Unable to obtain feed name")
		return
	}

	var f *feed.Feed
	var err error

	f, err = api.FeedManager.GetFeed(feedName)

	if err != nil {
		if errors.Is(err, feed.FeedErrorNotFound) {
			_, err = feed.NewFeed(path.Join(api.BasePath, feedName))
			if err != nil {
				utils.CloseWithCodeAndMessage(w, 500, err.Error())
			}
			f, _ = api.FeedManager.GetFeed(feedName)
		} else {
			utils.CloseWithCodeAndMessage(w, 500, err.Error())
			return
		}
	} else {
		secret, _ := utils.GetSecret(r)
		hL.Logger.Debug("secret", slog.String("secret", secret))

		err = f.IsSecretValid(secret)
		if err != nil {
			switch {
			case errors.Is(err, feed.FeedErrorInvalidSecret) || errors.Is(err, feed.FeedConfigErrorPinExpired):
				utils.CloseWithCodeAndMessage(w, 401, "Unauthorized")
			default:
				utils.CloseWithCodeAndMessage(w, 500, err.Error())
			}
			return
		}
	}

	publicFeed, err := f.Public()
	if err != nil {
		utils.CloseWithCodeAndMessage(w, 500, err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "Secret",
		Value:   publicFeed.Secret,
		Path:    fmt.Sprintf("/api/feeds/%s", feedName),
		Expires: time.Now().Add(time.Hour * 24 * 365 * 10),
	})

	j, err := json.Marshal(publicFeed)
	if err != nil {
		utils.CloseWithCodeAndMessage(w, 500, err.Error())
		return
	}
	if _, err = w.Write(j); err != nil {
		hL.Logger.Error("Error while writing HTTP response", slog.String("error", err.Error()))
	}
}

func (api *ApiHandler) feedPatchFunc(w http.ResponseWriter, r *http.Request) {
	hL.Logger.Debug("Feed API Set PIN request", slog.String("request_uri", r.RequestURI))
	secret, _ := utils.GetSecret(r)

	feedName, _ := url.QueryUnescape(chi.URLParam(r, "feedName"))
	if feedName == "" {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to obtain feed name")
	}

	f, err := api.FeedManager.GetFeedWithAuth(feedName, secret)

	if err != nil {
		switch {
		case errors.Is(err, feed.FeedErrorNotFound):
			utils.CloseWithCodeAndMessage(w, 404, fmt.Sprintf("feed '%s' not found", feedName))
		case errors.Is(err, feed.FeedErrorInvalidSecret):
			utils.CloseWithCodeAndMessage(w, 401, "Unauthorized")
		default:
			utils.CloseWithCodeAndMessage(w, 500, fmt.Sprintf("Error while getting feed: %s", err.Error()))
		}
		return
	}

	pin, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		if _, err = w.Write([]byte(err.Error())); err != nil {
			hL.Logger.Error("Error while writing HTTP response", slog.String("error", err.Error()))
		}
		return
	}

	if err = f.SetPIN(string(pin)); err != nil {
		if errors.Is(err, feed.FeedConfigErrorPinIncorrectLength) {
			utils.CloseWithCodeAndMessage(w, 400, "PIN should be 4 digits")
		}
		utils.CloseWithCodeAndMessage(w, 500, err.Error())
		return
	}
}
func (api *ApiHandler) itemsDeleteFunc(w http.ResponseWriter, r *http.Request) {
	hL.Logger.Debug("Item API EMPTY request", slog.String("request_uri", r.RequestURI))

	secret, _ := utils.GetSecret(r)

	feedName, _ := url.QueryUnescape(chi.URLParam(r, "feedName"))
	if feedName == "" {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to obtain feed name")
	}

	f, err := api.FeedManager.GetFeedWithAuth(feedName, secret)
	if err != nil {
		switch {
		case errors.Is(err, feed.FeedErrorNotFound):
			utils.CloseWithCodeAndMessage(w, 404, fmt.Sprintf("feed '%s' not found", feedName))
		case errors.Is(err, feed.FeedErrorInvalidSecret):
			utils.CloseWithCodeAndMessage(w, 401, "Unauthorized")
		default:
			utils.CloseWithCodeAndMessage(w, 500, fmt.Sprintf("Error while getting feed: %s", err.Error()))
		}
		return
	}

	err = f.Empty()
	if err != nil {
		utils.CloseWithCodeAndMessage(w, 500, fmt.Sprintf("Error while getting feed: %s", err.Error()))
		return
	}
}

func (api *ApiHandler) itemGetFunc(w http.ResponseWriter, r *http.Request) {
	hL.Logger.Debug("Item API GET request", slog.String("request_uri", r.RequestURI))

	secret, _ := utils.GetSecret(r)

	feedName, _ := url.QueryUnescape(chi.URLParam(r, "feedName"))
	if feedName == "" {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to obtain feed name")
		return
	}

	f, err := api.FeedManager.GetFeedWithAuth(feedName, secret)

	if err != nil {
		switch {
		case errors.Is(err, feed.FeedErrorNotFound):
			utils.CloseWithCodeAndMessage(w, 404, fmt.Sprintf("feed '%s' not found", feedName))
		case errors.Is(err, feed.FeedErrorInvalidSecret):
			utils.CloseWithCodeAndMessage(w, 401, "Unauthorized")
		default:
			utils.CloseWithCodeAndMessage(w, 500, fmt.Sprintf("Error while getting feed: %s", err.Error()))
		}
		return
	}

	feedItem, _ := url.QueryUnescape(chi.URLParam(r, "itemName"))

	if feedItem == "" {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to obtain feed item")
	}
	content, err := f.GetItemData(feedItem)

	if err != nil {
		switch {
		case errors.Is(err, feed.FeedErrorItemNotFound):
			utils.CloseWithCodeAndMessage(w, 404, err.Error())
		default:
			utils.CloseWithCodeAndMessage(w, 500, err.Error())
		}
		return
	}
	if _, err = w.Write(content); err != nil {
		hL.Logger.Error("Error while writing HTTP response", slog.String("error", err.Error()))
	}
}

func (api *ApiHandler) feedPostFunc(w http.ResponseWriter, r *http.Request) {
	hL.Logger.Debug("Item API POST request", slog.String("request_uri", r.RequestURI))

	secret, _ := utils.GetSecret(r)

	feedName, _ := url.QueryUnescape(chi.URLParam(r, "feedName"))
	if feedName == "" {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to obtain feed name")
	}

	f, err := api.FeedManager.GetFeedWithAuth(feedName, secret)

	if err != nil {
		switch {
		case errors.Is(err, feed.FeedErrorNotFound):
			utils.CloseWithCodeAndMessage(w, 404, fmt.Sprintf("feed '%s' not found", feedName))
		case errors.Is(err, feed.FeedErrorInvalidSecret):
			utils.CloseWithCodeAndMessage(w, 401, "Unauthorized")
		default:
			utils.CloseWithCodeAndMessage(w, 500, fmt.Sprintf("Error while getting feed: %s", err.Error()))
		}
		return
	}

	mr, err := r.MultipartReader()
	if err != nil {
		utils.CloseWithCodeAndMessage(w, 500, fmt.Sprintf("Error while getting parts: %s", err.Error()))
		return
	}

	np, err := mr.NextPart()
	if err != nil {
		utils.CloseWithCodeAndMessage(w, 500, fmt.Sprintf("Error while getting next part: %s", err.Error()))
		return
	}

	contentType := np.Header.Get("Content-Type")

	err = f.AddItem(contentType, np.FileName(), http.MaxBytesReader(w, np, int64(api.MaxBodySize)))

	if err != nil {
		switch {
		case errors.Is(err, feed.FeedErrorInvalidContentType):
			utils.CloseWithCodeAndMessage(w, 400, "Content-type is not supported")
		case errors.Is(err, feed.FeedErrorMaxBodySizeExceeded):
			utils.CloseWithCodeAndMessage(w, 413, "Max size exceeded")
		default:
			utils.CloseWithCodeAndMessage(w, 500, err.Error())
		}
		return
	}

	if _, err := w.Write([]byte("OK")); err != nil {
		slog.Error("Error while writing HTTP response", slog.String("error", err.Error()))
	}
}

func (api *ApiHandler) itemDeleteFunc(w http.ResponseWriter, r *http.Request) {
	hL.Logger.Debug("Item API DELETE request", slog.String("request_uri", r.RequestURI))

	secret, _ := utils.GetSecret(r)

	feedName, _ := url.QueryUnescape(chi.URLParam(r, "feedName"))
	if feedName == "" {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to obtain feed name")
	}

	f, err := api.FeedManager.GetFeedWithAuth(feedName, secret)

	if err != nil {
		switch {
		case errors.Is(err, feed.FeedErrorNotFound):
			utils.CloseWithCodeAndMessage(w, 404, fmt.Sprintf("feed '%s' not found", feedName))
		case errors.Is(err, feed.FeedErrorInvalidSecret):
			utils.CloseWithCodeAndMessage(w, 401, "Unauthorized")
		default:
			utils.CloseWithCodeAndMessage(w, 500, fmt.Sprintf("Error while getting feed: %s", err.Error()))
		}
		return
	}

	feedItem, _ := url.QueryUnescape(chi.URLParam(r, "itemName"))
	if feedItem == "" {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to obtain feed item")
	}

	err = f.RemoveItem(feedItem, true)
	if err != nil {
		switch {
		case errors.Is(err, feed.FeedErrorItemNotFound):
			utils.CloseWithCodeAndMessage(w, 404, "Item does not exists")
		default:
			utils.CloseWithCodeAndMessage(w, 500, err.Error())
		}
		return
	}

	if _, err = w.Write([]byte("Item Removed")); err != nil {
		hL.Logger.Error("Error while writing HTTP response", slog.String("error", err.Error()))
	}
}

func (api *ApiHandler) subscriptionPostFunc(w http.ResponseWriter, r *http.Request) {

	hL.Logger.Debug("Feed subscription request", slog.String("request_uri", r.RequestURI))

	secret, _ := utils.GetSecret(r)

	feedName, _ := url.QueryUnescape(chi.URLParam(r, "feedName"))
	if feedName == "" {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to obtain feed name")
	}

	f, err := api.FeedManager.GetFeedWithAuth(feedName, secret)

	if err != nil {
		switch {
		case errors.Is(err, feed.FeedErrorNotFound):
			utils.CloseWithCodeAndMessage(w, 404, fmt.Sprintf("feed '%s' not found", feedName))
		case errors.Is(err, feed.FeedErrorInvalidSecret):
			utils.CloseWithCodeAndMessage(w, 401, "Unauthorized")
		default:
			utils.CloseWithCodeAndMessage(w, 500, fmt.Sprintf("Error while getting feed: %s", err.Error()))
		}
		return
	}

	var s webpush.Subscription

	err = json.NewDecoder(r.Body).Decode(&s)

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

func (api *ApiHandler) subscriptionDeleteFunc(w http.ResponseWriter, r *http.Request) {

	hL.Logger.Debug("Feed subscription request", slog.String("request_uri", r.RequestURI))

	secret, _ := utils.GetSecret(r)

	feedName, _ := url.QueryUnescape(chi.URLParam(r, "feedName"))
	if feedName == "" {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to obtain feed name")
	}

	f, err := api.FeedManager.GetFeedWithAuth(feedName, secret)

	if err != nil {
		switch {
		case errors.Is(err, feed.FeedErrorNotFound):
			utils.CloseWithCodeAndMessage(w, 404, fmt.Sprintf("feed '%s' not found", feedName))
		case errors.Is(err, feed.FeedErrorInvalidSecret):
			utils.CloseWithCodeAndMessage(w, 401, "Unauthorized")
		default:
			utils.CloseWithCodeAndMessage(w, 500, fmt.Sprintf("Error while getting feed: %s", err.Error()))
		}
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

func (api *ApiHandler) postSecretsHandler(w http.ResponseWriter, r *http.Request) {
	api.FeedManager.DumpSecrets()
}
