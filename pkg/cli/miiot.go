package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/device"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/miioservice"
	"github.com/zeusro/miflow/miiot"
	"github.com/zeusro/miflow/pkg/i18n"
)

// NewMiiotCmd 创建 miiot 子命令。
func NewMiiotCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "miiot",
		Short: "Validate m list device specs against home.miot-spec.com",
		RunE:  runMiiot,
	}
}

func runMiiot(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	tokenPath := cfg.TokenPath
	if tokenPath == "" {
		tokenPath = os.ExpandEnv("$HOME/.mi.token")
	}
	store := &miaccount.TokenStore{Path: tokenPath}
	token := store.LoadOAuth()
	if token == nil || !token.IsValid() {
		return fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "cli.error.no_token", nil))
	}
	ioSvc, err := miioservice.New(token, tokenPath)
	if err != nil {
		return err
	}
	api := device.NewAPI(ioSvc)
	devs, err := api.List("", false, 0)
	if err != nil {
		return err
	}
	seen := make(map[string]bool)
	var models []string
	for _, d := range devs {
		if d != nil && d.Model != "" && !seen[d.Model] {
			seen[d.Model] = true
			models = append(models, d.Model)
		}
	}
	result := make(map[string]interface{})
	ok := make(map[string]string)
	failed := make(map[string]string)
	for _, model := range models {
		specURL, err := miiot.SpecURL(model)
		if err != nil {
			failed[model] = err.Error()
			continue
		}
		ok[model] = specURL
	}
	result["ok"] = ok
	result["failed"] = failed
	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(b))
	if len(failed) > 0 {
		return fmt.Errorf("%d models failed spec validation", len(failed))
	}
	return nil
}
