# PROJECT_INDEX.md — frp 项目快速索引（给 AI / 新人）

> 目的：让 AI 或新人在**不全局重读代码**的情况下，快速建立对项目架构、关键文件、改动位置的准确认知。
> 读这份文件 + `AGENTS.md`（构建命令）即可开工；只有要深入某模块时才去读对应的具体源文件。
> 核心心智模型：**frp = 一条常驻控制连接 + 按需建立的工作连接**。

---

## 0. 一句话定位

frp 是 C/S 架构的内网穿透/反向代理工具：
- **frps（服务端）**：部署在公网，监听客户端连接、鉴权、在公网侧 listen 端口/域名，把外部流量"搬"进内网。
- **frpc（客户端）**：部署在内网，主动连 frps 建立**控制连接**注册代理；外部流量到达时，frps 经控制连接要求 frpc 开**工作连接**，真正承载被代理的数据流。

**两个关键抽象（必须记住）：**
- **控制连接（control conn）**：frpc↔frps 之间常驻，只跑信令（`Login / NewProxy / Ping-Pong / ReqWorkConn` 等），复用 `msg.Dispatcher`，支持多路复用（yamux/quic）。
- **工作连接（work conn）**：真正转发用户 TCP/UDP 流量的管道，由 frpc 按需建立并"贴合"到某个代理。

---

## 1. 程序入口（先看这两个）

| 角色 | 入口文件 | 关键逻辑文件 | 初始化链路 |
|---|---|---|---|
| frps | `cmd/frps/main.go` → `cmd/frps/root.go` | `cmd/frps/root.go` 的 `RunE` | main → Execute → rootCmd.RunE → `config.LoadServerConfig` → `server.NewService` → `svr.Run` |
| frpc | `cmd/frpc/main.go` → `cmd/frpc/sub/root.go` | `cmd/frpc/sub/root.go` 的 `runClient` | main → sub.Execute → runClient → `config.LoadClientConfigResult` → `source.Aggregator` → `client.NewService` → `svr.Run` |

两个 main 还匿名 import `pkg/metrics`（监控）与 `web/frps`、`web/frpc`（内置 Dashboard 静态资源）。

---

## 2. 目录地图

```
cmd/frps/        frps 命令入口（cobra），仅 main + root
cmd/frpc/        frpc 命令入口，sub/ 下为子命令
client/          ★ 客户端总控：service.go(Service) / control_session.go(登录握手) /
                 control.go(Control 生命周期) / connector.go(物理连接抽象) /
                 proxy_manager.go / proxy/(每类代理一个文件) / visitor/
server/          ★ 服务端总控：service.go(Service, handleConnection) / control.go(Control) /
                 proxy/(每类代理一个文件) / controller/ / group/ / ports/ / registry/ /
                 visitor/ / metrics/ / http/(vhost 路由)
pkg/
  config/        ★ 配置解析中枢：load.go(入口) / v1/(当前配置结构体) / source/(动态配置源) /
                 validation/ / legacy/(旧版 INI 兼容) / types/
  msg/           ★ 协议消息定义与编解码：msg.go(所有消息) / handler.go+ctl.go(Dispatcher) / wire_v2.go
  proto/wire/    v2 线协议（ClientHello/ServerHello 加密协商、Frame、magic）
  auth/          鉴权：token.go / pass.go / oidc.go / auth.go（Setter 签名 / Verifier 验签）
  transport/     message.go(MessageTransporter 多路请求/响应) / tls.go
  plugin/        client/ server/ visitor/ 三套扩展点
  util/          基础设施：net / http / vhost / tcpmux / limit / log
  metrics/       mem/ prometheus/ aggregate/
  nathole/       P2P 打洞    vnet/ 虚拟网络    ssh/ SSH tunnel gateway
  sdk/ errors/ naming/ policy/(featuregate/security)
web/             frps / frpc 的 Dashboard 前端（需 make web 构建）
test/e2e/       端到端测试（见第 7 节）
doc/agents/     运维 runbook（如 release.md）
```

