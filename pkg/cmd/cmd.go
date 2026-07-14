// Package cmd 提供 miflow CLI 的根命令和子命令注册。
// 结构遵循 kubectl 风格：每个子命令对应 pkg/cmd/<subcommand>/。
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/device"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/miiocommand"
	"github.com/zeusro/miflow/internal/miioservice"
	"github.com/zeusro/miflow/internal/minaservice"
	"github.com/zeusro/miflow/internal/nlu"
	"github.com/zeusro/miflow/internal/ollama"
	"github.com/zeusro/miflow/pkg/cmd/mina"
	"github.com/zeusro/miflow/pkg/cmd/util"
	"github.com/zeusro/miflow/pkg/i18n"
)

const prefix = "m "

// Usage 打印 m 的简短用法。
func Usage() {
	lang := i18n.DefaultLang()
	fmt.Fprint(os.Stderr, i18n.T(lang, "cli.usage.title", nil))
	fmt.Fprint(os.Stderr, i18n.T(lang, "cli.usage.first_run", nil))
	fmt.Fprint(os.Stderr, i18n.T(lang, "cli.usage.device", nil))
	fmt.Fprint(os.Stderr, i18n.T(lang, "cli.usage.device_required", nil))
	fmt.Fprint(os.Stderr, i18n.T(lang, "cli.usage.nlu", nil))
	fmt.Fprint(os.Stderr, i18n.T(lang, "cli.usage.mina", nil))
	fmt.Fprint(os.Stderr, i18n.T(lang, "cli.usage.miio", nil))
	fmt.Fprint(os.Stderr, miiocommand.Help("", prefix, lang))
}

// FullHelp 返回完整帮助字符串。
func FullHelp() string {
	return i18n.T(i18n.DefaultLang(), "cli.full_help", nil)
}

// Run 使用给定参数执行 m 命令。
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

	token := (&miaccount.TokenStore{Path: tokenPath}).LoadOAuth()
	if token == nil || !token.IsValid() {
		fmt.Fprintln(os.Stderr, i18n.T(i18n.DefaultLang(), "cli.error.no_token", nil))
		Usage()
		os.Exit(1)
	}

	ioSvc, err := miioservice.New(token, tokenPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	did := cfg.DefaultDID
	if cmd == "ask" || cmd == "nlu" {
		if !cfg.Ollama.Enabled {
			fmt.Fprintln(os.Stderr, i18n.T(i18n.DefaultLang(), "nlu.not_enabled", nil))
			os.Exit(1)
		}
		if cfg.Ollama.Model == "" || cfg.Ollama.Host == "" {
			fmt.Fprintln(os.Stderr, i18n.T(i18n.DefaultLang(), "nlu.misconfigured", nil))
			os.Exit(1)
		}
		api := device.NewAPI(ioSvc)
		client := ollama.NewClient(cfg.Ollama.Host, cfg.Ollama.Model, 0)
		svc := nlu.NewService(api, client)
		text := strings.Join(args[1:], " ")
		res, err := svc.Execute(context.Background(), text)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		util.PrintResult(res)
		return
	}

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

	// MiIO/MIoT 命令
	text := strings.Join(args, " ")
	result, err := miiocommand.Run(ioSvc, did, text, prefix)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	util.PrintResult(result)
}
