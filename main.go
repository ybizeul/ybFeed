package main

import (
	"embed"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

//go:embed ui/build
var EMBED_UI embed.FS

var dataDir string

func main() {
	// Initialize file system
	initialize()

	// Start HTTP Server
	r := http.NewServeMux()
	//r.HandleFunc("/getFeed/", getFeedHandler)
	//r.Handle("/", getUiFs())
	r.HandleFunc("/api/", apiHandleFunc)
	r.HandleFunc("/", rootHandlerFunc)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}

func initialize() {
	if os.Getenv("DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
	}

	// Create data directory
	dataDir = os.Getenv("YBF_DATADIR")
	if len(dataDir) == 0 {
		dataDir = "data"
	}
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		err = os.Mkdir(dataDir, 0700)
	}
}
