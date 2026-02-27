// Package web 为 miflow 网页服务器提供内嵌的静态资源。
package web

import (
	"bytes"
	"embed"
	"html/template"
	"io"
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

// RenderLogin 将带 authURL 的登录页写入 w。
func RenderLogin(w io.Writer, authURL string) error {
	return Templates.ExecuteTemplate(w, "login.html", map[string]string{"AuthURL": authURL})
}

// RenderLoginBytes 返回登录页 HTML 的字节形式。
func RenderLoginBytes(authURL string) ([]byte, error) {
	var buf bytes.Buffer
	if err := RenderLogin(&buf, authURL); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderError 将带标题和消息的错误页写入 w。
func RenderError(w io.Writer, title, message string) error {
	return Templates.ExecuteTemplate(w, "error.html", map[string]string{"Title": title, "Message": message})
}

// RenderErrorBytes 返回错误页 HTML 的字节形式。
func RenderErrorBytes(title, message string) ([]byte, error) {
	var buf bytes.Buffer
	if err := RenderError(&buf, title, message); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderCallbackSuccess 将回调成功页写入 w。
func RenderCallbackSuccess(w io.Writer) error {
	return Templates.ExecuteTemplate(w, "callback-success.html", nil)
}

// RenderCallbackSuccessBytes 返回回调成功页 HTML 的字节形式。
func RenderCallbackSuccessBytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := RenderCallbackSuccess(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderDefault 将默认备用首页写入 w。
func RenderDefault(w io.Writer) error {
	return Templates.ExecuteTemplate(w, "default.html", nil)
}

// RenderDefaultBytes 返回默认备用首页 HTML 的字节形式。
func RenderDefaultBytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := RenderDefault(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderIndex 将带 authURL 的首页写入 w（参见 docs/auth.md）。
func RenderIndex(w io.Writer, authURL string) error {
	return Templates.ExecuteTemplate(w, "index.html", map[string]string{"AuthURL": authURL})
}

// RenderIndexBytes 返回首页 HTML 的字节形式。
func RenderIndexBytes(authURL string) ([]byte, error) {
	var buf bytes.Buffer
	if err := RenderIndex(&buf, authURL); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
