# 改动

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
