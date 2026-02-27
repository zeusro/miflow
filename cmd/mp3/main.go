// Command mp3 - HTTP 文件服务，将本地路径映射为可访问的 URL。
// 映射规则：/Users/zeusro/Music/QQ音乐/Taylor Swift-Red.flac -> http://本机ip:端口/Users/zeusro/Music/QQ音乐/Taylor%20Swift-Red.flac
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/mp3server"
	"github.com/zeusro/miflow/pkg/i18n"
)

// usage 打印命令行用法。
func usage() {
	lang := i18n.DefaultLang()
	fmt.Fprint(os.Stderr, i18n.T(lang, "mp3.usage", nil))
	flag.PrintDefaults()
}

func main() {
	lang := i18n.DefaultLang()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cfg := config.Get()
	flagAddr := flag.String("addr", cfg.Mp3.Addr, i18n.T(lang, "mp3.addr_desc", nil))
	flagHost := flag.String("host", cfg.Mp3.Host, i18n.T(lang, "mp3.host_desc", nil))
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		usage()
		os.Exit(1)
	}
	filePath := args[0]

	srv, err := mp3server.New(mp3server.Config{
		Addr:       *flagAddr,
		Host:       *flagHost,
		LogRequest: true,
	}, "/") // 根目录，实现完整路径映射
	if err != nil {
		log.Fatal(err)
	}
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	defer srv.Close()

	playURL, err := srv.PathToURL(filePath)
	if err != nil {
		log.Fatal(err)
	}
	absPath, _ := filepath.Abs(filePath)
	log.Println(i18n.T(lang, "mp3.mapped", map[string]interface{}{"Local": absPath, "URL": playURL}))
	fmt.Println(playURL)
	if srv.Host() == "127.0.0.1" {
		fmt.Fprintln(os.Stderr, i18n.T(lang, "mp3.host_hint", nil))
	}

	if !srv.WaitReady(5 * time.Second) {
		log.Fatal(i18n.T(lang, "mp3.port_not_ready", map[string]interface{}{"Port": srv.Port()}))
	}
	log.Println(i18n.T(lang, "mp3.ready", map[string]interface{}{"Port": srv.Port()}))

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
