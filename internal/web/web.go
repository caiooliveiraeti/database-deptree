package web

import (
	"embed"
	"io/fs"
)

//go:embed index.html html.js api.js colors.js App.js components
var files embed.FS

// FS returns the embedded web assets rooted at this package directory.
func FS() fs.FS {
	return files
}
