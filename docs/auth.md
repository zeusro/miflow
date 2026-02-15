# Xiaomi Home 集成 HA 流程与模拟

参考 [ha_xiaomi_home](https://github.com/XiaoMi/ha_xiaomi_home) 官方集成，介绍 Xiaomi 集成 Home Assistant 的完整流程及 miflow 的模拟实现。

## 一、整体架构

Xiaomi Home 集成 HA 的核心组件：

1. **OAuth 2.0 登录**：小米账号授权，获取 access_token / refresh_token
2. **HTTP API**：通过 `ha.api.io.mi.com` 控制设备、获取设备列表
3. **MQTT 消息**：订阅设备属性变化和事件（云端或本地网关）
4. **MIoT-Spec-V2**：设备规格与 HA 实体的映射

## 二、完整流程

### 1. OAuth 2.0 登录

```
用户 → 点击登录 → 生成 auth_url → 跳转小米 OAuth 页面
→ 用户授权 → 回调 redirect_uri?code=xxx&state=xxx
→ 用 code 换取 access_token / refresh_token
```

**关键参数：**

| 参数 | 值 |
|------|-----|
| Client ID | `2882303761520251711` |
| Auth URL | `https://account.xiaomi.com/oauth2/authorize` |
| Token API | `https://{region}.ha.api.io.mi.com/app/v2/ha/oauth/get_token` |
| Redirect URI | HA 用 `http://homeassistant.local:8123/api/webhook/{id}` |

**Token 换取：**

- 用 code 换 token：`GET /app/v2/ha/oauth/get_token?data={json}`
- 用 refresh_token 刷新：`data` 中传 `refresh_token` 而非 `code`

### 2. HTTP API 调用

拿到 access_token 后，请求头为：

```
Host: {region}.ha.api.io.mi.com
X-Client-BizId: haapi
Authorization: Bearer{access_token}
X-Client-AppId: 2882303761520251711
Content-Type: application/json
```

**常用接口：**

| 接口 | 方法 | 用途 |
|------|------|------|
| `/app/v2/homeroom/gethome` | POST | 获取家庭、房间、设备列表 |
| `/app/v2/home/device_list_page` | POST | 分页获取设备详情 |
| `/app/v2/miotspec/prop/get` | POST | 读取属性 |
| `/app/v2/miotspec/prop/set` | POST | 设置属性 |
| `/app/v2/miotspec/action` | POST | 执行动作 |
| `/app/v2/ha/oauth/get_central_crt` | POST | 获取本地网关证书（中国区） |

### 3. 设备控制模式

- **云端**：HTTP 控制 + MQTT 订阅（`{region}-ha.mqtt.io.mi.com:8883`）
- **本地**：通过小米中枢网关的 MQTT Broker，或 LAN 直连（仅 IP 设备）

## 三、miflow 模拟实现

### 1. OAuth 流程（`internal/miaccount/oauth.go`）

```go
const (
    OAuth2ClientID   = "2882303761520251711"
    OAuth2AuthURL    = "https://account.xiaomi.com/oauth2/authorize"
    OAuth2APIHost    = "ha.api.io.mi.com"
    OAuth2TokenPath  = "/app/v2/ha/oauth/get_token"
    DefaultCloudSvr = "cn"
    TokenExpireRatio = 0.7
)
```

- `GenAuthURL`：生成授权 URL
- `GetToken`：用 code 换 token
- `RefreshToken`：用 refresh_token 刷新
- `ServeCallback`：本地 HTTP 服务接收回调，拿到 code

### 2. HTTP 客户端（`internal/miaccount/haclient.go`）

```go
func (c *HAClient) setHeaders(req *http.Request) {
    req.Header.Set("Host", c.Host)
    req.Header.Set("X-Client-BizId", "haapi")
    req.Header.Set("Authorization", "Bearer"+c.AccessToken)
    req.Header.Set("X-Client-AppId", c.ClientID)
}
```

与 ha_xiaomi_home 的 `MIoTHttpClient` 一致。

### 3. MIoT API（`internal/mihomeapi/api.go`）

已实现：

- `DeviceList` → `/app/v2/home/device_list_page`
- `GetProps` → `/app/v2/miotspec/prop/get`
- `SetProps` → `/app/v2/miotspec/prop/set`
- `Action` → `/app/v2/miotspec/action`

## 四、模拟流程建议

1. **登录**：`m login` 或类似命令
   - 生成 auth_url
   - 打开浏览器让用户授权
   - 本地起 HTTP 服务接收 code
   - 调用 `GetToken` 获取 token 并持久化

2. **设备列表**：
   - 使用 `mihomeapi.Service.DeviceList`
   - 对应 HA 的 `get_homeinfos_async` + `get_devices_with_dids_async`

3. **属性读写**：
   - `GetProps` / `SetProps`，参数格式与 HA 一致

4. **动作执行**：
   - `Action(did, siid, aiid, in)`

5. **可选扩展**：
   - 实现 `get_homeinfos_async`（`/app/v2/homeroom/gethome`）获取家庭/房间结构
   - 实现 MQTT 订阅以接收设备上报（类似 `miot_mips.py`）
   - 中国区本地网关：实现证书申请和本地 MQTT 连接

## 五、与 HA 集成的对应关系

| HA 组件 | miflow 对应 |
|---------|-------------|
| `MIoTOauthClient` | `miaccount.OAuthClient` |
| `MIoTHttpClient` | `miaccount.HAClient` |
| `get_devices_async` | `mihomeapi.DeviceList` + `homeroom/gethome` |
| `get_props_async` / `set_prop_async` | `mihomeapi.GetProps` / `SetProps` |
| `action_async` | `mihomeapi.Action` |

## 六、区域支持

| 区域代码 | 说明 |
|----------|------|
| cn | 中国大陆 |
| de | Europe |
| i2 | India |
| ru | Russia |
| sg | Singapore |
| us | United States |

不同区域使用不同的 API 域名，如 `cn.ha.api.io.mi.com`、`us.ha.api.io.mi.com` 等。
