// Command web 使用 GoFrame 运行 miflow 网页服务器。
package main

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/constants"
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
		addr = constants.DefaultWebAddr
	}
	s.SetAddr(addr)

	api.RegisterRoutes(s, a)

	g.Log().Infof(context.Background(), "miflow web server listening on %s", addr)
	s.Run()
}
