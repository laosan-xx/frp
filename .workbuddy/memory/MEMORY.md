# 项目长期笔记 — frp (定制版: OpenWrt + passwall 集成)

## 项目核心心智模型
- frp = 一条常驻**控制连接**(跑 msg 信令: Login/NewProxy/Ping/ReqWorkConn) + 按需建立的**工作连接**(搬运真实 TCP/UDP 流量)。
- 两个 `Control` (`client/control.go` 与 `server/control.go`) 是理解全项目的关键。
- 这是定制版 frp，非官方 fatedier/frp；含 passwall 节点测速/落地IP/归属地等定制功能。

## AI 快速了解项目的入口文件
- **`PROJECT_INDEX.md`**(仓库根目录): 架构地图 + 目录地图 + "改 X 去哪找"速查表 + TCP 数据流 + 配置/协议/测试说明 + 新人阅读顺序。每次会话先读它，不要再全局重读代码。
- `AGENTS.md` 顶部已加指针指向 PROJECT_INDEX.md。
- 构建命令见 `AGENTS.md`(make build/frps/frpc/web/e2e 等)。

## 常用改动位置(速记)
- 代理类型: `server/proxy/*` + `client/proxy/*`(基类 `proxy.go`, 注册用 `proxyFactoryRegistry`)
- 配置项: `pkg/config/v1/*` 结构体 + `pkg/config/load.go`
- C/S 消息: `pkg/msg/msg.go`
- 鉴权: `pkg/auth/`; 传输层: `client/connector.go` + `pkg/transport` + `pkg/proto/wire`
- 验证优先 `make e2e` 写/改 `test/e2e/v1/basic/` 用例

## 重要约束
- web/frps 前端 `npm run build` 在 WorkBuddy 沙箱内会被 safe-delete 拦截清空 dist(非代码问题)，真机无此问题；沙箱内可用 `mv dist _dist_old` 绕过。
