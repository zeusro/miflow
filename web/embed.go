// Package web provides embedded static assets for the miflow web server.
package web

import (
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

// RenderError writes the error page with title and message to w.
func RenderError(w io.Writer, title, message string) error {
	return Templates.ExecuteTemplate(w, "error.html", map[string]string{"Title": title, "Message": message})
}

// RenderCallbackSuccess writes the callback success page to w.
func RenderCallbackSuccess(w io.Writer) error {
	return Templates.ExecuteTemplate(w, "callback-success.html", nil)
}
