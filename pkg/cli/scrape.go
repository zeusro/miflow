package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zeusro/miflow/miiot/specs"
	"github.com/zeusro/miflow/pkg/i18n"
)

// NewScrapeSpecsCmd 创建 scrape-specs 子命令。
func NewScrapeSpecsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scrape-specs",
		Short: "Scrape device spec URLs from m list, output Markdown table",
		RunE:  runScrapeSpecs,
	}
}

func runScrapeSpecs(cmd *cobra.Command, args []string) error {
	lang := i18n.DefaultLang()
	models, err := getModels()
	if err != nil {
		return fmt.Errorf("%s", i18n.T(lang, "scrape.get_models_err", map[string]interface{}{"Err": err.Error()}))
	}
	if len(models) == 0 {
		return fmt.Errorf("%s", i18n.T(lang, "scrape.no_models", nil))
	}
	sort.Strings(models)

	var rows [][]string
	for _, model := range models {
		specURL, err := specs.TechSpecURL(model)
		if err != nil {
			rows = append(rows, []string{model, productURL(model), "-"})
			fmt.Fprint(os.Stderr, i18n.T(lang, "scrape.warn", map[string]interface{}{"Model": model, "Err": err}))
			continue
		}
		rows = append(rows, []string{model, productURL(model), specURL})
	}

	fmt.Println(i18n.T(lang, "scrape.table_header", nil))
	fmt.Println("|-------|--------|---------------------|")
	for _, r := range rows {
		specCell := r[2]
		if specCell != "-" {
			specCell = fmt.Sprintf("[链接](%s)", specCell)
		}
		fmt.Printf("| %s | [链接](%s) | %s |\n", r[0], r[1], specCell)
	}
	return nil
}

func productURL(model string) string {
	return specs.ProductBaseURL + "/" + model
}

func getModels() ([]string, error) {
	// 尝试 m list 或 miflow m list
	var c *exec.Cmd
	if _, err := os.Stat("./m"); err == nil {
		c = exec.Command("./m", "list")
	} else if _, err := exec.LookPath("miflow"); err == nil {
		c = exec.Command("miflow", "m", "list")
	} else {
		c = exec.Command("m", "list")
	}
	c.Stderr = os.Stderr
	out, err := c.Output()
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
