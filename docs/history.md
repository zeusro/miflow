# 改动

## 初始化 

2026-02-11

- 使用 Go 初始化模块 `github.com/zeusro/miflow`

## Go 实现 MiService（m CLI） 

2026-02-11

- 基于 [MiService](https://github.com/yihong0618/MiService) 实现小米云服务 Go 版本：
  - 账号登录与 Token 持久化（`internal/miaccount`）
  - MiIO / MIoT 协议支持（`internal/miioservice`）
  - MiNA 小爱音箱控制（`internal/minaservice`）
  - 命令解析（`internal/miiocommand`）
- 新增命令行工具 `m`（替代原 `micli`），支持设备列表、属性读写、动作调用、TTS 播报与播放 URL 等功能
