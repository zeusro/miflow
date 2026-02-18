// Command scrape-specs - 从 ./m list 获取设备列表，爬取各型号技术说明 URL，输出 Markdown 表格。
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/zeusro/miflow/miiot/specs"
)

func main() {
	models, err := getModels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "get models: %v\n", err)
		os.Exit(1)
	}
	if len(models) == 0 {
		fmt.Fprintln(os.Stderr, "no models from m list")
		os.Exit(1)
	}
	sort.Strings(models)

	var rows [][]string
	for _, model := range models {
		specURL, err := specs.TechSpecURL(model)
		if err != nil {
			rows = append(rows, []string{model, productURL(model), "-"})
			fmt.Fprintf(os.Stderr, "warn: %s: %v\n", model, err)
			continue
		}
		rows = append(rows, []string{model, productURL(model), specURL})
	}

	// 输出 Markdown 表格
	fmt.Println("| model | 产品页 | 技术说明 (Spec URL) |")
	fmt.Println("|-------|--------|---------------------|")
	for _, r := range rows {
		specCell := r[2]
		if specCell != "-" {
			specCell = fmt.Sprintf("[链接](%s)", specCell)
		}
		fmt.Printf("| %s | [链接](%s) | %s |\n", r[0], r[1], specCell)
	}
}

func productURL(model string) string {
	return specs.ProductBaseURL + "/" + model
}

func getModels() ([]string, error) {
	mBin := "m"
	if _, err := os.Stat("./m"); err == nil {
		mBin = "./m"
	}
	cmd := exec.Command(mBin, "list")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var devs []struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(out, &devs); err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	var models []string
	for _, d := range devs {
		m := strings.TrimSpace(d.Model)
		if m != "" && !seen[m] {
			seen[m] = true
			models = append(models, m)
		}
	}
	return models, nil
}
