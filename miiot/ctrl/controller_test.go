package ctrl

import (
	"testing"

	"github.com/zeusro/miflow/internal/device"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/miioservice"
	"github.com/zeusro/miflow/internal/config"
	"os"
)

func TestSpecsCoverage(t *testing.T) {
	for _, model := range []string{
		"bean.switch.bln31", "chuangmi.plug.m3", "opple.light.bydceiling",
		"xiaomi.wifispeaker.oh2", "xiaomi.tv.eanfv1", "linp.sensor_occupy.hb01",
	} {
		s := spec(model)
		if s.SiidSwitch == 0 && s.SiidLight == 0 && s.SiidSpeaker == 0 &&
			s.SiidTV == 0 && s.SiidOccupancy == 0 {
			t.Errorf("model %s: no spec constants", model)
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
		if d != nil && (d.Model == "babai.plug.sk01a" || d.Model == "chuangmi.plug.m3") {
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

func TestControllerErrors(t *testing.T) {
	api := setupAPI(t)
	if api == nil {
		return
	}
	ctrl := New(api)
	devs, _ := api.List("", false, 0)
	var did string
	if len(devs) > 0 && devs[0] != nil {
		did = devs[0].DID
	} else {
		did = "123456789"
	}
	// 不支持的 model
	if err := ctrl.SetOn(did, "unknown.model.xyz", true); err == nil {
		t.Error("expected error for unknown model")
	}
	if _, err := ctrl.GetOn(did, "unknown.model.xyz"); err == nil {
		t.Error("expected error for unknown model")
	}
	// 传感器无 SetOn
	if err := ctrl.SetOn(did, "linp.sensor_occupy.hb01", true); err == nil {
		t.Error("expected error: sensor has no set")
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
