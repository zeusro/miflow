package api

import (
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/zeusro/miflow/web"
)

// RequireAuth returns 401 if no valid token.
func RequireAuth(a *web.App, r *ghttp.Request) bool {
	if a.DeviceAPI() == nil {
		Err(r, http.StatusUnauthorized, "请先登录 (run login or visit /login)")
		return false
	}
	return true
}
