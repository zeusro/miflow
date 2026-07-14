package nlu

import (
	"fmt"
	"strings"

	"github.com/zeusro/miflow/internal/device"
)

// Resolver 将 Intent 匹配到实际设备。
type Resolver struct {
	devices []*device.Device
	rooms   []*device.HomeWithRooms
}

// NewResolver 创建解析器。
func NewResolver(devices []*device.Device, rooms []*device.HomeWithRooms) *Resolver {
	return &Resolver{devices: devices, rooms: rooms}
}

// Resolve 根据 Intent 匹配设备。
// 返回匹配到的设备列表与是否模糊（多个设备）。
// 当 intent 为 list_devices、unknown 或不需要设备时返回空列表。
func (r *Resolver) Resolve(intent *Intent) ([]*device.Device, bool, error) {
	if intent == nil {
		return nil, false, fmt.Errorf("intent is nil")
	}
	if intent.Action == "list_devices" || intent.Action == "unknown" {
		return nil, false, nil
	}

	candidates := r.filter(intent)
	if len(candidates) == 0 {
		return nil, false, fmt.Errorf("未找到匹配设备")
	}
	if len(candidates) == 1 {
		return candidates, false, nil
	}
	// 多个候选时返回全部，由上层决定是否全部执行或报错。
	return candidates, true, nil
}

func (r *Resolver) filter(intent *Intent) []*device.Device {
	var out []*device.Device
	for _, d := range r.devices {
		if d == nil {
			continue
		}
		if !r.matchesRoom(d, intent.Room) {
			continue
		}
		if intent.DeviceName != "" && !matchName(d.Name, intent.DeviceName) {
			continue
		}
		if intent.DeviceType != "" && inferType(d.Model) != intent.DeviceType {
			continue
		}
		out = append(out, d)
	}
	return out
}

func (r *Resolver) matchesRoom(d *device.Device, room string) bool {
	if room == "" {
		return true
	}
	for _, home := range r.rooms {
		for _, rm := range home.Rooms {
			if !matchName(rm.RoomName, room) {
				continue
			}
			for _, dev := range rm.Devices {
				if dev.DID == d.DID {
					return true
				}
			}
		}
	}
	return false
}

func matchName(name, keyword string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return true
	}
	return strings.Contains(name, keyword)
}
