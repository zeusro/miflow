package minaservice

import (
	"fmt"
	"strings"

	"github.com/zeusro/miflow/internal/miioservice"
)

// Execute Text Directive: siid=5, aiid=5 for xiaomi.wifispeaker.* (text, silent).
// Ref: https://github.com/XiaoMi/ha_xiaomi_home
const (
	TTSsiid = 5
	TTSaiid = 5
)

// Service implements MiNA (speaker) control via MIoT actions (OAuth mode).
// TTS uses "Execute Text Directive" action; play/pause require device-specific MIoT actions.
type Service struct {
	MiIO *miioservice.Service
}

// New creates MiNA service backed by MiIO (OAuth).
func New(miio *miioservice.Service) *Service {
	return &Service{MiIO: miio}
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

// TextToSpeech sends TTS via MIoT "Execute Text Directive" action.
func (s *Service) TextToSpeech(did string, text string) (map[string]interface{}, error) {
	_, err := s.MiIO.MiotAction(did, TTSsiid, TTSaiid, []interface{}{text, false})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"code": 0}, nil
}

// GetMinaDeviceID returns device did for the given MI_DID (did or name).
func (s *Service) GetMinaDeviceID(miDID string) (string, error) {
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

// PlayByURL plays audio. OAuth mode: try MIoT action if speaker supports it.
func (s *Service) PlayByURL(deviceID, url string, _type int) (map[string]interface{}, error) {
	// Some speakers have play_url action; siid/aiid vary. Try common pattern.
	// e.g. siid=5 aiid=2 or similar - check spec for your model
	return nil, fmt.Errorf("play_url: use 'm spec <speaker_model>' then MIoT action, or use TTS: m 5 播放 %s", url)
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
