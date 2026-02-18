package switch_

import (
	"testing"

	"github.com/zeusro/miflow/miiot/ctrl"
)

func TestConstants(t *testing.T) {
	s := ctrl.Specs[Model]
	if s.SiidSwitch != 2 || s.PiidOn != 1 || s.AiidToggle != 1 {
		t.Errorf("bln31 spec: siid=%d piid=%d aiid=%d", s.SiidSwitch, s.PiidOn, s.AiidToggle)
	}
	s2 := ctrl.Specs[ModelBln33]
	if s2.SiidSwitch != 2 {
		t.Errorf("bln33 spec: siid=%d", s2.SiidSwitch)
	}
}
