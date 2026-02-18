package specs

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	m, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(m) == 0 {
		t.Fatal("expected non-empty model map")
	}
	// 验证 m list 中的型号存在
	models := []string{
		"xiaomi.wifispeaker.oh2", "bean.switch.bln31", "chuangmi.plug.m3",
		"opple.light.bydceiling", "giot.light.v5ssm",
	}
	for _, model := range models {
		urn, ok := m[model]
		if !ok {
			t.Errorf("model %s not found", model)
			continue
		}
		if !strings.HasPrefix(urn, "urn:miot-spec-v2:") {
			t.Errorf("model %s: invalid URN %s", model, urn)
		}
	}
}

func TestSpecURL(t *testing.T) {
	url, err := SpecURL("xiaomi.wifispeaker.oh2")
	if err != nil {
		t.Fatalf("SpecURL: %v", err)
	}
	if !strings.Contains(url, "home.miot-spec.com/spec") {
		t.Errorf("expected home.miot-spec.com/spec in URL, got %s", url)
	}
	if !strings.Contains(url, "urn") && !strings.Contains(url, "urn%3A") {
		t.Errorf("expected URN in URL, got %s", url)
	}
}

func TestProductURLForModel(t *testing.T) {
	url := ProductURLForModel("xiaomi.wifispeaker.oh2")
	if !strings.HasSuffix(url, "/xiaomi.wifispeaker.oh2") {
		t.Errorf("expected product URL to end with model, got %s", url)
	}
}
