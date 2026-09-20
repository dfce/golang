# go-module

基于 Gin、GORM 和 Redis 的 Go 服务端模块化脚手架。

项目按业务域组织代码。每个业务域作为一个独立 Module，内部包含路由、Handler、Service、Repository 和业务 DTO，基础设施能力由 `internal/platform` 提供，应用启动和依赖组装由 `internal/bootstrap` 负责。

当前仓库包含 `demo` 和 `user` 两个模块，主要用于验证：

- Gin 路由注册和统一响应
- Trace ID 传递
- Request Context 和协作式超时
- PostgreSQL、Redis 初始化
- JWT 生成与解析
- 用户创建接口和模块化业务装配
- Zap 日志和 GORM SQL 日志

## 技术栈

- Go 1.27.1
- Gin v1.12
- GORM v1.31
- PostgreSQL
- Redis
- Zap
- JWT v5

## 项目结构

```text
.
├── cmd/
│   └── server/
│       └── main.go                 # 程序入口
├── internal/
│   ├── bootstrap/
│   │   └── app.go                  # 配置、数据库、Redis、HTTP Server 初始化
│   ├── config/
│   │   └── config.go               # 环境变量加载和配置解析
│   ├── middleware/
│   │   ├── errorHandler.go         # Gin 错误统一处理
│   │   ├── recover.go              # Panic 恢复
│   │   ├── requestInfo.go          # 开发环境请求信息
│   │   ├── teace.go                # Trace ID 中间件
│   │   └── timeout.go              # 请求超时 Context
│   ├── models/
│   │   ├── base.go                 # 公共持久化字段
│   │   └── user.go                 # 当前示例持久化模型和遗留 DTO
│   ├── modules/
│   │   ├── module.go               # 模块依赖组装和全局路由注册
│   │   ├── demo/
│   │   │   ├── handler.go          # HTTP Handler
│   │   │   ├── module.go           # Demo 模块构造和路由挂载
│   │   │   ├── repository.go       # 数据访问
│   │   │   ├── router.go           # Demo 路由
│   │   │   ├── schema.go           # Demo 请求/响应结构
│   │   │   └── service.go          # Demo 业务逻辑
│   │   └── user/
│   │       ├── handler.go          # 用户 HTTP Handler
│   │       ├── module.go           # 用户模块构造和路由挂载
│   │       ├── repository.go       # 用户数据访问
│   │       ├── router.go           # 用户路由
│   │       ├── schema.go           # 用户请求/响应结构
│   │       └── service.go          # 用户业务逻辑
│   └── platform/
│       ├── datastore/
│       │   ├── database.go         # GORM 数据库连接和迁移
│       │   └── redis.go            # Redis 客户端
│       └── logging/
│           ├── logger.go           # Zap 日志初始化
│           └── gorm.logger.go      # GORM 日志适配器
├── pkg/
│   ├── apperror/                   # 应用错误类型
│   ├── base/                       # Service 基础能力
│   ├── constant/                   # 公共常量
│   ├── jwt/                        # JWT Service
│   ├── response/                   # 统一 HTTP 响应
│   └── util/                       # 通用工具
├── .env.example                    # 本地配置模板
├── go.mod
└── go.sum
```

## 模块开发约定

新增业务模块时，推荐使用以下结构：

```text
internal/modules/user/
├── handler.go
├── module.go
├── model.go
├── repository.go
├── router.go
├── schema.go
└── service.go
```

职责约定：

- `router.go`：只负责路由注册。
- `handler.go`：绑定 HTTP 参数、调用 Service、返回 HTTP 响应。
- `service.go`：实现业务逻辑，接收标准 `context.Context`，不依赖 `*gin.Context`。
- `repository.go`：访问数据库、Redis 或其他外部存储。
- `schema.go`：当前业务模块的请求 DTO、响应 DTO 和分页结构。
- `model.go`：当前业务模块的持久化模型或领域模型。
- `module.go`：组装本模块的 Repository、Service、Handler 和 Router。

业务模块自己的 DTO 应放在该模块的 `schema.go` 中，不建议将用户、订单等业务 DTO 继续集中放入全局 `internal/models`。数据库模型也不应直接作为 API 响应，尤其不能将密码等敏感字段返回给客户端。

## 本地启动

### 前置依赖

- Go 1.27.1 或兼容版本
- PostgreSQL
- Redis（可选；`REDIS_ENABLED=false` 时可以不启动）

