package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/zeusro/miflow/internal/miiocommand"
	"github.com/zeusro/miflow/miiot/ctrl"
	"github.com/zeusro/miflow/pkg/i18n"
	"github.com/zeusro/miflow/web"
)

// RoomsList 处理 GET /api/rooms - 列出房间及设备
func RoomsList(a *web.App, r *ghttp.Request) {
	if !RequireAuth(a, r) {
		return
	}
	list, err := a.DeviceAPI().RoomsWithDevices()
	if err != nil {
		Err(r, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(r, http.StatusOK, list)
}

// RoomsDevicesList 处理 GET /api/rooms/:roomId/devices - 按房间返回设备列表（含 status）
func RoomsDevicesList(a *web.App, r *ghttp.Request) {
	if !RequireAuth(a, r) {
		return
	}
	roomID := r.GetRouter("roomId").String()
	homeID := r.Get("homeId").String()
	if roomID == "" {
		Err(r, http.StatusBadRequest, "roomId required")
		return
	}
	homes, err := a.DeviceAPI().RoomsWithDevices()
	if err != nil {
		Err(r, http.StatusInternalServerError, err.Error())
		return
	}
	var devices []map[string]interface{}
	c := ctrl.New(a.DeviceAPI())
	for _, h := range homes {
		if homeID != "" && h.HomeID != homeID {
			continue
		}
		for _, room := range h.Rooms {
			if room.RoomID != roomID {
				continue
			}
			for _, d := range room.Devices {
				model := d.Model
				if model == "" {
					model = d.DID
				}
				status := map[string]interface{}{"online": false, "supported": []string{}}
				supported := []string{}
				if on, err := c.GetOn(d.DID, model); err == nil {
					status["online"] = true
					status["value"] = on
					status["on"] = on
					supported = append(supported, "on")
				}
				if brightness, err := c.GetBrightness(d.DID, model); err == nil {
					status["brightness"] = brightness
					supported = append(supported, "brightness")
				}
				if volume, err := c.GetVolume(d.DID, model); err == nil {
					status["volume"] = volume
					supported = append(supported, "volume")
				}
				status["supported"] = supported
				devices = append(devices, map[string]interface{}{
					"id": d.DID, "did": d.DID, "name": d.Name, "model": d.Model, "type": inferDeviceType(d.Model),
					"room_id": room.RoomID, "room_name": room.RoomName, "home_id": h.HomeID,
					"status": status,
				})
			}
		}
	}
	JSON(r, http.StatusOK, devices)
}

func inferDeviceType(model string) string {
	m := strings.ToLower(model)
	if strings.Contains(m, "light") || strings.Contains(m, "lamp") {
		return "light"
	}
	if strings.Contains(m, "switch") {
		return "switch"
	}
	if strings.Contains(m, "plug") || strings.Contains(m, "outlet") {
		return "plug"
	}
	if strings.Contains(m, "speaker") || strings.Contains(m, "wifispeaker") {
		return "speaker"
	}
	return "device"
}

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

// DeviceStatus 处理 GET /api/devices/:id/status - 获取设备状态（开关、亮度、音量等）
func DeviceStatus(a *web.App, r *ghttp.Request) {
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
	c := ctrl.New(a.DeviceAPI())
	model := d.Model
	if model == "" {
		model = d.DID
	}
	resp := map[string]interface{}{"supported": []string{}}
	supported := []string{}

	if on, err := c.GetOn(d.DID, model); err == nil {
		resp["on"] = on
		supported = append(supported, "on")
	}
	if brightness, err := c.GetBrightness(d.DID, model); err == nil {
		resp["brightness"] = brightness
		supported = append(supported, "brightness")
	}
	if volume, err := c.GetVolume(d.DID, model); err == nil {
		resp["volume"] = volume
		supported = append(supported, "volume")
	}
	if mute, err := c.GetMute(d.DID, model); err == nil {
		resp["mute"] = mute
		supported = append(supported, "mute")
	}
	if occ, err := c.GetOccupancy(d.DID, model); err == nil {
		resp["occupancy"] = occ
		supported = append(supported, "occupancy")
	}
	resp["supported"] = supported

	JSON(r, http.StatusOK, resp)
}

// DeviceSetStatus 处理 POST /api/devices/:id/status - 设置设备状态
func DeviceSetStatus(a *web.App, r *ghttp.Request) {
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
	var body struct {
		On         *bool `json:"on"`
		Brightness *int  `json:"brightness"`
		Volume     *int  `json:"volume"`
		Mute       *bool `json:"mute"`
	}
	if err := json.NewDecoder(r.Request.Body).Decode(&body); err != nil {
		Err(r, http.StatusBadRequest, i18n.T(lang, "web.api.invalid_json", nil))
		return
	}
	c := ctrl.New(a.DeviceAPI())
	model := d.Model
	if model == "" {
		model = d.DID
	}

	if body.On != nil {
		if err := c.SetOn(d.DID, model, *body.On); err != nil {
			Err(r, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if body.Brightness != nil {
		if err := c.SetBrightness(d.DID, model, *body.Brightness); err != nil {
			Err(r, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if body.Volume != nil {
		if err := c.SetVolume(d.DID, model, *body.Volume); err != nil {
			Err(r, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if body.Mute != nil {
		if err := c.SetMute(d.DID, model, *body.Mute); err != nil {
			Err(r, http.StatusInternalServerError, err.Error())
			return
		}
	}
	JSON(r, http.StatusOK, map[string]string{"status": "ok"})
}

// DeviceControlPatch 处理 PATCH /api/devices/:id/control - 控制设备（action: toggle/set_brightness/set_on 等）
func DeviceControlPatch(a *web.App, r *ghttp.Request) {
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
	var body struct {
		Action string      `json:"action"`
		Value  interface{} `json:"value"`
	}
	if err := json.NewDecoder(r.Request.Body).Decode(&body); err != nil {
		Err(r, http.StatusBadRequest, i18n.T(lang, "web.api.invalid_json", nil))
		return
	}
	c := ctrl.New(a.DeviceAPI())
	model := d.Model
	if model == "" {
		model = d.DID
	}

	switch body.Action {
	case "toggle":
		if err := c.Toggle(d.DID, model); err != nil {
			Err(r, http.StatusInternalServerError, err.Error())
			return
		}
	case "set_on":
		on, ok := body.Value.(bool)
		if !ok {
			Err(r, http.StatusBadRequest, "value must be boolean for set_on")
			return
		}
		if err := c.SetOn(d.DID, model, on); err != nil {
			Err(r, http.StatusInternalServerError, err.Error())
			return
		}
	case "set_brightness":
		var v int
		switch x := body.Value.(type) {
		case float64:
			v = int(x)
		case int:
			v = x
		default:
			Err(r, http.StatusBadRequest, "value must be number for set_brightness")
			return
		}
		if v < 0 {
			v = 0
		}
		if v > 100 {
			v = 100
		}
		if err := c.SetBrightness(d.DID, model, v); err != nil {
			Err(r, http.StatusInternalServerError, err.Error())
			return
		}
	case "set_volume":
		var v int
		switch x := body.Value.(type) {
		case float64:
			v = int(x)
		case int:
			v = x
		default:
			Err(r, http.StatusBadRequest, "value must be number for set_volume")
			return
		}
		if v < 0 {
			v = 0
		}
		if v > 100 {
			v = 100
		}
		if err := c.SetVolume(d.DID, model, v); err != nil {
			Err(r, http.StatusInternalServerError, err.Error())
			return
		}
	default:
		Err(r, http.StatusBadRequest, "unknown action: "+body.Action)
		return
	}
	JSON(r, http.StatusOK, map[string]string{"status": "ok"})
}
