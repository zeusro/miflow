package specs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// ScrapeProductPage 从 https://home.miot-spec.com/s/{model} 抓取页面，
// 解析嵌入的 JSON 提取 model->URN 映射（含当前型号及关联型号）。
// 页面中的 specs 数组包含 model 与 type(URN)。HTML 中可能使用 &quot; 表示引号。
func ScrapeProductPage(model string) (map[string]string, error) {
	url := ProductBaseURL + "/" + model
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("scraper: %s returned %d", url, resp.StatusCode)
	}
	body := make([]byte, 0, 512*1024)
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		body = append(body, buf[:n]...)
		if n == 0 || err != nil {
			break
		}
	}
	return parseProductPageBody(body)
}

// 匹配 HTML 中 model&quot;:&quot;xxx 与 type&quot;:&quot;urn:... 对（中间可有 status,version 等）
var modelTypeRE = regexp.MustCompile(`model&quot;:&quot;([a-z0-9._-]+)&quot;.*?type&quot;:&quot;(urn:miot-spec[^&]+)`)

func parseProductPageBody(body []byte) (map[string]string, error) {
	result := make(map[string]string)
	bodyStr := string(body)
	matches := modelTypeRE.FindAllStringSubmatch(bodyStr, -1)
	for _, m := range matches {
		if len(m) < 3 {
			continue
		}
		model := strings.TrimSpace(m[1])
		urn := strings.TrimSpace(m[2])
		var decoded string
		if json.Unmarshal([]byte(`"`+urn+`"`), &decoded) == nil {
			urn = decoded
		}
		if model != "" && urn != "" {
			result[model] = urn
		}
	}
	return result, nil
}

// dataPageProps 解析 data-page 嵌入的 Inertia 页面数据
type dataPageProps struct {
	Props struct {
		List []struct {
			Model string `json:"model"`
			Specs []struct {
				Model string `json:"model"`
				Type  string `json:"type"`
			} `json:"specs"`
		} `json:"list"`
	} `json:"props"`
}

var dataPageRE = regexp.MustCompile(`data-page="([^"]*)"`)

// TechSpecURL 从产品页 https://home.miot-spec.com/s/{model} 抓取，
// 解析 data-page 中 list[].specs 提取技术说明 URL（class="divider" 下的规格链接）。
// 优先使用与 model 完全匹配的 spec，否则使用 list 首项的 specs[0]。
func TechSpecURL(model string) (string, error) {
	pageURL := ProductBaseURL + "/" + url.PathEscape(model)
	resp, err := http.DefaultClient.Get(pageURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("scraper: %s returned %d", pageURL, resp.StatusCode)
	}
	body := make([]byte, 0, 512*1024)
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		body = append(body, buf[:n]...)
		if n == 0 || err != nil {
			break
		}
	}
	// 提取 data-page 内容
	matches := dataPageRE.FindSubmatch(body)
	if len(matches) < 2 {
		// 回退到 ScrapeProductPage 的 regex 解析
		m, err := parseProductPageBody(body)
		if err != nil {
			return "", err
		}
		if urn, ok := m[model]; ok && urn != "" {
			return SpecBaseURL + "?type=" + url.QueryEscape(urn), nil
		}
		return "", fmt.Errorf("model not found in page: %s", model)
	}
	// 将 &quot; 还原为 "
	raw := string(matches[1])
	raw = strings.ReplaceAll(raw, "&quot;", `"`)
	raw = strings.ReplaceAll(raw, "&amp;", "&")
	var page dataPageProps
	if err := json.Unmarshal([]byte(raw), &page); err != nil {
		return "", fmt.Errorf("parse data-page: %w", err)
	}
	list := page.Props.List
	if len(list) == 0 {
		return "", fmt.Errorf("no product in list: %s", model)
	}
	// 查找与 model 匹配的 spec
	for _, item := range list {
		if item.Model != model {
			continue
		}
		for _, s := range item.Specs {
			if s.Type != "" {
				return SpecBaseURL + "?type=" + url.QueryEscape(s.Type), nil
			}
		}
	}
	// 使用首项的首个 spec
	if len(list[0].Specs) > 0 && list[0].Specs[0].Type != "" {
		return SpecBaseURL + "?type=" + url.QueryEscape(list[0].Specs[0].Type), nil
	}
	return "", fmt.Errorf("no spec found for model: %s", model)
}
