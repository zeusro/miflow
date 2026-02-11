# miflow

小米IOT自定义控制流

## m - 小米云服务命令行 (Go 实现)

本仓库包含 [MiService](https://github.com/yihong0618/MiService) 的 Go 语言实现，命令行入口为 `m`（对应原 Python 版 `micli`）。

### 构建

```bash
go build -o m ./cmd/m
```

### 环境变量

```bash
export MI_USER=<小米账号>
export MI_PASS=<密码>
export MI_DID=<设备ID或名称>   # 部分命令需要
```

登录态会保存在 `~/.mi.token`。

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

- **小爱播报 / 播放**（需设置 `MI_DID`）  
  `m message 你好`  
  `m play https://example.com/audio.mp3`  
  `m pause`  
  `m mina`  # 查看当前设备信息

- **MIoT 规格**  
  `m spec speaker`  
  `m spec xiaomi.wifispeaker.lx04`

- **帮助**  
  `m help` 或 `m ?`
