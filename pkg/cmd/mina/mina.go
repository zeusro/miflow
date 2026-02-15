// Package mina implements m mina-related subcommands (mina, message, play, pause, stop, loop, play_list, suno, suno_random).
package mina

import (
	"fmt"
	"os"
	"strings"

	"github.com/zeusro/miflow/internal/minaservice"
	"github.com/zeusro/miflow/pkg/cmd/util"
)

// Mina runs mina subcommands.
type Mina struct {
	MinaSvc *minaservice.Service
	DID     string
	Cmd     string
	Args    []string
}

// Run executes the mina subcommand.
func (m Mina) Run() {
	if m.Cmd != "mina" && m.DID == "" {
		fmt.Fprintln(os.Stderr, "Error: MI_DID must be set for mina commands (message, play, pause, etc.)")
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
			fmt.Fprintln(os.Stderr, "Usage: m message <text>")
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
			fmt.Fprintln(os.Stderr, "Usage: m play <url>")
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
			fmt.Fprintln(os.Stderr, "Usage: m loop <url>")
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
			fmt.Fprintln(os.Stderr, "Usage: m play_list <file>")
			os.Exit(1)
		}
		runPlayList(m.MinaSvc, deviceID, m.Args[0])
		return
	case "suno", "suno_random":
		runSuno(m.MinaSvc, deviceID, m.Cmd == "suno_random")
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
	}
}

func runSuno(mina *minaservice.Service, deviceID string, random bool) {
	fmt.Fprintln(os.Stderr, "suno/suno_random: play suno.ai trending (optional, requires network)")
	fmt.Println("Will play suno trending list")
}
