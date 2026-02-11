// Command xiaomusic - minimal XiaoMusic-style CLI entry.
// Goal: eventually mirror https://github.com/hanxi/xiaomusic behaviour,
// but currently only provides a thin wrapper around the existing MiService-based
// playback (command `m`) for simple URL/local-file播放。
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

	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/minaservice"
)

var (
	flagMusicDir = flag.String("music_dir", "./music", "本地音乐目录，用于本地文件播放")
	flagAddr     = flag.String("addr", ":8090", "本地静态文件 HTTP 服务监听地址")
)

func usage() {
	fmt.Fprintf(os.Stderr, "xiaomusic - 使用小爱音箱播放本地或网络音乐（简化版）\n\n")
	fmt.Fprintf(os.Stderr, "环境变量：\n")
	fmt.Fprintf(os.Stderr, "  MI_USER=<小米账号>\n")
	fmt.Fprintf(os.Stderr, "  MI_PASS=<密码>\n")
	fmt.Fprintf(os.Stderr, "  MI_DID=<设备ID或名称>\n\n")
	fmt.Fprintf(os.Stderr, "用法示例：\n")
	fmt.Fprintf(os.Stderr, "  xiaomusic -music_dir=./music play-url https://example.com/a.mp3\n")
	fmt.Fprintf(os.Stderr, "  xiaomusic -music_dir=./music play-file song.mp3\n")
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	cmd := args[0]
	rest := args[1:]

	user := os.Getenv("MI_USER")
	pass := os.Getenv("MI_PASS")
	did := os.Getenv("MI_DID")
	if user == "" || pass == "" || did == "" {
		fmt.Fprintln(os.Stderr, "错误：必须设置环境变量 MI_USER / MI_PASS / MI_DID")
		usage()
		os.Exit(1)
	}

	tokenPath := filepath.Join(os.Getenv("HOME"), ".mi.token")
	account := miaccount.NewAccount(user, pass, tokenPath)
	mina := minaservice.New(account)

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

