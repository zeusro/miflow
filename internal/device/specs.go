package device

import (
	"encoding/json"
	"fmt"

	"github.com/zeusro/miflow/pkg/i18n"
)

// LoadSpec 从 API 获取指定型号的 SPEC 并解析为 ModelSpec。
func (a *API) LoadSpec(model string) (*ModelSpec, error) {
	raw, err := a.Spec(model, "json")
	if err != nil {
		return nil, err
	}
	m, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "device.invalid_spec", map[string]interface{}{"Model": model}))
	}
	return parseModelSpec(m)
}

// LoadAllModelSpecs 获取 m list 中所有唯一型号的 SPEC，遵循 docs/spec.md 流程。
// 返回 model -> ModelSpec 映射，未找到 SPEC 的型号会记录在 failed 中。
func (a *API) LoadAllModelSpecs() (map[string]*ModelSpec, map[string]error) {
	devs, err := a.List("", false, 0)
	if err != nil {
		return nil, map[string]error{"_list": err}
	}
	models := uniqueModels(devs)
	specs := make(map[string]*ModelSpec)
	failed := make(map[string]error)
	for _, m := range models {
		spec, err := a.LoadSpec(m)
		if err != nil {
			failed[m] = err
			continue
		}
		specs[m] = spec
	}
	return specs, failed
}

// uniqueModels 从设备列表提取唯一型号列表。
func uniqueModels(devs []*Device) []string {
	seen := make(map[string]bool)
	var out []string
	for _, d := range devs {
		if d != nil && d.Model != "" && !seen[d.Model] {
			seen[d.Model] = true
			out = append(out, d.Model)
		}
	}
	return out
}

// parseModelSpec 从 map 解析为 ModelSpec。
func parseModelSpec(m map[string]interface{}) (*ModelSpec, error) {
	spec := &ModelSpec{
		Type:        getStr(m, "type"),
		Description: getStr(m, "description"),
	}
	svcs, _ := m["services"].([]interface{})
	for _, s := range svcs {
		sm, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		svc := ServiceSpec{
			IID:         int(getFloat(sm, "iid")),
			Description: getStr(sm, "description"),
			Type:        getStr(sm, "type"),
		}
		for _, p := range toSlice(sm["properties"]) {
			if pm, ok := p.(map[string]interface{}); ok {
				svc.Properties = append(svc.Properties, PropSpec{
					IID:         int(getFloat(pm, "iid")),
					Description: getStr(pm, "description"),
					Format:      getStr(pm, "format"),
					Access:      getStrSlice(pm, "access"),
				})
			}
		}
		for _, a := range toSlice(sm["actions"]) {
			if am, ok := a.(map[string]interface{}); ok {
				svc.Actions = append(svc.Actions, ActionSpec{
					IID:         int(getFloat(am, "iid")),
					Description: getStr(am, "description"),
					In:          toSlice(am["in"]),
					Out:         toSlice(am["out"]),
				})
			}
		}
		for _, e := range toSlice(sm["events"]) {
			if em, ok := e.(map[string]interface{}); ok {
				svc.Events = append(svc.Events, EventSpec{
					IID:         int(getFloat(em, "iid")),
					Description: getStr(em, "description"),
					Arguments:   toSlice(em["arguments"]),
				})
			}
		}
		spec.Services = append(spec.Services, svc)
	}
	return spec, nil
}

// getStr 从 map 安全获取字符串。
func getStr(m map[string]interface{}, k string) string {
	if v, ok := m[k].(string); ok {
		return v
	}
	return ""
}

// getFloat 从 map 安全获取 float64。
func getFloat(m map[string]interface{}, k string) float64 {
	switch v := m[k].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	}
	return 0
}

// getStrSlice 从 map 安全获取字符串切片。
func getStrSlice(m map[string]interface{}, k string) []string {
	raw := toSlice(m[k])
	out := make([]string, 0, len(raw))
	for _, r := range raw {
		if s, ok := r.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// toSlice 将 interface 转为 []interface{}。
func toSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	if s, ok := v.([]interface{}); ok {
		return s
	}
	return nil
}

// Summary 返回 SPEC 的简要描述（服务数、属性数、动作数）。
func (s *ModelSpec) Summary() string {
	var props, actions int
	for _, svc := range s.Services {
		props += len(svc.Properties)
		actions += len(svc.Actions)
	}
	return i18n.T(i18n.DefaultLang(), "device.spec_summary", map[string]interface{}{
		"Services": len(s.Services), "Properties": props, "Actions": actions,
	})
}

// ToJSON 序列化为 JSON。
func (s *ModelSpec) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}
