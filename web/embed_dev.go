//go:build !embedfrontend

package web

import "io/fs"

// distFS returns nil when the frontend is not embedded (dev mode).
// Run `make build-all` to build with embedded frontend.
func distFS() fs.FS {
	return nil
}
