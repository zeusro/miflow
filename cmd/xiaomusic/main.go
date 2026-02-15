// Command xiaomusic - minimal XiaoMusic-style CLI entry (OAuth 2.0).
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/miioservice"
	"github.com/zeusro/miflow/internal/minaservice"
)

func usage() {
	fmt.Fprintf(os.Stderr, "xiaomusic - 使用小爱音箱播放本地或网络音乐（OAuth 模式）\n\n")
	fmt.Fprintf(os.Stderr, "请先运行: m login\n")
	fmt.Fprintf(os.Stderr, "环境变量: MI_DID=<设备ID或名称>\n\n")
	fmt.Fprintf(os.Stderr, "用法示例：\n")
	fmt.Fprintf(os.Stderr, "  xiaomusic -music_dir=./music play-url https://example.com/a.mp3\n")
	fmt.Fprintf(os.Stderr, "  xiaomusic -music_dir=./music play-file song.mp3\n")
	fmt.Fprintf(os.Stderr, "\n注：OAuth 模式下 play-url/play-file 可能需设备特定 MIoT 动作，\n")
	fmt.Fprintf(os.Stderr, "    可用 m spec <音箱型号> 查看支持的播放动作。\n")
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cfg := config.Get()
	flagMusicDir := flag.String("music_dir", cfg.Xiaomusic.MusicDir, "本地音乐目录，用于本地文件播放")
	flagAddr := flag.String("addr", cfg.Xiaomusic.Addr, "本地静态文件 HTTP 服务监听地址")
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
	mina := minaservice.New(ioSvc)

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
		if err := playFile(mina, did, *flagMusicDir, rest[0], *flagAddr); err != nil {
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
func playFile(mina *minaservice.Service, miDID, musicDir, filePath, addr string) error {
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
	if !strings.HasPrefix(target, root) {
		// 如果传的是文件名，就认为在 musicDir 下面
		target = filepath.Join(root, filePath)
	}
	if _, err := os.Stat(target); err != nil {
		return fmt.Errorf("文件不存在: %s (%v)", target, err)
	}

	// 在随机端口或指定端口起 HTTP 文件服务
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir(root))
	mux.Handle("/", fs)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("启动 HTTP 服务失败: %w", err)
	}
	defer ln.Close()

	// 构造相对 URL
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return err
	}
	rel = strings.ReplaceAll(rel, string(filepath.Separator), "/")

	hostPort := ln.Addr().String()
	url := fmt.Sprintf("http://%s/%s", hostPort, rel)
	fmt.Println("本地文件映射为 URL：", url)

	// 异步启动 HTTP 服务
	go func() {
		if err := http.Serve(ln, mux); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			log.Println("HTTP 服务错误:", err)
		}
	}()

	_, err = mina.PlayByURL(deviceID, url, 2)
	return err
}

