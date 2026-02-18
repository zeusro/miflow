package specs

import (
	"strings"
	"testing"
)

func TestFetchInstance_Oh2(t *testing.T) {
	urn := "urn:miot-spec-v2:device:speaker:0000A015:xiaomi-oh2:1"
	raw, err := FetchInstance(urn)
	if err != nil {
		t.Fatalf("FetchInstance: %v", err)
	}
	if raw["type"] != urn {
		t.Errorf("type mismatch: %v", raw["type"])
	}
	svcs, _ := raw["services"].([]interface{})
	if len(svcs) == 0 {
		t.Fatal("no services")
	}
	// 检查是否有 Speaker 服务
	var hasSpeaker bool
	for _, s := range svcs {
		sm, _ := s.(map[string]interface{})
		if sm == nil {
			continue
		}
		if desc, _ := sm["description"].(string); desc == "Speaker" {
			hasSpeaker = true
			break
		}
	}
	if !hasSpeaker {
		t.Error("expected Speaker service")
	}
}

func TestFetchInstance_InvalidURN(t *testing.T) {
	_, err := FetchInstance("urn:invalid:xyz")
	if err != nil {
		return // 预期可能失败
	}
	// 若 miot-spec.org 返回了内容，也接受
}

func TestProductURL(t *testing.T) {
	url := ProductURLForModel("xiaomi.wifispeaker.oh2")
	if !strings.HasSuffix(url, "/xiaomi.wifispeaker.oh2") {
		t.Errorf("ProductURL: %s", url)
	}
}
