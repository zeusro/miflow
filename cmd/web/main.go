// Command web 使用 GoFrame 运行 miflow 网页服务器。
package main

import (
	"context"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/web"
	"github.com/zeusro/miflow/web/api"
)

func main() {
	a, err := web.NewApp()
	if err != nil {
		g.Log().Fatalf(context.Background(), "init app: %v", err)
	}

	s := g.Server()
	addr := config.Get().Web.Addr
	if addr == "" {
		addr = ":8123"
	}
	s.SetAddr(addr)

	staticRoot, _ := fs.Sub(web.StaticFS, "static")
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.GET("/", func(r *ghttp.Request) {
			data, err := fs.ReadFile(staticRoot, "index.html")
			if err != nil {
				data, _ = web.RenderDefaultBytes()
			}
			if len(data) > 0 {
				r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
				r.Response.Write(data)
			}
		})
		group.GET("/app", func(r *ghttp.Request) {
			data, err := fs.ReadFile(staticRoot, "app.html")
			if err != nil {
				r.Response.WriteStatus(http.StatusNotFound)
				r.Response.Write([]byte("app.html not found"))
				return
			}
			r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			r.Response.Write(data)
		})
		group.GET("/login", func(r *ghttp.Request) { api.Login(a, r) })
		group.GET("/callback", func(r *ghttp.Request) { api.Callback(a, r) })
	})

	// API: 设备 (DDD - 设备领域)
	s.Group("/api/devices", func(group *ghttp.RouterGroup) {
		group.GET("/", func(r *ghttp.Request) { api.DevicesList(a, r) })
		group.GET("/{id}", func(r *ghttp.Request) { api.DeviceGet(a, r) })
		group.GET("/{id}/spec", func(r *ghttp.Request) { api.DeviceSpec(a, r) })
		group.POST("/{id}/control", func(r *ghttp.Request) { api.DeviceControl(a, r) })
	})

	// API: 工作流 (DDD - 工作流领域)
	s.Group("/api/workflows", func(group *ghttp.RouterGroup) {
		group.GET("/", func(r *ghttp.Request) { api.WorkflowsList(a, r) })
		group.GET("/{id}", func(r *ghttp.Request) { api.WorkflowGet(a, r) })
		group.POST("/", func(r *ghttp.Request) { api.WorkflowCreate(a, r) })
		group.PUT("/{id}", func(r *ghttp.Request) { api.WorkflowUpdate(a, r) })
		group.DELETE("/{id}", func(r *ghttp.Request) { api.WorkflowDelete(a, r) })
		group.POST("/{id}/run", func(r *ghttp.Request) { api.WorkflowRun(a, r) })
	})

	s.Group("/dist", func(group *ghttp.RouterGroup) {
		group.ALL("/*", func(r *ghttp.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/dist/")
			if path == "" {
				path = "output.css"
			}
			data, err := fs.ReadFile(staticRoot, "dist/"+path)
			if err != nil {
				r.Response.WriteStatus(http.StatusNotFound)
				return
			}
			r.Response.Header().Set("Content-Type", "text/css; charset=utf-8")
			r.Response.Write(data)
		})
	})

	g.Log().Infof(context.Background(), "miflow web server listening on %s", addr)
	s.Run()
}
