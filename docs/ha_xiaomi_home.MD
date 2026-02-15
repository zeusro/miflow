# ha_xiaomi_home 集成的米家 API

参考 [ha_xiaomi_home](https://github.com/XiaoMi/ha_xiaomi_home) 官方集成，整理项目中集成的米家相关 API。

## 一、OAuth 2.0 认证

| 接口 | 用途 |
|------|------|
| `https://account.xiaomi.com/oauth2/authorize` | 小米账号 OAuth 授权页面 |
| `https://{region}.ha.api.io.mi.com/app/v2/ha/oauth/get_token` | 用 code 换取 access_token / refresh_token |
| `https://open.account.xiaomi.com/user/profile` | 获取用户信息（昵称等） |

## 二、HTTP API（ha.api.io.mi.com）

所有 HTTP 接口均通过 `https://{region}.ha.api.io.mi.com` 调用，区域包括：`cn`、`de`、`i2`、`ru`、`sg`、`us`。

### 1. 家庭与设备

| 接口 | 方法 | 用途 |
|------|------|------|
| `/app/v2/homeroom/gethome` | POST | 获取家庭、房间、设备列表（含共享家庭） |
| `/app/v2/homeroom/get_dev_room_page` | POST | 分页获取设备与房间关系 |
| `/app/v2/home/device_list_page` | POST | 分页获取设备详情（did、model、spec_type、token 等） |

### 2. MIoT 设备控制

| 接口 | 方法 | 用途 |
|------|------|------|
| `/app/v2/miotspec/prop/get` | POST | 读取设备属性（siid、piid） |
| `/app/v2/miotspec/prop/set` | POST | 设置设备属性 |
| `/app/v2/miotspec/action` | POST | 执行设备动作（siid、aiid、in 参数） |

### 3. 本地网关（中国区）

| 接口 | 方法 | 用途 |
|------|------|------|
| `/app/v2/ha/oauth/get_central_crt` | POST | 获取本地中枢网关证书（用于本地 MQTT） |

## 三、MQTT 消息总线（MIPS）

通过 MQTT 订阅设备属性变化和事件，支持云端和本地两种模式。

### 云端模式

- **Broker**：`{region}-ha.mqtt.io.mi.com:8883`（TLS）
- **用途**：订阅设备属性变化、事件上报，无需轮询即可收到设备状态更新

### 本地模式（中国区）

- **Broker**：小米中枢网关内置 MQTT Broker
- **用途**：通过 `miot_mdns` 发现网关，使用 `get_central_crt` 获取证书后直连本地 MQTT

### MIPS 协议能力

- 订阅属性（`sub_prop`）：设备属性变化推送
- 订阅事件（`sub_event`）：设备事件推送
- 发布属性读写、动作执行等控制命令（与 HTTP 接口功能对应）

## 四、MIoT-Spec-V2 规格

- **规格来源**：`http://miot-spec.org/miot-spec-v2/instances`、`http://miot-spec.org/miot-spec-v2/instance?type=...`
- **用途**：解析设备 URN、服务、属性、事件、动作，并映射为 Home Assistant 实体

## 五、LAN 控制（可选）

- **模块**：`miot_lan`、`miot_mdns`
- **用途**：局域网内发现和控制 IP 设备（WiFi / 有线），不依赖云端
- **限制**：仅支持 IP 设备，不支持 BLE Mesh、ZigBee 等；有中枢网关时会被禁用

## 六、整体架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                    ha_xiaomi_home 集成架构                        │
├─────────────────────────────────────────────────────────────────┤
│  OAuth 2.0  →  access_token  →  HTTP API + MQTT                   │
│                                                                   │
│  HTTP API (ha.api.io.mi.com):                                    │
│    • homeroom/gethome, get_dev_room_page  → 家庭/房间/设备结构     │
│    • home/device_list_page                → 设备详情             │
│    • miotspec/prop/get, prop/set, action  → 属性读写、动作执行    │
│    • ha/oauth/get_central_crt             → 本地网关证书          │
│                                                                   │
│  MQTT (云端: *-ha.mqtt.io.mi.com / 本地: 中枢网关):               │
│    • 订阅设备属性变化、事件                                         │
│    • 发布控制命令                                                  │
│                                                                   │
│  MIoT-Spec-V2: 设备规格解析与 HA 实体映射                          │
└─────────────────────────────────────────────────────────────────┘
```

## 七、与 miflow 的对应关系

| 功能 | ha_xiaomi_home | miflow |
|------|----------------|--------|
| 登录方式 | OAuth 2.0 | OAuth 2.0 |
| API 域名 | ha.api.io.mi.com | ha.api.io.mi.com |
| 设备列表 | device_list_page | m list |
| 属性读写 | miotspec/prop | m siid,piid=val |
| 动作执行 | miotspec/action | m siid-aiid args |
| 小爱播报 | Execute Text Directive | m message |

## 八、MQTT 定义

### 什么是 MQTT

**MQTT**（Message Queuing Telemetry Transport，消息队列遥测传输）是一种轻量级、基于发布/订阅模式的消息传输协议，专为带宽有限、延迟较高、网络不稳定的物联网环境设计。

### 核心特性

| 特性 | 说明 |
|------|------|
| 轻量 | 控制消息最小仅 2 字节，头部开销小 |
| 发布/订阅 | 发布者与订阅者通过主题（Topic）解耦，无需直接连接 |
| QoS | 支持 0（至多一次）、1（至少一次）、2（精确一次）三种服务质量 |
| 安全 | 支持 TLS 加密、用户名/密码认证 |

### 在 ha_xiaomi_home 中的角色

- **Broker 主机**：`ha.mqtt.io.mi.com`，按区域前缀为 `{region}-ha.mqtt.io.mi.com`（如 `cn-ha.mqtt.io.mi.com`）
- **端口**：8883（TLS）
- **Keepalive**：60 秒
- **协议**：MIPS（MIoT Pub/Sub），基于 MQTT 的扩展，用于设备属性与事件的订阅/发布

### 消息流向

```
设备属性变化/事件 → 设备上报 → MIoT Cloud / 中枢网关 → MQTT Broker
                                                         ↓
                                              ha_xiaomi_home 订阅 Topic
                                                         ↓
                                              Home Assistant 实体更新

控制命令 → ha_xiaomi_home 发布 → MQTT Broker → 转发至设备
```
