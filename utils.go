package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

func getUiFs() fs.FS {
	embedRoot, err := fs.Sub(EMBED_UI, "ui")
	embedRoot, err = fs.Sub(embedRoot, "build")
	if err != nil {
		log.Fatal(err)
	}
	return embedRoot
	// return http.FileServer(http.FS(embedRoot))
}

func getSecret(r *http.Request) (string, bool) {
	var secret string
	var fromURL = false
	for _, c := range r.Cookies() {
		if c.Name == "Secret" {
			secret = c.Value
		}
	}
	if secret == "" {
		secret = r.URL.Query().Get("secret")
		if secret != "" {
			fromURL = true
		}
	}

	return secret, fromURL
}

func checkSecret(w http.ResponseWriter, request *http.Request) (string, error) {
	feed := strings.Split(request.URL.Path, "/")[3]
	feedPath := path.Join(dataDir, feed)

	// Feed exists, check secret

	feedToken, err := os.ReadFile(path.Join(feedPath, "secret"))
	if err != nil {
		log.Fatal(err)
	}

	var token string
	for _, c := range request.Cookies() {
		if c.Name == "Secret" {
			token = c.Value
		}
	}
	if token == "" {
		q := request.URL.Query().Get("secret")
		if q != "" {
			token = q
		}
	}

	if token != string(feedToken) {
		return "", &FeedError{
			Code:    401,
			Message: "Authorization failed",
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "Secret",
		Value:   token,
		Path:    fmt.Sprintf("/api/feed/%s", feed),
		Expires: time.Now().Add(time.Hour * 24 * 365 * 10),
	})

	return token, nil
}
