# miflow

小米 IOT 自定义控制流

## m - 小米云服务命令行 (OAuth 2.0)

本仓库包含基于 [ha_xiaomi_home](https://github.com/XiaoMi/ha_xiaomi_home) OAuth 2.0 接入方式的 Go 实现，命令行入口为 `m`。

### 构建

```bash
go build -o m ./cmd/m
```

### 配置 homeassistant.local 解析

OAuth 回调默认使用 `http://homeassistant.local:8123/callback`，需将 `homeassistant.local` 解析到本机（127.0.0.1）才能完成登录。

**macOS**

```bash
sudo nano /etc/hosts
```

在文件末尾添加一行：

```
127.0.0.1 homeassistant.local
```

保存后生效（无需重启）。

**Windows**

1. 以管理员身份打开记事本
2. 打开文件：`C:\Windows\System32\drivers\etc\hosts`
3. 在文件末尾添加一行：

```
127.0.0.1 homeassistant.local
```

4. 保存后生效（无需重启）

### 登录（OAuth 2.0）

首次使用需完成 OAuth 授权：

```bash
m login
```

将自动打开浏览器，使用小米账号登录后，Token 会保存在 `~/.mi.token`。无需输入密码，符合 OAuth 2.0 安全规范。

> **白名单说明**：默认使用与 Home Assistant Xiaomi Home 集成相同的 OAuth 配置。若需自定义，可通过配置文件或环境变量：
> - `MI_OAUTH_CLIENT_ID` - OAuth 客户端 ID
> - `MI_OAUTH_REDIRECT_URI` - 回调地址（需已加入小米开发者白名单）
> - `MI_CLOUD_SERVER` - 区域：cn（中国大陆）、de、i2、ru、sg、us

### 配置文件

参考 [go-template](https://github.com/zeusro/go-template)，支持从配置文件读取变量，获取不到才使用默认值。

```bash
cp configs/config-example.yaml .config.yaml
# 编辑 .config.yaml 修改配置
```

配置优先级：**环境变量 > 配置文件 > 默认值**。配置文件支持路径：`.config.yaml`、`config.yaml`、`~/.config/miflow/config.yaml`、`~/.miflow.yaml`。

### 环境变量

```bash
export MI_DID=<设备ID或名称>   # 部分命令需要，也可在配置 default_did
export MI_DEBUG=1              # 可选，打印 HTTP 请求/响应（调试用），或配置 debug: true
```

### 用法示例

- **设备列表**  
  `m list`  
  `m list full true 0`

- **MIoT 属性**  
  查: `m 1,1-2,2-1`  
  设: `m 2=#60,2-2=#false`

- **MIoT 动作**  
  `m 5 你好`  
  `m 5-4 查询天气 #1`

- **小爱播报**（需设置 `MI_DID`）  
  `m message 你好`  
  `m mina`  # 查看当前设备信息

- **MIoT 规格**  
  `m spec speaker`  
  `m spec xiaomi.wifispeaker.lx04`

- **帮助**  
  `m help` 或 `m ?`

### 与 ha_xiaomi_home 的对应关系

| 功能     | ha_xiaomi_home     | miflow         |
|----------|--------------------|-----------------|
| 登录方式 | OAuth 2.0          | OAuth 2.0      |
| API 域名 | ha.api.io.mi.com   | ha.api.io.mi.com |
| 设备列表 | device_list_page   | m list         |
| 属性读写 | miotspec/prop      | m siid,piid=val |
| 动作执行 | miotspec/action    | m siid-aiid args |
| 小爱播报 | Execute Text Directive | m message    |



/Users/zeusro/Music/h.flac

# 查看本机 IP（如 ifconfig 或 系统设置）
-host=192.168.1.100

MI_DID=978878303 ./xiaomusic  play-file "/Users/zeusro/Music/h.flac"