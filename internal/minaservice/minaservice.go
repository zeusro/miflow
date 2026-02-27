package minaservice

import (
	"fmt"
	"strings"

	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/miioservice"
	"github.com/zeusro/miflow/internal/minaapi"
	"github.com/zeusro/miflow/pkg/i18n"
)

// TTS 使用智能音箱服务 (siid=5)。
// - play-text (aiid=3): TTS 播报，in=[text]
// - execute-text-directive (aiid=4): 语音指令，in=[text, silent]，部分机型 aiid=5
// 参考: MIoT spec xiaomi.wifispeaker.*, ha_xiaomi_home issue #57
const (
	TTSsiid       = 5
	TTSaiidPlay   = 3 // play-text: 纯 TTS 播报
	TTSaiidDirect = 4 // execute-text-directive: 部分机型为 5
)

// Service 通过 MIoT 动作实现 MiNA（音箱）控制（OAuth 模式）。
// TTS 使用 "Execute Text Directive" 动作；play/pause 需要设备特定的 MIoT 动作。
// PlayByURL 在可用时使用 MinaAPI (api2.mina.mi.com)。
type Service struct {
	MiIO     *miioservice.Service
	MinaAPI  *minaapi.Client
}

// New 创建基于 MiIO 的 MiNA 服务（OAuth）。
func New(miio *miioservice.Service) *Service {
	return &Service{MiIO: miio}
}

// NewWithMinaAPI 创建带 MinaAPI 的服务，用于 play_by_url (api2.mina.mi.com)。
func NewWithMinaAPI(miio *miioservice.Service, token *miaccount.OAuthToken, tokenPath string) *Service {
	s := &Service{MiIO: miio}
	if token != nil && token.IsValid() {
		s.MinaAPI = minaapi.New(token, tokenPath)
	}
	return s
}

// DeviceList 从 MiIO 设备列表返回音箱设备。
func (s *Service) DeviceList(master int) ([]map[string]interface{}, error) {
	list, err := s.MiIO.DeviceList("", false, 0)
	if err != nil {
		return nil, err
	}
	// 过滤音箱（model 包含 wifispeaker、speaker 等）
	out := make([]map[string]interface{}, 0)
	for _, d := range list {
		model, _ := d["model"].(string)
		if isSpeaker(model) {
			out = append(out, d)
		}
	}
	return out, nil
}

// isSpeaker 判断型号是否为音箱类设备。
func isSpeaker(model string) bool {
	return strings.Contains(strings.ToLower(model), "speaker") ||
		strings.Contains(strings.ToLower(model), "wifispeaker") ||
		strings.Contains(strings.ToLower(model), "soundbar")
}

// TextToSpeech 通过 MIoT play-text 或 execute-text-directive 动作发送 TTS。
// play-text (aiid=3) 仅需 [text]；execute-text-directive 需 [text]，格式错误会导致不播放。
// 参考: ha_xiaomi_home issue #57 - 正确格式为 ["文本"]，不能多传 silent 等参数。
func (s *Service) TextToSpeech(did string, text string) (map[string]interface{}, error) {
	args := []interface{}{text}
	var lastErr error
	for _, aiid := range []int{TTSaiidPlay, TTSaiidDirect, 5} {
		_, err := s.MiIO.MiotAction(did, TTSsiid, aiid, args)
		if err == nil {
			return map[string]interface{}{"code": 0}, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "minasvc.tts_failed", map[string]interface{}{"Err": lastErr}))
}

// GetMinaDeviceID 返回给定 MI_DID（did 或 name）对应的设备 ID。
// 当 MinaAPI 可用时，使用 Mina 设备列表 (deviceID) 以兼容播放。
func (s *Service) GetMinaDeviceID(miDID string) (string, error) {
	if s.MinaAPI != nil {
		// Mina API 期望使用其自身设备列表中的 deviceID
		devices, err := s.MinaAPI.DeviceList(0)
		if err == nil {
			for _, d := range devices {
				deviceID, _ := d["deviceID"].(string)
				name, _ := d["name"].(string)
				did, _ := d["did"].(string)
				if deviceID == miDID || did == miDID || strings.Contains(name, miDID) {
					return deviceID, nil
				}
			}
		}
	}
	// 回退：ha 设备列表 (did)
	list, err := s.DeviceList(0)
	if err != nil {
		return "", err
	}
	for _, d := range list {
		did, _ := d["did"].(string)
		name, _ := d["name"].(string)
		if did == miDID || strings.Contains(name, miDID) {
			return did, nil
		}
	}
	return "", fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "minasvc.device_not_found", map[string]interface{}{"Did": miDID}))
}

// PlayerStop: OAuth 模式使用 MIoT。多数音箱有 play_control 动作；siid/aiid 因型号而异。
func (s *Service) PlayerStop(deviceID string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "minasvc.player_stop", nil))
}

// PlayerSetVolume: OAuth 模式未实现。
func (s *Service) PlayerSetVolume(deviceID string, volume int) (map[string]interface{}, error) {
	return nil, fmt.Errorf("player_set_volume: use MIoT prop for your speaker")
}

// PlayerPause: OAuth 模式未实现。
func (s *Service) PlayerPause(deviceID string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("player_pause: use MIoT action for your speaker")
}

// PlayerPlay: OAuth 模式未实现。
func (s *Service) PlayerPlay(deviceID string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("player_play: use MIoT action for your speaker")
}

// PlayerSetLoop: OAuth 模式未实现。
func (s *Service) PlayerSetLoop(deviceID string, loopType int) (map[string]interface{}, error) {
	return nil, fmt.Errorf("player_set_loop: use MIoT action for your speaker")
}

// PlayByURL 播放音频。可用时使用 MinaAPI (api2.mina.mi.com)。
// 参考: https://github.com/hanxi/xiaomusic, MiService minaservice.play_by_url
func (s *Service) PlayByURL(deviceID, url string, _type int) (map[string]interface{}, error) {
	if s.MinaAPI != nil {
		return s.MinaAPI.PlayByURL(deviceID, url, _type)
	}
	return nil, fmt.Errorf("play_url: MinaAPI not configured (use NewWithMinaAPI with OAuth token)")
}

// FindDeviceIDByMiotDID: OAuth 模式下，did 为主要标识符。
func FindDeviceIDByMiotDID(devices []map[string]interface{}, miotDID string) (string, error) {
	for _, d := range devices {
		if did, _ := d["did"].(string); did == miotDID {
			return did, nil
		}
	}
	return "", fmt.Errorf("device not found for miot DID %s", miotDID)
}
