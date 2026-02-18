package miiot

import (
	"strings"
	"testing"
)

func TestModelToPath(t *testing.T) {
	tests := []struct {
		model string
		want  string
	}{
		{"xiaomi.wifispeaker.oh2", "xiaomi/wifispeaker/oh2.go"},
		{"bean.switch.bln31", "bean/switch/bln31.go"},
		{"chuangmi.plug.m3", "chuangmi/plug/m3.go"},
	}
	for _, tt := range tests {
		got := ModelToPath(tt.model)
		if got != tt.want {
			t.Errorf("ModelToPath(%q) = %q, want %q", tt.model, got, tt.want)
		}
	}
}

func TestRegistry(t *testing.T) {
	// 导入 init.go 会触发各子包 init 注册
	models := All()
	if len(models) == 0 {
		t.Fatal("expected registered models")
	}
	// 验证 m list 中的型号已注册
	wantModels := map[string]bool{
		"xiaomi.wifispeaker.oh2": true, "xiaomi.wifispeaker.l05b": true,
		"xiaomi.wifispeaker.l05c": true, "bean.switch.bln31": true,
		"bean.switch.bln33": true, "chuangmi.plug.m3": true,
		"chuangmi.plug.v3": true, "babai.plug.sk01a": true,
		"giot.light.v5ssm": true, "opple.light.bydceiling": true,
		"lemesh.switch.sw3f13": true, "linp.sensor_occupy.hb01": true,
		"xiaomi.tv.eanfv1": true,
	}
	got := make(map[string]bool)
	for _, m := range models {
		got[m] = true
	}
	for m := range wantModels {
		if !got[m] {
			t.Errorf("model %s not registered", m)
		}
	}
}

func TestSpecURL(t *testing.T) {
	url, err := SpecURL("xiaomi.wifispeaker.oh2")
	if err != nil {
		t.Fatalf("SpecURL: %v", err)
	}
	if !strings.Contains(url, "speaker") || !strings.Contains(url, "xiaomi-oh2") {
		t.Errorf("unexpected SpecURL: %s", url)
	}
}

func TestGet(t *testing.T) {
	api, ok := Get("xiaomi.wifispeaker.oh2")
	if !ok {
		t.Fatal("Get xiaomi.wifispeaker.oh2: not found")
	}
	if api.Model() != "xiaomi.wifispeaker.oh2" {
		t.Errorf("Model() = %s", api.Model())
	}
	specURL, err := api.SpecURL()
	if err != nil {
		t.Fatalf("SpecURL: %v", err)
	}
	if !strings.Contains(specURL, "home.miot-spec.com") {
		t.Errorf("SpecURL = %s", specURL)
	}
	prodURL := api.ProductURL()
	if !strings.Contains(prodURL, "xiaomi.wifispeaker.oh2") {
		t.Errorf("ProductURL = %s", prodURL)
	}
}
