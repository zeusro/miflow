package ctrl

import (
	"testing"
)

func TestResolveSpec_Oh2(t *testing.T) {
	s, err := ResolveSpec("xiaomi.wifispeaker.oh2")
	if err != nil {
		t.Fatalf("ResolveSpec: %v", err)
	}
	if s.SiidSpeaker != 2 || s.PiidVolume != 1 || s.PiidMute != 2 {
		t.Errorf("speaker: siid=%d volume=%d mute=%d", s.SiidSpeaker, s.PiidVolume, s.PiidMute)
	}
	if s.SiidPlayControl != 3 || s.AiidPlay != 2 || s.AiidPause != 3 {
		t.Errorf("play: siid=%d play=%d pause=%d", s.SiidPlayControl, s.AiidPlay, s.AiidPause)
	}
	if s.SiidVoiceAssistant != 5 || s.AiidExecuteText == 0 {
		t.Errorf("TTS: siid=%d aiid=%d", s.SiidVoiceAssistant, s.AiidExecuteText)
	}
}

func TestResolveSpec_Switch(t *testing.T) {
	s, err := ResolveSpec("bean.switch.bln31")
	if err != nil {
		t.Fatalf("ResolveSpec: %v", err)
	}
	if s.SiidSwitch != 2 || s.PiidOn != 1 {
		t.Errorf("switch: siid=%d on=%d", s.SiidSwitch, s.PiidOn)
	}
	if s.AiidToggle == 0 {
		t.Error("expected toggle action")
	}
}

func TestResolveSpec_Plug(t *testing.T) {
	s, err := ResolveSpec("chuangmi.plug.m3")
	if err != nil {
		t.Fatalf("ResolveSpec: %v", err)
	}
	if s.SiidSwitch != 2 || s.PiidOn != 1 {
		t.Errorf("plug: siid=%d on=%d", s.SiidSwitch, s.PiidOn)
	}
}

func TestResolveSpec_Light(t *testing.T) {
	s, err := ResolveSpec("opple.light.bydceiling")
	if err != nil {
		t.Fatalf("ResolveSpec: %v", err)
	}
	if s.SiidLight == 0 || s.PiidBrightness == 0 {
		t.Errorf("light: siid=%d brightness=%d", s.SiidLight, s.PiidBrightness)
	}
}

func TestResolveSpec_TV(t *testing.T) {
	s, err := ResolveSpec("xiaomi.tv.eanfv1")
	if err != nil {
		t.Fatalf("ResolveSpec: %v", err)
	}
	if s.SiidTV != 2 || s.AiidTurnOff != 1 {
		t.Errorf("TV: siid=%d turnOff=%d", s.SiidTV, s.AiidTurnOff)
	}
}

func TestResolveSpec_Occupancy(t *testing.T) {
	s, err := ResolveSpec("linp.sensor_occupy.hb01")
	if err != nil {
		t.Fatalf("ResolveSpec: %v", err)
	}
	if s.SiidOccupancy != 2 || s.PiidStatus != 1 {
		t.Errorf("occupancy: siid=%d status=%d", s.SiidOccupancy, s.PiidStatus)
	}
}

func TestResolveSpec_UnknownModel(t *testing.T) {
	_, err := ResolveSpec("unknown.model.xyz999")
	if err == nil {
		t.Error("expected error for unknown model")
	}
}

func TestSpec_DynamicFallback(t *testing.T) {
	// spec() 应优先用静态 Specs，动态解析作为回退
	s1 := spec("xiaomi.wifispeaker.oh2")
	if s1.SiidSpeaker == 0 {
		t.Error("oh2 should have speaker spec")
	}
	// 动态解析的型号（若 instances 中有）
	s2 := spec("xiaomi.wifispeaker.l05b")
	if s2.SiidSpeaker == 0 {
		t.Error("l05b should have speaker spec")
	}
}
