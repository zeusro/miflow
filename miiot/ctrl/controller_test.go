package ctrl

import (
	"testing"

	"github.com/zeusro/miflow/internal/device"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/miioservice"
	"github.com/zeusro/miflow/internal/config"
	"os"
)

// ModelsFromReadme 为 readme.md 中 ./m list 的所有型号。
var ModelsFromReadme = []string{
	"babai.plug.sk01a", "bean.switch.bln31", "bean.switch.bln33",
	"chuangmi.plug.m3", "chuangmi.plug.v3", "giot.light.v5ssm",
	"lemesh.switch.sw3f13", "linp.sensor_occupy.hb01",
	"opple.light.bydceiling", "xiaomi.tv.eanfv1",
	"xiaomi.wifispeaker.l05b", "xiaomi.wifispeaker.l05c", "xiaomi.wifispeaker.oh2",
}

func TestSpecsCoverage(t *testing.T) {
	for _, model := range ModelsFromReadme {
		s := spec(model)
		if s.SiidSwitch == 0 && s.SiidLight == 0 && s.SiidSpeaker == 0 &&
			s.SiidTV == 0 && s.SiidOccupancy == 0 {
			t.Errorf("model %s: no spec constants", model)
		}
	}
}

func TestSpecsCoverage_DeviceSpecific(t *testing.T) {
	tests := []struct {
		model      string
		hasSwitch  bool
		hasLight   bool
		hasSpeaker bool
		hasTV      bool
		hasOccupancy bool
		hasToggle  bool
		hasBrightness bool
		hasChannels bool
	}{
		{"babai.plug.sk01a", true, false, false, false, false, false, false, false},
		{"bean.switch.bln31", true, false, false, false, false, true, false, false},
		{"bean.switch.bln33", true, false, false, false, false, true, false, false},
		{"chuangmi.plug.m3", true, false, false, false, false, false, false, false},
		{"chuangmi.plug.v3", true, false, false, false, false, false, false, false},
		{"giot.light.v5ssm", false, true, false, false, false, false, true, false},
		{"lemesh.switch.sw3f13", true, false, false, false, false, true, false, true},
		{"linp.sensor_occupy.hb01", false, false, false, false, true, false, false, false},
		{"opple.light.bydceiling", false, true, false, false, false, false, true, false},
		{"xiaomi.tv.eanfv1", false, false, false, true, false, false, false, false},
		{"xiaomi.wifispeaker.l05b", false, false, true, false, false, false, false, false},
		{"xiaomi.wifispeaker.l05c", false, false, true, false, false, false, false, false},
		{"xiaomi.wifispeaker.oh2", false, false, true, false, false, false, false, false},
	}
	for _, tt := range tests {
		s := spec(tt.model)
		if tt.hasSwitch && s.SiidSwitch == 0 {
			t.Errorf("%s: expected switch", tt.model)
		}
		if tt.hasLight && s.SiidLight == 0 {
			t.Errorf("%s: expected light", tt.model)
		}
		if tt.hasSpeaker && s.SiidSpeaker == 0 {
			t.Errorf("%s: expected speaker", tt.model)
		}
		if tt.hasTV && s.SiidTV == 0 {
			t.Errorf("%s: expected TV", tt.model)
		}
		if tt.hasOccupancy && s.SiidOccupancy == 0 {
			t.Errorf("%s: expected occupancy", tt.model)
		}
		if tt.hasToggle && s.AiidToggle == 0 {
			t.Errorf("%s: expected toggle", tt.model)
		}
		if tt.hasBrightness && s.PiidBrightness == 0 {
			t.Errorf("%s: expected brightness", tt.model)
		}
		if tt.hasChannels && len(s.SwitchChannels) == 0 {
			t.Errorf("%s: expected switch channels", tt.model)
		}
	}
}

