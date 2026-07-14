package nlu

// Intent 表示从自然语言解析出的设备控制意图。
type Intent struct {
	// Action 为控制动作，如 turn_on、turn_off、toggle、set_brightness、set_volume 等。
	Action string `json:"action"`
	// DeviceName 为用户提到的设备名称（可能为空）。
	DeviceName string `json:"device_name"`
	// Room 为用户提到的房间（可能为空）。
	Room string `json:"room"`
	// DeviceType 为用户提到的设备类型，如 light、switch、speaker、plug（可能为空）。
	DeviceType string `json:"device_type"`
	// Value 为动作对应的数值或布尔值，如亮度 50、音量 30、开关 true/false。
	Value interface{} `json:"value"`
	// Channel 为多通道开关的通道索引（从 0 开始），可能为空。
	Channel *int `json:"channel"`
	// Text 为 TTS 或播放等动作需要播报/传递的文本。
	Text string `json:"text"`
	// Reasoning 为模型的推理说明，便于用户理解。
	Reasoning string `json:"reasoning"`
}

// IsControl 判断意图是否为控制类动作（需要解析并执行）。
func (i *Intent) IsControl() bool {
	if i == nil {
		return false
	}
	switch i.Action {
	case "turn_on", "turn_off", "toggle", "set_brightness", "set_volume",
		"set_mute", "set_channel", "tts", "play", "pause", "next", "previous":
		return true
	}
	return false
}

// IsQuery 判断意图是否为查询类动作（不需要执行控制）。
func (i *Intent) IsQuery() bool {
	if i == nil {
		return false
	}
	switch i.Action {
	case "query_status", "list_devices":
		return true
	}
	return false
}
