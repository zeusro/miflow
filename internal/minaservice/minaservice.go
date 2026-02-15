package minaservice

import (
	"fmt"
	"strings"

	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/miioservice"
	"github.com/zeusro/miflow/internal/minaapi"
)

// TTS uses intelligent-speaker service (siid=5).
// - play-text (aiid=3): TTS 播报，in=[text]
// - execute-text-directive (aiid=4): 语音指令，in=[text, silent]，部分机型 aiid=5
// Ref: MIoT spec xiaomi.wifispeaker.*, ha_xiaomi_home issue #57
const (
	TTSsiid       = 5
	TTSaiidPlay   = 3 // play-text: 纯 TTS 播报
	TTSaiidDirect = 4 // execute-text-directive: 部分机型为 5
)

// Service implements MiNA (speaker) control via MIoT actions (OAuth mode).
// TTS uses "Execute Text Directive" action; play/pause require device-specific MIoT actions.
// PlayByURL uses MinaAPI (api2.mina.mi.com) when available.
type Service struct {
	MiIO     *miioservice.Service
	MinaAPI  *minaapi.Client
}

// New creates MiNA service backed by MiIO (OAuth).
func New(miio *miioservice.Service) *Service {
	return &Service{MiIO: miio}
}

// NewWithMinaAPI creates service with MinaAPI for play_by_url (api2.mina.mi.com).
func NewWithMinaAPI(miio *miioservice.Service, token *miaccount.OAuthToken, tokenPath string) *Service {
	s := &Service{MiIO: miio}
	if token != nil && token.IsValid() {
		s.MinaAPI = minaapi.New(token, tokenPath)
	}
	return s
}

// DeviceList returns speaker devices from MiIO device list.
func (s *Service) DeviceList(master int) ([]map[string]interface{}, error) {
	list, err := s.MiIO.DeviceList("", false, 0)
	if err != nil {
		return nil, err
	}
	// Filter speakers (model contains wifispeaker, speaker, etc.)
	out := make([]map[string]interface{}, 0)
	for _, d := range list {
		model, _ := d["model"].(string)
		if isSpeaker(model) {
			out = append(out, d)
		}
	}
	return out, nil
}

func isSpeaker(model string) bool {
	return strings.Contains(strings.ToLower(model), "speaker") ||
		strings.Contains(strings.ToLower(model), "wifispeaker") ||
		strings.Contains(strings.ToLower(model), "soundbar")
}

// TextToSpeech sends TTS via MIoT play-text or execute-text-directive action.
// play-text (aiid=3) 仅需 [text]；execute-text-directive 需 [text]，格式错误会导致不播放。
// Ref: ha_xiaomi_home issue #57 - 正确格式为 ["文本"]，不能多传 silent 等参数。
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
	return nil, fmt.Errorf("TTS failed: %w", lastErr)
}

// GetMinaDeviceID returns device ID for the given MI_DID (did or name).
// When MinaAPI is available, uses Mina device list (deviceID) for play compatibility.
func (s *Service) GetMinaDeviceID(miDID string) (string, error) {
	if s.MinaAPI != nil {
		// Mina API expects deviceID from its own device list
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
	// Fallback: ha device list (did)
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
	return "", fmt.Errorf("device not found: %s (use 'm mina' to list)", miDID)
}

// PlayerStop: OAuth mode uses MIoT. Many speakers have play_control action; siid/aiid vary by model.
func (s *Service) PlayerStop(deviceID string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("player_stop: use MIoT action for your speaker (m spec <model>)")
}

// PlayerSetVolume: not implemented in OAuth mode.
func (s *Service) PlayerSetVolume(deviceID string, volume int) (map[string]interface{}, error) {
	return nil, fmt.Errorf("player_set_volume: use MIoT prop for your speaker")
}

// PlayerPause: not implemented in OAuth mode.
func (s *Service) PlayerPause(deviceID string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("player_pause: use MIoT action for your speaker")
}

// PlayerPlay: not implemented in OAuth mode.
func (s *Service) PlayerPlay(deviceID string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("player_play: use MIoT action for your speaker")
}

// PlayerSetLoop: not implemented in OAuth mode.
func (s *Service) PlayerSetLoop(deviceID string, loopType int) (map[string]interface{}, error) {
	return nil, fmt.Errorf("player_set_loop: use MIoT action for your speaker")
}

// PlayByURL plays audio. Uses MinaAPI (api2.mina.mi.com) when available.
// Ref: https://github.com/hanxi/xiaomusic, MiService minaservice.play_by_url
func (s *Service) PlayByURL(deviceID, url string, _type int) (map[string]interface{}, error) {
	if s.MinaAPI != nil {
		return s.MinaAPI.PlayByURL(deviceID, url, _type)
	}
	return nil, fmt.Errorf("play_url: MinaAPI not configured (use NewWithMinaAPI with OAuth token)")
}

// FindDeviceIDByMiotDID: for OAuth, did is the primary identifier.
func FindDeviceIDByMiotDID(devices []map[string]interface{}, miotDID string) (string, error) {
	for _, d := range devices {
		if did, _ := d["did"].(string); did == miotDID {
			return did, nil
		}
	}
	return "", fmt.Errorf("device not found for miot DID %s", miotDID)
}
