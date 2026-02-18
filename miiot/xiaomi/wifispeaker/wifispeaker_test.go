package wifispeaker

import (
	"os"
	"testing"

	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/device"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/miioservice"
	"github.com/zeusro/miflow/miiot/ctrl"
)

var wifispeakerModels = []string{Model, ModelL05B, ModelL05C}

func setupAPIForModel(t *testing.T, model string) (*device.API, *device.Device) {
	t.Helper()
	tokenPath := config.Get().TokenPath
	if tokenPath == "" {
		tokenPath = os.ExpandEnv("$HOME/.mi.token")
	}
	store := &miaccount.TokenStore{Path: tokenPath}
	token := store.LoadOAuth()
	if token == nil || !token.IsValid() {
		t.Skip("no valid OAuth token, run 'm login' first")
		return nil, nil
	}
	ioSvc, err := miioservice.New(token, tokenPath)
	if err != nil {
		t.Fatalf("miioservice.New: %v", err)
		return nil, nil
	}
	api := device.NewAPI(ioSvc)
	devs, err := api.List("", false, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
		return nil, nil
	}
	for _, d := range devs {
		if d == nil {
			continue
		}
		if model == "" {
			for _, m := range wifispeakerModels {
				if d.Model == m {
					return api, d
				}
			}
		} else if d.Model == model {
			return api, d
		}
	}
	if model == "" {
		t.Skip("no wifispeaker device (oh2/l05b/l05c) in list")
	} else {
		t.Skipf("no %s device in list", model)
	}
	return nil, nil
}

func TestConstants(t *testing.T) {
	for _, model := range wifispeakerModels {
		s := ctrl.Specs[model]
		if s.SiidVoiceAssistant != 5 || s.AiidExecuteText != 1 {
			t.Errorf("%s TTS: siid=%d aiid=%d", model, s.SiidVoiceAssistant, s.AiidExecuteText)
		}
		if s.SiidSpeaker != 2 || s.PiidVolume != 1 {
			t.Errorf("%s speaker: siid=%d piid=%d", model, s.SiidSpeaker, s.PiidVolume)
		}
		if s.SiidPlayControl != 3 || s.AiidPlay != 2 || s.AiidPause != 3 {
			t.Errorf("%s play: siid=%d play=%d pause=%d", model, s.SiidPlayControl, s.AiidPlay, s.AiidPause)
		}
		if s.PiidMute != 2 || s.AiidNext != 6 || s.AiidPrevious != 5 {
			t.Errorf("%s mute/next/prev: piidMute=%d aiidNext=%d aiidPrev=%d", model, s.PiidMute, s.AiidNext, s.AiidPrevious)
		}
	}
}

func TestModelConstantsInSpecs(t *testing.T) {
	models := []string{Model, ModelL05B, ModelL05C}
	for _, m := range models {
		s, ok := ctrl.Specs[m]
		if !ok {
			t.Errorf("model %s not in ctrl.Specs", m)
			continue
		}
		if s.SiidSpeaker == 0 || s.PiidVolume == 0 {
			t.Errorf("model %s: speaker spec incomplete", m)
		}
	}
}

func TestUnsupportedModel(t *testing.T) {
	api, dev := setupAPIForModel(t, "")
	if api == nil || dev == nil {
		return
	}
	c := ctrl.New(api)
	unknown := "unknown.speaker.xyz"
	if _, err := c.GetVolume(dev.DID, unknown); err == nil {
		t.Error("expected error for unknown model GetVolume")
	}
	if _, err := c.GetMute(dev.DID, unknown); err == nil {
		t.Error("expected error for unknown model GetMute")
	}
	if err := c.SetVolume(dev.DID, unknown, 50); err == nil {
		t.Error("expected error for unknown model SetVolume")
	}
	if err := c.SetMute(dev.DID, unknown, true); err == nil {
		t.Error("expected error for unknown model SetMute")
	}
	if err := c.TTS(dev.DID, unknown, "test"); err == nil {
		t.Error("expected error for unknown model TTS")
	}
	if err := c.Play(dev.DID, unknown); err == nil {
		t.Error("expected error for unknown model Play")
	}
	if err := c.Pause(dev.DID, unknown); err == nil {
		t.Error("expected error for unknown model Pause")
	}
	if err := c.Next(dev.DID, unknown); err == nil {
		t.Error("expected error for unknown model Next")
	}
	if err := c.Previous(dev.DID, unknown); err == nil {
		t.Error("expected error for unknown model Previous")
	}
}

