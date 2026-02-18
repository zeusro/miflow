// Command web runs the miflow web server with GoFrame.
package main

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/web"
)

func main() {
	a, err := newApp()
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
				r.Response.WriteStatus(http.StatusNotFound)
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

	// API: devices
	s.Group("/api/devices", func(group *ghttp.RouterGroup) {
		group.GET("/", a.handleDevicesList)
		group.GET("/{id}", a.handleDeviceGet)
		group.GET("/{id}/spec", a.handleDeviceSpec)
		group.POST("/{id}/control", a.handleDeviceControl)
	})

	// API: workflows (SQLite)
	s.Group("/api/workflows", func(group *ghttp.RouterGroup) {
		group.GET("/", a.handleWorkflowsList)
		group.GET("/{id}", a.handleWorkflowGet)
		group.POST("/", a.handleWorkflowCreate)
		group.PUT("/{id}", a.handleWorkflowUpdate)
		group.DELETE("/{id}", a.handleWorkflowDelete)
		group.POST("/{id}/run", a.handleWorkflowRun)
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
	// Use config's redirect_uri (must be whitelisted in Xiaomi developer console)
	authURL := oc.GenAuthURL("", "", true)

	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>登录 - miflow</title>
  <link href="/dist/output.css" rel="stylesheet">
</head>
<body class="min-h-screen bg-slate-50 flex flex-col items-center justify-center p-6">
  <main class="max-w-md w-full">
    <div class="rounded-xl bg-white shadow-sm border border-slate-200 p-6 text-center">
      <p class="text-slate-600">正在跳转到小米账号授权页面...</p>
      <p class="mt-2 text-sm text-slate-500">如未自动跳转，请点击下方按钮</p>
      <a id="auth-link" href="` + authURL + `" class="mt-4 inline-block rounded-lg bg-emerald-600 px-4 py-2 text-white hover:bg-emerald-700 transition-colors">
        前往授权
      </a>
    </div>
  </main>
  <script>
    document.getElementById('auth-link').href = "` + authURL + `";
    window.location.href = "` + authURL + `";
  </script>
</body>
</html>`
	r.Response.Write(html)
}

func handleCallback(r *ghttp.Request) {
	code := r.Get("code").String()
	if code == "" {
		r.Response.WriteStatus(http.StatusBadRequest)
		r.Response.Write(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>错误</title>
<link href="/dist/output.css" rel="stylesheet"></head>
<body class="min-h-screen bg-slate-50 flex items-center justify-center p-6">
<div class="rounded-xl bg-white shadow-sm border border-red-200 p-6 max-w-md">
<h2 class="text-xl font-semibold text-red-600">授权失败</h2>
<p class="mt-2 text-slate-600">缺少授权码 (code)，请重新登录。</p>
<a href="/" class="mt-4 inline-block text-emerald-600 hover:underline">返回首页</a>
</div></body></html>`)
		return
	}

	oc := miaccount.NewOAuthClient()
	// redirect_uri must match config (whitelisted in Xiaomi)
	token, err := oc.GetToken(code)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError)
		r.Response.Write(fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>错误</title>
<link href="/dist/output.css" rel="stylesheet"></head>
<body class="min-h-screen bg-slate-50 flex items-center justify-center p-6">
<div class="rounded-xl bg-white shadow-sm border border-red-200 p-6 max-w-md">
<h2 class="text-xl font-semibold text-red-600">Token 获取失败</h2>
<p class="mt-2 text-slate-600">%s</p>
<a href="/" class="mt-4 inline-block text-emerald-600 hover:underline">返回首页</a>
</div></body></html>`, err.Error()))
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
		r.Response.Write(fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>错误</title>
<link href="/dist/output.css" rel="stylesheet"></head>
<body class="min-h-screen bg-slate-50 flex items-center justify-center p-6">
<div class="rounded-xl bg-white shadow-sm border border-red-200 p-6 max-w-md">
<h2 class="text-xl font-semibold text-red-600">Token 保存失败</h2>
<p class="mt-2 text-slate-600">%s</p>
<a href="/" class="mt-4 inline-block text-emerald-600 hover:underline">返回首页</a>
</div></body></html>`, err.Error()))
		return
	}

	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.Response.Write(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <title>登录成功</title>
  <link href="/dist/output.css" rel="stylesheet">
</head>
<body class="min-h-screen bg-slate-50 flex flex-col items-center justify-center p-6">
  <div class="rounded-xl bg-white shadow-sm border border-slate-200 p-8 max-w-md text-center">
    <div class="w-12 h-12 rounded-full bg-emerald-100 flex items-center justify-center mx-auto">
      <svg class="w-6 h-6 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
      </svg>
    </div>
    <h2 class="mt-4 text-xl font-semibold text-slate-800">登录成功</h2>
    <p class="mt-2 text-slate-600">米家 OAuth 授权已完成，token 已保存。</p>
    <p id="countdown" class="mt-4 text-sm text-slate-500">5 秒后自动关闭此页面...</p>
    <a href="/" class="mt-4 inline-block text-emerald-600 hover:underline">返回首页</a>
  </div>
  <script>
    (function(){
      var n=5;
      var el=document.getElementById('countdown');
      var t=setInterval(function(){
        n--;
        if(n>0) el.textContent=n+' 秒后自动关闭此页面...';
        else {
          clearInterval(t);
          el.textContent='正在关闭...';
          try{window.close()}catch(e){}
        }
      }, 1000);
    })();
  </script>
</body>
</html>`)
}
