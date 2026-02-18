# miiot - 产品规格 API

按 `./m list` 的 model 与 [home.miot-spec.com](https://home.miot-spec.com/s/) 1:1 匹配，实现各型号的规格 API。

## 数据来源

model→URN 映射从 miot-spec.org API 获取（与 home.miot-spec.com 规格页同源）：
- `http://miot-spec.org/miot-spec-v2/instances?status=all`
- 规格页 URL：`https://home.miot-spec.com/spec?type={urn}`

## 文件规则

model 格式 `vendor.category.suffix` → 路径 `miiot/vendor/category/suffix.go`

| model | 文件 |
|-------|------|
| xiaomi.wifispeaker.oh2 | miiot/xiaomi/wifispeaker/oh2.go |
| bean.switch.bln31 | miiot/bean/switch/bln31.go |
| chuangmi.plug.m3 | miiot/chuangmi/plug/m3.go |

## 使用

```bash
# 验证 m list 中所有型号的 SpecURL
go run ./cmd/miiot
```

## API

- `miiot.SpecURL(model)` - 规格页 URL
- `miiot.ProductURL(model)` - 产品页 URL (home.miot-spec.com/s/{model})
- `miiot.Get(model)` - 获取 ModelAPI

## 属性与动作控制 (miiot/ctrl)

```go
import "github.com/zeusro/miflow/miiot/ctrl"

c := ctrl.New(deviceAPI)
c.SetOn(did, model, true)           // 开关/插座/灯
c.GetOn(did, model)
c.Toggle(did, model)
c.SetBrightness(did, model, 80)      // 灯光
c.GetBrightness(did, model)
c.TTS(did, model, "你好")             // 音箱
c.SetVolume(did, model, 50)
c.Play(did, model) / c.Pause(did, model)
c.TVTurnOff(did, model)              // 电视
c.GetOccupancy(did, model)           // 人体传感器
c.SetSwitchChannel(did, model, 0, true)  // 多通道开关
```