★ = 最常改动/最该先读。

---

## 3. 「我要改 X，去哪找」速查表

| 需求 | 主要文件 |
|---|---|
| 加/改代理类型（tcp/udp/http/https/stcp/sudp/xtcp/tcpmux） | 服务端 `server/proxy/{proxy.go + <type>.go}`；客户端 `client/proxy/{proxy.go + <type>.go}`；先读基类 `BaseProxy` |
| 新增代理类型 | 照抄 `server/proxy/tcp.go` + `client/proxy/general_tcp.go`，并在各自 `init()` 用 `RegisterProxyFactory` 注册（按 `reflect.Type`） |
| 加配置项 | 先改 `pkg/config/v1/{server,client,proxy,visitor}.go` 结构体（注意 `json` tag = TOML/YAML 字段名），再到 `pkg/config/load.go` 确认解析路径；别忘了 `Complete`/`Validate` |
| 加 C/S 通信消息 | 先改 `pkg/msg/msg.go`（含单字节 `type` 分发码），再在两侧 `registerMsgHandlers` 注册 |
| 加鉴权方式 | `pkg/auth/`（参考 `token.go` / `oidc.go`） |
| 加插件 | `pkg/plugin/client` 或 `pkg/plugin/server` 或 `pkg/plugin/visitor` |
| 加负载均衡组 | `server/group/`（TCP/HTTP/HTTPS/TCPMux） |
| 加端口分配逻辑 | `server/ports/` |
| 加 HTTP 路由/vhost | `server/http/` + `pkg/util/vhost` |
| 改传输层（KCP/QUIC/WS/TCPMux） | `client/connector.go` + `pkg/transport` + `pkg/proto/wire` |
| 改监控指标 | `pkg/metrics/` + `server/metrics/` |
| 改 Dashboard 前端 | `web/frps/src` / `web/frpc/src`（需 `make web`） |

---

## 4. 一次 TCP 代理请求的端到端数据流

1. 外部用户连接 frps 公网 `remote_port`（frpc 登录时已注册，frps 已 listen）。
2. frps `HandleListener` → `handleConnection` 识别为普通用户流量 → 对应 proxy 的 `handleUserTCPConnection`。
3. `Control.GetWorkConnFromPool` 取工作连接：池中有（frpc 按 `poolCount` 预建）则复用；否则 frps 发 `ReqWorkConn` 信令（走控制连接）。
4. frpc `handleReqWorkConn` 新建到 frps 的连接并发 `NewWorkConn`；frps `RegisterWorkConn` 入该客户端的 `workConnCh`。
5. frps 在该 work conn 上写 `StartWorkConn` 标注 proxy 归属。
6. frpc 收 `StartWorkConn` → `pm.HandleWorkConn` → 代理 `InWorkConn` → 用 `LocalIP:LocalPort` 连内网真实后端。
7. 两端分别把"用户连接↔work conn"、"work conn↔本地后端"双向 `Join` 拷贝（含可选加密/压缩/限速/ProxyProtocol）→ 流量贯通。

> HTTP/HTTPS 额外经 `vhost.Routers` 按 Host/路径路由；UDP 走 `UDPPacket` 封装；stcp/sudp/xtcp 由 visitor 端先 `NewVisitorConn` 接入再走同样的工作连接机制。

---

## 5. 配置体系

- **解析入口**：`pkg/config/load.go`（`LoadServerConfig` / `LoadClientConfigResult`）。先 TOML 解析再转 JSON，strict 模式拒绝未知字段；支持 Go `text/template` 预渲染（可用环境变量）；旧版 INI 由 `legacy/` 转换。
- **当前结构体**：`pkg/config/v1/`（`server.go` / `client.go` / `proxy.go` / `visitor.go` / `common.go` / `proxy_plugin.go`）。每个代理类型实现 `ProxyConfigurer` 接口（`GetBaseConfig`/`Complete`/`UnmarshalFromMsg`/`ValidateForServer`）。
- **动态配置源**：`pkg/config/source/`（`Aggregator` 聚合文件/store 等多源，支持 frpc 热更新）。
- 字段名由结构体 `json:"..."` tag 决定；代理是 `[[proxies]]` 数组，每项 `type = "tcp"`。

