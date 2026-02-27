// Package web 为 miflow 网页服务器提供内嵌的静态资源。
package web

import (
	"bytes"
	"embed"
	"html/template"
	"io"

	"github.com/zeusro/miflow/pkg/i18n"
)

//go:embed static templates
var StaticFS embed.FS

// Templates 是用于服务端渲染页面的已解析 Go 模板。
var Templates *template.Template

func init() {
	var err error
	Templates, err = template.New("").ParseFS(StaticFS, "templates/*.html")
	if err != nil {
		panic("parse templates: " + err.Error())
	}
}

// RenderLogin 将带 authURL 的登录页写入 w。lang 为空时使用默认语言。
func RenderLogin(w io.Writer, authURL, lang string) error {
	if lang == "" {
		lang = i18n.DefaultLang()
	}
	data := map[string]string{
		"AuthURL":     authURL,
		"Title":       i18n.T(lang, "web.login.title", nil),
		"Redirecting": i18n.T(lang, "web.login.redirecting", nil),
		"ClickIfNot":  i18n.T(lang, "web.login.click_if_not", nil),
		"GoAuth":      i18n.T(lang, "web.login.go_auth", nil),
	}
	return Templates.ExecuteTemplate(w, "login.html", data)
}

// RenderLoginBytes 返回登录页 HTML 的字节形式。
func RenderLoginBytes(authURL, lang string) ([]byte, error) {
	var buf bytes.Buffer
	if err := RenderLogin(&buf, authURL, lang); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderError 将带标题和消息的错误页写入 w。
func RenderError(w io.Writer, title, message, lang string) error {
	if lang == "" {
		lang = i18n.DefaultLang()
	}
	data := map[string]string{
		"Title":    title,
		"Message":  message,
		"BackHome": i18n.T(lang, "web.error.back_home", nil),
	}
	return Templates.ExecuteTemplate(w, "error.html", data)
}

// RenderErrorBytes 返回错误页 HTML 的字节形式。
func RenderErrorBytes(title, message, lang string) ([]byte, error) {
	var buf bytes.Buffer
	if err := RenderError(&buf, title, message, lang); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderCallbackSuccess 将回调成功页写入 w。
func RenderCallbackSuccess(w io.Writer, lang string) error {
	if lang == "" {
		lang = i18n.DefaultLang()
	}
	data := map[string]string{
		"Title":           i18n.T(lang, "web.callback.title", nil),
		"Success":         i18n.T(lang, "web.callback.success", nil),
		"Done":            i18n.T(lang, "web.callback.done", nil),
		"Countdown":      i18n.T(lang, "web.callback.countdown", map[string]interface{}{"N": 5}),
		"CountdownSuffix": i18n.T(lang, "web.callback.countdown_suffix", nil),
		"Closing":         i18n.T(lang, "web.callback.closing", nil),
		"BackHome":       i18n.T(lang, "web.error.back_home", nil),
	}
	return Templates.ExecuteTemplate(w, "callback-success.html", data)
}

// RenderCallbackSuccessBytes 返回回调成功页 HTML 的字节形式。
func RenderCallbackSuccessBytes(lang string) ([]byte, error) {
	var buf bytes.Buffer
	if err := RenderCallbackSuccess(&buf, lang); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderCallbackCLI 将 CLI OAuth 回调成功页（无外部 CSS）写入 w。
func RenderCallbackCLI(w io.Writer, lang string) error {
	if lang == "" {
		lang = i18n.DefaultLang()
	}
	data := map[string]string{
		"Title":           i18n.T(lang, "web.callback.title", nil),
		"Success":         i18n.T(lang, "web.callback.success", nil),
		"Done":            i18n.T(lang, "web.callback.done", nil),
		"Countdown":       i18n.T(lang, "web.callback.countdown", map[string]interface{}{"N": 5}),
		"CountdownSuffix": i18n.T(lang, "web.callback.countdown_suffix", nil),
		"Closing":         i18n.T(lang, "web.callback.closing", nil),
	}
	return Templates.ExecuteTemplate(w, "callback-cli.html", data)
}

// RenderCallbackCLIBytes 返回 CLI 回调成功页 HTML 的字节形式。
func RenderCallbackCLIBytes(lang string) ([]byte, error) {
	var buf bytes.Buffer
	if err := RenderCallbackCLI(&buf, lang); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderDefault 将默认备用首页写入 w。
func RenderDefault(w io.Writer, lang string) error {
	if lang == "" {
		lang = i18n.DefaultLang()
	}
	data := map[string]string{
		"Title":       i18n.T(lang, "web.default.title", nil),
		"Subtitle":   i18n.T(lang, "web.default.subtitle", nil),
		"DevicesLink": i18n.T(lang, "web.default.devices_link", nil),
		"LoginLink":   i18n.T(lang, "web.default.login_link", nil),
	}
	return Templates.ExecuteTemplate(w, "default.html", data)
}

// RenderDefaultBytes 返回默认备用首页 HTML 的字节形式。
func RenderDefaultBytes(lang string) ([]byte, error) {
	var buf bytes.Buffer
	if err := RenderDefault(&buf, lang); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderIndex 将首页写入 w，支持 i18n。
func RenderIndex(w io.Writer, lang string) error {
	if lang == "" {
		lang = i18n.DefaultLang()
	}
	data := map[string]string{
		"Title":       i18n.T(lang, "web.default.title", nil),
		"Subtitle":   i18n.T(lang, "web.default.subtitle", nil),
		"DevicesLink": i18n.T(lang, "web.default.devices_link", nil),
		"LoginLink":   i18n.T(lang, "web.default.login_link", nil),
		"TokenHint":  i18n.T(lang, "web.index.token_hint", nil),
	}
	return Templates.ExecuteTemplate(w, "index.html", data)
}

// RenderIndexBytes 返回首页 HTML 的字节形式。
func RenderIndexBytes(lang string) ([]byte, error) {
	var buf bytes.Buffer
	if err := RenderIndex(&buf, lang); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
