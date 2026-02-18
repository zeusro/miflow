package light

import (
	"testing"

	"github.com/zeusro/miflow/miiot/ctrl"
)

func TestConstants(t *testing.T) {
	s := ctrl.Specs[Model]
	if s.SiidLight != 2 || s.PiidOn != 1 || s.PiidBrightness != 3 {
		t.Errorf("bydceiling: siid=%d on=%d brightness=%d", s.SiidLight, s.PiidOn, s.PiidBrightness)
	}
}
