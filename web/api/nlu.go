package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/nlu"
	"github.com/zeusro/miflow/internal/ollama"
	"github.com/zeusro/miflow/pkg/i18n"
	"github.com/zeusro/miflow/web"
)

// NLUControl 处理 POST /api/nlu/control - 自然语言控制设备。
func NLUControl(a *web.App, r *ghttp.Request) {
	if !RequireAuth(a, r) {
		return
	}
	lang := i18n.AcceptLanguage(r.Header.Get("Accept-Language"))

	cfg := config.Get().Ollama
	if !cfg.Enabled {
		Err(r, http.StatusServiceUnavailable, i18n.T(lang, "nlu.not_enabled", nil))
		return
	}
	if cfg.Host == "" || cfg.Model == "" {
		Err(r, http.StatusServiceUnavailable, i18n.T(lang, "nlu.misconfigured", nil))
		return
	}

	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Request.Body).Decode(&body); err != nil {
		Err(r, http.StatusBadRequest, i18n.T(lang, "web.api.invalid_json", nil))
		return
	}
	if body.Text == "" {
		Err(r, http.StatusBadRequest, i18n.T(lang, "web.api.command_required", nil))
		return
	}

	api := a.DeviceAPI()
	if api == nil {
		Err(r, http.StatusServiceUnavailable, i18n.T(lang, "web.auth.please_login", nil))
		return
	}

	client := ollama.NewClient(cfg.Host, cfg.Model, 0)
	svc := nlu.NewService(api, client)
	res, err := svc.Execute(context.Background(), body.Text)
	if err != nil {
		Err(r, http.StatusInternalServerError, err.Error())
		return
	}

	// 对外返回时不暴露 DID 等敏感信息，仅保留名称与型号。
	JSON(r, http.StatusOK, sanitizeNLUResult(res))
}

func sanitizeNLUResult(res *nlu.Result) map[string]interface{} {
	out := map[string]interface{}{
		"intent":    res.Intent,
		"executed":  res.Executed,
		"ambiguous": res.Ambiguous,
	}
	if res.Error != "" {
		out["error"] = res.Error
	}
	if len(res.Devices) > 0 {
		devs := make([]map[string]interface{}, 0, len(res.Devices))
		for _, d := range res.Devices {
			devs = append(devs, map[string]interface{}{
				"name":  d.Name,
				"model": d.Model,
			})
		}
		out["devices"] = devs
	}
	if len(res.Results) > 0 {
		clean := make([]map[string]interface{}, 0, len(res.Results))
		for _, it := range res.Results {
			c := map[string]interface{}{
				"name":  it["name"],
				"model": it["model"],
				"ok":    it["ok"],
			}
			if e, ok := it["error"]; ok {
				c["error"] = e
			}
			if d, ok := it["data"]; ok {
				c["data"] = d
			}
			clean = append(clean, c)
		}
		out["results"] = clean
	}
	return out
}


