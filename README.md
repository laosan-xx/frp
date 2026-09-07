# frp 自用构建产物

> 本仓库**只发布编译好的 `frpc` / `frps` 二进制文件**，不包含源码。
>
> 基于 [fatedier/frp](https://github.com/fatedier/frp) 的个人定制版本，主要用于 **OpenWrt 路由器运行 frpc** 与 **Linux 服务器运行 frps**，仅自用，不保证与上游同步更新，也不保证稳定性与安全性。
> 需要官方稳定版请使用 [fatedier/frp](https://github.com/fatedier/frp)。

## 下载

前往 [Releases](https://github.com/laosan-xx/frp/releases) 页面，按设备架构选择对应包：

| 包名 | 适用设备 |
|------|----------|
| `frp_<版本>_linux_amd64.tar.gz` | x86_64 服务器、软路由 |
| `frp_<版本>_linux_arm64.tar.gz` | ARMv8 / aarch64（如 R2S、N1） |
| `frp_<版本>_linux_arm_hf.tar.gz` | ARMv7 硬浮点 |
| `frp_<版本>_linux_arm.tar.gz` | ARMv5 / 软浮点老设备 |
| `frp_<版本>_linux_mipsle.tar.gz` | 小端 MIPS（如 MT7621） |
| `frp_<版本>_linux_mips.tar.gz` | 大端 MIPS（如 MT7620） |

每个压缩包内含：`frpc`、`frps`、`frpc.toml`、`frps.toml`、`LICENSE`。

不确定架构时，在设备上执行：

```bash
uname -m
```

## frps（Linux 服务端）

```bash
tar -zxf frp_0.80.7_linux_amd64.tar.gz
cd frp_0.80.7_linux_amd64
```

最小配置 `frps.toml`：

```toml
bindPort = 7000
auth.token = "改成你自己的密码"

# 可选：仪表盘
webServer.addr = "0.0.0.0"
webServer.port = 7500
webServer.user = "admin"
webServer.password = "admin"
```

启动：

```bash
./frps -c frps.toml
```

以 systemd 托管（`/etc/systemd/system/frps.service`）：

```ini
[Unit]
Description=frps service
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/frps -c /etc/frp/frps.toml
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

```bash
systemctl enable --now frps
```

## frpc（OpenWrt 路由器）

```bash
tar -zxf frp_0.80.7_linux_mipsle.tar.gz
chmod +x frpc
```

最小配置 `frpc.toml`：

```toml
serverAddr = "服务器公网 IP"
serverPort = 7000
auth.token = "改成你自己的密码"

[[proxies]]
name = "ssh"
type = "tcp"
localIP = "127.0.0.1"
localPort = 22
remotePort = 6000
```

前台测试：

```bash
./frpc -c frpc.toml
```

确认无误后放到 `/usr/bin/frpc` 并在 OpenWrt 中配置为服务（或使用 procd init 脚本）常驻。

## 配置说明

完整配置文档参考官方站点 [gofrp.org](https://gofrp.org)。本版本沿用上游 TOML 格式（`frpc.toml` / `frps.toml`），不再支持旧的 ini 格式。

## 更新

frpc 内置自更新能力，会从本仓库的 Release 拉取与当前架构匹配的最新包，无需手动下载。

## 问题反馈

在本仓库 [issues](https://github.com/laosan-xx/frp/issues) 中反馈。

## 许可

基于 [fatedier/frp](https://github.com/fatedier/frp)，遵循 Apache-2.0 许可，详见包内 `LICENSE`。
