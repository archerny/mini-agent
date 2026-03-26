//go:build embedfrontend

package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var embeddedDist embed.FS

func distFS() fs.FS {
	sub, err := fs.Sub(embeddedDist, "dist")
	if err != nil {
		return nil
	}
	// Quick sanity check: index.html must exist.
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return nil
	}
	return sub
}
