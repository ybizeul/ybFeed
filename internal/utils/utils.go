package utils

import (
	"net/http"

	"golang.org/x/exp/slog"
)

func GetSecret(r *http.Request) (string, bool) {
	var secret string
	var fromURL = false
	secret = r.URL.Query().Get("secret")
	if secret != "" {
		slog.Debug("Found secret in URL", slog.Int("secret_len", len(secret)))
		fromURL = true
	}

	if secret == "" {
		for _, c := range r.Cookies() {
			if c.Name == "Secret" {
				secret = c.Value
				slog.Debug("Found secret in Cookie", slog.Int("secret_len", len(secret)))
			}
		}
	}

	return secret, fromURL
}

func CloseWithCodeAndMessage(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	w.Write([]byte(message))
}
