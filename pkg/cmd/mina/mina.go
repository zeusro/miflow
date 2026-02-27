// Package mina 实现 m 的 mina 相关子命令（mina, message, play, pause, stop, loop, play_list, suno, suno_random）。
package mina

import (
	"fmt"
	"os"
	"strings"

	"github.com/zeusro/miflow/internal/minaservice"
	"github.com/zeusro/miflow/pkg/cmd/util"
	"github.com/zeusro/miflow/pkg/i18n"
)

// Mina 运行 mina 子命令。
type Mina struct {
	MinaSvc *minaservice.Service
	DID     string
	Cmd     string
	Args    []string
}

// Run 执行 mina 子命令。
func (m Mina) Run() {
	lang := i18n.DefaultLang()
	if m.Cmd != "mina" && m.DID == "" {
		fmt.Fprintln(os.Stderr, i18n.T(lang, "mina.error.no_did", nil))
		os.Exit(1)
	}

	deviceID, err := m.MinaSvc.GetMinaDeviceID(m.DID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch m.Cmd {
	case "mina":
		list, err := m.MinaSvc.DeviceList(0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if len(list) > 0 {
			util.PrintResult(list[0])
		} else {
			fmt.Println("[]")
		}
		return
	case "pause", "stop":
		_, err := m.MinaSvc.PlayerStop(deviceID)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("Stop")
		return
	case "message":
		if len(m.Args) < 1 {
			fmt.Fprintln(os.Stderr, i18n.T(lang, "mina.usage.message", nil))
			os.Exit(1)
		}
		_, err := m.MinaSvc.TextToSpeech(deviceID, strings.Join(m.Args, " "))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	case "play":
		if len(m.Args) < 1 {
			fmt.Fprintln(os.Stderr, i18n.T(lang, "mina.usage.play", nil))
			os.Exit(1)
		}
		_, err := m.MinaSvc.PlayByURL(deviceID, m.Args[0], 2)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		m.MinaSvc.PlayerSetLoop(deviceID, 1)
		return
	case "loop":
		if len(m.Args) < 1 {
			fmt.Fprintln(os.Stderr, i18n.T(lang, "mina.usage.loop", nil))
			os.Exit(1)
		}
		_, err := m.MinaSvc.PlayByURL(deviceID, m.Args[0], 2)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		m.MinaSvc.PlayerSetLoop(deviceID, 0)
		return
	case "play_list":
		if len(m.Args) < 1 {
			fmt.Fprintln(os.Stderr, i18n.T(lang, "mina.usage.play_list", nil))
			os.Exit(1)
		}
		runPlayList(m.MinaSvc, deviceID, m.Args[0])
		return
	case "suno", "suno_random":
		runSuno(m.MinaSvc, deviceID, m.Cmd == "suno_random")
		return
	}
}

// runPlayList 按文件中的 URL 列表顺序播放（每行一个，# 为注释）。
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
		fmt.Println(i18n.T(i18n.DefaultLang(), "mina.will_play", nil), line)
		_, err := mina.PlayByURL(deviceID, line, 2)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
	}
}

// runSuno 播放 Suno trending 列表（需网络，random 为随机模式）。
func runSuno(mina *minaservice.Service, deviceID string, random bool) {
	lang := i18n.DefaultLang()
	fmt.Fprintln(os.Stderr, i18n.T(lang, "mina.suno_hint", nil))
	fmt.Println(i18n.T(lang, "mina.suno_will_play", nil))
}
