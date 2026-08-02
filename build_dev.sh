#!/usr/bin/env bash
set -e

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export GO111MODULE=on
LDFLAGS="-s -w"
OUTDIR=bin
mkdir -p "$OUTDIR"

# ---------- 1) frpc (linux/arm64, no web) ----------
echo "[1/5] 构建 frpc  linux/arm64  (无 web) ..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
  go build -trimpath -ldflags "$LDFLAGS" -tags "frpc,noweb" \
  -o "$OUTDIR/frpc" ./cmd/frpc

# ---------- 2) frps (windows/amd64, with web) ----------
# 先构建 frps 前端 (web/frps/dist)，Go 编译期会把 dist 内嵌进二进制
echo "构建 frps 前端 web ..."
if command -v make >/dev/null 2>&1; then
  make -C web/frps build
else
  cd web/frps && npm run build && cd "$ROOT"  # 构建完 frps 前端后回到仓库根目录
fi
echo "[2/5] 构建 frps  windows/amd64  (含 web) ..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
  go build -trimpath -ldflags "$LDFLAGS" -tags "frps" \
  -o "$OUTDIR/frps.exe" ./cmd/frps

echo "[完成] 产物:"
echo "  $OUTDIR/frpc"
echo "  $OUTDIR/frps.exe"

# ---------- 部署相关配置 (可用环境变量覆盖) ----------
#   OPENWRT_HOST / OPENWRT_USER : 路由器地址与账号
#   DEPLOY=0                    : 只本地构建不部署
#   SSH_KEY                     : 本地私钥路径 (默认 ~/.ssh/id_ed25519)
OPENWRT_HOST="${OPENWRT_HOST:-192.168.99.1}"
OPENWRT_USER="${OPENWRT_USER:-root}"
DEPLOY="${DEPLOY:-1}"
SSH_KEY="${SSH_KEY:-$HOME/.ssh/id_ed25519}"
SSH_OPTS="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"

# ---------- 3) 部署前置: 配置 OpenWrt SSH 免密 ----------
if [ "$DEPLOY" = "1" ]; then
  echo ""
  echo "[3/5] 配置 OpenWrt SSH 免密 ($OPENWRT_USER@$OPENWRT_HOST) ..."
  # 1) 若无密钥则生成 (空密码短语 -> 后续全免交互)
  if [ ! -f "$SSH_KEY" ]; then
    echo "  未找到 SSH 密钥, 生成 $SSH_KEY ..."
    mkdir -p "$(dirname "$SSH_KEY")"
    ssh-keygen -t ed25519 -N "" -f "$SSH_KEY"
  fi
  # 2) 已免密则跳过; 否则推送公钥 (仅需输入一次路由器密码)
  if ssh -i "$SSH_KEY" -o BatchMode=yes -o ConnectTimeout=8 $SSH_OPTS ${OPENWRT_USER}@${OPENWRT_HOST} "true" 2>/dev/null; then
    echo "  SSH 已免密, 跳过公钥推送"
  else
    echo "  尚未免密, 推送公钥到路由器 (请输入一次 $OPENWRT_USER@$OPENWRT_HOST 密码) ..."
    # OpenWrt 用 Dropbear, 公钥需写入 /etc/dropbear/authorized_keys 而非 ~/.ssh
    cat "$SSH_KEY.pub" | ssh -i "$SSH_KEY" $SSH_OPTS ${OPENWRT_USER}@${OPENWRT_HOST} \
      "mkdir -p /etc/dropbear && cat >> /etc/dropbear/authorized_keys && echo OK_PUSHED"
    echo "  公钥已推送, 后续部署将免密"
  fi
fi

# ---------- 4) 部署 frpc 到 OpenWrt ----------
if [ "$DEPLOY" = "1" ]; then
  echo ""
  echo "[4/5] 部署 frpc 到 OpenWrt ($OPENWRT_HOST) ..."
  # 删除远端旧文件
  ssh $SSH_OPTS ${OPENWRT_USER}@${OPENWRT_HOST} "rm -f /usr/bin/frpc"
  echo "  已删除远端 /usr/bin/frpc"
  # 上传新文件
  scp $SSH_OPTS "$OUTDIR/frpc" ${OPENWRT_USER}@${OPENWRT_HOST}:/usr/bin/frpc
  echo "  已上传 $OUTDIR/frpc -> /usr/bin/frpc"
  # 赋予可执行权限并重启服务
  ssh $SSH_OPTS ${OPENWRT_USER}@${OPENWRT_HOST} "chmod +x /usr/bin/frpc && /etc/init.d/frpc restart"
  echo "  已赋予执行权限并重启 frpc 服务"
fi

# ---------- 5) 启动本地 frps ----------
echo ""
echo "[5/5] 启动本地 frps ..."
./bin/frps.exe -c frps.toml
