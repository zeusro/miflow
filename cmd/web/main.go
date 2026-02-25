// Command web runs the miflow web server with GoFrame.
package main

import (
	"context"
	"io/fs"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/web"
	"github.com/zeusro/miflow/web/api"
)

// pendingOAuth stores OAuthClient by state for callback; device_id must match login.
var (
	pendingOAuthMu sync.Mutex
	pendingOAuth   = make(map[string]*pendingOAuthEntry)
)

type pendingOAuthEntry struct {
	oc        *miaccount.OAuthClient
	createdAt time.Time
}

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
				r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
				_ = web.RenderDefault(r.Response.Writer)
				return
			}
			r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			r.Response.Write(data)
		})
		group.GET("/app", func(r *ghttp.Request) {
			data, err := fs.ReadFile(staticRoot, "app.html")
			if err != nil {
				r.Response.WriteStatus(http.StatusNotFound)
				return
			}
			r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			r.Response.Write(data)
		})
		group.GET("/login", handleLogin)
		group.GET("/callback", handleCallback)
	})

	// API: devices (DDD - device domain)
	s.Group("/api/devices", func(group *ghttp.RouterGroup) {
		group.GET("/", func(r *ghttp.Request) { api.DevicesList(a, r) })
		group.GET("/{id}", func(r *ghttp.Request) { api.DeviceGet(a, r) })
		group.GET("/{id}/spec", func(r *ghttp.Request) { api.DeviceSpec(a, r) })
		group.POST("/{id}/control", func(r *ghttp.Request) { api.DeviceControl(a, r) })
	})

	// API: workflows (DDD - workflow domain)
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

func handleLogin(r *ghttp.Request) {
	oc := miaccount.NewOAuthClient()
	pendingOAuthMu.Lock()
	// Clean up entries older than 10 minutes
	for state, e := range pendingOAuth {
		if time.Since(e.createdAt) > 10*time.Minute {
			delete(pendingOAuth, state)
		}
	}
	pendingOAuth[oc.State] = &pendingOAuthEntry{oc: oc, createdAt: time.Now()}
	pendingOAuthMu.Unlock()

	authURL := oc.GenAuthURL("", "", true)
	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = web.RenderLogin(r.Response.Writer, authURL)
}

func handleCallback(r *ghttp.Request) {
	code := r.Get("code").String()
	state := r.Get("state").String()
	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	if code == "" {
		r.Response.WriteStatus(http.StatusBadRequest)
		_ = web.RenderError(r.Response.Writer, "授权失败", "缺少授权码 (code)，请重新登录。")
		return
	}

	var oc *miaccount.OAuthClient
	pendingOAuthMu.Lock()
	if e, ok := pendingOAuth[state]; ok {
		oc = e.oc
		delete(pendingOAuth, state)
	}
	pendingOAuthMu.Unlock()

	if oc == nil {
		r.Response.WriteStatus(http.StatusBadRequest)
		_ = web.RenderError(r.Response.Writer, "授权失败", "未找到对应的登录会话 (state 不匹配)，请从 /login 重新发起授权。")
		return
	}

	token, err := oc.GetToken(code)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError)
		_ = web.RenderError(r.Response.Writer, "Token 获取失败", err.Error())
		return
	}

	cfg := config.Get()
	tokenPath := cfg.TokenPath
	if tokenPath == "" {
		tokenPath = ".mi.token"
	}
	store := &miaccount.TokenStore{Path: tokenPath}
	if err := store.SaveOAuth(token); err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError)
		_ = web.RenderError(r.Response.Writer, "Token 保存失败", err.Error())
		return
	}

	_ = web.RenderCallbackSuccess(r.Response.Writer)
}
