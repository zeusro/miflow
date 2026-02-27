# 改动

## 房间页支持多家庭层级展示

2026-02-27

- 更新 rooms.html：按「家庭 → 房间 → 设备」层级结构展示
- 新增家庭 Tab（全部、家庭1、家庭2…），房间 Tab 根据所选家庭过滤
- 主内容区：每个家庭为独立 section，下列房间分区，每房间下为设备卡片网格
- 多家庭时房间 Tab 显示「房间名 (家庭名)」以区分
- 后端 GET /api/rooms 已支持多家庭，无需新增接口

## 房间页智能家居控制面板全面升级

2026-02-27

- 重设计 rooms.html 为现代化智能家居控制面板，参考 Tailwind Showcase、TailAdmin、Flowbite 等
- 设计：深色/浅色模式切换、柔和渐变、玻璃态、圆润卡片、微交互 hover、响应式网格（桌面 4 列、平板 2 列、手机 1 列）
- 设备卡片：5:4 矩形、头部（图标 48-64px、名称、状态徽章 Online/Offline）、主体（亮度/音量/状态数值）、底部（Toggle、亮度滑块）
- 房间 Tab 切换、多选批量开关（全部开启/全部关闭）
- 骨架屏加载、空状态插图
- 新增 GET /api/rooms/:roomId/devices?homeId= 按房间返回设备（含 status）
- 新增 PATCH /api/devices/:id/control（action: toggle/set_on/set_brightness/set_volume）
- Tailwind 增加 @custom-variant dark 支持 class 切换暗色模式
- 更新 openAPI3.yaml：RoomDevice、DeviceControlPatchRequest

## 房间页设备卡片与状态控制

2026-02-27

- 更新 `rooms.html`：设备以矩形卡片展示，支持状态显示与控制
- 设备卡片：矩形布局、DM Sans 字体、渐变背景、悬停阴影、状态徽章（开/关）
- 控制逻辑：支持开关的设备显示 Toggle 按钮，点击切换状态并实时更新
- 新增 `GET /api/devices/{id}/status`：获取设备状态（on/brightness/volume/mute/occupancy、supported 能力列表）
- 新增 `POST /api/devices/{id}/status`：设置设备状态（on/brightness/volume/mute）
- 使用 `miiot/ctrl` Controller 根据设备型号解析 MIoT 规格并读写属性
- 更新 openAPI3.yaml：DeviceStatus、DeviceStatusInput schema

## 登录成功跳转房间与设备页

2026-02-27

- OAuth 回调成功后改为重定向到 `/rooms` 页面，不再显示倒计时关闭页
- 新增 `/rooms` 页面：展示所有家庭、房间及对应设备
- 新增 `/api/rooms` 接口：调用 homeroom/gethome 获取家庭房间结构，返回按房间分组的设备列表
- 新增 `mihomeapi.GetHome()`、`miioservice.GetHome()`、`device.API.RoomsWithDevices()`
- 新增 `web.App.RefreshToken()`：登录后重新加载 token 并初始化 deviceAPI
- 新增 i18n：`web.rooms.title`、`web.rooms.other`
- 更新 openAPI3.yaml：增加 /api/rooms 及 HomeWithRooms、RoomWithDevices schema

## Cobra 命令行架构重构

2026-02-27

- 使用 Cobra 重构整个项目的命令行架构
- 新增依赖：`github.com/spf13/cobra`
- 新增统一入口 `miflow`（`cmd/miflow`），支持子命令：
  - `miflow m`：小米云 CLI（MiIO/MIoT/Mina）
  - `miflow flow`：Flow 可视化控制流服务
  - `miflow web`：Web 服务（OAuth 登录 + 设备管理）
  - `miflow miiot`：MiIoT 规格校验
  - `miflow mp3`：本地音乐 HTTP 映射
  - `miflow scrape-specs`：爬取设备规格 URL 表格
- 新增 `pkg/cli` 包：root、m、flow、web、miiot、mp3、scrape 子命令
- 新增 `internal/flowserver` 包：Flow HTTP 服务逻辑与内嵌 HTML，供 `cmd/flow` 与 `pkg/cli/flow` 共用
- 向后兼容：`cmd/m` 仍可构建为 `m`，内部调用 `cli.Execute`；`cmd/flow` 改为使用 `flowserver.Run`
- Makefile 新增 `miflow` 构建目标
- Cobra 能力：子命令结构、自动 `--help`、`miflow completion bash/zsh` 补全、统一 flag 解析

## LikeC4 项目架构图

2026-02-27

- 按 oracle.md 规范，使用 LikeC4 DSL 生成项目架构，导出到 likec4.c4
- 模型包含：用户、miflow 系统（m CLI、Web、Flow、mp3、miiot、scrape-specs）、内部组件（config、miaccount、miioservice、device、miiot）、外部依赖（Xiaomi OAuth、Xiaomi Home API、MIoT Spec）
- 定义 3 个视图：index（全架构）、commands（命令入口）、external（外部依赖）
- 导出 PNG 至 `docs/architecture/`，README.md 增加架构图展示
- 参考：https://github.com/likec4/likec4

