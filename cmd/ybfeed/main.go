package main

import (
	"os"

	"github.com/urfave/cli/v2"
	"github.com/ybizeul/ybfeed/internal/handlers"
	"golang.org/x/exp/slog"
)

var HTTP_PORT int
var DEBUG bool
var dataDir string
var maxBodySize int

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
			&cli.IntFlag{
				Name:        "max-upload-size",
				Aliases:     []string{"m"},
				Value:       5,
				EnvVars:     []string{"YBF_MAX_UPLOAD_SIZE"},
				Usage:       "Max upload size in MB",
				Destination: &maxBodySize,
			},
		},
		Action: func(cCtx *cli.Context) error {
			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: &logLevel}))
			slog.SetDefault(logger)
			if DEBUG {
				logLevel.Set(slog.LevelDebug)
			}
			slog.Debug("Running in DEBUG mode")
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
	makeDataDirectory(dataDir)

	// Start HTTP Server
	api, err := handlers.NewApiHandler(dataDir)
	if err != nil {
		return
	}
	api.Version = version
	api.MaxBodySize = maxBodySize * 1024 * 1024
	api.HttpPort = HTTP_PORT

	api.StartServer()
}

func makeDataDirectory(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		slog.Debug("Creating data directory", slog.String("path", dir))
		if err = os.Mkdir(dir, 0700); err != nil {
			slog.Error("Unable to create data directory", slog.String("path", dataDir), slog.String("error", err.Error()))
			os.Exit(1)
		}
	}
}
