**Xiaomi Miio 集成（Home Assistant）的前提条件总结**（基于官方文档 https://www.home-assistant.io/integrations/xiaomi_miio/#prerequisites，截至2026年2月仍为主要内容，没有重大变更迹象）：

### 核心要求
- 设备**必须**通过**小米官方米家App（Mi Home / Xiaomi Home）**完成初始配网和设置  
  → 千万不要使用厂商专用App（如Roborock、Viomi、Roidmi等），否则大概率无法正常接入或获取token
- 大部分设备支持通过Home Assistant **图形界面**直接添加  
  → 少数例外（如小米设备追踪器 Mi WiFi Repeater 2、小米红外遥控器）仍需手动写 `configuration.yaml`

### Token（令牌）要求
以下几类设备必须提供32位十六进制token才能正常工作：
- 扫地机器人（大部分型号）
- 米家插座、Philips智睿灯泡
- 小米红外遥控器
- 小米WiFi设备追踪器
- 部分网关的子设备控制等

**获取token的几种常见方法**（官方列出的方式，从推荐到复杂程度排序）：
1. 使用 **Xiaomi Cloud Tokens Extractor** 工具（最推荐）  
   → 输入小米账号密码 + 国家/地区 → 一次性拉取账号下所有设备的token（可执行文件或Python脚本）
2. 米家App内提取（Android/iOS）  
   → 开启开发者模式（连点版本号）→ 查看JSON或特定界面  
   → 注意：新版App经常取不到或取错多设备时的token
3. **Get Mi Home Devices Token** 桌面工具（Win/mac）
4. 老版本米家App（Android v5.4.49）日志提取法  
   → 新版App已不再明文存储token
5. Rooted Android / Linux ADB 提取
6. iOS 通过未加密备份 + sqlite + OpenSSL 解密
7. BlueStacks模拟器 + 老版本App日志提取
8. miio命令行工具（`npx miio discover`）  
   → **仅适用于设备还未被米家App绑定**的情况，一旦绑定通常显示`???`

**安全建议**：优先使用账号密码方式让HA自动获取（集成流程中支持），尽量避免手动提取token并长期保存。

### 服务器地区（Server Region）
配置时必须正确选择米家App使用的服务器区域，否则连不上：
- `cn`（中国大陆）
- `de`（德国）
- `i2`（新加坡/印度？）
- `ru`（俄罗斯）
- `sg`（新加坡）
- `us`（美国）

### 其他重要注意事项
- 复杂网络环境（VLAN、多子网）可能导致设备发现/通信失败（UDP广播问题）
- 部分功能（如某些网关子设备）仍需互联网连接小米云
- 固件兼容性：确保设备固件不是最新但不兼容的版本（极少数情况）
- 已知问题：网络阻断小米设备本地通信会导致集成卡死或反复重试

一句话总结：  
想稳定接入小米设备 → 用米家App配网 → 用账号密码方式让HA自动拉取token → 选对服务器地区 → 大部分就能直接UI集成。手动抠token是备选方案，且越来越麻烦。

（如果你的设备是比较新的型号或特殊品牌子品牌，建议同时查看对应设备的具体支持状态，因为部分型号已迁移到别的集成方式。）