### 配置环境变量

```bash
cp .env.example .env
```

至少需要确认以下配置：

```dotenv
ENV=dev
APP_PORT=8080

DB_CONNECTIONS=primary
DB_PRIMARY_DRIVER=postgres
DB_PRIMARY_DSN="host=localhost user=app password=<change-me> dbname=test port=5432 sslmode=disable"

# 没有 Redis 时设为 false
REDIS_ENABLED=false
```

开发环境会自动加载当前工作目录下的 `.env`。已有系统环境变量不会被 `.env` 覆盖。

### 安装依赖并启动

```bash
go mod download
go test ./...
go vet ./...
go run ./cmd/server
```

不需要重复执行 `go mod init`，项目已经包含 `go.mod`。

## 主要配置

| 配置项 | 说明 |
| --- | --- |
| `ENV` | 运行环境。`dev`、`develop`、`development`、`local` 会按开发环境处理 |
| `APP_NAME` | 应用名称，也用于日志文件名 |
| `APP_PORT` | HTTP 服务端口 |
| `HTTP_READ_TIMEOUT` | HTTP Header 读取超时 |
| `HTTP_WRITE_TIMEOUT` | HTTP 响应写入超时 |
| `HTTP_IDLE_TIMEOUT` | Keep-Alive 空闲连接超时 |
| `HTTP_REQUEST_TIMEOUT` | 业务请求 Context 超时，默认 2 秒 |
| `LOG_LEVEL` | `debug`、`info`、`warn`、`error` 等 |
| `LOG_DIR` | 日志目录 |
| `LOG_MOD` | `0` 输出到控制台和文件，`1` 仅文件，`2` 仅控制台 |
| `AUTH_ENABLED` | 是否启用认证能力 |
| `JWT_SECRET_KEY` | JWT 密钥，非开发环境必须配置且至少 32 个字符 |
| `DB_CONNECTIONS` | 数据库名称列表，逗号分隔 |
| `DB_<NAME>_DSN` | 对应数据库连接串 |
| `DB_<NAME>_AUTO_MIGRATE` | 开发环境是否执行 GORM AutoMigrate |
| `REDIS_ENABLED` | 是否连接 Redis |
| `REDIS_ADDR` | Redis 地址 |

Duration 配置支持 Go duration 格式，例如 `2s`、`30m`、`2h`；纯数字按秒处理。

生产环境不要使用 `.env.example` 中的默认密钥、密码或本地连接配置。

## HTTP 接口

默认地址为 `http://127.0.0.1:8080`，实际端口以 `APP_PORT` 为准。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/health` | 进程存活检查，当前不检查数据库和 Redis |
| `GET` | `/api/v1/demo/ready` | Demo 就绪检查，当前主要检查 Redis |
| `GET` | `/api/v1/demo/test-timeout` | 协作式请求超时示例 |
| `GET` | `/api/v1/demo/test-jwt` | JWT 测试接口，仅用于开发验证 |
| `GET` | `/api/v1/demo/` | Demo 基础路由测试 |
| `GET` | `/api/v1/demo/test` | Demo 测试接口 |
| `POST` | `/api/v1/user` | 创建用户 |
| `GET` | `/swagger/index.html` | 开发环境 Swagger UI |

当用户名已存在时，`POST /api/v1/user` 返回 `409 Conflict`，不会把 PostgreSQL 的 `23505` 原始错误暴露给客户端。服务端仍依赖数据库唯一索引处理并发请求。

请求示例：

```bash
curl -i http://127.0.0.1:8080/health
curl -i http://127.0.0.1:8080/api/v1/demo/ready
curl -i http://127.0.0.1:8080/api/v1/demo/test-timeout
curl -i -X POST http://127.0.0.1:8080/api/v1/user \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"change-me","email":"alice@example.com"}'
```

开发环境 Swagger 文档：

```bash
make apidoc
open http://127.0.0.1:8080/swagger/index.html
```

`make apidoc` 会扫描 `cmd/server`、`internal/modules` 和 `pkg/response`，其中最后一个目录用于解析公共响应结构 `response.Body`。
Swagger JSON 地址为 `http://127.0.0.1:8080/swagger/doc.json`。文档的 `basePath` 是 `/api/v1`，因此业务接口路径仍需包含该前缀。

### Swagger 参数和响应模型