## Web API OpenAPI 3 文档

2026-02-27

- 按 oracle.md 规范，整理 web/api 接口，生成 openAPI3.yaml
- 文档覆盖：设备域（/api/devices 列表、详情、spec、control）与工作流域（/api/workflows 增删改查、run）
- 遵循 OpenAPI 3.0.3 标准，包含请求/响应 schema、认证说明、错误码

## 移除 xiaomusic 命令

2026-02-27

- 删除 `cmd/xiaomusic/main.go` 及 `pkg/cmd/xiaomusic` 包（play-url、play-file 子命令）
- 配置：`XiaomusicConfig` 改为 `Mp3Config`，仅保留 mp3 命令所需的 addr、host
- Makefile 移除 xiaomusic 构建目标
- `cmd/mp3` 改为使用 `cfg.Mp3` 作为默认配置

## OAuth /callback 8123 程序 token 响应与存储修复

2026-02-25

- 修复 Web 端 `/callback` 未正确响应和储存 token 的问题
- 根因：`handleLogin` 与 `handleCallback` 各自创建新的 `OAuthClient`，导致 device_id 不一致；小米 auth code 与发起授权时的 device_id 绑定，换 token 时 device_id 不匹配会失败
- 新增 `pendingOAuth` 缓存：以 state 为 key 存储 login 时的 `OAuthClient`
- `handleLogin`：创建 OAuthClient 后存入 `pendingOAuth[state]`，并清理 10 分钟前的旧记录
- `handleCallback`：从 URL 取 state，从缓存取出对应 OAuthClient 再调用 GetToken；未找到时提示「请从 /login 重新发起授权」

## miiot 全设备属性操作与测试用例

2026-02-18

- ctrl TestSpecsCoverage 扩展为 readme 中全部 13 个型号
- 新增 TestSpecsCoverage_DeviceSpecific：按设备类型校验 Spec（switch/light/speaker/TV/occupancy/toggle/brightness/channels）
- TestControllerErrors 补充所有操作的 unknown model 错误校验，以及 SetSwitchChannel 通道越界
- 新增集成测试：TestControllerToggle、TestControllerSetSwitchChannel、TestControllerTVTurnOff
- TestControllerSetOnRoundtrip 增加 chuangmi.plug.v3 覆盖
- 新增设备包测试：babai/plug、chuangmi/plug、giot/light、lemesh/switch、xiaomi/tv 的 Spec 常量校验

## miiot/xiaomi/wifispeaker 属性控制与 API 测试

2026-02-18

- ctrl 新增 SetMute/GetMute、Next/Previous
- l05b、l05c 补充完整 siid/piid/aiid 常量（与 oh2 一致）
- wifispeaker 测试：GetVolume、SetVolumeGetVolume、GetMute、SetMuteGetMute、TTS、Play、Pause、Next、Previous、UnsupportedModel

## miiot 产品规格 API（model 与 home.miot-spec.com 1:1 匹配）

2026-02-18

- 新增 `miiot/specs`：从 miot-spec.org/instances 加载 model→URN，生成 home.miot-spec.com/spec?type={urn}
- 新增 `miiot/registry`：ModelAPI 注册表，支持 SpecURL、ProductURL
- 按文件规则实现 13 个型号：xiaomi.wifispeaker.{oh2,l05b,l05c}、xiaomi.tv.eanfv1、bean.switch.{bln31,bln33}、chuangmi.plug.{m3,v3}、babai.plug.sk01a、giot.light.v5ssm、opple.light.bydceiling、lemesh.switch.sw3f13、linp.sensor_occupy.hb01
- 新增 `cmd/miiot`：对照 m list 验证所有型号 SpecURL
- 新增 `miiot/ctrl`：Controller 封装属性与动作操作（SetOn/GetOn、SetBrightness、TTS、SetVolume、Play/Pause、TVTurnOff、GetOccupancy、SetSwitchChannel）
- 各型号文件补充 siid/piid/aiid 常量
- 新增测试：ctrl 集成测试、bean/switch、opple/light、xiaomi/wifispeaker、linp/sensor_occupy 常量校验

## m list 全型号 SPEC 功能（按 docs/spec.md）

2026-02-18

- 新增 `internal/device/specs.go`：ModelSpec 结构体及 LoadSpec、LoadAllModelSpecs，按 spec.md 流程从 miot-spec.org 获取 instances→instance
- 新增 `m spec_all` 命令：获取 m list 中所有唯一型号的 SPEC，输出 ok/failed 汇总
- 修复 miioservice.MiotSpec：精确 model 匹配优先，避免 oh2 误匹配 oh21/oh27/oh2p
- 新增 `internal/device/specs_test.go`：TestLoadAllModelSpecs、TestLoadSpec

## OAuth /callback 登录成功页优化

2026-02-18

