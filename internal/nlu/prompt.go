package nlu

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zeusro/miflow/internal/device"
)

// DefaultSystemPrompt 返回默认的系统提示词，包含设备上下文与 JSON schema 说明。
func DefaultSystemPrompt(devices []*device.Device, rooms []*device.HomeWithRooms) string {
	var b strings.Builder
	b.WriteString("你是 miflow 智能家居助手，擅长把用户的自然语言指令解析为结构化的设备控制意图。\n")
	b.WriteString("你只能使用下面列出的设备、房间和动作，输出必须是合法的 JSON，不要包含任何其他内容。\n\n")

	b.WriteString("可用设备（按房间分组）：\n")
	written := false
	for _, home := range rooms {
		for _, room := range home.Rooms {
			if len(room.Devices) == 0 {
				continue
			}
			b.WriteString(fmt.Sprintf("房间 [%s]：\n", room.RoomName))
			for _, d := range room.Devices {
				b.WriteString(fmt.Sprintf("  - 名称: %s, DID: %s, 型号: %s, 类型: %s\n",
					d.Name, d.DID, d.Model, inferType(d.Model)))
				written = true
			}
		}
	}
	if !written {
		for _, d := range devices {
			b.WriteString(fmt.Sprintf("  - 名称: %s, DID: %s, 型号: %s, 类型: %s\n",
				d.Name, d.DID, d.Model, inferType(d.Model)))
		}
	}

	b.WriteString("\n支持的 action 值：\n")
	b.WriteString("  turn_on         打开设备\n")
	b.WriteString("  turn_off        关闭设备\n")
	b.WriteString("  toggle          切换开关状态\n")
	b.WriteString("  set_brightness  设置灯光亮度，value 为 0-100 的整数\n")
	b.WriteString("  set_volume      设置音箱音量，value 为 0-100 的整数\n")
	b.WriteString("  set_mute        设置静音，value 为 true/false\n")
	b.WriteString("  set_channel     设置多通道开关，value 为 true/false，channel 为通道索引（0 开始）\n")
	b.WriteString("  tts             让音箱播报文本，text 为要播报的内容\n")
	b.WriteString("  play            播放，音箱专用\n")
	b.WriteString("  pause           暂停播放，音箱专用\n")
	b.WriteString("  next            下一曲\n")
	b.WriteString("  previous        上一曲\n")
	b.WriteString("  query_status    查询设备状态\n")
	b.WriteString("  list_devices    列出设备\n")
	b.WriteString("  unknown         无法理解或不支持的指令\n")

	b.WriteString("\n输出 JSON 格式：\n")
	example := &Intent{
		Action:     "turn_on",
		DeviceName: "客厅灯",
		Room:       "客厅",
		DeviceType: "light",
		Reasoning:  "用户要求打开客厅的灯",
	}
	data, _ := json.MarshalIndent(example, "", "  ")
	b.Write(data)
	b.WriteString("\n\n规则：\n")
	b.WriteString("1. device_name 应尽量匹配设备名称中的关键词。\n")
	b.WriteString("2. 若用户指定了房间，请填写 room。\n")
	b.WriteString("3. 若用户只说类型（如“打开所有灯”），device_type 填 light，device_name 可为空。\n")
	b.WriteString("4. value 用于 set_brightness、set_volume、set_mute、set_channel。\n")
	b.WriteString("5. channel 用于多通道开关，从 0 开始计数。\n")
	b.WriteString("6. text 用于 tts 动作。\n")
	b.WriteString("7. 如果指令不明确或没有匹配设备，action 使用 unknown，并在 reasoning 中说明。\n")

	return b.String()
}

func inferType(model string) string {
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
	if strings.Contains(m, "tv") {
		return "tv"
	}
	if strings.Contains(m, "sensor") {
		return "sensor"
	}
	return "device"
}
