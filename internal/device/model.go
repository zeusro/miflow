package device

import (
	"github.com/zeusro/miflow/internal/miioservice"
)

// Device 表示 m list 导出的接入设备。
// 与 m list 的 JSON 输出字段一致。
type Device struct {
	DID   string `json:"did"`   // 设备 ID
	Model string `json:"model"` // 型号，用于查询 miot-spec.org SPEC
	Name  string `json:"name"`  // 设备名称
	Token string `json:"token"` // 设备 token
}

// API 封装接入设备的操作，基于 m list 设备列表与 docs/spec.md 的 SPEC 查询流程。
type API struct {
	io *miioservice.Service
}

// ModelSpec 表示单个型号的 MIoT SPEC，按 docs/spec.md 从 miot-spec.org 获取。
type ModelSpec struct {
	Type        string        `json:"type"`
	Description string        `json:"description"`
	Services    []ServiceSpec `json:"services"`
}

// ServiceSpec 表示 SPEC 中的服务（siid）。
type ServiceSpec struct {
	IID         int          `json:"iid"`
	Description string       `json:"description"`
	Type        string       `json:"type"`
	Properties  []PropSpec    `json:"properties"`
	Actions     []ActionSpec `json:"actions"`
	Events      []EventSpec  `json:"events"`
}

// PropSpec 表示属性（piid）。
type PropSpec struct {
	IID         int      `json:"iid"`
	Description string   `json:"description"`
	Format      string   `json:"format"`
	Access      []string `json:"access"`
}

// ActionSpec 表示动作（aiid）。
type ActionSpec struct {
	IID         int           `json:"iid"`
	Description string        `json:"description"`
	In          []interface{} `json:"in"`
	Out         []interface{} `json:"out"`
}

// EventSpec 表示事件（eid）。
type EventSpec struct {
	IID         int           `json:"iid"`
	Description string        `json:"description"`
	Arguments   []interface{} `json:"arguments"`
}
