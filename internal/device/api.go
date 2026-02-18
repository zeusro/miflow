package device

import (
	"fmt"
	"strings"

	"github.com/zeusro/miflow/internal/miioservice"
)

// NewAPI 创建设备 API，需要已初始化的 miioservice.Service。
func NewAPI(io *miioservice.Service) *API {
	return &API{io: io}
}

// List 列出设备，name 为空时返回全部，否则按 did/name 模糊匹配。
// getVirtualModel、getHuamiDevices 与 m list 参数一致。
func (a *API) List(name string, getVirtualModel bool, getHuamiDevices int) ([]*Device, error) {
	raw, err := a.io.DeviceList(name, getVirtualModel, getHuamiDevices)
	if err != nil {
		return nil, err
	}
	out := make([]*Device, 0, len(raw))
	for _, m := range raw {
		if d := FromMap(m); d != nil && d.DID != "" {
			out = append(out, d)
		}
	}
	return out, nil
}

// Get 按 did 或 name 获取单个设备，未找到返回 nil 和错误。
func (a *API) Get(didOrName string) (*Device, error) {
	if didOrName == "" {
		return nil, fmt.Errorf("device: did or name required")
	}
	list, err := a.List(didOrName, false, 0)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("device not found: %s", didOrName)
	}
	return list[0], nil
}

// Spec 按设备型号查询 MIoT SPEC，遵循 docs/spec.md 流程：
// 1. 从 miot-spec.org/instances 获取 model→URN 映射
// 2. 用 URN 请求 miot-spec.org/instance 获取完整 SPEC
// typ 可为 model 关键词或完整 URN；format 为 text|python|json。
func (a *API) Spec(typ, format string) (interface{}, error) {
	return a.io.MiotSpec(typ, format)
}

// SpecForDevice 获取指定设备的 SPEC，使用其 model。
func (a *API) SpecForDevice(d *Device, format string) (interface{}, error) {
	if d == nil || d.Model == "" {
		return nil, fmt.Errorf("device: model required")
	}
	return a.Spec(d.Model, format)
}

// GetProps 获取 MIoT 属性，iids 为 [siid, piid] 对。
func (a *API) GetProps(did string, iids [][2]int) ([]interface{}, error) {
	return a.io.MiotGetProps(did, iids)
}

// SetProps 设置 MIoT 属性，props 为 [siid, piid, value] 三元组。
func (a *API) SetProps(did string, props [][3]interface{}) ([]int, error) {
	return a.io.MiotSetProps(did, props)
}

// Action 执行 MIoT 动作。
func (a *API) Action(did string, siid, aiid int, in []interface{}) (int, error) {
	return a.io.MiotAction(did, siid, aiid, in)
}

// ResolveDID 将 name 解析为 did，若已是纯数字 did 则原样返回。
func (a *API) ResolveDID(didOrName string) (string, error) {
	if didOrName == "" {
		return "", fmt.Errorf("device: did or name required")
	}
	if isDigits(didOrName) {
		return didOrName, nil
	}
	d, err := a.Get(didOrName)
	if err != nil {
		return "", err
	}
	return d.DID, nil
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}

// MatchName 检查设备名称或 did 是否包含关键词（不区分大小写）。
func MatchName(d *Device, keyword string) bool {
	if d == nil || keyword == "" {
		return false
	}
	k := strings.ToLower(keyword)
	return strings.Contains(strings.ToLower(d.Name), k) || strings.Contains(d.DID, k)
}
