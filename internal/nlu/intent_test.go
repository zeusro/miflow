package nlu

import (
	"testing"
)

func TestParseIntent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    Intent
	}{
		{
			name:    "plain json",
			content: `{"action":"turn_on","device_name":"客厅灯","room":"客厅","device_type":"light","reasoning":"打开客厅灯"}`,
			want:    Intent{Action: "turn_on", DeviceName: "客厅灯", Room: "客厅", DeviceType: "light", Reasoning: "打开客厅灯"},
		},
		{
			name:    "markdown json",
			content: "```json\n{\"action\":\"turn_off\",\"device_name\":\"卧室灯\"}\n```",
			want:    Intent{Action: "turn_off", DeviceName: "卧室灯"},
		},
		{
			name:    "normalize action",
			content: `{"action":"on","device_name":"灯"}`,
			want:    Intent{Action: "turn_on", DeviceName: "灯"},
		},
		{
			name:    "with value",
			content: `{"action":"set_brightness","device_name":"客厅灯","value":50}`,
			want:    Intent{Action: "set_brightness", DeviceName: "客厅灯", Value: float64(50)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIntent(tt.content)
			if err != nil {
				t.Fatalf("parseIntent error: %v", err)
			}
			if got.Action != tt.want.Action {
				t.Errorf("action = %q, want %q", got.Action, tt.want.Action)
			}
			if got.DeviceName != tt.want.DeviceName {
				t.Errorf("device_name = %q, want %q", got.DeviceName, tt.want.DeviceName)
			}
			if got.Room != tt.want.Room {
				t.Errorf("room = %q, want %q", got.Room, tt.want.Room)
			}
		})
	}
}

func TestNormalizeAction(t *testing.T) {
	cases := map[string]string{
		"on":          "turn_on",
		"turn_on":     "turn_on",
		"off":         "turn_off",
		"open":        "turn_on",
		"close":       "turn_off",
		"brightness":  "set_brightness",
		"volume":      "set_volume",
		"speak":       "tts",
		"status":      "query_status",
		"unknown":     "unknown",
		"unsupported": "unknown",
	}
	for in, want := range cases {
		got := normalizeAction(in)
		if got != want {
			t.Errorf("normalizeAction(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIntentIsControl(t *testing.T) {
	controlActions := []string{"turn_on", "turn_off", "toggle", "set_brightness", "set_volume", "set_mute", "set_channel", "tts", "play", "pause", "next", "previous"}
	for _, a := range controlActions {
		i := &Intent{Action: a}
		if !i.IsControl() {
			t.Errorf("IsControl(%q) = false, want true", a)
		}
	}
	query := &Intent{Action: "query_status"}
	if query.IsControl() {
		t.Error("query_status should not be control")
	}
}
