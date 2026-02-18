package miiot

import (
	"fmt"
	"sync"

	"github.com/zeusro/miflow/miiot/specs"
)

// Models 为 m list 中所有型号，与 home.miot-spec.com 1:1 匹配。
var Models = []string{
	"babai.plug.sk01a", "bean.switch.bln31", "bean.switch.bln33",
	"chuangmi.plug.m3", "chuangmi.plug.v3", "giot.light.v5ssm",
	"lemesh.switch.sw3f13", "linp.sensor_occupy.hb01",
	"opple.light.bydceiling", "xiaomi.tv.eanfv1",
	"xiaomi.wifispeaker.l05b", "xiaomi.wifispeaker.l05c", "xiaomi.wifispeaker.oh2",
}

// ModelAPI 表示单个型号的规格 API。
type ModelAPI interface {
	Model() string
	SpecURL() (string, error)
	ProductURL() string
}

var (
	registry   = make(map[string]ModelAPI)
	registryMu sync.RWMutex
	initOnce   sync.Once
)

func initRegistry() {
	initOnce.Do(func() {
		for _, model := range Models {
			Register(NewModelAPI(model))
		}
	})
}

// Register 注册型号的 API。
func Register(api ModelAPI) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[api.Model()] = api
}

// Get 获取型号的 API。
func Get(model string) (ModelAPI, bool) {
	initRegistry()
	registryMu.RLock()
	defer registryMu.RUnlock()
	api, ok := registry[model]
	return api, ok
}

// All 返回所有已注册的型号。
func All() []string {
	initRegistry()
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]string, 0, len(registry))
	for m := range registry {
		out = append(out, m)
	}
	return out
}

// baseModelAPI 为通用实现，使用 specs 包获取 URL。
type baseModelAPI struct {
	model string
}

func (b *baseModelAPI) Model() string {
	return b.model
}

func (b *baseModelAPI) SpecURL() (string, error) {
	return specs.SpecURL(b.model)
}

func (b *baseModelAPI) ProductURL() string {
	return specs.ProductURLForModel(b.model)
}

// NewModelAPI 创建基础 ModelAPI，仅包含 model/SpecURL/ProductURL。
func NewModelAPI(model string) ModelAPI {
	return &baseModelAPI{model: model}
}

// SpecURL 便捷函数：获取型号的规格页 URL。
func SpecURL(model string) (string, error) {
	return specs.SpecURL(model)
}

// ProductURL 便捷函数：获取型号的产品页 URL。
func ProductURL(model string) string {
	return specs.ProductURLForModel(model)
}

// URN 便捷函数：获取型号的 URN。
func URN(model string) (string, error) {
	return specs.URN(model)
}

// MustLoad 加载 model->URN 映射，失败时 panic。
func MustLoad() map[string]string {
	m, err := specs.Load()
	if err != nil {
		panic(fmt.Sprintf("miiot: load specs: %v", err))
	}
	return m
}
