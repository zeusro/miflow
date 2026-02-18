package light

import (
	"testing"

	"github.com/zeusro/miflow/miiot/ctrl"
)

func TestConstants(t *testing.T) {
	s := ctrl.Specs[Model]
	if s.SiidLight != 2 || s.PiidOn != 1 || s.PiidBrightness != 2 {
		t.Errorf("v5ssm spec: siid=%d on=%d brightness=%d", s.SiidLight, s.PiidOn, s.PiidBrightness)
	}
}

func TestSpecInRegistry(t *testing.T) {
	if _, ok := ctrl.Specs[Model]; !ok {
		t.Errorf("model %s not in ctrl.Specs", Model)
	}
}
