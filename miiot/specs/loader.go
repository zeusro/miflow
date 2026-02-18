package specs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/zeusro/miflow/internal/config"
)

const (
	InstancesURL   = "http://miot-spec.org/miot-spec-v2/instances?status=all"
	SpecBaseURL    = "https://home.miot-spec.com/spec"
	ProductBaseURL = "https://home.miot-spec.com/s"
)

var (
	modelToURN map[string]string
	loadOnce   sync.Once
	loadErr    error
)

// Load 从 miot-spec.org 加载 model->URN 映射，与 home.miot-spec.com 的规格链接一致。
func Load() (map[string]string, error) {
	loadOnce.Do(func() {
		modelToURN, loadErr = loadInstances()
	})
	return modelToURN, loadErr
}

func loadInstances() (map[string]string, error) {
	cachePath := config.Get().MiIO.SpecsCachePath
	if cachePath == "" {
		cachePath = filepath.Join(os.TempDir(), "miservice_miot_specs.json")
	}
	if data, err := os.ReadFile(cachePath); err == nil {
		var m map[string]string
		if json.Unmarshal(data, &m) == nil && len(m) > 0 {
			return m, nil
		}
	}
	resp, err := http.DefaultClient.Get(InstancesURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var inst struct {
		Instances []struct {
			Model string `json:"model"`
			Type  string `json:"type"`
		} `json:"instances"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&inst); err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, i := range inst.Instances {
		m[i.Model] = i.Type
	}
	if b, err := json.Marshal(m); err == nil {
		os.WriteFile(cachePath, b, 0644)
	}
	return m, nil
}

// SpecURL 返回 model 对应的 home.miot-spec.com 规格页 URL。
func SpecURL(model string) (string, error) {
	m, err := Load()
	if err != nil {
		return "", err
	}
	urn, ok := m[model]
	if !ok {
		return "", fmt.Errorf("model not found: %s", model)
	}
	return SpecBaseURL + "?type=" + url.QueryEscape(urn), nil
}

// ProductURL 返回 model 对应的 home.miot-spec.com 产品页 URL。
func ProductURLForModel(model string) string {
	return ProductBaseURL + "/" + model
}

// URN 返回 model 对应的 URN。
func URN(model string) (string, error) {
	return URNWithScrape(model)
}

// URNWithScrape 返回 URN，若 instances 中无则尝试从产品页抓取。
func URNWithScrape(model string) (string, error) {
	m, err := Load()
	if err != nil {
		return "", err
	}
	urn, ok := m[model]
	if !ok {
		if scraped, err := ScrapeProductPage(model); err == nil && scraped[model] != "" {
			return scraped[model], nil
		}
		return "", fmt.Errorf("model not found: %s", model)
	}
	return urn, nil
}
