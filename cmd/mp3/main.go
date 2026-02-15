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
)

func usage() {
	fmt.Fprintf(os.Stderr, "mp3 - 将本地音乐文件映射为 HTTP 可访问链接\n\n")
	fmt.Fprintf(os.Stderr, "用法：\n")
	fmt.Fprintf(os.Stderr, "  mp3 [选项] <文件路径>\n")
	fmt.Fprintf(os.Stderr, "  mp3 -addr=:8090 /Users/zeusro/Music/QQ音乐/Taylor Swift-Red.flac\n\n")
	fmt.Fprintf(os.Stderr, "选项：\n")
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cfg := config.Get()
	flagAddr := flag.String("addr", cfg.Xiaomusic.Addr, "HTTP 服务监听地址")
	flagHost := flag.String("host", cfg.Xiaomusic.Host, "本机 IP，供局域网访问，空则自动检测")
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
	log.Printf("[映射] 本地: %s -> URL: %s", absPath, playURL)
	fmt.Println(playURL)
	if srv.Host() == "127.0.0.1" {
		fmt.Fprintln(os.Stderr, "提示：未检测到局域网 IP，请用 -host=本机IP 指定，如 -host=192.168.1.100")
	}

	if !srv.WaitReady(5 * time.Second) {
		log.Fatalf("HTTP 服务未能就绪，端口 %s 未监听", srv.Port())
	}
	log.Printf("HTTP 服务就绪，端口 %s，按 Ctrl+C 退出", srv.Port())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
