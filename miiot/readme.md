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

## ./m list 设备技术说明

以下为 `./m list` 中所有型号的产品页与技术说明 (Spec) 对应关系，技术说明 URL 从产品页 class="divider" 下的规格链接抓取。

| model | 产品页 | 技术说明 (Spec URL) |
|-------|--------|---------------------|
| babai.plug.sk01a | [链接](https://home.miot-spec.com/s/babai.plug.sk01a) | [链接](https://home.miot-spec.com/spec?type=urn%3Amiot-spec-v2%3Adevice%3Aoutlet%3A0000A002%3Ababai-sk01a%3A1%3A0000C816) |
| bean.switch.bln31 | [链接](https://home.miot-spec.com/s/bean.switch.bln31) | [链接](https://home.miot-spec.com/spec?type=urn%3Amiot-spec-v2%3Adevice%3Aswitch%3A0000A003%3Abean-bln31%3A1%3A0000C808) |
| bean.switch.bln33 | [链接](https://home.miot-spec.com/s/bean.switch.bln33) | [链接](https://home.miot-spec.com/spec?type=urn%3Amiot-spec-v2%3Adevice%3Aswitch%3A0000A003%3Abean-bln33%3A1%3A0000C810) |
| chuangmi.plug.m3 | [链接](https://home.miot-spec.com/s/chuangmi.plug.m3) | [链接](https://home.miot-spec.com/spec?type=urn%3Amiot-spec-v2%3Adevice%3Aoutlet%3A0000A002%3Achuangmi-m3%3A1) |
| chuangmi.plug.v3 | [链接](https://home.miot-spec.com/s/chuangmi.plug.v3) | [链接](https://home.miot-spec.com/spec?type=urn%3Amiot-spec-v2%3Adevice%3Aoutlet%3A0000A002%3Achuangmi-v3%3A1) |
| giot.light.v5ssm | [链接](https://home.miot-spec.com/s/giot.light.v5ssm) | [链接](https://home.miot-spec.com/spec?type=urn%3Amiot-spec-v2%3Adevice%3Alight%3A0000A001%3Agiot-v5ssm%3A1%3A0000C802) |
| lemesh.switch.sw3f13 | [链接](https://home.miot-spec.com/s/lemesh.switch.sw3f13) | [链接](https://home.miot-spec.com/spec?type=urn%3Amiot-spec-v2%3Adevice%3Aswitch%3A0000A003%3Alemesh-sw3f13%3A1%3A0000C810) |
| linp.sensor_occupy.hb01 | [链接](https://home.miot-spec.com/s/linp.sensor_occupy.hb01) | [链接](https://home.miot-spec.com/spec?type=urn%3Amiot-spec-v2%3Adevice%3Aoccupancy-sensor%3A0000A0BF%3Alinp-hb01%3A1%3A0000C824) |
| opple.light.bydceiling | [链接](https://home.miot-spec.com/s/opple.light.bydceiling) | [链接](https://home.miot-spec.com/spec?type=urn%3Amiot-spec-v2%3Adevice%3Alight%3A0000A001%3Aopple-bydceiling%3A1) |
| xiaomi.tv.eanfv1 | [链接](https://home.miot-spec.com/s/xiaomi.tv.eanfv1) | [链接](https://home.miot-spec.com/spec?type=urn%3Amiot-spec-v2%3Adevice%3Atelevision%3A0000A010%3Axiaomi-eanfv1%3A1) |
| xiaomi.wifispeaker.l05b | [链接](https://home.miot-spec.com/s/xiaomi.wifispeaker.l05b) | [链接](https://home.miot-spec.com/spec?type=urn%3Amiot-spec-v2%3Adevice%3Aspeaker%3A0000A015%3Axiaomi-l05b%3A1) |
| xiaomi.wifispeaker.l05c | [链接](https://home.miot-spec.com/s/xiaomi.wifispeaker.l05c) | [链接](https://home.miot-spec.com/spec?type=urn%3Amiot-spec-v2%3Adevice%3Aspeaker%3A0000A015%3Axiaomi-l05c%3A1) |
| xiaomi.wifispeaker.oh2 | [链接](https://home.miot-spec.com/s/xiaomi.wifispeaker.oh2) | [链接](https://home.miot-spec.com/spec?type=urn%3Amiot-spec-v2%3Adevice%3Aspeaker%3A0000A015%3Axiaomi-oh2%3A1) |

> 表格可通过 `go run ./cmd/scrape-specs` 重新生成（需在项目根目录执行，依赖 `./m list`）。

## 使用

```bash
# 验证 m list 中所有型号的 SpecURL
go run ./cmd/miiot

# 重新爬取并生成设备技术说明表格（输出可追加到本文档）
go run ./cmd/scrape-specs
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