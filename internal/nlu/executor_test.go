package nlu

import (
	"testing"

	"github.com/zeusro/miflow/internal/device"
)

func TestExecutorInvalidInputs(t *testing.T) {
	// 使用 nil API 仅验证参数检查分支，不会真正调用控制接口。
	e := NewExecutor(nil)

	cases := []struct {
		name   string
		intent *Intent
		dev    *device.Device
	}{
		{"nil intent", nil, &device.Device{DID: "1", Model: "opple.light.bydceiling"}},
		{"nil device", &Intent{Action: "turn_on"}, nil},
		{"brightness no value", &Intent{Action: "set_brightness"}, &device.Device{DID: "1", Model: "opple.light.bydceiling"}},
		{"volume no value", &Intent{Action: "set_volume"}, &device.Device{DID: "1", Model: "xiaomi.wifispeaker.l05c"}},
		{"mute no value", &Intent{Action: "set_mute"}, &device.Device{DID: "1", Model: "xiaomi.wifispeaker.l05c"}},
		{"channel no index", &Intent{Action: "set_channel", Value: true}, &device.Device{DID: "1", Model: "lemesh.switch.sw3f13"}},
		{"tts no text", &Intent{Action: "tts"}, &device.Device{DID: "1", Model: "xiaomi.wifispeaker.l05c"}},
		{"unsupported action", &Intent{Action: "fly"}, &device.Device{DID: "1", Model: "opple.light.bydceiling"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := e.Execute(c.dev, c.intent)
			if err == nil {
				t.Error("expected error for invalid input")
			}
		})
	}
}

func TestToInt(t *testing.T) {
	cases := []struct {
		v    interface{}
		want int
		ok   bool
	}{
		{float64(50), 50, true},
		{int(30), 30, true},
		{"80", 80, true},
		{"abc", 0, false},
		{true, 0, false},
	}
	for _, c := range cases {
		got, ok := toInt(c.v)
		if ok != c.ok {
			t.Errorf("toInt(%v) ok = %v, want %v", c.v, ok, c.ok)
		}
		if ok && got != c.want {
			t.Errorf("toInt(%v) = %d, want %d", c.v, got, c.want)
		}
	}
}

func TestToBool(t *testing.T) {
	cases := []struct {
		v    interface{}
		want bool
		ok   bool
	}{
		{true, true, true},
		{false, false, true},
		{"true", true, true},
		{"on", true, true},
		{"开", true, true},
		{"0", false, true},
		{"off", false, true},
		{"maybe", false, false},
		{float64(1), true, true},
		{float64(0), false, true},
	}
	for _, c := range cases {
		got, ok := toBool(c.v)
		if ok != c.ok {
			t.Errorf("toBool(%v) ok = %v, want %v", c.v, ok, c.ok)
		}
		if ok && got != c.want {
			t.Errorf("toBool(%v) = %v, want %v", c.v, got, c.want)
		}
	}
}
