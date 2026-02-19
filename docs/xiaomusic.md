# xiaomusic

https://github.com/hanxi/xiaomusic

使用小爱音箱播放音乐，音乐使用 yt-dlp 下载。

## 小米米家 API 实现

xiaomusic 基于 [MiService](https://github.com/yihong0618/MiService)（小米云服务接口）实现与小米小爱音箱的交互。

### 核心依赖

- **miservice-fork**：MiService 的 fork 版本，用于调用小米云 API

### MiService 架构

```
MiService：XiaoMi Cloud Service
├── MiAccount：账号服务
├── MiIOService：MiIO 服务 (sid=xiaomiio)
│   └── MIoT_xxx：MIoT 服务
├── MiNAService：MiAI 服务 (sid=micoapi)
└── MiIOCommand：MiIO 命令风格接口
```

### 使用的 API

| 功能 | API 方法 | 说明 |
|------|----------|------|
| 播放控制 | `mina_service.player_pause(device_id)` | 暂停播放 |
| 播放控制 | `mina_service.player_stop(device_id)` | 停止播放 |
| 播放控制 | `mina_service.player_get_status(device_id)` | 获取播放状态 |
| 音量控制 | `mina_service.player_set_volume(device_id, volume)` | 设置音量 |
| TTS | `mina_service.text_to_speech(device_id, value)` | 文字转语音播报 |
| 播放音乐 | `mina_service.play_by_url(device_id, url)` | 通过 URL 播放 |
| 播放音乐 | `mina_service.play_by_music_url(device_id, url, ...)` | 通过音乐 API 播放（支持续播） |
| 音乐搜索 | `mina_service.mina_request("/music/search", params)` | 搜索歌曲获取 audioID |
| MiIO 命令 | `miio_command(miio_service, did, cmd)` | 部分设备 TTS 使用 MiIO 命令 |

### 实现原理参考

- [不用 root 使用小爱同学和 ChatGPT 交互折腾记](https://github.com/yihong0618/gitblog/issues/258)
- [awesome-xiaoai](https://github.com/zzz6519003/awesome-xiaoai)

### 技术栈

- 后端：Python + FastAPI
- 小米服务：miservice-fork
- 音乐下载：yt-dlp
- 容器化：Docker

### miflow xiaomusic 限制

- `api2.mina.mi.com` 需 **micoapi** 认证（MiAccount 账号密码），OAuth（m login）不被支持
- 若出现 `http 401, 返回 HTML 非 JSON`，需使用 [hanxi/xiaomusic](https://github.com/hanxi/xiaomusic) 或后续添加 MiAccount 支持
- **获取小爱对话**：`userprofile.mina.mi.com` 需 Cookie（userId、deviceId、serviceToken），OAuth 不支持。实现步骤见 [docs/conversation.md](conversation.md)
