# 高架题库阿里云部署

- **服务器**: 8.156.79.217（Alibaba Cloud Linux 3）
- **端口**: 5173（nginx：前端 + /api 反向代理到后端 5170）
- **目录**: 应用 `/opt/gaojia`，前端静态 `/opt/gaojia/web`

## 一键部署

在项目根目录（gaojia-assembly）执行：

```bash
./00-doc/deploy/deploy.sh
```

前提：本机可免密 `ssh root@8.156.79.217`。

## 阿里云安全组

若外网无法访问，请在 **阿里云控制台 → ECS → 安全组** 中添加入方向规则：

- 端口：5173/tcp
- 授权对象：0.0.0.0/0（或按需限制）

## 后端 SQLite 与 CGO

Go 服务使用 **modernc.org/sqlite**（纯 Go），支持 `CGO_ENABLED=0` 交叉编译，无需在服务器上安装 C 库或 Go 环境。

## 常用命令（服务器）

```bash
ssh root@8.156.79.217
systemctl status gaojia-server nginx
systemctl restart gaojia-server nginx
journalctl -u gaojia-server -f
```
