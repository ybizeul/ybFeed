package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"io/fs"
	"net/http"
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
	BasePath string
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
		} else if yberr.Code == 401 {
			w.WriteHeader(401)
			w.Write([]byte("Access denied"))
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
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
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
		w.WriteHeader(401)
		w.Write([]byte(yberr.Error()))
		return
	}
	pin, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
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
		if yberr.Code == 404 {
			w.WriteHeader(401)
			w.Write([]byte("No such feed"))
			return
		} else if yberr.Code == 401 {
			w.WriteHeader(401)
			w.Write([]byte("Access denied"))
			return
		}
	}

	feedItem := html.UnescapeString(strings.Split(r.URL.Path, "/")[4])

	content, err := f.GetItem(feedItem)

	if err != nil {
		yberr := err.(*feed.FeedError)
		if yberr.Code == 500 {
			w.WriteHeader(500)
			w.Write([]byte(yberr.Error()))
			return
		}
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
		if yberr.Code == 404 {
			w.WriteHeader(401)
			w.Write([]byte("No such feed"))
			return
		} else if yberr.Code == 401 {
			w.WriteHeader(401)
			w.Write([]byte(err.Error()))
			return
		}
	}

	contentType := r.Header.Get("Content-type")

	f.AddItem(contentType, r.Body)

	w.Write([]byte("OK"))
}

func (api *ApiHandler) feedItemDeleteHandlerFunc(w http.ResponseWriter, r *http.Request) {
	slog.Default().WithGroup("http").Debug("Item API DELETE request", slog.Any("request", r))

	secret, _ := utils.GetSecret(r)

	feedName := strings.Split(r.URL.Path, "/")[3]

	f, err := feed.GetFeed(api.BasePath, feedName, secret)

	if err != nil {
		yberr := err.(*feed.FeedError)
		if yberr.Code == 404 {
			w.WriteHeader(401)
			w.Write([]byte("No such feed"))
			return
		} else if yberr.Code == 401 {
			w.WriteHeader(401)
			w.Write([]byte("Access denied"))
			return
		}
	}

	item := strings.Split(r.URL.Path, "/")[4]

	err = f.RemoveItem(item)
	w.Write([]byte("Item Removed"))
}
