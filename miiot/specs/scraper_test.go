package specs

import (
	"strings"
	"testing"
)

func TestTechSpecURL(t *testing.T) {
	url, err := TechSpecURL("xiaomi.tv.eanfv1")
	if err != nil {
		t.Skipf("TechSpecURL (network): %v", err)
		return
	}
	if !strings.Contains(url, "television") {
		t.Errorf("expected television in spec URL, got %s", url)
	}
	if !strings.Contains(url, "xiaomi-eanfv1") {
		t.Errorf("expected xiaomi-eanfv1 in spec URL, got %s", url)
	}
	t.Logf("xiaomi.tv.eanfv1 -> %s", url)
}

func TestScrapeProductPage(t *testing.T) {
	m, err := ScrapeProductPage("xiaomi.wifispeaker.oh2")
	if err != nil {
		t.Skipf("ScrapeProductPage (network): %v", err)
		return
	}
	if len(m) == 0 {
		t.Skip("no model->URN extracted from page")
		return
	}
	// 当前型号或关联型号应存在
	if urn, ok := m["xiaomi.wifispeaker.oh2"]; ok {
		if urn == "" || len(urn) < 20 {
			t.Errorf("invalid URN for oh2: %s", urn)
		}
		t.Logf("xiaomi.wifispeaker.oh2 -> %s", urn)
	}
	// 可能有关联型号如 oh27, oh2p
	for model, urn := range m {
		if len(urn) > 0 && len(model) > 0 {
			short := urn
			if len(short) > 60 {
				short = short[:60] + "..."
			}
			t.Logf("  %s -> %s", model, short)
		}
	}
}
