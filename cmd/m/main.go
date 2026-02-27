// Command m - 小米云服务命令行工具 (OAuth 2.0, 参考: ha_xiaomi_home)。
package main

import (
	"os"

	"github.com/zeusro/miflow/pkg/cmd"
)

func main() {
	cmd.Run(os.Args[1:])
}
