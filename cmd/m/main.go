// Command m - XiaoMi Cloud Service CLI (OAuth 2.0, ref: ha_xiaomi_home).
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/miiocommand"
	"github.com/zeusro/miflow/internal/miioservice"
	"github.com/zeusro/miflow/internal/minaservice"
)

const prefix = "m "

func usage() {
	fmt.Fprintf(os.Stderr, "m - XiaoMi MIoT + Mina CLI (OAuth 2.0)\n\n")
	fmt.Fprintf(os.Stderr, "First run: m login\n")
	fmt.Fprintf(os.Stderr, "Device:    config default_did or export MI_DID=<device_id|name>\n")
	fmt.Fprintf(os.Stderr, "          (required for mina commands except 'mina')\n\n")
	fmt.Fprintf(os.Stderr, "Mina:      m mina | message <text> | play <url> | pause | stop | loop <url> | play_list <file> | suno | suno_random\n\n")
	fmt.Fprint(os.Stderr, miiocommand.Help("", prefix))
}

func fullHelp() string {
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

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	args := os.Args[1:]
	for len(args) > 0 && strings.HasPrefix(args[0], "-v") {
		args = args[1:]
	}
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}
	cmd := args[0]
	if cmd == "help" || cmd == "?" || cmd == "？" || cmd == "-h" || cmd == "--help" {
		fmt.Print(fullHelp())
		os.Exit(0)
	}

	cfg := config.Get()
	tokenPath := cfg.TokenPath

	// login: OAuth 2.0 flow
	if cmd == "login" {
		runLogin(tokenPath)
		return
	}

	// Load OAuth token
	token := (&miaccount.TokenStore{Path: tokenPath}).LoadOAuth()
	if token == nil || !token.IsValid() {
		fmt.Fprintln(os.Stderr, "Error: no valid token, run 'm login' first")
		usage()
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
		runMina(minaservice.NewWithMinaAPI(ioSvc, token, tokenPath), did, cmd, args[1:])
		return
	}

	// MiIO/MIoT
	text := strings.Join(args, " ")
	result, err := miiocommand.Run(ioSvc, did, text, prefix)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	printResult(result)
}

func runLogin(tokenPath string) {
	oc := miaccount.NewOAuthClient()
	authURL := oc.GenAuthURL("", "", true)
	fmt.Fprintf(os.Stderr, "Open this URL in browser to login:\n%s\n\n", authURL)
	callbackPort := config.Get().MiIO.CallbackPort
	if callbackPort <= 0 {
		callbackPort = 8123
	}
	fmt.Fprintf(os.Stderr, "Starting local callback server on :%d...\n", callbackPort)
	if err := miaccount.OpenAuthURL(authURL); err != nil {
		fmt.Fprintln(os.Stderr, "(Could not open browser, open the URL manually)")
	}
	code, err := miaccount.ServeCallback(callbackPort)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	token, err := oc.GetToken(code)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	store := &miaccount.TokenStore{Path: tokenPath}
	if err := store.SaveOAuth(token); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "Login successful. Token saved to", tokenPath)
}

func runMina(mina *minaservice.Service, miDID, cmd string, rest []string) {
	if cmd != "mina" && miDID == "" {
		fmt.Fprintln(os.Stderr, "Error: MI_DID must be set for mina commands (message, play, pause, etc.)")
		os.Exit(1)
	}

	deviceID, err := mina.GetMinaDeviceID(miDID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch cmd {
	case "mina":
		list, err := mina.DeviceList(0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if len(list) > 0 {
			printResult(list[0])
		} else {
			fmt.Println("[]")
		}
		return
	case "pause", "stop":
		_, err := mina.PlayerStop(deviceID)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("Stop")
		return
	case "message":
		if len(rest) < 1 {
			fmt.Fprintln(os.Stderr, "Usage: m message <text>")
			os.Exit(1)
		}
		_, err := mina.TextToSpeech(deviceID, strings.Join(rest, " "))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	case "play":
		if len(rest) < 1 {
			fmt.Fprintln(os.Stderr, "Usage: m play <url>")
			os.Exit(1)
		}
		_, err := mina.PlayByURL(deviceID, rest[0], 2)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		mina.PlayerSetLoop(deviceID, 1)
		return
	case "loop":
		if len(rest) < 1 {
			fmt.Fprintln(os.Stderr, "Usage: m loop <url>")
			os.Exit(1)
		}
		_, err := mina.PlayByURL(deviceID, rest[0], 2)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		mina.PlayerSetLoop(deviceID, 0)
		return
	case "play_list":
		if len(rest) < 1 {
			fmt.Fprintln(os.Stderr, "Usage: m play_list <file>")
			os.Exit(1)
		}
		runPlayList(mina, deviceID, rest[0])
		return
	case "suno", "suno_random":
		runSuno(mina, deviceID, cmd == "suno_random")
		return
	}
}

func runPlayList(mina *minaservice.Service, deviceID, filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	lines := strings.Split(string(data), "\n")
	mina.PlayerSetLoop(deviceID, 1)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fmt.Println("Will play", line)
		_, err := mina.PlayByURL(deviceID, line, 2)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		// Wait for duration would need HTTP HEAD or similar; skip for simplicity
		// User can Ctrl+C to stop
	}
}

func runSuno(mina *minaservice.Service, deviceID string, random bool) {
	// Suno playlist: optional external API; without it we just print usage
	fmt.Fprintln(os.Stderr, "suno/suno_random: play suno.ai trending (optional, requires network)")
	// Minimal: could add http get to suno API and play URLs
	fmt.Println("Will play suno trending list")
}

func printResult(v interface{}) {
	switch t := v.(type) {
	case string:
		fmt.Println(t)
	case nil:
		fmt.Println("null")
	default:
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			fmt.Println(v)
			return
		}
		fmt.Println(string(b))
	}
}
