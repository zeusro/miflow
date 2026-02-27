package api

import (
	"net/http"
	"net/url"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/zeusro/miflow/internal/constants"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/pkg/i18n"
	"github.com/zeusro/miflow/web"
)

// RequireAuth 若无有效 token 则返回 401。
func RequireAuth(a *web.App, r *ghttp.Request) bool {
	if a.DeviceAPI() == nil {
		lang := i18n.AcceptLanguage(r.Header.Get("Accept-Language"))
		Err(r, http.StatusUnauthorized, i18n.T(lang, "web.auth.please_login", nil))
		return false
	}
	return true
}

// Login 处理 GET /login - 发起 OAuth 流程。
func Login(a *web.App, r *ghttp.Request) {
	oc := miaccount.NewOAuthClient()
	// 按 docs/auth.md，Web 流程固定使用 HA 回调 URL，与小米 OAuth 白名单一致
	// oc.RedirectURI = constants.DefaultOAuthRedirectURI
	a.OAuthStore().Put(oc.State, oc)

	// 按 docs/auth.md 流程，使用固定 Auth URL
	params := url.Values{
		"redirect_uri":  {oc.RedirectURI},
		"client_id":     {oc.ClientID},
		"response_type": {"code"},
		"device_id":     {oc.DeviceID},
		"state":         {oc.State},
		"skip_confirm":  {"true"},
	}
	authURL := constants.OAuth2AuthURL + "?" + params.Encode()
	lang := i18n.AcceptLanguage(r.Header.Get("Accept-Language"))
	data, err := web.RenderLoginBytes(authURL, lang)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError)
		r.Response.Write([]byte(i18n.T(lang, "web.auth.template_failed", nil)))
		return
	}
	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.Response.Write(data)
}

// Callback 处理 GET /callback - OAuth 回调，用 code 换取 token 并保存。
func Callback(a *web.App, r *ghttp.Request) {
	lang := i18n.AcceptLanguage(r.Header.Get("Accept-Language"))
	code := r.Get("code").String()
	state := r.Get("state").String()
	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")

	writeError := func(status int, title, message string) {
		data, err := web.RenderErrorBytes(title, message, lang)
		if err != nil {
			r.Response.WriteStatus(status)
			r.Response.Write([]byte(title + ": " + message))
			return
		}
		r.Response.WriteStatus(status)
		r.Response.Write(data)
	}

	if code == "" {
		writeError(http.StatusBadRequest, i18n.T(lang, "web.auth.auth_failed", nil), i18n.T(lang, "web.auth.missing_code", nil))
		return
	}

	oc := a.OAuthStore().Pop(state)
	if oc == nil {
		writeError(http.StatusBadRequest, i18n.T(lang, "web.auth.auth_failed", nil), i18n.T(lang, "web.auth.state_mismatch", nil))
		return
	}

	token, err := oc.GetToken(code)
	if err != nil {
		writeError(http.StatusInternalServerError, i18n.T(lang, "web.auth.token_fetch_failed", nil), err.Error())
		return
	}

	tokenPath := a.TokenPath()
	if tokenPath == "" {
		tokenPath = ".mi.token"
	}
	store := &miaccount.TokenStore{Path: tokenPath}
	if err := store.SaveOAuth(token); err != nil {
		writeError(http.StatusInternalServerError, i18n.T(lang, "web.auth.token_save_failed", nil), err.Error())
		return
	}

	data, err := web.RenderCallbackSuccessBytes(lang)
	if err != nil {
		r.Response.WriteStatus(http.StatusOK)
		r.Response.Write([]byte(i18n.T(lang, "web.auth.login_success", nil)))
		return
	}
	r.Response.WriteStatus(http.StatusOK)
	r.Response.Write(data)
}
