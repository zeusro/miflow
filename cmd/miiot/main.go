// Command miiot - 列出 m list 中所有型号的规格 API，验证与 home.miot-spec.com 的 1:1 匹配。
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/device"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/miioservice"
	"github.com/zeusro/miflow/miiot"
)

func main() {
	cfg := config.Get()
	tokenPath := cfg.TokenPath
	if tokenPath == "" {
		tokenPath = os.ExpandEnv("$HOME/.mi.token")
	}
	store := &miaccount.TokenStore{Path: tokenPath}
	token := store.LoadOAuth()
	if token == nil || !token.IsValid() {
		fmt.Fprintln(os.Stderr, "Error: no valid token, run 'm login' first")
		os.Exit(1)
	}
	ioSvc, err := miioservice.New(token, tokenPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	api := device.NewAPI(ioSvc)
	devs, err := api.List("", false, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// 提取唯一 model
	seen := make(map[string]bool)
	var models []string
	for _, d := range devs {
		if d != nil && d.Model != "" && !seen[d.Model] {
			seen[d.Model] = true
			models = append(models, d.Model)
		}
	}
	// 验证每个 model 的 SpecURL
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
		os.Exit(1)
	}
}
