package device

// FromMap 从 m list 返回的 map 解析为 Device。
func FromMap(m map[string]interface{}) *Device {
	if m == nil {
		return nil
	}
	did, _ := m["did"].(string)
	model, _ := m["model"].(string)
	name, _ := m["name"].(string)
	token, _ := m["token"].(string)
	return &Device{
		DID:   did,
		Model: model,
		Name:  name,
		Token: token,
	}
}

// ToMap 转为 map，便于 JSON 序列化或与现有 API 兼容。
func (d *Device) ToMap() map[string]interface{} {
	if d == nil {
		return nil
	}
	return map[string]interface{}{
		"did":   d.DID,
		"model": d.Model,
		"name":  d.Name,
		"token": d.Token,
	}
}
