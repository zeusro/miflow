https://github.com/home-assistant/core/tree/dev/homeassistant/components/xiaomi_miio 遍历，介绍一下这个目录下面的文件干什么用的

miio 协议 = UDP 54321 + AES-128-CBC + JSON-RPC + 32 字节固定 header + 设备专属 token，它是小米早期（2015–2019）Wi-Fi 智能设备本地控制的事实标准，后来逐渐被更规范的 MIoT 协议在 payload 层取代，但 UDP + AES 的传输层框架至今（2026 年）仍然被广泛沿用。


**homeassistant/components/xiaomi_miio** 是 Home Assistant 中负责本地控制部分小米/米家/石头等设备的集成（**Xiaomi Miio**），主要支持通过 **miio 协议** 直接局域网通信的设备（不需要经过小米云）。

2026 年 2 月的 dev 分支下，这个目录典型的结构和主要文件作用如下（基于目前主流版本的实际情况）：

| 文件名                        | 主要作用                                                                 | 备注 |
|-------------------------------|--------------------------------------------------------------------------|------|
| `__init__.py`                 | 集成入口文件，负责加载、初始化、配置校验、发现设备、统一协调各个平台     | 几乎所有集成都有 |
| `config_flow.py`              | 配置向导（UI 添加强制走这个流程）                                        | 支持自动发现 + 手动输入 IP/token |
| `const.py`                    | 存放所有常量：DOMAIN、CONF_xxx、ATTR_xxx、默认值、支持的设备型号等       | 集中管理魔法字符串 |
| `device.py` / `devices.py`    | 设备基类、设备信息解析、型号 → 实现类映射关系                            | 核心：决定用哪个子类来控制具体设备 |
| `exceptions.py`               | 自定义异常类（连接失败、token 错误、设备不支持等）                       | - |
| `manifest.json`               | 组件元数据（名称、依赖、文档链接、支持的 HA 版本、iot-class 等）         | 必备 |
| `fan.py`                      | fan 平台实现，主要支持：空气净化器、加湿器、风扇（Pedestal Fan 等）      | 目前使用最多的平台之一 |
| `vacuum.py`                   | vacuum 平台实现，支持小米/石头扫地机器人（吸尘器）                       | 支持地图、房间、分区清扫等 |
| `humidifier.py`               | humidifier 平台（部分加湿器走这个，而不是 fan）                          | 较新设备可能使用 |
| `light.py`                    | light 平台（支持米家台灯、床头灯、吸顶灯等 miio 协议灯具）               | - |
| `switch.py`                   | switch 平台（插座、智能插排、部分净化器开关等）                          | - |
| `sensor.py`                   | sensor 平台（温度、湿度、PM2.5、滤芯寿命、水箱状态、错误代码等传感器）   | 非常多实体来源 |
| `binary_sensor.py`            | 二元传感器（滤芯是否需要更换、水箱是否已满、错误状态等）                 | - |
| `remote.py`                   | remote 平台（少数红外遥控器设备）                                        | 使用较少 |
| `xiaomi_cloud.py` / `cloud.py`| 云端辅助功能（部分设备需要云端 token 提取、本地发现失败时的 fallback）   | 视版本而定 |
| `gateway.py`                  | 网关相关实现（Aqara 网关的部分 miio 模式支持）                           | 早期较多，现在逐渐减少 |
| `helpers.py` / `utils.py`     | 通用工具函数（token 计算、属性转换、设备信息解析、枚举映射等）           | - |
| `translations/` 文件夹        | 多语言字符串（en、zh-Hans 等）                                           | UI 显示用 |
| `services.yaml`               | 定义额外服务（如 vacuum.goto、fan.set_favorite_level 等）                | 部分版本有 |
| `strings.json`                | 旧版多语言文件（新版多用 translations/）                                | - |

### 总结常见的典型组合（2025–2026 年主流情况）

- **空气净化器** → fan + sensor + switch + binary_sensor
- **扫地机器人** → vacuum + sensor + binary_sensor
- **加湿器** → humidifier / fan + sensor
- **米家台灯 / 吸顶灯** → light + sensor
- **插座** → switch + sensor（功率、电流等）

如果你看到某个具体型号不支持，大概率是：

1. `device.py` 里缺少对应型号的映射
2. 缺少特定子类实现（例如某个新风扇型号没加到 fan.py）
3. 协议变了（小米经常升级加密或字段）

建议直接看 `device.py` 里的 **MODELS** 或 **SUPPORTED_DEVICES** 字典，就能知道当前 dev 分支到底支持哪些具体型号。

需要我帮你重点解释某个文件（比如 vacuum.py 或 fan.py）里面的具体实现逻辑吗？