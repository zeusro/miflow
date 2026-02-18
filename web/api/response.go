// Package api provides HTTP API handlers organized by domain (DDD).
package api

import (
	"encoding/json"

	"github.com/gogf/gf/v2/net/ghttp"
)

// JSON writes JSON response.
func JSON(r *ghttp.Request, code int, v interface{}) {
	r.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
	r.Response.WriteStatus(code)
	if v != nil {
		b, _ := json.Marshal(v)
		r.Response.Write(b)
	}
}

// Err writes error JSON.
func Err(r *ghttp.Request, code int, msg string) {
	JSON(r, code, map[string]string{"error": msg})
}
