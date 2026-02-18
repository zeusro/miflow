// Package web provides embedded static assets for the miflow web server.
package web

import "embed"

//go:embed static
var StaticFS embed.FS
