# cmd 模块说明

miflow 的命令行入口与 HTTP 服务，按功能分为核心命令与工具命令。

## 核心命令

### m

小米云服务 CLI（OAuth 2.0），整合 MiIO/MIoT 设备控制与 Mina 小爱音箱。

- **认证**：`m login` 完成 OAuth 授权，token 存于 `~/.mi.token`
- **Mina**：`mina`、`message`、`play`、`pause`、`loop`、`play_list`、`suno` 等
- **MiIO/MIoT**：`list`、`spec`、`spec_all`、`decode`、属性读写、动作调用、原始 `/uri` 调用

设备通过 `config default_did` 或环境变量 `MI_DID` 指定。

### flow

可视化控制流服务，提供 HTTP API 与拖拽式 Web UI。

- 步骤类型：`tts`、`play_url`、`miio`、`delay`
- 配置存于 `flowdata/`，支持创建、更新、删除、执行
- 默认监听 `:18090`

### web

Web 服务，基于 GoFrame，提供 OAuth 登录、设备管理、工作流管理。

- `/login`、`/callback`：OAuth 2.0 授权流程
- `/api/devices`：设备列表、规格、控制
- `/api/workflows`：工作流 CRUD 与执行
- 默认监听 `:8123`，与 `oauth.redirect_uri` 一致

### mp3

HTTP 文件服务，将本地路径映射为可访问 URL。

- 根目录 `/` 映射，支持完整路径（含空格等）
- 输出 `http://本机IP:端口/路径` 供局域网设备访问
- 默认监听 `:8090`，`-host` 可指定本机 IP

## 工具命令

### miiot

验证 `m list` 中所有型号与 home.miot-spec.com 的 Spec 1:1 匹配。

- 需先执行 `m login`
- 输出 JSON：`ok` / `failed` 按 model 分类

### scrape-specs

从 `m list` 获取设备列表，爬取各型号技术说明 URL，输出 Markdown 表格。

- 依赖 `m` 可执行文件（当前目录或 PATH）
- 输出：`| model | 产品页 | 技术说明 (Spec URL) |`

## 构建

```bash
make b
```

构建产物：`m`、`mp3`、`flow`、`miflow-web`（web 命令）。

工具命令 `miiot`、`scrape-specs` 可按需单独构建：

```bash
go build -o miiot ./cmd/miiot
go build -o scrape-specs ./cmd/scrape-specs
```
