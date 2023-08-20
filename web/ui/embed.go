package ui

import (
	"embed"
	"io/fs"
	"os"

	"golang.org/x/exp/slog"
)

//go:embed build
var EMBED_UI embed.FS

func GetUiFs() fs.FS {
	//embedRoot, err := fs.Sub(EMBED_UI, "ui")
	embedRoot, err := fs.Sub(EMBED_UI, "build")
	if err != nil {
		slog.Error("Unable to get root for web ui", slog.String("error", err.Error()))
		os.Exit(1)
	}
	return embedRoot
	// return http.FileServer(http.FS(embedRoot))
}
