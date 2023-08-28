package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/exp/slog"

	"github.com/ybizeul/ybfeed/internal/feed"
	"github.com/ybizeul/ybfeed/internal/utils"
	"github.com/ybizeul/ybfeed/web/ui"
)

var handler = http.FileServer(http.FS(ui.GetUiFs()))

func RootHandlerFunc(w http.ResponseWriter, r *http.Request) {
	slog.Default().WithGroup("http").Debug("Root request", slog.Any("request", r))
	p := r.URL.Path

	ui := ui.GetUiFs()

	//
	// Serve path from web UI if file exists
	//

	// Strip "/" at the beginning of path
	p = p[1:]

	matches, err := fs.Glob(ui, p)

	if err != nil {
		slog.Error("Unable to get web ui fs", slog.String("error", err.Error()))
	}

	if len(matches) == 1 {
		handler.ServeHTTP(w, r)
		return
	}

	//
	// For everything else, it goes to index.html
	//

	content, err := fs.ReadFile(ui, "index.html")
	if err != nil {
		slog.Error("Unable to read index.html from web ui", slog.String("error", err.Error()))
	}
	w.Write(content)
}

// Handle requests to /api
type ApiHandler struct {
	BasePath    string
	Version     string
	MaxBodySize int
}

func NewApiHandler(basePath string) *ApiHandler {
	os.MkdirAll(basePath, 0700)
	return &ApiHandler{
		BasePath: basePath,
	}
}

func (api *ApiHandler) ApiHandleFunc(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimSuffix(r.URL.Path, "/")
	split := strings.Split(p, "/")

	w.Header().Add("ybFeed-Version", api.Version)

	if len(split) == 4 {
		if r.Method == "GET" {
			api.feedHandlerFunc(w, r)
		} else if r.Method == "POST" {
			api.feedPostHandlerFunc(w, r)
		} else if r.Method == "PATCH" {
			api.feedPatchHandlerFunc(w, r)
		}
	} else if len(split) == 5 {
		if r.Method == "GET" {
			api.feedItemHandlerFunc(w, r)
		} else if r.Method == "DELETE" {
			api.feedItemDeleteHandlerFunc(w, r)
		}
	} else if len(split) == 2 {
		w.WriteHeader(200)
		return
	} else {
		utils.CloseWithCodeAndMessage(w, 400, "Malformed request")
	}
}

func (api *ApiHandler) feedHandlerFunc(w http.ResponseWriter, r *http.Request) {
	slog.Default().WithGroup("http").Debug("Feed API request", slog.Any("request", r))

	secret, fromURL := utils.GetSecret(r)

	feedName := strings.Split(r.URL.Path, "/")[3]

	f, err := feed.GetFeed(api.BasePath, feedName, secret)

	if err != nil {
		yberr := err.(*feed.FeedError)
		if yberr.Code == 404 {
			f, err = feed.NewFeed(api.BasePath, feedName)

			http.SetCookie(w, &http.Cookie{
				Name:    "Secret",
				Value:   f.Secret,
				Path:    fmt.Sprintf("/api/feed/%s", feedName),
				Expires: time.Now().Add(time.Hour * 24 * 365 * 10),
			})
		} else {
			utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
			return
		}
	}

	if fromURL {
		http.SetCookie(w, &http.Cookie{
			Name:    "Secret",
			Value:   f.Secret,
			Path:    fmt.Sprintf("/api/feed/%s", feedName),
			Expires: time.Now().Add(time.Hour * 24 * 365 * 10),
		})
	}

	j, err := json.Marshal(f)
	if err != nil {
		utils.CloseWithCodeAndMessage(w, 500, err.Error())
		return
	}
	w.Write(j)
}

func (api *ApiHandler) feedPatchHandlerFunc(w http.ResponseWriter, r *http.Request) {
	slog.Default().WithGroup("http").Debug("Feed API Set PIN request", slog.Any("request", r))
	secret, _ := utils.GetSecret(r)

	feedName := strings.Split(r.URL.Path, "/")[3]

	f, err := feed.GetFeed(api.BasePath, feedName, secret)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}
	pin, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	if len(pin) != 4 {
		utils.CloseWithCodeAndMessage(w, 400, "Malformed PIN")
		return
	}
	f.SetPIN(string(pin))
}

func (api *ApiHandler) feedItemHandlerFunc(w http.ResponseWriter, r *http.Request) {
	slog.Default().WithGroup("http").Debug("Item API GET request", slog.Any("request", r))

	secret, _ := utils.GetSecret(r)

	feedName := strings.Split(r.URL.Path, "/")[3]

	f, err := feed.GetFeed(api.BasePath, feedName, secret)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	fileNameElement := strings.Split(r.URL.Path, "/")[4]
	feedItem, err := url.QueryUnescape(fileNameElement)

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("Unable to parse file name '%s'", fileNameElement)))
	}

	content, err := f.GetItem(feedItem)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}
	w.Write(content)
}

func (api *ApiHandler) feedPostHandlerFunc(w http.ResponseWriter, r *http.Request) {
	slog.Default().WithGroup("http").Debug("Item API POST request", slog.Any("request", r))

	secret, _ := utils.GetSecret(r)

	feedName := strings.Split(r.URL.Path, "/")[3]

	f, err := feed.GetFeed(api.BasePath, feedName, secret)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	contentType := r.Header.Get("Content-type")

	err = f.AddItem(contentType, http.MaxBytesReader(w, r.Body, int64(api.MaxBodySize)))
	//err = f.AddItem(contentType, r.Body)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	w.Write([]byte("OK"))
}

func (api *ApiHandler) feedItemDeleteHandlerFunc(w http.ResponseWriter, r *http.Request) {
	slog.Default().WithGroup("http").Debug("Item API DELETE request", slog.Any("request", r))

	secret, _ := utils.GetSecret(r)

	feedName := strings.Split(r.URL.Path, "/")[3]

	f, err := feed.GetFeed(api.BasePath, feedName, secret)

	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}

	item, err := url.QueryUnescape(strings.Split(r.URL.Path, "/")[4])
	if err != nil {
		utils.CloseWithCodeAndMessage(w, 500, "Unable to unescape query string")
		return
	}
	err = f.RemoveItem(item)
	if err != nil {
		yberr := err.(*feed.FeedError)
		utils.CloseWithCodeAndMessage(w, yberr.Code, yberr.Error())
		return
	}
	w.Write([]byte("Item Removed"))
}
