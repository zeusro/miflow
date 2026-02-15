// Package cmd provides the root command and subcommand registration for miflow CLI.
// Structure follows kubectl-style: pkg/cmd/<subcommand>/ for each subcommand.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/miiocommand"
	"github.com/zeusro/miflow/internal/miioservice"
	"github.com/zeusro/miflow/internal/minaservice"
	"github.com/zeusro/miflow/pkg/cmd/login"
	"github.com/zeusro/miflow/pkg/cmd/mina"
	"github.com/zeusro/miflow/pkg/cmd/util"
)

const prefix = "m "

// Usage prints short usage for m.
func Usage() {
	fmt.Fprintf(os.Stderr, "m - XiaoMi MIoT + Mina CLI (OAuth 2.0)\n\n")
	fmt.Fprintf(os.Stderr, "First run: m login\n")
	fmt.Fprintf(os.Stderr, "Device:    config default_did or export MI_DID=<device_id|name>\n")
	fmt.Fprintf(os.Stderr, "          (required for mina commands except 'mina')\n\n")
	fmt.Fprintf(os.Stderr, "Mina:      m mina | message <text> | play <url> | pause | stop | loop <url> | play_list <file> | suno | suno_random\n\n")
	fmt.Fprint(os.Stderr, miiocommand.Help("", prefix))
}

// FullHelp returns the complete help string.
func FullHelp() string {
	return `m - XiaoMi MIoT + Mina CLI (OAuth 2.0)

USAGE
  m <command> [args...]

AUTH
  login              首次使用需执行 OAuth 2.0 登录，在浏览器中完成授权后保存 token

DEVICE
  通过 config 的 default_did 或环境变量 MI_DID 指定设备（device_id 或设备名称）
  mina 命令（除 mina 外）均需指定设备

MINA（小爱音箱 / 语音设备）
  mina              列出 Mina 设备列表
  message <text>    设备 TTS 播报指定文本
  play <url>        播放指定 URL 的音频
  pause             暂停播放
  stop              停止播放（同 pause）
  loop <url>        循环播放指定 URL
  play_list <file>  按文件中的 URL 列表顺序播放（每行一个 URL，# 开头为注释）
  suno              播放 Suno trending 列表（需网络）
  suno_random       随机播放 Suno 列表（需网络）

MIoT / MiIO（设备属性与控制）
  list [name] [getVirtualModel] [getHuamiDevices]
                    列出设备，可选按名称筛选、是否含虚拟设备、华米设备数量
  spec [model] [format]
                    查询 MIoT 规格，format 可选 text|python|json
  decode <ssecurity> <nonce> <data> [gzip]
                    解码 MIoT 加密数据

  siid-piid          获取属性，如 m 1,1-2,1-3,2-1
  siid-piid=value    设置属性，如 m 2=#60,2-2=#false,3=test
  siid-aiid args     执行动作，如 m 5 Hello 或 m 5-4 Hello #1

  prop/get|prop/set|action <params>
                    原始 MIoT 调用，params 为 JSON
  /<uri> <data>     原始 MiIO 调用，如 m /home/device_list '{"getVirtualModel":false}'

EXAMPLES
  m login
  m mina
  m message 你好世界
  m play https://example.com/audio.mp3
  m list Light true 0
  m 2=#60
`
}

// Run executes the m command with given args.
func Run(args []string) {
	for len(args) > 0 && strings.HasPrefix(args[0], "-v") {
		args = args[1:]
	}
	if len(args) == 0 {
		Usage()
		os.Exit(1)
	}

	cmd := args[0]
	if cmd == "help" || cmd == "?" || cmd == "？" || cmd == "-h" || cmd == "--help" {
		fmt.Print(FullHelp())
		os.Exit(0)
	}

	cfg := config.Get()
	tokenPath := cfg.TokenPath

	if cmd == "login" {
		login.Login{TokenPath: tokenPath}.Run()
		return
	}

	token := (&miaccount.TokenStore{Path: tokenPath}).LoadOAuth()
	if token == nil || !token.IsValid() {
		fmt.Fprintln(os.Stderr, "Error: no valid token, run 'm login' first")
		Usage()
		os.Exit(1)
	}

	ioSvc, err := miioservice.New(token, tokenPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	did := cfg.DefaultDID
	minaLikes := map[string]bool{
		"message": true, "play": true, "mina": true, "pause": true, "stop": true,
		"loop": true, "play_list": true, "suno": true, "suno_random": true,
	}
	if minaLikes[cmd] {
		mina.Mina{
			MinaSvc: minaservice.NewWithMinaAPI(ioSvc, token, tokenPath),
			DID:     did,
			Cmd:     cmd,
			Args:    args[1:],
		}.Run()
		return
	}

	// MiIO/MIoT
	text := strings.Join(args, " ")
	result, err := miiocommand.Run(ioSvc, did, text, prefix)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	util.PrintResult(result)
}
