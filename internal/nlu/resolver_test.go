package nlu

import (
	"testing"

	"github.com/zeusro/miflow/internal/device"
)

func TestResolver(t *testing.T) {
	devs := []*device.Device{
		{DID: "1", Name: "客厅灯", Model: "opple.light.bydceiling"},
		{DID: "2", Name: "卧室灯", Model: "opple.light.bydceiling"},
		{DID: "3", Name: "客厅插座", Model: "chuangmi.plug.m3"},
		{DID: "4", Name: "小爱音箱", Model: "xiaomi.wifispeaker.l05c"},
	}
	rooms := []*device.HomeWithRooms{
		{
			HomeID: "h1", HomeName: "家",
			Rooms: []*device.RoomWithDevices{
				{RoomID: "r1", RoomName: "客厅", Devices: []*device.Device{devs[0], devs[2], devs[3]}},
				{RoomID: "r2", RoomName: "卧室", Devices: []*device.Device{devs[1]}},
			},
		},
	}

	r := NewResolver(devs, rooms)

	tests := []struct {
		name      string
		intent    *Intent
		wantCount int
		wantErr   bool
		ambig     bool
	}{
		{
			name:      "by exact name",
			intent:    &Intent{Action: "turn_on", DeviceName: "客厅灯"},
			wantCount: 1,
		},
		{
			name:      "by room and type",
			intent:    &Intent{Action: "turn_on", Room: "卧室", DeviceType: "light"},
			wantCount: 1,
		},
		{
			name:      "by type only multiple",
			intent:    &Intent{Action: "turn_on", DeviceType: "light"},
			wantCount: 2,
			ambig:     true,
		},
		{
			name:      "no match",
			intent:    &Intent{Action: "turn_on", DeviceName: "厨房灯"},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "list devices no resolve",
			intent:    &Intent{Action: "list_devices"},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ambig, err := r.Resolve(tt.intent)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(got) != tt.wantCount {
				t.Errorf("matched count = %d, want %d", len(got), tt.wantCount)
			}
			if ambig != tt.ambig {
				t.Errorf("ambiguous = %v, want %v", ambig, tt.ambig)
			}
		})
	}
}

func TestMatchName(t *testing.T) {
	cases := []struct {
		name, kw string
		want     bool
	}{
		{"客厅灯", "客厅", true},
		{"客厅灯", "灯", true},
		{"客厅灯", "卧室", false},
		{"", "", true},
	}
	for _, c := range cases {
		got := matchName(c.name, c.kw)
		if got != c.want {
			t.Errorf("matchName(%q,%q) = %v, want %v", c.name, c.kw, got, c.want)
		}
	}
}
