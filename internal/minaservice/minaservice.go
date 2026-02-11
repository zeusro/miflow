package minaservice

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/zeusro/miflow/internal/miaccount"
)

const sidMicoAPI = "micoapi"
const minaBase = "https://api2.mina.mi.com"

var usePlayMusicAPI = map[string]bool{
	"LX04": true, "LX05": true, "L05B": true, "L05C": true, "L06": true, "L06A": true,
	"X08A": true, "X10A": true, "X08C": true, "X08E": true, "X8F": true, "X4B": true,
	"OH2": true, "OH2P": true, "X6A": true,
}

const minaUserAgent = "MiHome/6.0.103 (com.xiaomi.mihome; build:6.0.103.1; iOS 14.4.0) Alamofire/6.0.103 MICO/iOSApp/appStore/6.0.103"

// Service implements MiNA (micoapi) for speakers.
type Service struct {
	Account          *miaccount.Account
	DeviceToHardware map[string]string
}

// New creates MiNA service.
func New(account *miaccount.Account) *Service {
	return &Service{
		Account:          account,
		DeviceToHardware: make(map[string]string),
	}
}

func (s *Service) minaRequest(uri string, data map[string]string) (map[string]interface{}, error) {
	requestID := "app_ios_" + miaccount.RandString(30)
	if data != nil {
		data["requestId"] = requestID
	} else {
		uri += "&requestId=" + requestID
	}
	var body interface{}
	if data != nil {
		form := url.Values{}
		for k, v := range data {
			form.Set(k, v)
		}
		body = form.Encode()
	}
	oldUA := s.Account.UserAgent
	s.Account.UserAgent = minaUserAgent
	raw, err := s.Account.Request(sidMicoAPI, "POST", minaBase+uri, body, true)
	s.Account.UserAgent = oldUA
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if code, _ := out["code"].(float64); code != 0 {
		return nil, fmt.Errorf("mina api error: %v", out)
	}
	return out, nil
}

// DeviceList returns list of devices (master 0 or 1).
func (s *Service) DeviceList(master int) ([]map[string]interface{}, error) {
	result, err := s.minaRequest(fmt.Sprintf("/admin/v2/device_list?master=%d", master), nil)
	if err != nil {
		return nil, err
	}
	if data, ok := result["data"]; ok {
		if arr, ok := data.([]interface{}); ok {
			out := make([]map[string]interface{}, 0, len(arr))
			for _, it := range arr {
				if m, ok := it.(map[string]interface{}); ok {
					out = append(out, m)
				}
			}
			return out, nil
		}
	}
	return nil, nil
}

// UbusRequest calls remote ubus.
func (s *Service) UbusRequest(deviceID, method, path string, message interface{}) (map[string]interface{}, error) {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	return s.minaRequest("/remote/ubus", map[string]string{
		"deviceId": deviceID,
		"message":  string(msgBytes),
		"method":   method,
		"path":     path,
	})
}

// TextToSpeech sends TTS to device.
func (s *Service) TextToSpeech(deviceID, text string) (map[string]interface{}, error) {
	return s.UbusRequest(deviceID, "text_to_speech", "mibrain", map[string]string{"text": text})
}

// PlayerSetVolume sets volume (0-100).
func (s *Service) PlayerSetVolume(deviceID string, volume int) (map[string]interface{}, error) {
	return s.UbusRequest(deviceID, "player_set_volume", "mediaplayer", map[string]interface{}{
		"volume": volume,
		"media":  "app_ios",
	})
}

// PlayerPause pauses playback.
func (s *Service) PlayerPause(deviceID string) (map[string]interface{}, error) {
	return s.UbusRequest(deviceID, "player_play_operation", "mediaplayer", map[string]string{
		"action": "pause",
		"media":  "app_ios",
	})
}

// PlayerStop stops playback.
func (s *Service) PlayerStop(deviceID string) (map[string]interface{}, error) {
	return s.UbusRequest(deviceID, "player_play_operation", "mediaplayer", map[string]string{
		"action": "stop",
		"media":  "app_ios",
	})
}

