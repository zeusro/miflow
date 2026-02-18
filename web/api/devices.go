package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/zeusro/miflow/internal/miiocommand"
	"github.com/zeusro/miflow/web"
)

// DevicesList handles GET /api/devices - list devices
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

// DeviceGet handles GET /api/devices/:id - get device detail
func DeviceGet(a *web.App, r *ghttp.Request) {
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
	JSON(r, http.StatusOK, d)
}

// DeviceControl handles POST /api/devices/:id/control - control device (miot command)
func DeviceControl(a *web.App, r *ghttp.Request) {
	if !RequireAuth(a, r) {
		return
	}
	id := r.GetRouter("id").String()
	if id == "" {
		Err(r, http.StatusBadRequest, "device id required")
		return
	}
	var body struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Request.Body).Decode(&body); err != nil {
		Err(r, http.StatusBadRequest, "invalid JSON")
		return
	}
	cmd := strings.TrimSpace(body.Command)
	if cmd == "" {
		Err(r, http.StatusBadRequest, "command required")
		return
	}
	_, err := miiocommand.Run(a.Miio(), id, cmd, "web ")
	if err != nil {
		Err(r, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(r, http.StatusOK, map[string]string{"status": "ok"})
}

// DeviceSpec handles GET /api/devices/:id/spec - get device MIoT spec (for control UI)
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
