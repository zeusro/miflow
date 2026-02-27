// Package web provides embedded static assets for the miflow web server.
package web

import (
	"bytes"
	"embed"
	"html/template"
	"io"
)

//go:embed static templates
var StaticFS embed.FS

// Templates are parsed Go templates for server-rendered pages.
var Templates *template.Template

func init() {
	var err error
	Templates, err = template.New("").ParseFS(StaticFS, "templates/*.html")
	if err != nil {
		panic("parse templates: " + err.Error())
	}
}

// RenderLogin writes the login page with authURL to w.
func RenderLogin(w io.Writer, authURL string) error {
	return Templates.ExecuteTemplate(w, "login.html", map[string]string{"AuthURL": authURL})
}

// RenderLoginBytes returns the login page HTML as bytes.
func RenderLoginBytes(authURL string) ([]byte, error) {
	var buf bytes.Buffer
	if err := RenderLogin(&buf, authURL); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderError writes the error page with title and message to w.
func RenderError(w io.Writer, title, message string) error {
	return Templates.ExecuteTemplate(w, "error.html", map[string]string{"Title": title, "Message": message})
}

// RenderErrorBytes returns the error page HTML as bytes.
func RenderErrorBytes(title, message string) ([]byte, error) {
	var buf bytes.Buffer
	if err := RenderError(&buf, title, message); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderCallbackSuccess writes the callback success page to w.
func RenderCallbackSuccess(w io.Writer) error {
	return Templates.ExecuteTemplate(w, "callback-success.html", nil)
}

// RenderCallbackSuccessBytes returns the callback success page HTML as bytes.
func RenderCallbackSuccessBytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := RenderCallbackSuccess(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderDefault writes the default fallback index page to w.
func RenderDefault(w io.Writer) error {
	return Templates.ExecuteTemplate(w, "default.html", nil)
}

// RenderDefaultBytes returns the default fallback index page HTML as bytes.
func RenderDefaultBytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := RenderDefault(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
