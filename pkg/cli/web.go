package cli

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/spf13/cobra"
	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/constants"
	"github.com/zeusro/miflow/web"
	"github.com/zeusro/miflow/web/api"
)

// NewWebCmd 创建 web 子命令。
func NewWebCmd() *cobra.Command {
	var addr string

	webCmd := &cobra.Command{
		Use:   "web",
		Short: "Run miflow web server (OAuth login + device management)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebServer(addr)
		},
	}

	cfg := config.Get()
	defaultAddr := cfg.Web.Addr
	if defaultAddr == "" {
		defaultAddr = constants.DefaultWebAddr
	}
	webCmd.Flags().StringVar(&addr, "addr", defaultAddr, "HTTP listen address")

	return webCmd
}

func runWebServer(addr string) error {
	a, err := web.NewApp()
	if err != nil {
		g.Log().Fatalf(context.Background(), "init app: %v", err)
	}

	s := g.Server()
	s.SetAddr(addr)
	api.RegisterRoutes(s, a)

	g.Log().Infof(context.Background(), "miflow web server listening on %s", addr)
	s.Run()
	return nil
}
