# m login 命令

`login` 命令实现 OAuth 2.0 登录流程，首次使用需执行以获取并保存小米账号 token。

## 入口

- **注册位置**：`pkg/cmd/cmd.go`
- **实现位置**：`pkg/cmd/login/login.go`

当用户执行 `m login` 时，创建 `login.Login` 并调用 `Run()`。`login` 是唯一在加载 token 之前执行的子命令，因此不需要先有 token。

## 核心流程

`Run()` 分为 5 步：

| 步骤 | 说明 |
|------|------|
| 1. 创建 OAuth 客户端 | `miaccount.NewOAuthClient()`，从配置读取 client_id、redirect_uri、device_id 等 |
| 2. 生成授权 URL | `GenAuthURL()` 生成小米 OAuth 授权链接，并打印给用户 |
| 3. 启动本地回调服务 | 在 `config.MiIO.CallbackPort`（默认 8123）上启动 HTTP 服务，等待授权回调 |
| 4. 打开浏览器 | `OpenAuthURL()` 尝试用系统默认浏览器打开授权页（失败则提示手动打开） |
| 5. 换取并保存 token | 收到 `code` 后调用 `GetToken()` 换取 access/refresh token，再用 `TokenStore.SaveOAuth()` 写入 `TokenPath` |

## OAuth 2.0 流程（internal/miaccount）

- **授权 URL**：`https://account.xiaomi.com/oauth2/authorize`，参数包括 `redirect_uri`、`client_id`、`response_type=code`、`device_id`、`state`、`skip_confirm=true` 等。
- **回调服务**：监听 `/callback`，从 `?code=xxx` 中取出授权码，返回“登录成功”页面，并在 5 秒后尝试关闭窗口。
- **换取 token**：向 `https://ha.api.io.mi.com/app/v2/ha/oauth/get_token` 发送 `code`，获取 `access_token` 和 `refresh_token`。
- **保存**：将 token 写入 `TokenPath`（由 `config.TokenPath` 决定）。

## 配置

- **CallbackPort**：`config.MiIO.CallbackPort`，默认 8123，需与 `redirect_uri` 中的端口一致（如 `http://homeassistant.local:8123/callback`）。
- **TokenPath**：由 `config.Get().TokenPath` 提供，用于保存 OAuth token。

## 流程图

```
用户执行 m login
    → 生成授权 URL
    → 启动本地 HTTP 服务（:8123）
    → 打开浏览器访问小米授权页
    → 用户登录并授权
    → 小米重定向到 localhost:8123/callback?code=xxx
    → 回调服务拿到 code
    → 用 code 换取 access_token / refresh_token
    → 写入 TokenPath
    → 提示 "Login successful"
```