---

## 6. 协议消息（`pkg/msg/msg.go` 关键类型）

- `Login` / `LoginResp`：登录、分配 `RunID`
- `NewProxy` / `NewProxyResp` / `CloseProxy`：代理注册/注销
- `NewWorkConn` / `ReqWorkConn` / `StartWorkConn`：工作连接三方握手（要→建→标注）
- `NewVisitorConn` / `NewVisitorConnResp`：visitor 接入
- `Ping` / `Pong`：心跳
- `UDPPacket`：UDP 数据封装
- `NatHole*`：P2P 打洞信令
- `ServerCommand` / `ServerCommandResp`：frps 反向让 frpc 执行命令
- 线协议：v1 = JSON over 加密 ReadWriter；v2 = magic + ClientHello/ServerHello AEAD 协商 + 分帧（`pkg/proto/wire`、`pkg/msg/wire_v2.go`）。

---

## 7. 测试与构建

- **构建**：`make build`（frps+frpc，带 `-tags frps/frpc`，可选 `noweb`）、`make frps`、`make frpc`、`make web`（构建 Dashboard）、`make alltest`（vet+unit+e2e）。
- **单元测试**：分散在各包同目录 `*_test.go`（如 `client/control_session_test.go`、`server/control_test.go`、`pkg/config/load_test.go`、`pkg/msg/*_test.go`）。
- **E2E（Ginkgo/Gomega）**：`test/e2e/`
  - `framework/framework.go`：每个 spec `BeforeEach` 建临时目录、起 mock、分配端口；`AfterEach` 停进程。端口分配器 `test/e2e/pkg/port`（按并行节点分区间 10000–30000）。
  - `framework/process.go`：把编译好的 `bin/frps`、`bin/frpc` 当子进程启停（路径由 `-frps-path`/`-frpc-path` 传入，可指定旧版本做兼容测试）。
  - `mock/server/`：模拟后端服务、OIDC 等。
  - 用例：`v1/basic/`（tcp/http/udp/tcpmux/xtcp/oidc/配置）、`v1/features/`（带宽/心跳/监控/SSH tunnel/store/control 替换）、`v1/plugin/`、`compatibility/`（跨版本）。
  - 入口：`test/e2e/e2e_test.go` → `RunE2ETests`。
- **验证习惯**：改控制面/协议优先写或改一个 `test/e2e/v1/basic/` 下的 E2E 用例，比单测更能暴露链路问题。

---

## 8. 新人第一次读代码的推荐顺序

1. `cmd/frps/root.go` 与 `cmd/frpc/sub/root.go`（总入口）
2. `pkg/config/v1/{server,client,proxy}.go`（配置结构体 = 一切输入）
3. `pkg/msg/msg.go`（C/S 词汇表）
4. `client/service.go` → `client/control_session.go` → `client/control.go`（登录、持控制连接、响应 ReqWorkConn、管代理）
5. `client/connector.go`（物理连接与多路复用抽象）
6. `server/service.go`（`handleConnection` 与监听初始化）→ `server/control.go`（`handleNewProxy`/`GetWorkConn`/`RegisterWorkConn`）
7. **`server/proxy/proxy.go`+`tcp.go` 配 `client/proxy/proxy.go`+`general_tcp.go`**（把 TCP 代理从注册到端到端转发读通——关键一跃）
8. `pkg/auth/` · `pkg/proto/wire/` · `pkg/transport/message.go`（鉴权与 v2 线协议）
9. `pkg/config/load.go`（配置解析、模板、legacy、动态源）
10. 跑一遍 `test/e2e/v1/basic/tcp.go` 把抽象落到行为

> 精髓回顾：**先吃透两个 `Control`（`client/control.go` 与 `server/control.go`），整个项目就理顺了。**
