package nlu

import (
	"strings"
	"testing"

	"github.com/zeusro/miflow/internal/device"
)

func TestDefaultSystemPrompt(t *testing.T) {
	devs := []*device.Device{
		{DID: "1", Name: "客厅灯", Model: "opple.light.bydceiling"},
		{DID: "2", Name: "小爱音箱", Model: "xiaomi.wifispeaker.l05c"},
	}
	rooms := []*device.HomeWithRooms{
		{
			HomeID: "h1", HomeName: "家",
			Rooms: []*device.RoomWithDevices{
				{RoomID: "r1", RoomName: "客厅", Devices: devs},
			},
		},
	}

	p := DefaultSystemPrompt(devs, rooms)
	checks := []string{
		"客厅灯",
		"小爱音箱",
		"turn_on",
		"set_brightness",
		"输出 JSON 格式",
		"客厅",
	}
	for _, c := range checks {
		if !strings.Contains(p, c) {
			t.Errorf("prompt missing %q", c)
		}
	}
}

func TestInferType(t *testing.T) {
	cases := map[string]string{
		"opple.light.bydceiling":  "light",
		"bean.switch.bln31":       "switch",
		"chuangmi.plug.m3":        "plug",
		"xiaomi.wifispeaker.l05c": "speaker",
		"xiaomi.tv.eanfv1":        "tv",
		"linp.sensor_occupy.hb01": "sensor",
		"unknown.model.xyz":       "device",
	}
	for model, want := range cases {
		got := inferType(model)
		if got != want {
			t.Errorf("inferType(%q) = %q, want %q", model, got, want)
		}
	}
}
