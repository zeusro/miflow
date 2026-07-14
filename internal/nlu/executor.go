package nlu

import (
	"fmt"

	"github.com/zeusro/miflow/internal/device"
	"github.com/zeusro/miflow/miiot/ctrl"
)

// Executor 负责将 Intent 转换为实际的设备控制调用。
type Executor struct {
	api *device.API
}

// NewExecutor 创建执行器。
func NewExecutor(api *device.API) *Executor {
	return &Executor{api: api}
}

// Execute 对单个设备执行意图，返回结果或错误。
func (e *Executor) Execute(d *device.Device, intent *Intent) (interface{}, error) {
	if intent == nil || d == nil {
		return nil, fmt.Errorf("executor: invalid arguments")
	}
	c := ctrl.New(e.api)
	model := d.Model
	if model == "" {
		model = d.DID
	}

	switch intent.Action {
	case "turn_on":
		return nil, c.SetOn(d.DID, model, true)
	case "turn_off":
		return nil, c.SetOn(d.DID, model, false)
	case "toggle":
		return nil, c.Toggle(d.DID, model)
	case "set_brightness":
		v, ok := toInt(intent.Value)
		if !ok {
			return nil, fmt.Errorf("executor: brightness value must be integer")
		}
		return nil, c.SetBrightness(d.DID, model, v)
	case "set_volume":
		v, ok := toInt(intent.Value)
		if !ok {
			return nil, fmt.Errorf("executor: volume value must be integer")
		}
		return nil, c.SetVolume(d.DID, model, v)
	case "set_mute":
		v, ok := toBool(intent.Value)
		if !ok {
			return nil, fmt.Errorf("executor: mute value must be boolean")
		}
		return nil, c.SetMute(d.DID, model, v)
	case "set_channel":
		if intent.Channel == nil {
			return nil, fmt.Errorf("executor: channel required")
		}
		v, ok := toBool(intent.Value)
		if !ok {
			return nil, fmt.Errorf("executor: channel value must be boolean")
		}
		return nil, c.SetSwitchChannel(d.DID, model, *intent.Channel, v)
	case "tts":
		if intent.Text == "" {
			return nil, fmt.Errorf("executor: tts text required")
		}
		return nil, c.TTS(d.DID, model, intent.Text)
	case "play":
		return nil, c.Play(d.DID, model)
	case "pause":
		return nil, c.Pause(d.DID, model)
	case "next":
		return nil, c.Next(d.DID, model)
	case "previous":
		return nil, c.Previous(d.DID, model)
	case "query_status":
		return e.queryStatus(d, model)
	default:
		return nil, fmt.Errorf("executor: unsupported action %s", intent.Action)
	}
}

func (e *Executor) queryStatus(d *device.Device, model string) (map[string]interface{}, error) {
	c := ctrl.New(e.api)
	out := map[string]interface{}{}
	if on, err := c.GetOn(d.DID, model); err == nil {
		out["on"] = on
	}
	if brightness, err := c.GetBrightness(d.DID, model); err == nil {
		out["brightness"] = brightness
	}
	if volume, err := c.GetVolume(d.DID, model); err == nil {
		out["volume"] = volume
	}
	if mute, err := c.GetMute(d.DID, model); err == nil {
		out["mute"] = mute
	}
	if occ, err := c.GetOccupancy(d.DID, model); err == nil {
		out["occupancy"] = occ
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("executor: no readable status")
	}
	return out, nil
}

func toInt(v interface{}) (int, bool) {
	switch x := v.(type) {
	case float64:
		return int(x), true
	case int:
		return x, true
	case int64:
		return int(x), true
	case string:
		var n int
		_, err := fmt.Sscanf(x, "%d", &n)
		return n, err == nil
	}
	return 0, false
}

func toBool(v interface{}) (bool, bool) {
	switch x := v.(type) {
	case bool:
		return x, true
	case string:
		switch x {
		case "true", "1", "on", "开":
			return true, true
		case "false", "0", "off", "关":
			return false, true
		}
	case float64:
		return x != 0, true
	case int:
		return x != 0, true
	}
	return false, false
}
