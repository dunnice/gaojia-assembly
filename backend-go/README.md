# 高架题库 Go 后端

基于 Go + Gin + SQLite 的后端服务，API 与原有 Java/Spring Boot 版本兼容。

## 特性

- **SQLite 内置**：数据库文件直接随服务部署，无需单独安装 MySQL
- **单文件部署**：编译后可生成单一可执行文件
- **API 兼容**：与前端完全兼容，无需修改前端调用

## 依赖安装

若网络受限，可设置国内代理：

```bash
go env -w GOPROXY=https://goproxy.cn,direct
go mod tidy
```

### 1. 数据迁移（从现有 MySQL 迁移）

确保 MySQL 服务可访问，设置连接串后执行迁移：

```bash
# 使用默认连接（与 application.yml 中的配置一致）
MYSQL_DSN="ruankao_user:密码@tcp(127.0.0.1:3306)/ruankao_gaojia?charset=utf8mb4&parseTime=true" \
  go run ./cmd/migrate -db gaojia.db

# 或指定输出路径
go run ./cmd/migrate -db ./data/gaojia.db
```

### 2. 启动服务

```bash
go run ./cmd/server -db gaojia.db -port 5170
```

或编译后运行：

```bash
go build -o gaojia-server ./cmd/server
./gaojia-server -db gaojia.db -port 5170
```

### 3. 启动前端

```bash
cd ../frontend && pnpm dev
```

前端默认请求 `http://localhost:5170/api`，与 Go 服务端口一致。

## 配置

| 参数 | 环境变量 | 默认值 | 说明 |
|------|----------|--------|------|
| `-db` | - | gaojia.db | SQLite 数据库文件路径 |
| `-port` | - | 5170 | HTTP 端口 |
| - | DEFAULT_USER_ID | 1 | 默认用户 ID |

## 项目结构

```
backend-go/
├── cmd/
│   ├── server/    # 主服务
│   └── migrate/   # MySQL → SQLite 迁移工具
├── internal/
│   ├── db/        # 数据库初始化与 schema
│   ├── dto/       # 数据模型
│   ├── handler/   # HTTP 处理
│   ├── repository/# 数据访问
│   └── service/   # 业务逻辑
├── go.mod
└── README.md
```

## API 端点

与 Java 版本一致：

- `GET /api/chapters/tree` - 章节树
- `GET /api/chapters/:chapterId/questions` - 章节题目
- `GET /api/chapters/:chapterId/favorites` - 收藏题目
- `GET /api/questions/:questionId` - 题目详情
- `POST /api/questions/:questionId/status` - 更新状态
- `POST /api/questions/:questionId/answer` - 提交答案
- `GET /api/review/wrongs?chapterId=` - 错题列表
- `GET /api/review/summary` - 复习统计
