// Command xiaomusic - minimal XiaoMusic-style CLI entry (OAuth 2.0).
package main

import (
	"flag"
	"log"
	"os"

	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/pkg/cmd/xiaomusic"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cfg := config.Get()
	flagMusicDir := flag.String("music_dir", cfg.Xiaomusic.MusicDir, "本地音乐目录，用于本地文件播放")
	flagAddr := flag.String("addr", cfg.Xiaomusic.Addr, "本地静态文件 HTTP 服务监听地址")
	flagHost := flag.String("host", cfg.Xiaomusic.Host, "本机 IP，供音箱访问 play-file 的 HTTP 服务")
	flag.Usage = xiaomusic.Usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		xiaomusic.Usage()
		os.Exit(1)
	}

	subcmd := args[0]
	rest := args[1:]

	opts := xiaomusic.Options{
		MusicDir: *flagMusicDir,
		Addr:     *flagAddr,
		Host:     *flagHost,
	}
	if err := xiaomusic.Run(subcmd, rest, opts); err != nil {
		log.Fatal(err)
	}
}
