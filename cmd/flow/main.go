// Command flow - Flow 可视化控制流 HTTP 服务。
package main

import (
	"flag"
	"log"

	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/flowserver"
	"github.com/zeusro/miflow/pkg/i18n"
)

func main() {
	lang := i18n.DefaultLang()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg := config.Get()
	addr := flag.String("addr", cfg.Flow.Addr, i18n.T(lang, "flow.addr_desc", nil))
	dataDir := flag.String("data_dir", cfg.Flow.DataDir, i18n.T(lang, "flow.datadir_desc", nil))
	flag.Parse()

	if err := flowserver.Run(*addr, *dataDir); err != nil {
		log.Fatal(err)
	}
}
