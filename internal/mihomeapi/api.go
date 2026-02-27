package mihomeapi

import (
	"fmt"
	"strings"

	"github.com/zeusro/miflow/internal/miaccount"
)

// Service 通过 ha.api.io.mi.com 实现小米 MIoT API（OAuth）。
// 参考: https://github.com/XiaoMi/ha_xiaomi_home
type Service struct {
	Client *miaccount.HAClient
}

// New 使用 OAuth 支持的 HAClient 创建服务。
func New(token *miaccount.OAuthToken, tokenPath string) (*Service, error) {
	if token == nil {
		return nil, fmt.Errorf("oauth token required")
	}
	store := &miaccount.TokenStore{Path: tokenPath}
	return &Service{
		Client: miaccount.NewHAClient(token, store),
	}, nil
}

// DeviceList 获取设备列表。name 按 did/name 筛选；"full" 返回完整信息。
func (s *Service) DeviceList(name string, getVirtualModel bool, getHuamiDevices int) ([]map[string]interface{}, error) {
	return s.deviceListPage(name, nil, nil, getVirtualModel, getHuamiDevices)
}

// deviceListPage 分页获取设备列表，支持按 name 筛选及 startDID 游标。
func (s *Service) deviceListPage(name string, dids []string, startDID *string, getVirtualModel bool, getHuamiDevices int) ([]map[string]interface{}, error) {
	data := map[string]interface{}{
		"limit":             200,
		"get_split_device":  true,
		"get_third_device":  true,
		"dids":              dids,
	}
	if startDID != nil {
		data["start_did"] = *startDID
	}
	res, err := s.Client.Post("/app/v2/home/device_list_page", data)
	if err != nil {
		return nil, err
	}
	result, _ := res["result"].(map[string]interface{})
	if result == nil {
		return nil, fmt.Errorf("invalid device_list_page response")
	}
	list, _ := result["list"].([]interface{})
	hasMore, _ := result["has_more"].(bool)
	nextStart, _ := result["next_start_did"].(string)

	out := make([]map[string]interface{}, 0)
	for _, it := range list {
		m, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		did, _ := m["did"].(string)
		model, _ := m["model"].(string)
		if strings.HasPrefix(did, "miwifi.") {
			continue
		}
		if name == "full" {
			out = append(out, m)
			continue
		}
		n, _ := m["name"].(string)
		if name != "" && !strings.Contains(did, name) && !strings.Contains(n, name) {
			continue
		}
		out = append(out, map[string]interface{}{
			"name":   m["name"],
			"model":  model,
			"did":    did,
			"token":  m["token"],
		})
	}
	if hasMore && nextStart != "" {
		more, err := s.deviceListPage(name, dids, &nextStart, getVirtualModel, getHuamiDevices)
		if err != nil {
			return out, nil
		}
		out = append(out, more...)
	}
	return out, nil
}

// GetProps 获取 MIoT 属性。
func (s *Service) GetProps(params []map[string]interface{}) ([]map[string]interface{}, error) {
	res, err := s.Client.Post("/app/v2/miotspec/prop/get", map[string]interface{}{
		"datasource": 1,
		"params":     params,
	})
	if err != nil {
		return nil, err
	}
	result, _ := res["result"]
	arr, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid prop/get response")
	}
	out := make([]map[string]interface{}, len(arr))
	for i, it := range arr {
		if m, ok := it.(map[string]interface{}); ok {
			out[i] = m
		}
	}
	return out, nil
}

// SetProps 设置 MIoT 属性。
func (s *Service) SetProps(params []map[string]interface{}) ([]map[string]interface{}, error) {
	res, err := s.Client.Post("/app/v2/miotspec/prop/set", map[string]interface{}{
		"params": params,
	})
	if err != nil {
		return nil, err
	}
	result, _ := res["result"]
	arr, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid prop/set response")
	}
	out := make([]map[string]interface{}, len(arr))
	for i, it := range arr {
		if m, ok := it.(map[string]interface{}); ok {
			out[i] = m
		}
	}
	return out, nil
}

// GetHome 获取家庭、房间、设备结构。参考 ha_xiaomi_home homeroom/gethome。
func (s *Service) GetHome() ([]map[string]interface{}, error) {
	data := map[string]interface{}{
		"limit":           150,
		"fetch_share":     true,
		"fetch_share_dev": true,
		"plat_form":       0,
		"app_ver":         9,
	}
	res, err := s.Client.Post("/app/v2/homeroom/gethome", data)
	if err != nil {
		return nil, err
	}
	result, _ := res["result"].(map[string]interface{})
	if result == nil {
		return nil, fmt.Errorf("invalid gethome response")
	}
	out := make([]map[string]interface{}, 0)
	for _, key := range []string{"homelist", "share_home_list"} {
		list, _ := result[key].([]interface{})
		for _, it := range list {
			m, ok := it.(map[string]interface{})
			if !ok {
				continue
			}
			if _, hasID := m["id"]; !hasID {
				continue
			}
			if _, hasName := m["name"]; !hasName {
				continue
			}
			if _, hasRoom := m["roomlist"]; !hasRoom {
				continue
			}
			out = append(out, m)
		}
	}
	return out, nil
}

// Action 执行 MIoT 动作。
func (s *Service) Action(did string, siid, aiid int, in []interface{}) (map[string]interface{}, error) {
	res, err := s.Client.Post("/app/v2/miotspec/action", map[string]interface{}{
		"params": map[string]interface{}{
			"did":  did,
			"siid": siid,
			"aiid": aiid,
			"in":   in,
		},
	})
	if err != nil {
		return nil, err
	}
	result, _ := res["result"]
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, nil
}
