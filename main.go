package main

import (
	"embed"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

//go:embed ui/build
var EMBED_UI embed.FS

var HTTP_PORT int

var dataDir string

func main() {
	app := &cli.App{
		Name:    "ybFeed",
		Version: version,
		Usage:   "Microfeeds for personal use",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "port",
				Aliases:     []string{"p"},
				Value:       8080,
				EnvVars:     []string{"YBF_HTTP_PORT"},
				Usage:       "TCP Port to listen",
				Destination: &HTTP_PORT,
			},
			&cli.StringFlag{
				Name:        "dir",
				Aliases:     []string{"d"},
				Value:       "./data",
				EnvVars:     []string{"YBF_DATA_DIR"},
				Usage:       "Data directory path",
				Destination: &dataDir,
			},
		},
		Action: func(cCtx *cli.Context) error {
			run()
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
func run() {
	// Initialize file system
	initialize()

	// Start HTTP Server
	r := http.NewServeMux()

	r.HandleFunc("/api/", apiHandleFunc)
	r.HandleFunc("/", rootHandlerFunc)

	log.Infof("ybFeed v%s starting serving from %s on port %d", version, dataDir, HTTP_PORT)
	err := http.ListenAndServe(fmt.Sprintf(":%d", HTTP_PORT), r)
	if err != nil {
		log.Fatal(err)
	}
}

func initialize() {
	if os.Getenv("DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
	}

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		err = os.Mkdir(dataDir, 0700)
	}
}
