// Command m - XiaoMi Cloud Service CLI (OAuth 2.0, ref: ha_xiaomi_home).
package main

import (
	"os"

	"github.com/zeusro/miflow/pkg/cmd"
)

func main() {
	cmd.Run(os.Args[1:])
}
