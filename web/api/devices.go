package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/zeusro/miflow/internal/miiocommand"
	"github.com/zeusro/miflow/pkg/i18n"
	"github.com/zeusro/miflow/web"
)

// DevicesList 处理 GET /api/devices - 列出设备
func DevicesList(a *web.App, r *ghttp.Request) {
	if !RequireAuth(a, r) {
		return
	}
	name := r.Get("name").String()
	getVirtual := r.Get("getVirtual").Bool()
	getHuami := r.Get("getHuami").Int()
	list, err := a.DeviceAPI().List(name, getVirtual, getHuami)
	if err != nil {
		Err(r, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(r, http.StatusOK, list)
}

// DeviceGet 处理 GET /api/devices/:id - 获取设备详情
func DeviceGet(a *web.App, r *ghttp.Request) {
	if !RequireAuth(a, r) {
		return
	}
	lang := i18n.AcceptLanguage(r.Header.Get("Accept-Language"))
	id := r.GetRouter("id").String()
	if id == "" {
		Err(r, http.StatusBadRequest, i18n.T(lang, "web.api.device_id_required", nil))
		return
	}
	d, err := a.DeviceAPI().Get(id)
	if err != nil {
		Err(r, http.StatusNotFound, err.Error())
		return
	}
	JSON(r, http.StatusOK, d)
}

// DeviceControl 处理 POST /api/devices/:id/control - 控制设备 (miot 命令)
func DeviceControl(a *web.App, r *ghttp.Request) {
	if !RequireAuth(a, r) {
		return
	}
	lang := i18n.AcceptLanguage(r.Header.Get("Accept-Language"))
	id := r.GetRouter("id").String()
	if id == "" {
		Err(r, http.StatusBadRequest, i18n.T(lang, "web.api.device_id_required", nil))
		return
	}
	var body struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Request.Body).Decode(&body); err != nil {
		Err(r, http.StatusBadRequest, i18n.T(lang, "web.api.invalid_json", nil))
		return
	}
	cmd := strings.TrimSpace(body.Command)
	if cmd == "" {
		Err(r, http.StatusBadRequest, i18n.T(lang, "web.api.command_required", nil))
		return
	}
	_, err := miiocommand.Run(a.Miio(), id, cmd, "web ")
	if err != nil {
		Err(r, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(r, http.StatusOK, map[string]string{"status": "ok"})
}

// DeviceSpec 处理 GET /api/devices/:id/spec - 获取设备 MIoT 规格（用于控制 UI）
func DeviceSpec(a *web.App, r *ghttp.Request) {
	if !RequireAuth(a, r) {
		return
	}
	id := r.GetRouter("id").String()
	if id == "" {
		Err(r, http.StatusBadRequest, "device id required")
		return
	}
	d, err := a.DeviceAPI().Get(id)
	if err != nil {
		Err(r, http.StatusNotFound, err.Error())
		return
	}
	spec, err := a.DeviceAPI().SpecForDevice(d, "json")
	if err != nil {
		Err(r, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(r, http.StatusOK, spec)
}
