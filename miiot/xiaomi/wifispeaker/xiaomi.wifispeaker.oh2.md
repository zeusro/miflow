# xiaomi.wifispeaker.oh2 规格说明

> 数据来源：[miot-spec.org](https://miot-spec.org/miot-spec-v2/instance?type=urn:miot-spec-v2:device:speaker:0000A015:xiaomi-oh2:1) | [规格页](https://home.miot-spec.com/spec?type=urn%3Amiot-spec-v2%3Adevice%3Aspeaker%3A0000A015%3Axiaomi-oh2%3A1)

**URN**: `urn:miot-spec-v2:device:speaker:0000A015:xiaomi-oh2:1`

---

## 服务 (Services) 与 属性/动作

### 1. Device Information (siid=1) - 设备信息

| piid | 属性 | 类型 | 访问 | 说明 |
|------|------|------|------|------|
| 1 | manufacturer | string | read | 厂商 |
| 2 | model | string | read | 型号 |
| 3 | serial-number | string | read | 设备 ID |
| 4 | firmware-revision | string | read | 固件版本 |
| 5 | serial-no | string | read, notify | 序列号 |

---

### 2. Speaker (siid=2) - 音箱

| piid | 属性 | 类型 | 访问 | 说明 |
|------|------|------|------|------|
| 1 | volume | uint8 | read, write, notify | 音量，范围 4-100，步进 1，单位 % |
| 2 | mute | bool | read, write, notify | 静音 |

**功能 control**:
- `get_properties` 读取音量/静音
- `set_properties` 设置音量、静音

---

### 3. Play Control (siid=3) - 播放控制

| piid | 属性 | 类型 | 访问 | 说明 |
|------|------|------|------|------|
| 1 | playing-state | uint8 | read, notify | 0=Stop, 1=Playing, 2=Pause |

| aiid | 动作 | 入参 | 说明 |
|------|------|------|------|
| 2 | play | - | 播放 |
| 3 | pause | - | 暂停 |
| 5 | previous | - | 上一曲 |
| 6 | next | - | 下一曲 |

**功能 control**:
- `action(3, 2)` 播放
- `action(3, 3)` 暂停
- `action(3, 5)` 上一曲
- `action(3, 6)` 下一曲

---

### 4. Microphone (siid=4) - 麦克风

| piid | 属性 | 类型 | 访问 | 说明 |
|------|------|------|------|------|
| 1 | mute | bool | read, write, notify | 麦克风静音 |

---

### 5. Intelligent Speaker (siid=5) - 智能音箱 / 语音助手

| piid | 属性 | 类型 | 访问 | 说明 |
|------|------|------|------|------|
| 1 | text-content | string | - | 文本内容（TTS 输入） |
| 2 | silent-execution | bool | - | 静默执行 |
| 3 | sleep-mode | bool | read, write, notify | 休眠模式 |
| 4 | audio-id | string | read, notify | 音频 ID |

| aiid | 动作 | 入参 | 说明 |
|------|------|------|------|
| 1 | wake-up | - | 唤醒 |
| 2 | play-radio | - | 播放电台 |
| 3 | play-text | in[1] | TTS 播报，入参 1=文本 |
| 4 | execute-text-directive | in[1,2] | 执行文本指令，入参 1=文本，2=静默执行 |
| 5 | play-music | - | 播放音乐 |

**功能 control**:
- `action(5, 3, [text])` TTS 播报（play-text）
- `action(5, 4, [text, silent])` 执行 TTS 指令（execute-text-directive）
- `action(5, 1)` 唤醒
- `action(5, 2)` 播放电台
- `action(5, 5)` 播放音乐

---

### 6. Clock (siid=6) - 闹钟

| aiid | 动作 | 入参 | 说明 |
|------|------|------|------|
| 1 | stop-alarm | - | 停止闹钟 |

---

### 7. No Disturb (siid=7) - 勿扰模式

| piid | 属性 | 类型 | 访问 | 说明 |
|------|------|------|------|------|
| 1 | no-disturb | bool | read, write, notify | 勿扰开关 |
| 2 | enable-time-period | string | read, write, notify | 勿扰时间段 |

---

### 8. tv-switch (siid=8) - 电视开关 [小米私有]

| aiid | 动作 | 入参 | 说明 |
|------|------|------|------|
| 1 | tv-switchon | - | 电视开机 |

---

## miiot ctrl 映射 (oh2.go / constants.go)

| ctrl 方法 | siid | piid/aiid |
|-----------|------|-----------|
| TTS / ExecuteText | 5 | aiid=1 (play-text) 或 4 (execute-text-directive) |
| Volume | 2 | piid=1 |
| Mute | 2 | piid=2 |
| Play | 3 | aiid=2 |
| Pause | 3 | aiid=3 |
| Next | 3 | aiid=6 |
| Previous | 3 | aiid=5 |
