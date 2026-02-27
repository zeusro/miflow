// Command miflow - 统一入口，基于 Cobra 的命令行架构。
package main

import (
	"os"
	"path/filepath"

	"github.com/zeusro/miflow/pkg/cli"
)

func main() {
	args := os.Args[1:]
	// 若以 m 为名运行（go build -o m ./cmd/miflow），则自动补全 m 子命令
	if filepath.Base(os.Args[0]) == "m" {
		args = append([]string{"m"}, args...)
	}
	cli.Execute(args)
}
