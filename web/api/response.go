// Package api 提供按领域组织的 HTTP API 处理器 (DDD)。
package api

import (
	"encoding/json"

	"github.com/gogf/gf/v2/net/ghttp"
)

// JSON 写入 JSON 响应。
func JSON(r *ghttp.Request, code int, v interface{}) {
	r.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
	r.Response.WriteStatus(code)
	if v != nil {
		b, _ := json.Marshal(v)
		r.Response.Write(b)
	}
}

// Err 写入错误 JSON。
func Err(r *ghttp.Request, code int, msg string) {
	JSON(r, code, map[string]string{"error": msg})
}