func TestControllerSetOnGetOn(t *testing.T) {
	api := setupAPI(t)
	if api == nil {
		return
	}
	ctrl := New(api)
	// 找一个 switch 设备
	devs, err := api.List("", false, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	var switchDev *device.Device
	for _, d := range devs {
		if d != nil {
			s := spec(d.Model)
			if s.SiidSwitch != 0 || s.SiidLight != 0 {
				switchDev = d
				break
			}
		}
	}
	if switchDev == nil {
		t.Skip("no switch/light device in list")
	}
	// 读当前状态
	on, err := ctrl.GetOn(switchDev.DID, switchDev.Model)
	if err != nil {
		t.Fatalf("GetOn: %v", err)
	}
	t.Logf("device %s (%s) on=%v", switchDev.Name, switchDev.Model, on)
	// 切换后读回（会改变设备状态，仅在有明确测试设备时启用）
	// ctrl.SetOn(switchDev.DID, switchDev.Model, !on)
}

func TestControllerGetOccupancy(t *testing.T) {
	api := setupAPI(t)
	if api == nil {
		return
	}
	ctrl := New(api)
	devs, err := api.List("", false, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, d := range devs {
		if d == nil || d.Model != "linp.sensor_occupy.hb01" {
			continue
		}
		val, err := ctrl.GetOccupancy(d.DID, d.Model)
		if err != nil {
			t.Fatalf("GetOccupancy: %v", err)
		}
		t.Logf("occupancy %s: %v", d.Name, val)
		return
	}
	t.Skip("no linp.sensor_occupy.hb01 in list")
}

func TestControllerGetVolume(t *testing.T) {
	api := setupAPI(t)
	if api == nil {
		return
	}
	ctrl := New(api)
	devs, err := api.List("", false, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, d := range devs {
		if d == nil {
			continue
		}
		s := spec(d.Model)
		if s.SiidSpeaker == 0 {
			continue
		}
		vol, err := ctrl.GetVolume(d.DID, d.Model)
		if err != nil {
			t.Fatalf("GetVolume %s: %v", d.Model, err)
		}
		t.Logf("speaker %s (%s) volume=%d", d.Name, d.Model, vol)
		return
	}
	t.Skip("no speaker in list")
}

func TestControllerSetOnRoundtrip(t *testing.T) {
	api := setupAPI(t)
	if api == nil {
		return
	}
	ctrl := New(api)
	devs, err := api.List("", false, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	var target *device.Device
	for _, d := range devs {
		if d != nil && (d.Model == "babai.plug.sk01a" || d.Model == "chuangmi.plug.m3" || d.Model == "chuangmi.plug.v3") {
			target = d
			break
		}
	}
	if target == nil {
		t.Skip("no plug device for SetOn test")
	}
	on, err := ctrl.GetOn(target.DID, target.Model)
	if err != nil {
		t.Fatalf("GetOn: %v", err)
	}
	// 切换并验证 API 调用成功
	if err := ctrl.SetOn(target.DID, target.Model, !on); err != nil {
		t.Fatalf("SetOn: %v", err)
	}
	got, err := ctrl.GetOn(target.DID, target.Model)
	if err != nil {
		t.Fatalf("GetOn after SetOn: %v", err)
	}
	// 部分设备有延迟，仅记录结果
	if got != !on {
		t.Logf("SetOn(%v) then GetOn: got %v (device may have delay)", !on, got)
	}
	// 恢复原状态
	_ = ctrl.SetOn(target.DID, target.Model, on)
}

func TestControllerGetBrightness(t *testing.T) {
	api := setupAPI(t)
	if api == nil {
		return
	}
	ctrl := New(api)
	devs, err := api.List("", false, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, d := range devs {
		if d == nil {
			continue
		}
		s := spec(d.Model)
		if s.SiidLight == 0 || s.PiidBrightness == 0 {
			continue
		}
		level, err := ctrl.GetBrightness(d.DID, d.Model)
		if err != nil {
			t.Fatalf("GetBrightness %s: %v", d.Model, err)
		}
		t.Logf("light %s (%s) brightness=%d", d.Name, d.Model, level)
		return
	}
	t.Skip("no light with brightness in list")
}

func TestControllerToggle(t *testing.T) {
	api := setupAPI(t)
	if api == nil {
		return
	}
	c := New(api)
	devs, err := api.List("", false, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, d := range devs {
		if d == nil {
			continue
		}
		s := spec(d.Model)
		if s.SiidSwitch == 0 || s.AiidToggle == 0 {
			continue
		}
		// bean.switch / lemesh.switch 支持 Toggle
		if err := c.Toggle(d.DID, d.Model); err != nil {
			t.Fatalf("Toggle %s: %v", d.Model, err)
		}
		t.Logf("Toggle %s (%s) ok", d.Name, d.Model)
		return
	}
	t.Skip("no switch with toggle in list")
}

func TestControllerSetSwitchChannel(t *testing.T) {
	api := setupAPI(t)
	if api == nil {
		return
	}
	c := New(api)
	devs, err := api.List("", false, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, d := range devs {
		if d == nil || d.Model != "lemesh.switch.sw3f13" {
			continue
		}
		on, err := c.GetOn(d.DID, d.Model)
		if err != nil {
			t.Fatalf("GetOn: %v", err)
		}
		// 切换通道 0 的状态
		if err := c.SetSwitchChannel(d.DID, d.Model, 0, !on); err != nil {
			t.Fatalf("SetSwitchChannel: %v", err)
		}
		t.Logf("SetSwitchChannel %s channel 0 -> %v", d.Name, !on)
		_ = c.SetSwitchChannel(d.DID, d.Model, 0, on)
		return
	}
	t.Skip("no lemesh.switch.sw3f13 in list")
}

func TestControllerTVTurnOff(t *testing.T) {
	api := setupAPI(t)
	if api == nil {
		return
	}
	c := New(api)
	devs, err := api.List("", false, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, d := range devs {
		if d == nil || d.Model != "xiaomi.tv.eanfv1" {
			continue
		}
		if err := c.TVTurnOff(d.DID, d.Model); err != nil {
			t.Fatalf("TVTurnOff: %v", err)
		}
		t.Logf("TVTurnOff %s ok", d.Name)
		return
	}
	t.Skip("no xiaomi.tv.eanfv1 in list")
}

func TestControllerErrors(t *testing.T) {
	api := setupAPI(t)
	if api == nil {
		return
	}
	c := New(api)
	devs, _ := api.List("", false, 0)
	var did string
	if len(devs) > 0 && devs[0] != nil {
		did = devs[0].DID
	} else {
		did = "123456789"
	}
	unknown := "unknown.model.xyz"
	// 不支持的 model
	if err := c.SetOn(did, unknown, true); err == nil {
		t.Error("expected error for unknown model SetOn")
	}
	if _, err := c.GetOn(did, unknown); err == nil {
		t.Error("expected error for unknown model GetOn")
	}
	if err := c.Toggle(did, unknown); err == nil {
		t.Error("expected error for unknown model Toggle")
	}
	if err := c.SetBrightness(did, unknown, 50); err == nil {
		t.Error("expected error for unknown model SetBrightness")
	}
	if _, err := c.GetBrightness(did, unknown); err == nil {
		t.Error("expected error for unknown model GetBrightness")
	}
	if err := c.TTS(did, unknown, "test"); err == nil {
		t.Error("expected error for unknown model TTS")
	}
	if err := c.SetVolume(did, unknown, 50); err == nil {
		t.Error("expected error for unknown model SetVolume")
	}
	if _, err := c.GetVolume(did, unknown); err == nil {
		t.Error("expected error for unknown model GetVolume")
	}
	if err := c.SetMute(did, unknown, true); err == nil {
		t.Error("expected error for unknown model SetMute")
	}
	if _, err := c.GetMute(did, unknown); err == nil {
		t.Error("expected error for unknown model GetMute")
	}
	if err := c.Play(did, unknown); err == nil {
		t.Error("expected error for unknown model Play")
	}
	if err := c.Pause(did, unknown); err == nil {
		t.Error("expected error for unknown model Pause")
	}
	if err := c.TVTurnOff(did, unknown); err == nil {
		t.Error("expected error for unknown model TVTurnOff")
	}
	if _, err := c.GetOccupancy(did, unknown); err == nil {
		t.Error("expected error for unknown model GetOccupancy")
	}
	if err := c.SetSwitchChannel(did, unknown, 0, true); err == nil {
		t.Error("expected error for unknown model SetSwitchChannel")
	}
	// 传感器无 SetOn
	if err := c.SetOn(did, "linp.sensor_occupy.hb01", true); err == nil {
		t.Error("expected error: sensor has no set")
	}
	// SetSwitchChannel 通道越界
	if err := c.SetSwitchChannel(did, "lemesh.switch.sw3f13", 99, true); err == nil {
		t.Error("expected error for channel out of range")
	}
}

func setupAPI(t *testing.T) *device.API {
	t.Helper()
	cfg := config.Get()
	tokenPath := cfg.TokenPath
	if tokenPath == "" {
		tokenPath = os.ExpandEnv("$HOME/.mi.token")
	}
	store := &miaccount.TokenStore{Path: tokenPath}
	token := store.LoadOAuth()
	if token == nil || !token.IsValid() {
		t.Skip("no valid OAuth token, run 'm login' first")
		return nil
	}
	ioSvc, err := miioservice.New(token, tokenPath)
	if err != nil {
		t.Fatalf("miioservice.New: %v", err)
		return nil
	}
	return device.NewAPI(ioSvc)
}
