#!/bin/bash
# 部署高架题库到阿里云：Go 服务 + 前端静态 + nginx 代理，暴露 5173
# 用法：在 gaojia-assembly 项目根目录执行 ./00-doc/deploy/deploy.sh
# 前提：本机可免密 ssh root@8.156.79.217

set -e
HOST="8.156.79.217"
DEPLOY_DIR="/opt/gaojia"
WEB_DIR="$DEPLOY_DIR/web"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BACKEND="$ROOT/backend-go"
FRONTEND="$ROOT/frontend"

echo "=== 1. 本地构建 ==="
cd "$BACKEND"
GOOS=linux GOARCH=amd64 go build -o gaojia-server ./cmd/server
echo "Go binary: $(ls -la gaojia-server)"

cd "$FRONTEND"
VITE_API_BASE=/api npm run build --if-present
echo "Frontend dist: $(ls -la dist 2>/dev/null || true)"

echo "=== 2. 上传到 $HOST ==="
ssh "root@$HOST" "mkdir -p $WEB_DIR"
# 打包前端 dist 再上传解压，避免依赖 rsync
(cd "$FRONTEND/dist" && tar cf - .) | ssh "root@$HOST" "rm -rf $WEB_DIR/* 2>/dev/null; mkdir -p $WEB_DIR; cd $WEB_DIR && tar xf -"
scp "$BACKEND/gaojia-server" "root@$HOST:$DEPLOY_DIR/"
[ -f "$BACKEND/gaojia.db" ] && scp "$BACKEND/gaojia.db" "root@$HOST:$DEPLOY_DIR/"
scp "$SCRIPT_DIR/nginx-gaojia.conf" "root@$HOST:/tmp/gaojia-nginx.conf"
scp "$SCRIPT_DIR/gaojia-server.service" "root@$HOST:/tmp/gaojia-server.service"

echo "=== 3. 远程安装与启动 ==="
ssh "root@$HOST" bash -s << 'REMOTE'
set -e
# 安装 nginx（如未安装）
if ! command -v nginx &>/dev/null; then
  yum install -y nginx 2>/dev/null || dnf install -y nginx
fi
mkdir -p /etc/nginx/conf.d
mv /tmp/gaojia-nginx.conf /etc/nginx/conf.d/gaojia.conf 2>/dev/null || true
mv /tmp/gaojia-server.service /etc/systemd/system/gaojia-server.service
mkdir -p /opt/gaojia
chmod +x /opt/gaojia/gaojia-server 2>/dev/null || true
systemctl daemon-reload
systemctl enable gaojia-server
systemctl restart gaojia-server
systemctl enable nginx
systemctl restart nginx
# 放行 5173（firewalld）
if command -v firewall-cmd &>/dev/null; then
  firewall-cmd --permanent --add-port=5173/tcp 2>/dev/null || true
  firewall-cmd --reload 2>/dev/null || true
fi
echo "--- 服务状态 ---"
systemctl is-active gaojia-server nginx || true
REMOTE

echo "=== 部署完成 ==="
echo "访问: http://$HOST:5173"
echo "若外网无法访问，请在阿里云控制台 安全组 中放行入方向 TCP 5173。"