// PlayerPlay resumes playback.
func (s *Service) PlayerPlay(deviceID string) (map[string]interface{}, error) {
	return s.UbusRequest(deviceID, "player_play_operation", "mediaplayer", map[string]string{
		"action": "play",
		"media":  "app_ios",
	})
}

// PlayerSetLoop sets loop type (0=single, 1=list, etc.).
func (s *Service) PlayerSetLoop(deviceID string, loopType int) (map[string]interface{}, error) {
	return s.UbusRequest(deviceID, "player_set_loop", "mediaplayer", map[string]interface{}{
		"media": "common",
		"type":  loopType,
	})
}

func (s *Service) initDevices() error {
	if len(s.DeviceToHardware) > 0 {
		return nil
	}
	list, err := s.DeviceList(0)
	if err != nil {
		return err
	}
	for _, h := range list {
		deviceID, _ := h["deviceID"].(string)
		hardware, _ := h["hardware"].(string)
		if deviceID != "" && hardware != "" {
			s.DeviceToHardware[deviceID] = hardware
		}
	}
	return nil
}

// PlayByURL plays audio from URL. _type: 1=music, 2=default.
func (s *Service) PlayByURL(deviceID, url string, _type int) (map[string]interface{}, error) {
	if err := s.initDevices(); err != nil {
		return nil, err
	}
	hardware := s.DeviceToHardware[deviceID]
	if usePlayMusicAPI[hardware] {
		return s.playByMusicURL(deviceID, url, _type)
	}
	return s.UbusRequest(deviceID, "player_play_url", "mediaplayer", map[string]interface{}{
		"url":   url,
		"type":  _type,
		"media": "app_ios",
	})
}

func (s *Service) playByMusicURL(deviceID, url string, _type int) (map[string]interface{}, error) {
	audioType := ""
	if _type == 1 {
		audioType = "MUSIC"
	}
	music := map[string]interface{}{
		"payload": map[string]interface{}{
			"audio_type": audioType,
			"audio_items": []map[string]interface{}{
				{
					"item_id": map[string]interface{}{
						"audio_id": "1582971365183456177",
						"cp": map[string]interface{}{
							"album_id":   "-1",
							"episode_index": 0,
							"id":         "355454500",
							"name":       "xiaowei",
						},
					},
					"stream": map[string]string{"url": url},
				},
			},
			"list_params": map[string]interface{}{
				"listId":         "-1",
				"loadmore_offset": 0,
				"origin":         "xiaowei",
				"type":           "MUSIC",
			},
		},
		"play_behavior": "REPLACE_ALL",
	}
	musicJSON, _ := json.Marshal(music)
	return s.UbusRequest(deviceID, "player_play_music", "mediaplayer", map[string]string{
		"startaudioid": "1582971365183456177",
		"music":        string(musicJSON),
	})
}

// FindDeviceIDByMiotDID finds mina deviceID from miot DID (from device_list).
func FindDeviceIDByMiotDID(hardwareList []map[string]interface{}, miotDID string) (string, error) {
	for _, h := range hardwareList {
		if d, _ := h["miotDID"].(string); d == miotDID {
			if id, _ := h["deviceID"].(string); id != "" {
				return id, nil
			}
		}
	}
	return "", fmt.Errorf("device not found for miot DID %s", miotDID)
}

// GetMinaDeviceID returns deviceID for the given MI_DID (which can be DID or name).
func (s *Service) GetMinaDeviceID(miDID string) (string, error) {
	list, err := s.DeviceList(0)
	if err != nil {
		return "", err
	}
	// miDID can be miotDID (number) or device name
	for _, h := range list {
		miotDID, _ := h["miotDID"].(string)
		deviceID, _ := h["deviceID"].(string)
		name, _ := h["name"].(string)
		if miotDID == miDID || deviceID == miDID || strings.Contains(name, miDID) {
			return deviceID, nil
		}
	}
	return "", fmt.Errorf("device not found: %s (use 'm mina' to list)", miDID)
}
