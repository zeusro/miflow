package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/miiocommand"
	"github.com/zeusro/miflow/internal/miioservice"
	"github.com/zeusro/miflow/internal/minaservice"
	"github.com/zeusro/miflow/pkg/cmd/mina"
	"github.com/zeusro/miflow/pkg/cmd/util"
	"github.com/zeusro/miflow/pkg/i18n"
)

// minaLikes 需要走 Mina 服务的子命令。
var minaLikes = map[string]bool{
	"message": true, "play": true, "mina": true, "pause": true, "stop": true,
	"loop": true, "play_list": true, "suno": true, "suno_random": true,
}

// NewMCmd 创建 m 子命令（小米云服务 CLI）。
func NewMCmd() *cobra.Command {
	m := &cobra.Command{
		Use:   "m [command] [args...]",
		Short: "XiaoMi MIoT + Mina CLI (OAuth 2.0)",
		Long:  i18n.T(i18n.DefaultLang(), "cli.full_help", nil),
		Args:  cobra.ArbitraryArgs,
		RunE:  runM,
	}
	// 兼容 -v 等，Cobra 会解析
	m.Flags().BoolP("verbose", "v", false, "verbose output")
	m.Flags().Lookup("verbose").Hidden = true
	return m
}

func runM(cmd *cobra.Command, args []string) error {
	// 跳过 -v 前缀（保持与旧版兼容）
	for len(args) > 0 && strings.HasPrefix(args[0], "-v") {
		args = args[1:]
	}
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	sub := args[0]
	if sub == "help" || sub == "?" || sub == "？" || sub == "-h" || sub == "--help" {
		fmt.Print(i18n.T(i18n.DefaultLang(), "cli.full_help", nil))
		return nil
	}

	cfg := config.Get()
	tokenPath := cfg.TokenPath
	token := (&miaccount.TokenStore{Path: tokenPath}).LoadOAuth()
	if token == nil || !token.IsValid() {
		fmt.Fprintln(os.Stderr, i18n.T(i18n.DefaultLang(), "cli.error.no_token", nil))
		printUsage()
		os.Exit(1)
	}

	ioSvc, err := miioservice.New(token, tokenPath)
	if err != nil {
		return err
	}

	did := cfg.DefaultDID
	if minaLikes[sub] {
		mina.Mina{
			MinaSvc: minaservice.NewWithMinaAPI(ioSvc, token, tokenPath),
			DID:     did,
			Cmd:     sub,
			Args:    args[1:],
		}.Run()
		return nil
	}

	// MiIO/MIoT 命令
	text := strings.Join(args, " ")
	result, err := miiocommand.Run(ioSvc, did, text, prefix)
	if err != nil {
		return err
	}
	util.PrintResult(result)
	return nil
}

func printUsage() {
	lang := i18n.DefaultLang()
	fmt.Fprint(os.Stderr, i18n.T(lang, "cli.usage.title", nil))
	fmt.Fprint(os.Stderr, i18n.T(lang, "cli.usage.first_run", nil))
	fmt.Fprint(os.Stderr, i18n.T(lang, "cli.usage.device", nil))
	fmt.Fprint(os.Stderr, i18n.T(lang, "cli.usage.device_required", nil))
	fmt.Fprint(os.Stderr, i18n.T(lang, "cli.usage.mina", nil))
	fmt.Fprint(os.Stderr, i18n.T(lang, "cli.usage.miio", nil))
	fmt.Fprint(os.Stderr, miiocommand.Help("", prefix, lang))
}
