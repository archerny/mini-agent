// Package web provides the embedded frontend filesystem.
//
// When the dist/ directory exists at build time, the frontend assets are
// embedded into the binary. Use DistFS() to access them.
package web

import (
	"io/fs"
)

// DistFS returns the embedded frontend filesystem.
// Returns nil if the frontend was not embedded (dev mode / dist not built).
func DistFS() fs.FS {
	return distFS()
}

// DistAvailable returns true if the embedded frontend is available.
func DistAvailable() bool {
	return DistFS() != nil
}

// HasFile checks if a specific file exists in the embedded frontend.
func HasFile(name string) bool {
	f := DistFS()
	if f == nil {
		return false
	}
	_, err := fs.Stat(f, name)
	return err == nil
}