// 设备测试辅助函数，供 oh2_test.go / l05b_test.go / l05c_test.go 调用
func testGetVolume(t *testing.T, api *device.API, dev *device.Device) {
	c := ctrl.New(api)
	vol, err := c.GetVolume(dev.DID, dev.Model)
	if err != nil {
		t.Fatalf("GetVolume: %v", err)
	}
	if vol < 0 || vol > 100 {
		t.Errorf("volume %d out of range [0,100]", vol)
	}
	t.Logf("%s (%s) volume=%d", dev.Name, dev.Model, vol)
}

func testSetVolumeGetVolume(t *testing.T, api *device.API, dev *device.Device) {
	c := ctrl.New(api)
	orig, err := c.GetVolume(dev.DID, dev.Model)
	if err != nil {
		t.Fatalf("GetVolume: %v", err)
	}
	if err := c.SetVolume(dev.DID, dev.Model, 50); err != nil {
		t.Fatalf("SetVolume: %v", err)
	}
	got, err := c.GetVolume(dev.DID, dev.Model)
	if err != nil {
		t.Fatalf("GetVolume after SetVolume: %v", err)
	}
	if got != 50 {
		t.Logf("SetVolume(50) then GetVolume: got %d (device may have delay)", got)
	}
	_ = c.SetVolume(dev.DID, dev.Model, orig)
}

func testSetVolumeBoundary(t *testing.T, api *device.API, dev *device.Device) {
	c := ctrl.New(api)
	orig, err := c.GetVolume(dev.DID, dev.Model)
	if err != nil {
		t.Fatalf("GetVolume: %v", err)
	}
	defer func() { _ = c.SetVolume(dev.DID, dev.Model, orig) }()
	for _, vol := range []int{0, 100} {
		if err := c.SetVolume(dev.DID, dev.Model, vol); err != nil {
			t.Fatalf("SetVolume(%d): %v", vol, err)
		}
		got, err := c.GetVolume(dev.DID, dev.Model)
		if err != nil {
			t.Fatalf("GetVolume after SetVolume(%d): %v", vol, err)
		}
		if got != vol {
			t.Logf("SetVolume(%d) then GetVolume: got %d (device may have delay)", vol, got)
		}
	}
}

func testGetMute(t *testing.T, api *device.API, dev *device.Device) {
	c := ctrl.New(api)
	mute, err := c.GetMute(dev.DID, dev.Model)
	if err != nil {
		t.Fatalf("GetMute: %v", err)
	}
	t.Logf("%s (%s) mute=%v", dev.Name, dev.Model, mute)
}

func testSetMuteGetMute(t *testing.T, api *device.API, dev *device.Device) {
	c := ctrl.New(api)
	orig, err := c.GetMute(dev.DID, dev.Model)
	if err != nil {
		t.Fatalf("GetMute: %v", err)
	}
	if err := c.SetMute(dev.DID, dev.Model, !orig); err != nil {
		t.Fatalf("SetMute: %v", err)
	}
	got, err := c.GetMute(dev.DID, dev.Model)
	if err != nil {
		t.Fatalf("GetMute after SetMute: %v", err)
	}
	if got != !orig {
		t.Logf("SetMute(%v) then GetMute: got %v (device may have delay)", !orig, got)
	}
	_ = c.SetMute(dev.DID, dev.Model, orig)
}

func testTTS(t *testing.T, api *device.API, dev *device.Device) {
	c := ctrl.New(api)
	if err := c.TTS(dev.DID, dev.Model, "测试"); err != nil {
		t.Fatalf("TTS: %v", err)
	}
}

func testPlay(t *testing.T, api *device.API, dev *device.Device) {
	c := ctrl.New(api)
	if err := c.Play(dev.DID, dev.Model); err != nil {
		t.Fatalf("Play: %v", err)
	}
}

func testPause(t *testing.T, api *device.API, dev *device.Device) {
	c := ctrl.New(api)
	if err := c.Pause(dev.DID, dev.Model); err != nil {
		t.Fatalf("Pause: %v", err)
	}
}

func testPlayPauseSequence(t *testing.T, api *device.API, dev *device.Device) {
	c := ctrl.New(api)
	if err := c.Play(dev.DID, dev.Model); err != nil {
		t.Fatalf("Play: %v", err)
	}
	if err := c.Pause(dev.DID, dev.Model); err != nil {
		t.Fatalf("Pause: %v", err)
	}
	if err := c.Play(dev.DID, dev.Model); err != nil {
		t.Fatalf("Play again: %v", err)
	}
}

func testNext(t *testing.T, api *device.API, dev *device.Device) {
	c := ctrl.New(api)
	if err := c.Next(dev.DID, dev.Model); err != nil {
		t.Fatalf("Next: %v", err)
	}
}

func testPrevious(t *testing.T, api *device.API, dev *device.Device) {
	c := ctrl.New(api)
	if err := c.Previous(dev.DID, dev.Model); err != nil {
		t.Fatalf("Previous: %v", err)
	}
}
