# m - XiaoMi MIoT + Mina CLI

## 使用示例 / Usage Examples

### 语言切换 / Language

CLI 根据环境变量 `LANG` 或 `LC_ALL` 自动选择语言：

```bash
# 中文环境
LANG=zh_CN.UTF-8 m help

# 英文环境
LANG=en_US.UTF-8 m help
```

### 基本命令 / Basic Commands

```bash
# 首次使用需登录
m login

# 查看帮助
m help
```

### Mina 命令 / Mina Commands

```bash
# 列出 Mina 设备
m mina

# TTS 播报
m message 你好世界

# 播放音频
m play https://example.com/audio.mp3

# 循环播放
m loop https://example.com/audio.mp3
```

### MiIO/MIoT 命令 / MiIO/MIoT Commands

```bash
# 列出设备
m list

# 查询规格
m spec speaker
m spec xiaomi.wifispeaker.lx04 json

# 获取/设置属性
m 1,1-2,2-1
m 2=#60
```
