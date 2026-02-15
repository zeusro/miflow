深度遍历 https://iot.mi.com/v2/new/doc/home 。假设有一个新的米家设备，如何 找到它内部的SPEC 定义


**对于一个新的米家设备（已接入米家平台的），其内部 SPEC 定义（即 MIoT-Spec-V2 的功能描述，包括服务、服务下的属性/方法/事件、siid/piid/eid 等）几乎都公开在小米官方的 MIoT Spec 数据库中。** 官方开发者文档（https://iot.mi.com/v2/new/doc/home）里也反复指向这个数据库，让开发者“从 www.miot-spec.org 使用标准 HTTP/GET 请求读取并参考规范”。

### 1. 先拿到设备的型号（Model）
这是查询 SPEC 的关键钥匙。

- **最简单**：在米家 App 里进入该设备 → 设置 → 关于设备，通常直接显示型号（如 `yeelink.light.nl1`、`xiaomi.aircondition.v1`、`zhimi.toilet.sa1` 等）。
- **本地发现**：用 `miio discover`、`python-miio`、`XiaomiGateway3` 等工具扫描局域网设备，会返回 model。
- **BLE/Zigbee 等**：绑定时或在网关设备列表里也能看到 model。

### 2. 通过用户友好界面查询（推荐新手）
访问 **https://home.miot-spec.com/s/**（小米/米家产品库搜索页）。

- 输入设备名称或型号搜索。
- 找到对应产品后，点击右侧的 **“规格”** 按钮。
- 弹窗里选择固件版本（一般选最大的那个），就会进入详细 SPEC 页面，显示：
  - 完整 URN（如 `urn:miot-spec-v2:device:night-light:0000A0AB:yeelink-nl1:1`）
  - 所有服务（siid）、属性（piid + 读写通知权限）、方法、事件等树状结构
  - 可直接复制 JSON 或查看每项的 type、format、access 等

这个站点就是官方维护的“小米/米家产品大全 + SPEC 查询工具”。

### 3. 通过 API 直接获取 JSON（开发者/自动化首选）
小米官方提供的公开接口（文档里多次提到用 GET 请求读取）：

1. 先获取所有实例列表（方便找到你的型号对应的 URN）：
   ```
   https://miot-spec.org/miot-spec-v2/instances?status=all
   ```
   （页面很大，可用浏览器搜索你的 model）

2. 拿到 URN 后，访问具体 SPEC：
   ```
   https://miot-spec.org/miot-spec-v2/instance?type=urn:miot-spec-v2:device:xxx:xxxx:xxxx:1
   ```
   返回标准的 JSON，就是设备完整的内部 SPEC 定义。

示例（米家夜灯2）：
- 搜索后得到 URN → `https://miot-spec.org/miot-spec-v2/instance?type=urn:miot-spec-v2:device:night-light:0000A0AB:yeelink-nl1:1`

### 4. 如果你是设备开发者（在小米IoT平台创建的新产品）
在开发者平台控制台 → 产品 → 功能定义 里：
- 优先选用“标准 Spec 功能”（平台已为 190+ 品类预置）。
- 没有的可以自定义服务（必须符合 MIoT-Spec 格式）。
- 定义完成后，平台会自动生成对应的 instance/spec，你在产品详情里就能看到，也会同步到上面的公开数据库。

### 5. 官方文档里的相关路径（从 https://iot.mi.com/v2/new/doc/home 出发）
- **知识库 → MIoT Spec 协议详解**：解释什么是 Spec、属性/方法/事件/服务结构。
- **设计 → 定义产品功能**：明确写着“开发者可从 www.miot-spec.org 上使用标准 HTTP/GET 请求读取并参考规范”。
- **其他地方**（接入指南、控制端 API 等）也反复提到用 miot-spec.org 获取实例。

**总结最快路径**（适用于任何新米家设备）：
1. 米家 App 抄型号 →  
2. 打开 https://home.miot-spec.com/s/ 搜索 →  
3. 点击“规格” → 选最新固件 → 得到完整 SPEC。

或者直接走 API 两步拿 JSON。

这样就能拿到设备“内部”的完整功能定义，后面写 HA 集成、自定义插件、逆向控制等都靠它。需要具体某个设备的 SPEC 示例也可以告诉我型号，我可以帮你给出链接。