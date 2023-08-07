package main

import (
	"io/fs"
	"net/http"

	log "github.com/sirupsen/logrus"
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
	secret = r.URL.Query().Get("secret")
	if secret != "" {
		log.Debugf("Found secret in URL (%d)", len(secret))
		fromURL = true
	}

	if secret == "" {
		for _, c := range r.Cookies() {
			if c.Name == "Secret" {
				secret = c.Value
				log.Debugf("Found secret in Cookie (%d)", len(secret))
			}
		}
	}

	return secret, fromURL
}

// func checkSecret(w http.ResponseWriter, request *http.Request) (string, error) {
// 	feed := strings.Split(request.URL.Path, "/")[3]
// 	feedPath := path.Join(dataDir, feed)

// 	// Feed exists, check secret

// 	feedToken, err := os.ReadFile(path.Join(feedPath, "secret"))
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	var token string

// 	q := request.URL.Query().Get("secret")
// 	if q != "" {
// 		log.Debugf("Found token in URL (%d)", len(q))
// 		token = q
// 	}

// 	if token == "" {
// 		for _, c := range request.Cookies() {
// 			if c.Name == "Secret" {
// 				token = c.Value
// 			}
// 		}
// 	}

// 	if token != string(feedToken) {
// 		return "", &FeedError{
// 			Code:    401,
// 			Message: "Authorization failed",
// 		}
// 	}

// 	http.SetCookie(w, &http.Cookie{
// 		Name:    "Secret",
// 		Value:   token,
// 		Path:    fmt.Sprintf("/api/feed/%s", feed),
// 		Expires: time.Now().Add(time.Hour * 24 * 365 * 10),
// 	})

// 	return token, nil
// }