`swag` 只会生成被接口注解引用的模型，不会把所有 `schema.go` 中的结构体自动放入 `definitions`。请求 DTO 应通过 `@Param` 引用，响应 DTO 应通过 `@Success` 或 `@Failure` 引用。

请求字段的描述和约束可以写在结构体字段注释和标签中：

```go
type CreateUser struct {
    // 用户名，长度 3-50 个字符，且必须唯一。
    Username string `json:"username" binding:"required,min=3,max=50" minLength:"3" maxLength:"50" example:"alice"`
    // 登录密码，长度 8-72 个字符。
    Password string `json:"password" binding:"required,min=8,max=72" minLength:"8" maxLength:"72" format:"password"`
    // 用户邮箱，可选；填写时必须是合法邮箱地址。
    Email string `json:"email" binding:"omitempty,email,max=100" maxLength:"100" format:"email"`
}
```

其中：

- `binding:"required,min=...,max=..."` 同时用于 Gin 参数校验和 Swagger 的必填、长度限制。
- `format:"email"`、`format:"password"` 描述字段格式。
- `example:"..."` 提供 Swagger UI 示例值。
- 字段前的 Go 注释会生成字段 `description`。

统一响应结构中的 `data` 是 `any`，需要在接口注解中指定真实类型：

```go
// @Success 201 {object} response.Body{data=CreateUserRes} "创建成功"
```

这会生成类似以下结构：

```json
{
  "code": 201,
  "message": "Created",
  "data": {
    "id": 1
  },
  "trace_id": "..."
}
```

如果接口返回 Demo 就绪结果，则使用：

```go
// @Success 200 {object} response.Body{data=HealthStatusRes} "检查完成"
```

不要只写 `response.Body`，否则 Swagger 只能把 `data` 生成为空对象 `{}`。

统一响应格式：

```json
{
  "code": 200,
  "message": "OK",
  "data": {},
  "trace_id": "..."
}
```

服务会读取或生成 `X-Trace-Id` 请求头，并在响应中返回相同的 Trace ID。

## 请求超时约定

`Timeout` 中间件只负责向 `Request.Context()` 注入截止时间。Go 的 Context 是协作式取消，不能强制终止正在运行的 goroutine。

Service 必须接收标准 `context.Context`：

```go
func (s *UserService) Query(ctx context.Context) ([]UserResponse, error) {
    users, err := s.repo.Query(ctx)
    if err != nil {
        return nil, err
    }
    return users, nil
}
```

数据库和 Redis 操作必须继续传递 Context：

```go
db.WithContext(ctx).Find(&users)
redisClient.Ping(ctx)
```

耗时循环或计算也要主动检查：

```go
select {
case <-ctx.Done():
    return nil, ctx.Err()
default:
}
```

如果业务使用不可取消的 `time.Sleep`、不传 Context 的 SQL，或不响应 `ctx.Done()` 的复杂计算，Context 超时后业务仍可能继续执行，客户端也可能在业务完成后才收到响应。

## 日志

应用日志使用 Zap，GORM SQL 日志通过自定义适配器写入 Zap。

当前非生产环境还启用了 Gin 默认请求日志，开发环境的 `RequestInfo` 也可能记录请求信息。请求体、密码、Token 和 Authorization Header 不应进入日志；`/api/v1/demo/test-jwt` 仅用于开发验证，不应暴露到生产环境。

后续建议统一为单一 Zap Access Log，记录以下字段：

- Trace ID
- HTTP method
- 路由模板
- HTTP status
- 请求耗时
- 响应字节数
- 客户端 IP

## 数据库迁移

当前支持开发环境下通过 `DB_<NAME>_AUTO_MIGRATE=true` 执行 GORM AutoMigrate。

生产环境建议改用版本化迁移工具，并将迁移文件纳入版本控制。不要依赖生产启动时自动修改数据库结构。

## 验证命令

```bash
gofmt -l .
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/server
```

当前已包含 Swagger 文档回归测试；后续应继续补充配置、JWT、中间件、健康检查、数据库和模块业务测试。

## 后续优化项

- 清理 `internal/models/user.go` 中遗留的业务 DTO，只保留持久化模型。
- 将业务持久化模型按模块归属，减少全局 `internal/models` 耦合。
- 将 `/health` 和依赖检查拆分为 `livez`、`readyz`。
- 用单一 Zap Access Log 替代 Gin 默认请求日志。
- 增加生产环境版本化数据库迁移。
- 增加 CI、集成测试和依赖服务容器化配置。
