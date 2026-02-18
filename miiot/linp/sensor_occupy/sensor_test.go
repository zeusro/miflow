package sensor_occupy

import (
	"testing"

	"github.com/zeusro/miflow/miiot/ctrl"
)

func TestConstants(t *testing.T) {
	s := ctrl.Specs[Model]
	if s.SiidOccupancy != 2 || s.PiidStatus != 1 {
		t.Errorf("hb01: siid=%d piid=%d", s.SiidOccupancy, s.PiidStatus)
	}
}
