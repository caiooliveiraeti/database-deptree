package web

import (
	"embed"
	"io/fs"
)

//go:embed index.html
var files embed.FS

// FS returns the embedded web assets as an fs.FS rooted at the web directory.
func FS() fs.FS {
	return files
}
