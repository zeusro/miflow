# ×获取小爱对话 - 实现步骤说明

## 验证结果（2025-02）

在 miflow 中尝试使用 OAuth 2.0 获取小爱对话，**无法实现**。

调用 `userprofile.mina.mi.com/device_profile/v2/conversation` 时返回：

```
{"code":601,"message":"illegal argument exception","data":"MissingRequestCookieException: Required cookie 'userId' for method parameter type Long is not present"}
```

说明该 API **必须**使用 Cookie 认证（`userId`、`deviceId`、`serviceToken`），OAuth Bearer 不被支持。

## xiaomusic/xiaogpt 实现方式

### 1. 小爱对话记录 API（userprofile.mina.mi.com）

- **URL**: `https://userprofile.mina.mi.com/device_profile/v2/conversation?source=dialogu&hardware={hardware}&timestamp={timestamp}&limit=2`
- **认证**: Cookie
  ```
  deviceId={device_id}; serviceToken={service_token}; userId={user_id}
  ```
- **来源**: MiAccount 账号密码登录，获取 `micoapi` 的 `serviceToken`、`userId`，配合 `mina_service.device_list()` 的 `deviceID`
- **代码**: [hanxi/xiaomusic/conversation.py](https://github.com/hanxi/xiaomusic/blob/main/xiaomusic/conversation.py)、[yihong0618/xiaogpt/config.py](https://github.com/yihong0618/xiaogpt/blob/main/xiaogpt/config.py)

### 2. Mina ubus nlp_result_get（api2.mina.mi.com）

- **接口**: `POST /remote/ubus`，`method=nlp_result_get`, `path=mibrain`
- **认证**: micoapi（MiAccount 账号密码）
- **适用**: 部分设备如 M01
- **代码**: [MiService/minaservice.py get_latest_ask](https://github.com/yihong0618/MiService/blob/main/miservice/minaservice.py)

## 若要在 miflow 中实现

需要增加 **MiAccount 账号密码**登录支持：

1. **MiAccount 登录**：实现类似 [MiService/miaccount.py](https://github.com/yihong0618/MiService) 的 `login("micoapi")`，获取 `~/.mi.token` 中的 `userId`、`micoapi[1]`（serviceToken）
2. **设备列表**：调用 api2.mina.mi.com 的 `device_list`（需 micoapi），拿到 `deviceID`
3. **构造 Cookie**：`deviceId={deviceID}; serviceToken={serviceToken}; userId={userId}`
4. **轮询对话**：GET `userprofile.mina.mi.com/.../conversation`，带上述 Cookie
5. **解析 records**：`data.records[0]` 含 `time`、`query`、`answers[0].tts.text`

参考项目：

- [hanxi/xiaomusic](https://github.com/hanxi/xiaomusic) - conversation.py
- [yihong0618/xiaogpt](https://github.com/yihong0618/xiaogpt) - xiaogpt.py, config.py
- [yihong0618/MiService](https://github.com/yihong0618/MiService) - miservice

## 与 OAuth 的差异

| 认证方式       | miflow 当前 | 小爱对话 API 需要 |
|----------------|-------------|-------------------|
| 登录           | OAuth 2.0   | MiAccount 账号密码 |
| Token 类型     | AccessToken | serviceToken (micoapi) |
| 设备 ID 来源   | ha.api.io.mi.com | api2.mina.mi.com device_list |
| API 域名       | ha.api.io.mi.com | userprofile.mina.mi.com, api2.mina.mi.com |

详见 [docs/xiaomusic.md](xiaomusic.md) 中的 miflow xiaomusic 限制说明。
