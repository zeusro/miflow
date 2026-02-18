package ctrl

// Spec 为型号的 MIoT siid/piid/aiid 常量，从 home.miot-spec.com 规格提取。
type Spec struct {
	// Switch/Outlet 开关
	SiidSwitch int
	PiidOn     int
	AiidToggle int
	// Light 灯光（部分型号）
	SiidLight      int
	PiidBrightness int
	// Speaker 音箱
	SiidVoiceAssistant int
	AiidExecuteText    int
	SiidSpeaker        int
	PiidVolume         int
	PiidMute           int
	SiidPlayControl    int
	AiidPlay           int
	AiidPause          int
	AiidNext           int
	AiidPrevious       int
	// TV
	SiidTV      int
	AiidTurnOff int
	// Occupancy  occupancy sensor
	SiidOccupancy int
	PiidStatus    int
	// 多通道开关的 siid 列表（如 lemesh.switch.sw3f13 左中右）
	SwitchChannels []int
}

// Specs 为各型号的规格常量。
var Specs = map[string]Spec{
	// Switch
	"bean.switch.bln31": {
		SiidSwitch: 2, PiidOn: 1, AiidToggle: 1,
	},
	"bean.switch.bln33": {
		SiidSwitch: 2, PiidOn: 1, AiidToggle: 1,
	},
	"lemesh.switch.sw3f13": {
		SiidSwitch: 2, PiidOn: 1, AiidToggle: 1,
		SwitchChannels: []int{2, 3, 4},
	},
	// Plug/Outlet
	"chuangmi.plug.m3": {
		SiidSwitch: 2, PiidOn: 1,
	},
	"chuangmi.plug.v3": {
		SiidSwitch: 2, PiidOn: 1,
	},
	"babai.plug.sk01a": {
		SiidSwitch: 2, PiidOn: 1,
	},
	// Light
	"opple.light.bydceiling": {
		SiidLight: 2, PiidOn: 1, PiidBrightness: 3,
	},
	"giot.light.v5ssm": {
		SiidLight: 2, PiidOn: 1, PiidBrightness: 2,
	},
	// Speaker
	"xiaomi.wifispeaker.oh2": {
		SiidVoiceAssistant: 5, AiidExecuteText: 1,
		SiidSpeaker: 2, PiidVolume: 1, PiidMute: 2,
		SiidPlayControl: 3, AiidPlay: 2, AiidPause: 3, AiidNext: 6, AiidPrevious: 5,
	},
	"xiaomi.wifispeaker.l05b": {
		SiidVoiceAssistant: 5, AiidExecuteText: 1,
		SiidSpeaker: 2, PiidVolume: 1, PiidMute: 2,
		SiidPlayControl: 3, AiidPlay: 2, AiidPause: 3, AiidNext: 6, AiidPrevious: 5,
	},
	"xiaomi.wifispeaker.l05c": {
		SiidVoiceAssistant: 5, AiidExecuteText: 1,
		SiidSpeaker: 2, PiidVolume: 1, PiidMute: 2,
		SiidPlayControl: 3, AiidPlay: 2, AiidPause: 3, AiidNext: 6, AiidPrevious: 5,
	},
	// TV
	"xiaomi.tv.eanfv1": {
		SiidTV: 2, AiidTurnOff: 1,
	},
	// Occupancy Sensor
	"linp.sensor_occupy.hb01": {
		SiidOccupancy: 2, PiidStatus: 1,
	},
}
