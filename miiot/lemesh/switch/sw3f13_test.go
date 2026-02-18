package switch_

import (
	"testing"

	"github.com/zeusro/miflow/miiot/ctrl"
)

func TestConstants(t *testing.T) {
	s := ctrl.Specs[Model]
	if s.SiidSwitch != 2 || s.PiidOn != 1 || s.AiidToggle != 1 {
		t.Errorf("sw3f13 spec: siid=%d piid=%d aiid=%d", s.SiidSwitch, s.PiidOn, s.AiidToggle)
	}
	if len(s.SwitchChannels) != 3 {
		t.Errorf("sw3f13 expected 3 channels, got %d", len(s.SwitchChannels))
	}
	if s.SwitchChannels[0] != SiidLeft || s.SwitchChannels[1] != SiidMiddle || s.SwitchChannels[2] != SiidRight {
		t.Errorf("sw3f13 channels: got %v", s.SwitchChannels)
	}
}

func TestSpecInRegistry(t *testing.T) {
	if _, ok := ctrl.Specs[Model]; !ok {
		t.Errorf("model %s not in ctrl.Specs", Model)
	}
}
