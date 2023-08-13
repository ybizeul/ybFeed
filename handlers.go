package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"golang.org/x/exp/slog"
)

var handler = http.FileServer(http.FS(getUiFs()))

func rootHandlerFunc(w http.ResponseWriter, r *http.Request) {
	slog.Default().WithGroup("http").Debug("Root request", slog.Any("request", r))
	p := r.URL.Path

	ui := getUiFs()

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

//
// Handle requests to /api
//

func apiHandleFunc(w http.ResponseWriter, r *http.Request) {
	slog.Default().WithGroup("http").Debug("API request", slog.Any("request", r))
	split := strings.Split(r.URL.Path, "/")
	if len(split) == 4 {
		if r.Method == "GET" {
			feedHandlerFunc(w, r)
		} else if r.Method == "POST" {
			feedPostHandlerFunc(w, r)
		} else if r.Method == "PATCH" {
			feedPatchHandlerFunc(w, r)
		}
	} else if len(split) == 5 {
		if r.Method == "GET" {
			feedItemHandlerFunc(w, r)
		} else if r.Method == "DELETE" {
			feedItemDeleteHandlerFunc(w, r)
		}
	}
}

func feedHandlerFunc(w http.ResponseWriter, r *http.Request) {
	slog.Default().WithGroup("http").Debug("Feed API request", slog.Any("request", r))

	secret, fromURL := getSecret(r)

	feedName := strings.Split(r.URL.Path, "/")[3]

	feed, err := GetFeed(feedName, secret)

	if err != nil {
		yberr := err.(*FeedError)
		if yberr.Code == 404 {
			feed, err = NewFeed(feedName)

			http.SetCookie(w, &http.Cookie{
				Name:    "Secret",
				Value:   feed.Secret,
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
			Value:   feed.Secret,
			Path:    fmt.Sprintf("/api/feed/%s", feedName),
			Expires: time.Now().Add(time.Hour * 24 * 365 * 10),
		})
	}

	j, err := json.Marshal(feed)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(j)
}

func feedPatchHandlerFunc(w http.ResponseWriter, r *http.Request) {
	slog.Default().WithGroup("http").Debug("Feed API Set PIN request", slog.Any("request", r))
	secret, _ := getSecret(r)

	feedName := strings.Split(r.URL.Path, "/")[3]

	feed, err := GetFeed(feedName, secret)

	if err != nil {
		yberr := err.(*FeedError)
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
	feed.SetPIN(string(pin))
}

func feedItemHandlerFunc(w http.ResponseWriter, r *http.Request) {
	slog.Default().WithGroup("http").Debug("Item API GET request", slog.Any("request", r))

	secret, _ := getSecret(r)

	feedName := strings.Split(r.URL.Path, "/")[3]

	feed, err := GetFeed(feedName, secret)

	if err != nil {
		yberr := err.(*FeedError)
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

	content, err := feed.GetItem(feedItem)

	if err != nil {
		yberr := err.(*FeedError)
		if yberr.Code == 500 {
			w.WriteHeader(500)
			w.Write([]byte(yberr.Error()))
			return
		}
	}
	w.Write(content)
}

func feedPostHandlerFunc(w http.ResponseWriter, r *http.Request) {
	slog.Default().WithGroup("http").Debug("Item API POST request", slog.Any("request", r))

	secret, _ := getSecret(r)

	feedName := strings.Split(r.URL.Path, "/")[3]

	feed, err := GetFeed(feedName, secret)

	if err != nil {
		yberr := err.(*FeedError)
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

	feed.AddItem(contentType, r.Body)

	w.Write([]byte("OK"))
}

func feedItemDeleteHandlerFunc(w http.ResponseWriter, r *http.Request) {
	slog.Default().WithGroup("http").Debug("Item API DELETE request", slog.Any("request", r))

	secret, _ := getSecret(r)

	feedName := strings.Split(r.URL.Path, "/")[3]

	feed, err := GetFeed(feedName, secret)

	if err != nil {
		yberr := err.(*FeedError)
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

	err = feed.RemoveItem(item)
	w.Write([]byte("Item Removed"))
}
