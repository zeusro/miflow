// Command m - 小米云服务命令行工具 (OAuth 2.0)。
// 构建为 m 时作为 m 的入口；也可通过 miflow m 调用。
package main

import (
	"os"

	"github.com/zeusro/miflow/pkg/cli"
)

func main() {
	cli.Execute(append([]string{"m"}, os.Args[1:]...))
}