- `/callback` 登录成功时返回完整 HTML 页面，显示「✓ 登录成功」及「米家 OAuth 授权已完成，token 已保存」
- 新增 5 秒倒计时提示，倒计时结束后自动调用 `window.close()` 尝试关闭页面

## xiaomusic 完整实现（基于 hanxi/xiaomusic）

2026-02-15

- 新增 `internal/minaapi` 包，对接 api2.mina.mi.com，使用 m login 的 OAuth token
- 实现 `PlayByURL`：L06A、LX05 等机型走 `play_by_music_url`（player_play_music），其余走 `player_play_url`
- `minaservice` 新增 `NewWithMinaAPI`，`GetMinaDeviceID` 优先从 Mina device list 解析 deviceID
- `play-file` 支持绝对路径（如 `/Users/xxx/Music/xxx.mp3`），文件不在 musicDir 时以文件所在目录为 HTTP 根目录
- URL 路径段编码以支持空格等特殊字符
- 新增 `xiaomusic.host` 配置及 `-host` 参数，供音箱访问 play-file 的 HTTP 服务

## xiaomusic 局域网 IP 与端口处理

2026-02-15

- `getListenHost`：优先 UDP 探测（8.8.8.8）获取默认路由 IP，过滤 VPN 地址（198.18.x）；回退时遍历网卡，优先 192.168.x、10.x，排除 Docker 虚拟网段
- `play-file` 启动前自动释放端口：`killProcessOnPort` 用 lsof + kill 终止占用 8090 的进程
- 新增 `waitPortReady`：轮询确认端口已监听（5 秒超时）后再调用 `PlayByURL`，避免音箱请求时服务未就绪

## m 命令帮助与提示优化

2026-02-15

- 根据实际实现更新 `usage()` 提示：标题改为「m - XiaoMi MIoT + Mina CLI」，明确设备配置（config default_did 或 MI_DID）及 mina 命令的设备要求
- 新增 `m help` 完整帮助，支持 `help`、`-h`、`--help`、`?`、`？` 触发
- 帮助内容按 AUTH、DEVICE、MINA、MIoT/MiIO 分类，逐条描述子命令功能，并附 EXAMPLES

## 配置文件支持

2026-02-15

- 参考 [go-template](https://github.com/zeusro/go-template)，将项目中涉及变量的部分改为从配置文件获取，获取不到才赋予默认值
- 新增 `internal/config` 配置包，支持 YAML 配置加载
- 配置文件查找顺序：`.config.yaml`、`config.yaml`、`~/.config/miflow/config.yaml`、`~/.miflow.yaml`
- 配置优先级：环境变量 > 配置文件 > 默认值
- 新增 `configs/config-example.yaml` 示例配置
- 可配置项：OAuth（client_id、redirect_uri、cloud_server 等）、token_path、default_did、debug、flow/addr、flow/data_dir、xiaomusic/music_dir、xiaomusic/addr、miio/callback_port、miio/specs_cache_path、http/timeout_seconds

## OAuth 2.0 接入（替换密码登录）

2026-02-13

- 参考 [ha_xiaomi_home](https://github.com/XiaoMi/ha_xiaomi_home) 接入方式，放弃原始密码登录，改用 OAuth 2.0
- 假设白名单域名，使用 `ha.api.io.mi.com` 与 miotspec 接口
- 新增 `m login` 完成 OAuth 授权，Token 保存在 `~/.mi.token`
- 小爱播报通过 MIoT「执行文本指令」动作实现（siid=5, aiid=5）

## 初始化 

2026-02-11

- 使用 Go 初始化模块 `github.com/zeusro/miflow`

## Go 实现 MiService（m CLI） 

2026-02-11

- 基于 [MiService](https://github.com/yihong0618/MiService) 实现小米云服务 Go 版本：
  - 账号认证（`internal/miaccount`，现为 OAuth 2.0）
  - MiIO / MIoT 协议支持（`internal/miioservice`，现对接 ha.api.io.mi.com）
  - MiNA 小爱音箱控制（`internal/minaservice`，TTS 使用 MIoT 动作）
  - 命令解析（`internal/miiocommand`）
- 新增命令行工具 `m`，支持设备列表、属性读写、动作调用、TTS 播报等功能

## Flow 可视化控制流（flow CLI）

2026-02-11

- 新增命令行入口 `flow`（`cmd/flow`），提供基于 HTTP 的可视化控制流配置与执行服务
- 定义线性控制流模型 `Flow` / `FlowStep`，当前支持的步骤类型：
  - `delay`：等待指定毫秒数
  - `tts`：通过 MiNA 对小爱音箱进行 TTS 播报
  - `play_url`：通过 MiNA 播放指定音频 URL
  - `miio`：通过 `miiocommand` 发送等价于 `m` CLI 的 MiIO/MIoT 命令文本
- 后端提供 REST API（`/api/flows`）用于创建、更新、删除、执行控制流，前端为内嵌的极简单页界面，用于拖拽/编辑步骤并一键触发运行
