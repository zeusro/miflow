package plug

import (
	"testing"

	"github.com/zeusro/miflow/miiot/ctrl"
)

func TestM3Constants(t *testing.T) {
	s := ctrl.Specs[Model]
	if s.SiidSwitch != 2 || s.PiidOn != 1 {
		t.Errorf("m3 spec: siid=%d piid=%d", s.SiidSwitch, s.PiidOn)
	}
	if s.AiidToggle != 0 {
		t.Errorf("m3 plug should not have toggle, got aiid=%d", s.AiidToggle)
	}
}

func TestV3Constants(t *testing.T) {
	s := ctrl.Specs[ModelV3]
	if s.SiidSwitch != 2 || s.PiidOn != 1 {
		t.Errorf("v3 spec: siid=%d piid=%d", s.SiidSwitch, s.PiidOn)
	}
}

func TestSpecsInRegistry(t *testing.T) {
	for _, model := range []string{Model, ModelV3} {
		if _, ok := ctrl.Specs[model]; !ok {
			t.Errorf("model %s not in ctrl.Specs", model)
		}
	}
}
