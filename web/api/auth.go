package api

import (
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/web"
)

// RequireAuth 若无有效 token 则返回 401。
func RequireAuth(a *web.App, r *ghttp.Request) bool {
	if a.DeviceAPI() == nil {
		Err(r, http.StatusUnauthorized, "请先登录 (run login or visit /login)")
		return false
	}
	return true
}

// Login 处理 GET /login - 发起 OAuth 流程。
func Login(a *web.App, r *ghttp.Request) {
	oc := miaccount.NewOAuthClient()
	a.OAuthStore().Put(oc.State, oc)

	authURL := oc.GenAuthURL("", "", true)
	data, err := web.RenderLoginBytes(authURL)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError)
		r.Response.Write([]byte("模板渲染失败"))
		return
	}
	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.Response.Write(data)
}

// Callback 处理 GET /callback - OAuth 回调，用 code 换取 token 并保存。
func Callback(a *web.App, r *ghttp.Request) {
	code := r.Get("code").String()
	state := r.Get("state").String()
	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")

	writeError := func(status int, title, message string) {
		data, err := web.RenderErrorBytes(title, message)
		if err != nil {
			r.Response.WriteStatus(status)
			r.Response.Write([]byte(title + ": " + message))
			return
		}
		r.Response.WriteStatus(status)
		r.Response.Write(data)
	}

	if code == "" {
		writeError(http.StatusBadRequest, "授权失败", "缺少授权码 (code)，请重新登录。")
		return
	}

	oc := a.OAuthStore().Pop(state)
	if oc == nil {
		writeError(http.StatusBadRequest, "授权失败", "未找到对应的登录会话 (state 不匹配)，请从 /login 重新发起授权。")
		return
	}

	token, err := oc.GetToken(code)
	if err != nil {
		writeError(http.StatusInternalServerError, "Token 获取失败", err.Error())
		return
	}

	tokenPath := a.TokenPath()
	if tokenPath == "" {
		tokenPath = ".mi.token"
	}
	store := &miaccount.TokenStore{Path: tokenPath}
	if err := store.SaveOAuth(token); err != nil {
		writeError(http.StatusInternalServerError, "Token 保存失败", err.Error())
		return
	}

	data, err := web.RenderCallbackSuccessBytes()
	if err != nil {
		r.Response.WriteStatus(http.StatusOK)
		r.Response.Write([]byte("登录成功，token 已保存。"))
		return
	}
	r.Response.WriteStatus(http.StatusOK)
	r.Response.Write(data)
}
