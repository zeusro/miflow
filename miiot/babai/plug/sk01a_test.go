package plug

import (
	"testing"

	"github.com/zeusro/miflow/miiot/ctrl"
)

func TestConstants(t *testing.T) {
	s := ctrl.Specs[Model]
	if s.SiidSwitch != 2 || s.PiidOn != 1 {
		t.Errorf("sk01a spec: siid=%d piid=%d", s.SiidSwitch, s.PiidOn)
	}
	// 插座无 Toggle
	if s.AiidToggle != 0 {
		t.Errorf("sk01a should not have toggle, got aiid=%d", s.AiidToggle)
	}
}

func TestSpecInRegistry(t *testing.T) {
	if _, ok := ctrl.Specs[Model]; !ok {
		t.Errorf("model %s not in ctrl.Specs", Model)
	}
}
