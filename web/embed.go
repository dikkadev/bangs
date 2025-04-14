package web

import (
	"embed"
	"io/fs"
)

//go:embed all:frontend/dist
var embeddedFilesystem embed.FS

// FrontendFS provides access to the embedded frontend filesystem,
// rooted at the "frontend/dist" directory.
func FrontendFS() (fs.FS, error) {
	return fs.Sub(embeddedFilesystem, "frontend/dist")
}
