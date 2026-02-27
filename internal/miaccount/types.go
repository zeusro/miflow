package miaccount

import "encoding/json"

// Token 保存小米账号认证数据。
// 键：deviceId、userId、passToken，以及按服务分的键如 "micoapi"、"xiaomiio"，
// 值为 [ssecurity, serviceToken]。
type Token struct {
	DeviceID  string            `json:"deviceId"`
	UserID    string            `json:"userId,omitempty"`
	PassToken string            `json:"passToken,omitempty"`
	Services  map[string][]string `json:"-"` // sid -> [ssecurity, serviceToken]；序列化为扁平键
}

// UnmarshalJSON 支持旧版格式（sids 作为顶层键，值为 [ssecurity, serviceToken]）
// 以及 deviceId/userId/passToken.
func (t *Token) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if t.Services == nil {
		t.Services = make(map[string][]string)
	}
	for k, v := range raw {
		switch k {
		case "deviceId":
			if s, ok := v.(string); ok {
				t.DeviceID = s
			}
		case "userId":
			if s, ok := v.(string); ok {
				t.UserID = s
			}
		case "passToken":
			if s, ok := v.(string); ok {
				t.PassToken = s
			}
		default:
			if arr, ok := v.([]interface{}); ok && len(arr) >= 2 {
				var pair []string
				for _, a := range arr {
					if s, ok := a.(string); ok {
						pair = append(pair, s)
					}
				}
				if len(pair) >= 2 {
					t.Services[k] = pair
				}
			}
		}
	}
	return nil
}

func (t *Token) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"deviceId":  t.DeviceID,
		"userId":    t.UserID,
		"passToken": t.PassToken,
	}
	for sid, pair := range t.Services {
		m[sid] = pair
	}
	return json.Marshal(m)
}
