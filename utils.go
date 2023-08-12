package main

import (
	"io/fs"
	"net/http"
	"os"

	"golang.org/x/exp/slog"
)

func getUiFs() fs.FS {
	embedRoot, err := fs.Sub(EMBED_UI, "ui")
	embedRoot, err = fs.Sub(embedRoot, "build")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	return embedRoot
	// return http.FileServer(http.FS(embedRoot))
}

func getSecret(r *http.Request) (string, bool) {
	var secret string
	var fromURL = false
	secret = r.URL.Query().Get("secret")
	if secret != "" {
		slog.Debug("Found secret in URL", "secret_len", slog.IntValue(len(secret)))
		fromURL = true
	}

	if secret == "" {
		for _, c := range r.Cookies() {
			if c.Name == "Secret" {
				secret = c.Value
				slog.Debug("Found secret in Cookie", slog.IntValue(len(secret)))
			}
		}
	}

	return secret, fromURL
}
