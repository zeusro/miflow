// Package cli 提供基于 Cobra 的 miflow 命令行架构。
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const prefix = "m "

// NewRootCmd 创建 miflow 根命令。
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "miflow",
		Short: "MiFlow - XiaoMi MIoT + Mina CLI & Services",
		Long: `MiFlow 提供小米设备控制与工作流编排：

  m            小米云服务 CLI (MiIO/MIoT/Mina)
  flow         Flow 可视化控制流服务
  web          Web 服务 (OAuth 登录 + 设备管理)
  miiot        MiIoT 规格校验
  mp3          本地音乐 HTTP 映射
  scrape-specs 爬取设备规格 URL 表格`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return fmt.Errorf("unknown command: %s (use 'miflow --help')", args[0])
		},
	}

	root.AddCommand(NewMCmd())
	root.AddCommand(NewFlowCmd())
	root.AddCommand(NewWebCmd())
	root.AddCommand(NewMiiotCmd())
	root.AddCommand(NewMp3Cmd())
	root.AddCommand(NewScrapeSpecsCmd())

	return root
}

// Execute 执行根命令，args 为空时使用 os.Args[1:]。
func Execute(args []string) {
	root := NewRootCmd()
	if len(args) > 0 {
		root.SetArgs(args)
	}
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
