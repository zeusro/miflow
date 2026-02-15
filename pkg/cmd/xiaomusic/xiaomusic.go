// Package xiaomusic implements xiaomusic subcommands (play-url, play-file).
package xiaomusic

import (
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

// Options holds xiaomusic command options.
type Options struct {
	MusicDir string
	Addr     string
	Host     string
}

// Run executes the xiaomusic command.
func Run(cmd string, args []string, opts Options) error {
	cfg := config.Get()
	if opts.MusicDir == "" {
		opts.MusicDir = cfg.Xiaomusic.MusicDir
	}
	if opts.Addr == "" {
		opts.Addr = cfg.Xiaomusic.Addr
	}
	if opts.Host == "" {
		opts.Host = cfg.Xiaomusic.Host
	}

	did := cfg.DefaultDID
	if did == "" {
		return fmt.Errorf("必须设置 default_did（配置文件）或环境变量 MI_DID")
	}

	tokenPath := cfg.TokenPath
	token := (&miaccount.TokenStore{Path: tokenPath}).LoadOAuth()
	if token == nil || !token.IsValid() {
		return fmt.Errorf("未登录，请先运行 m login")
	}

	ioSvc, err := miioservice.New(token, tokenPath)
	if err != nil {
		return err
	}
	mina := minaservice.NewWithMinaAPI(ioSvc, token, tokenPath)

	switch cmd {
	case "play-url":
		if len(args) < 1 {
			return fmt.Errorf("用法：xiaomusic play-url <mp3_url>")
		}
		return playURL(mina, did, args[0])
	case "play-file":
		if len(args) < 1 {
			return fmt.Errorf("用法：xiaomusic play-file <相对或绝对文件路径>")
		}
		return playFile(mina, did, opts.MusicDir, args[0], opts.Addr, opts.Host)
	default:
		return fmt.Errorf("未知子命令：%s", cmd)
	}
}

// Usage prints xiaomusic usage.
func Usage() {
	fmt.Fprintf(os.Stderr, "xiaomusic - 使用小爱音箱播放本地或网络音乐（OAuth 模式）\n\n")
	fmt.Fprintf(os.Stderr, "请先运行: m login\n")
	fmt.Fprintf(os.Stderr, "环境变量: MI_DID=<设备ID或名称>\n\n")
	fmt.Fprintf(os.Stderr, "用法示例：\n")
	fmt.Fprintf(os.Stderr, "  xiaomusic -music_dir=./music play-url https://example.com/a.mp3\n")
	fmt.Fprintf(os.Stderr, "  xiaomusic -music_dir=./music play-file song.mp3\n")
	fmt.Fprintf(os.Stderr, "  xiaomusic -host=192.168.1.100 play-file /path/to/music.mp3  # 指定本机 IP\n")
	fmt.Fprintf(os.Stderr, "\n注：基于 MiNA API (api2.mina.mi.com)，参考 https://github.com/hanxi/xiaomusic\n")
}

func playURL(mina *minaservice.Service, miDID, url string) error {
	deviceID, err := mina.GetMinaDeviceID(miDID)
	if err != nil {
		return err
	}
	fmt.Println("播放 URL：", url)
	_, err = mina.PlayByURL(deviceID, strings.TrimSpace(url), 2)
	return err
}

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
	if !srv.WaitPortReady(5 * time.Second) {
		return fmt.Errorf("mp3 服务未就绪，请先启动: mp3 -addr=%s <音乐目录>", addr)
	}
	_, err = mina.PlayByURL(deviceID, playURL, 2)
	return err
}
