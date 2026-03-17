# MySQL 同步到 SQLite 与 Go 数据源切换

## 1. 从 MySQL 同步到 SQLite（可重构）

在 `backend-go` 目录下执行。数据库账号与 `crawler/config.json` 一致（可设置 `MYSQL_DSN` 覆盖）。

```bash
cd backend-go

# 使用 config 中的账号（或设置环境变量 MYSQL_DSN）
export MYSQL_DSN="ruankao_user:ruankao_user123@tcp(127.0.0.1:3306)/ruankao_gaojia?charset=utf8mb4&parseTime=true"

# 全量同步并重建 SQLite（删除已有 gaojia.db 后重新创建）
go run ./cmd/migrate -db gaojia.db -rebuild

# 仅增量覆盖（不删库，只清表后重新导入）
go run ./cmd/migrate -db gaojia.db
```

- **-db**：输出的 SQLite 文件路径，默认 `gaojia.db`。
- **-rebuild**：先删除该文件再创建，相当于重构 SQLite 数据库。

## 2. Go 服务使用 SQLite（默认）

服务默认已使用 **SQLite**，数据文件为 `gaojia.db`（与 migrate 的 `-db` 对应）。

```bash
cd backend-go
go run ./cmd/server
# 等价于
go run ./cmd/server -driver sqlite -db gaojia.db -port 5170
```

如需改用 MySQL，需显式指定：

```bash
go run ./cmd/server -driver mysql -mysql-dsn "user:pass@tcp(127.0.0.1:3306)/ruankao_gaojia"
# 或
MYSQL_DSN="..." go run ./cmd/server -driver mysql
```

## 3. 流程小结

1. **首次或重构**：`go run ./cmd/migrate -db gaojia.db -rebuild`（保证 MySQL 已可连）。
2. **启动后端**：`go run ./cmd/server`（使用当前目录下 `gaojia.db`）。
3. 之后若 MySQL 有更新，可再执行一次 migrate（不加 `-rebuild` 为覆盖同步，加 `-rebuild` 为完全重建）。
