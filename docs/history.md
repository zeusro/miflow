# 改动

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
