// Command xiaomusic - minimal XiaoMusic-style CLI entry (OAuth 2.0).
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/miioservice"
	"github.com/zeusro/miflow/internal/minaservice"
	"github.com/zeusro/miflow/internal/mp3server"
)

func usage() {
	fmt.Fprintf(os.Stderr, "xiaomusic - 使用小爱音箱播放本地或网络音乐（OAuth 模式）\n\n")
	fmt.Fprintf(os.Stderr, "请先运行: m login\n")
	fmt.Fprintf(os.Stderr, "环境变量: MI_DID=<设备ID或名称>\n\n")
	fmt.Fprintf(os.Stderr, "用法示例：\n")
	fmt.Fprintf(os.Stderr, "  xiaomusic -music_dir=./music play-url https://example.com/a.mp3\n")
	fmt.Fprintf(os.Stderr, "  xiaomusic -music_dir=./music play-file song.mp3\n")
	fmt.Fprintf(os.Stderr, "  xiaomusic -host=192.168.1.100 play-file /path/to/music.mp3  # 指定本机 IP\n")
	fmt.Fprintf(os.Stderr, "\n注：基于 MiNA API (api2.mina.mi.com)，参考 https://github.com/hanxi/xiaomusic\n")
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cfg := config.Get()
	flagMusicDir := flag.String("music_dir", cfg.Xiaomusic.MusicDir, "本地音乐目录，用于本地文件播放")
	flagAddr := flag.String("addr", cfg.Xiaomusic.Addr, "本地静态文件 HTTP 服务监听地址")
	flagHost := flag.String("host", cfg.Xiaomusic.Host, "本机 IP，供音箱访问 play-file 的 HTTP 服务")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	cmd := args[0]
	rest := args[1:]

	did := cfg.DefaultDID
	if did == "" {
		fmt.Fprintln(os.Stderr, "错误：必须设置 default_did（配置文件）或环境变量 MI_DID")
		usage()
		os.Exit(1)
	}

	tokenPath := cfg.TokenPath
	token := (&miaccount.TokenStore{Path: tokenPath}).LoadOAuth()
	if token == nil || !token.IsValid() {
		fmt.Fprintln(os.Stderr, "错误：未登录，请先运行 m login")
		os.Exit(1)
	}

	ioSvc, err := miioservice.New(token, tokenPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	mina := minaservice.NewWithMinaAPI(ioSvc, token, tokenPath)

	switch cmd {
	case "play-url":
		if len(rest) < 1 {
			fmt.Fprintln(os.Stderr, "用法：xiaomusic play-url <mp3_url>")
			os.Exit(1)
		}
		if err := playURL(mina, did, rest[0]); err != nil {
			log.Fatal(err)
		}
	case "play-file":
		if len(rest) < 1 {
			fmt.Fprintln(os.Stderr, "用法：xiaomusic play-file <相对或绝对文件路径>")
			os.Exit(1)
		}
		if err := playFile(mina, did, *flagMusicDir, rest[0], *flagAddr, *flagHost); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Fprintf(os.Stderr, "未知子命令：%s\n\n", cmd)
		usage()
		os.Exit(1)
	}
}

// playURL 直接通过 MiNA 播放远程 URL。
func playURL(mina *minaservice.Service, miDID, url string) error {
	deviceID, err := mina.GetMinaDeviceID(miDID)
	if err != nil {
		return err
	}
	fmt.Println("播放 URL：", url)
	_, err = mina.PlayByURL(deviceID, strings.TrimSpace(url), 2)
	return err
}

// playFile 在本机起一个简单静态 HTTP 服务，将本地文件映射为 URL 再通过 MiNA 播放。
// 映射规则：/Users/xxx/Music/QQ音乐/Taylor Swift-Red.flac -> http://本机ip:端口/Users/xxx/Music/QQ音乐/Taylor%20Swift-Red.flac
// 支持绝对路径或相对于 musicDir 的路径。
func playFile(mina *minaservice.Service, miDID, musicDir, filePath, addr, host string) error {
	deviceID, err := mina.GetMinaDeviceID(miDID)
	if err != nil {
		return err
	}

	root, err := filepath.Abs(musicDir)
	if err != nil {
		return err
	}
	target, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	if !filepath.IsAbs(filePath) {
		target = filepath.Join(root, filePath)
	}

	// mp3 可单独启动（如 ./mp3 -addr=:8090 /Users/zeusro/Music），xiaomusic 仅生成 URL 并通知音箱
	srv, err := mp3server.New(mp3server.Config{Addr: addr, Host: host, LogRequest: false}, "/")
	if err != nil {
		return err
	}
	playURL, err := srv.PathToURL(target)
	if err != nil {
		return err
	}
	log.Printf("[映射] 本地: %s -> URL: %s", target, playURL)
	fmt.Println("本地文件映射为 URL：", playURL)
	if srv.Host() == "127.0.0.1" {
		fmt.Fprintln(os.Stderr, "提示：未检测到局域网 IP，音箱可能无法访问。请用 -host=本机IP 指定，如 -host=192.168.1.100")
	}
	// 确认 mp3 服务已就绪
	if !srv.WaitPortReady(5 * time.Second) {
		return fmt.Errorf("mp3 服务未就绪，请先启动: mp3 -addr=%s <音乐目录>", addr)
	}
	_, err = mina.PlayByURL(deviceID, playURL, 2)
	return err
}
