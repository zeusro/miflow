// Package api - 路由注册，供 cmd/web 与 pkg/cli 共用。
package api

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/zeusro/miflow/pkg/i18n"
	"github.com/zeusro/miflow/web"
)

// RegisterRoutes 向 s 注册所有 Web 路由。
func RegisterRoutes(s *ghttp.Server, a *web.App) {
	staticRoot, _ := fs.Sub(web.StaticFS, "static")

	s.Group("/", func(group *ghttp.RouterGroup) {
		group.GET("/", func(r *ghttp.Request) {
			lang := i18n.AcceptLanguage(r.Header.Get("Accept-Language"))
			data, err := web.RenderIndexBytes(lang)
			if err != nil {
				data, _ = web.RenderDefaultBytes(lang)
			}
			if len(data) > 0 {
				r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
				r.Response.Write(data)
			}
		})
		group.GET("/app", func(r *ghttp.Request) {
			data, err := fs.ReadFile(staticRoot, "app.html")
			if err != nil {
				r.Response.WriteStatus(http.StatusNotFound, []byte("app.html not found"))
				return
			}
			r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			r.Response.Write(data)
		})
		group.GET("/rooms", func(r *ghttp.Request) {
			data, err := fs.ReadFile(staticRoot, "rooms.html")
			if err != nil {
				r.Response.WriteHeader(http.StatusNotFound)
				r.Response.Write([]byte("rooms.html not found"))
				return
			}
			r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			r.Response.Write(data)
		})
		group.GET("/login", func(r *ghttp.Request) { Login(a, r) })
		group.GET("/callback", func(r *ghttp.Request) { Callback(a, r) })
	})

	s.Group("/api/rooms", func(group *ghttp.RouterGroup) {
		group.GET("/", func(r *ghttp.Request) { RoomsList(a, r) })
	})
	s.Group("/api/devices", func(group *ghttp.RouterGroup) {
		group.GET("/", func(r *ghttp.Request) { DevicesList(a, r) })
		group.GET("/{id}", func(r *ghttp.Request) { DeviceGet(a, r) })
		group.GET("/{id}/spec", func(r *ghttp.Request) { DeviceSpec(a, r) })
		group.POST("/{id}/control", func(r *ghttp.Request) { DeviceControl(a, r) })
	})
	s.Group("/api/workflows", func(group *ghttp.RouterGroup) {
		group.GET("/", func(r *ghttp.Request) { WorkflowsList(a, r) })
		group.GET("/{id}", func(r *ghttp.Request) { WorkflowGet(a, r) })
		group.POST("/", func(r *ghttp.Request) { WorkflowCreate(a, r) })
		group.PUT("/{id}", func(r *ghttp.Request) { WorkflowUpdate(a, r) })
		group.DELETE("/{id}", func(r *ghttp.Request) { WorkflowDelete(a, r) })
		group.POST("/{id}/run", func(r *ghttp.Request) { WorkflowRun(a, r) })
	})

	s.Group("/dist", func(group *ghttp.RouterGroup) {
		group.ALL("/*", func(r *ghttp.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/dist/")
			if path == "" {
				path = "output.css"
			}
			data, err := fs.ReadFile(staticRoot, "dist/"+path)
			if err != nil {
				r.Response.WriteHeader(http.StatusNotFound)
				return
			}
			r.Response.Header().Set("Content-Type", "text/css; charset=utf-8")
			r.Response.Write(data)
		})
	})
}
