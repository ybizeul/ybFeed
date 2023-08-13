package main

import (
	"embed"
	"fmt"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slog"
)

//go:embed ui/build
var EMBED_UI embed.FS

var HTTP_PORT int
var DEBUG bool
var dataDir string

var logLevel slog.LevelVar

func main() {
	logLevel.Set(slog.LevelInfo)
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
			&cli.BoolFlag{
				Name:        "debug",
				Value:       false,
				EnvVars:     []string{"DEBUG"},
				Usage:       "Debug Logging",
				Destination: &DEBUG,
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
			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: &logLevel}))
			slog.SetDefault(logger)
			if DEBUG {
				logLevel.Set(slog.LevelDebug)
			}
			slog.Info("Debugging", slog.Bool("status", DEBUG))
			run()
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("Unable to run app with arguments", slog.String("error", err.Error()), slog.Any("args", os.Args))
		os.Exit(1)
	}
}
func run() {
	// Initialize file system
	initialize()

	// Start HTTP Server
	r := http.NewServeMux()

	r.HandleFunc("/api/", apiHandleFunc)
	r.HandleFunc("/", rootHandlerFunc)

	slog.Info("ybFeed starting", slog.String("version", version), slog.String("data_dir", dataDir), slog.Int("port", HTTP_PORT))
	err := http.ListenAndServe(fmt.Sprintf(":%d", HTTP_PORT), r)
	if err != nil {
		slog.Error("Unable to start HTTP server", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func initialize() {
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		slog.Debug("Creating data directory", slog.String("path", dataDir))
		if err = os.Mkdir(dataDir, 0700); err != nil {
			slog.Error("Unable to create data directory", slog.String("path", dataDir), slog.String("error", err.Error()))
			os.Exit(1)
		}
	}
}
