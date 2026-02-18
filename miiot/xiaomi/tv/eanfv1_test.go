package tv

import (
	"testing"

	"github.com/zeusro/miflow/miiot/ctrl"
)

func TestConstants(t *testing.T) {
	s := ctrl.Specs[Model]
	if s.SiidTV != 2 || s.AiidTurnOff != 1 {
		t.Errorf("eanfv1 spec: siid=%d turnOff=%d", s.SiidTV, s.AiidTurnOff)
	}
}

func TestSpecInRegistry(t *testing.T) {
	if _, ok := ctrl.Specs[Model]; !ok {
		t.Errorf("model %s not in ctrl.Specs", Model)
	}
}
