// Command m - XiaoMi Cloud Service CLI (OAuth 2.0, ref: ha_xiaomi_home).
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/miiocommand"
	"github.com/zeusro/miflow/internal/miioservice"
	"github.com/zeusro/miflow/internal/minaservice"
)

const prefix = "m "

func usage() {
	fmt.Fprintf(os.Stderr, "MiService (m) - XiaoMi MIoT CLI (OAuth 2.0)\n\n")
	fmt.Fprintf(os.Stderr, "First run: m login\n")
	fmt.Fprintf(os.Stderr, "Optional:  export MI_DID=<device_id|name>\n\n")
	fmt.Fprint(os.Stderr, miiocommand.Help("", prefix))
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
		fmt.Print(miiocommand.Help("", prefix))
		os.Exit(0)
	}

	tokenPath := filepath.Join(os.Getenv("HOME"), ".mi.token")

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

	did := os.Getenv("MI_DID")

	minaLikes := map[string]bool{
		"message": true, "play": true, "mina": true, "pause": true, "stop": true,
		"loop": true, "play_list": true, "suno": true, "suno_random": true,
	}
	if minaLikes[cmd] {
		runMina(minaservice.New(ioSvc), did, cmd, args[1:])
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
	fmt.Fprintln(os.Stderr, "Starting local callback server on :8765...")
	if err := miaccount.OpenAuthURL(authURL); err != nil {
		fmt.Fprintln(os.Stderr, "(Could not open browser, open the URL manually)")
	}
	code, err := miaccount.ServeCallback(8765)
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

