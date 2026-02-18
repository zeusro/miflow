package ctrl

import (
	"strings"
	"sync"

	"github.com/zeusro/miflow/miiot/specs"
)

var (
	resolveCache   = make(map[string]Spec)
	resolveCacheMu sync.RWMutex
)

// ResolveSpec 根据 model 从 miot-spec.org 解析规格，映射为 Spec。
// 先查 model->URN，再拉取 instance，按服务/属性/动作描述匹配标准能力。
func ResolveSpec(model string) (Spec, error) {
	resolveCacheMu.RLock()
	if s, ok := resolveCache[model]; ok {
		resolveCacheMu.RUnlock()
		return s, nil
	}
	resolveCacheMu.RUnlock()

	urn, err := specs.URNWithScrape(model)
	if err != nil {
		return Spec{}, err
	}
	raw, err := specs.FetchInstance(urn)
	if err != nil {
		return Spec{}, err
	}
	s, err := parseInstanceToSpec(raw)
	if err != nil {
		return Spec{}, err
	}
	// 多通道开关等需保留静态配置
	if existing, ok := Specs[model]; ok && len(existing.SwitchChannels) > 0 {
		s.SwitchChannels = existing.SwitchChannels
	}
	resolveCacheMu.Lock()
	resolveCache[model] = s
	resolveCacheMu.Unlock()
	return s, nil
}

func parseInstanceToSpec(m map[string]interface{}) (Spec, error) {
	s := Spec{}
	svcs, _ := m["services"].([]interface{})
	for _, x := range svcs {
		sm, ok := x.(map[string]interface{})
		if !ok {
			continue
		}
		desc := strings.ToLower(getStr(sm, "description"))
		svcType := strings.ToLower(getStr(sm, "type"))
		siid := int(getFloat(sm, "iid"))

		// Switch / Outlet
		if strings.Contains(desc, "switch") || desc == "outlet" {
			if s.SiidSwitch == 0 {
				s.SiidSwitch = siid
			}
			for _, p := range toSlice(sm["properties"]) {
				pm, _ := p.(map[string]interface{})
				if pm == nil {
					continue
				}
				pdesc := strings.ToLower(getStr(pm, "description"))
				if strings.Contains(pdesc, "switch status") || pdesc == "on" {
					s.PiidOn = int(getFloat(pm, "iid"))
					break
				}
			}
			for _, a := range toSlice(sm["actions"]) {
				am, _ := a.(map[string]interface{})
				if am == nil {
					continue
				}
				if strings.ToLower(getStr(am, "description")) == "toggle" {
					s.AiidToggle = int(getFloat(am, "iid"))
					break
				}
			}
		}

		// Light (排除 night-light 等)
		if strings.Contains(desc, "light") && !strings.Contains(desc, "night") {
			if s.SiidLight == 0 {
				s.SiidLight = siid
			}
			for _, p := range toSlice(sm["properties"]) {
				pm, _ := p.(map[string]interface{})
				if pm == nil {
					continue
				}
				pdesc := strings.ToLower(getStr(pm, "description"))
				if (pdesc == "on" || strings.Contains(pdesc, "switch status")) && s.PiidOn == 0 {
					s.PiidOn = int(getFloat(pm, "iid"))
				}
				if strings.Contains(pdesc, "brightness") {
					s.PiidBrightness = int(getFloat(pm, "iid"))
				}
			}
		}

		// Speaker
		if desc == "speaker" {
			s.SiidSpeaker = siid
			for _, p := range toSlice(sm["properties"]) {
				pm, _ := p.(map[string]interface{})
				if pm == nil {
					continue
				}
				pdesc := strings.ToLower(getStr(pm, "description"))
				if pdesc == "volume" {
					s.PiidVolume = int(getFloat(pm, "iid"))
				}
				if pdesc == "mute" {
					s.PiidMute = int(getFloat(pm, "iid"))
				}
			}
		}

		// Play Control
		if strings.Contains(desc, "play control") {
			s.SiidPlayControl = siid
			for _, a := range toSlice(sm["actions"]) {
				am, _ := a.(map[string]interface{})
				if am == nil {
					continue
				}
				adesc := strings.ToLower(getStr(am, "description"))
				aiid := int(getFloat(am, "iid"))
				switch adesc {
				case "play":
					s.AiidPlay = aiid
				case "pause":
					s.AiidPause = aiid
				case "next":
					s.AiidNext = aiid
				case "previous":
					s.AiidPrevious = aiid
				}
			}
		}

		// Intelligent Speaker / Voice Assistant (TTS)
		if strings.Contains(desc, "intelligent") || strings.Contains(desc, "voice") {
			s.SiidVoiceAssistant = siid
			for _, a := range toSlice(sm["actions"]) {
				am, _ := a.(map[string]interface{})
				if am == nil {
					continue
				}
				adesc := strings.ToLower(getStr(am, "description"))
				if strings.Contains(adesc, "play text") || strings.Contains(adesc, "execute text") {
					s.AiidExecuteText = int(getFloat(am, "iid"))
					break
				}
			}
		}

		// TV (television / tv-switch)
		if strings.Contains(desc, "television") || strings.Contains(desc, "tv") {
			s.SiidTV = siid
			for _, a := range toSlice(sm["actions"]) {
				am, _ := a.(map[string]interface{})
				if am == nil {
					continue
				}
				adesc := strings.ToLower(getStr(am, "description"))
				if adesc == "turn off" || adesc == "tv-switchon" {
					s.AiidTurnOff = int(getFloat(am, "iid"))
					break
				}
			}
		}

		// Occupancy Sensor
		if strings.Contains(desc, "occupancy") || strings.Contains(svcType, "occupancy") {
			s.SiidOccupancy = siid
			for _, p := range toSlice(sm["properties"]) {
				pm, _ := p.(map[string]interface{})
				if pm == nil {
					continue
				}
				pdesc := strings.ToLower(getStr(pm, "description"))
				if strings.Contains(pdesc, "occupancy") || pdesc == "status" {
					s.PiidStatus = int(getFloat(pm, "iid"))
					break
				}
			}
		}
	}
	return s, nil
}

func getStr(m map[string]interface{}, k string) string {
	if v, ok := m[k].(string); ok {
		return v
	}
	return ""
}

func getFloat(m map[string]interface{}, k string) float64 {
	switch v := m[k].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	}
	return 0
}

func toSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	if s, ok := v.([]interface{}); ok {
		return s
	}
	return nil
}
