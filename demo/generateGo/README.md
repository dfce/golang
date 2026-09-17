# 2026-06-26  封装go scaffold 脚手架
gin zap 
- 集成 RequestInfo、TraceId、日志、鉴权 ... 常用中间件
- gorm 数据库 Redis

## 
```sh
# 生产swager 文档
swag init -g cmd/server/main.go -d . -o swdocs
# 指定稳定生成目录，不全部扫描 internal/httpserver/handler
# 将-g 所在目录放在-d 的第一项(cmd/server)，带上行response的目录(pkg)
# 减少扫描目录
swag init  -g main.go  -d cmd/server,internal/httpserver/handler,pkg -o swdocs
```



# Architecture 结构图 


## 结构说明
```
cmd
└── server
    └── main.go


internal
├── bootstrap
│   └── app.go
├── config
│   └── config.go
├── health
│   └── health.go
├── platform
│   └── datastore
│   └── database.go
│     └── redis.go
│   └── healthcheck
│   └── logging
│     └── gorm.logger.go
│     └── logger.go
├── port
│   └── ports.go
├── repository
├── tokenservice
│   └── jwt.go
├── service
├── httpserver
│   ├── middleware
│   └── handler


pkg
├── response
├── errno
├── jwt
└── constant
```

## Directory Graph

```mermaid
flowchart TD
  CMD[cmd/server] --> BOOT[internal/bootstrap]
  BOOT --> CFG[internal/config]
  BOOT --> DS[internal/platform/datastore]
  BOOT --> HC[internal/platform/healthcheck]
  BOOT --> PORT[internal/port]
  BOOT --> REPO[internal/repository]
  BOOT --> SVC[internal/service]
  BOOT --> HTTP[internal/httpserver]
  HTTP --> MW[internal/httpserver/middleware]
  HTTP --> HDL[internal/httpserver/handler]
  HDL --> RESP[pkg/response]
  SVC --> PORT
  REPO --> PORT
  REPO --> DS
  PORT --> MODEL[internal/model]
```

## Request Flow

```mermaid
sequenceDiagram
  participant Client
  participant Gin
  participant Middleware
  participant Handler
  participant Service
  participant Repository
  participant DBRedis as DB / Redis

  Client->>Gin: HTTP request
  Gin->>Middleware: traceId, requestInfo, access log, auth
  Middleware->>Handler: request context
  Handler->>Service: business call
  Service->>Port: interface call
  Port->>Repository: adapter implementation
  Repository->>DBRedis: ping/query/cache
  DBRedis-->>Repository: result
  Repository-->>Service: data
  Service-->>Handler: response data
  Handler-->>Client: JSON + X-Trace-Id
```

## Bootstrap Flow

```mermaid
flowchart LR
  A[config.Load] --> B[logging.New]
  B --> C[datastore.OpenDatabases]
  C --> D[datastore.OpenRedis]
  D --> E[repository adapters]
  E --> F[service.NewRegistry]
  F --> G[explicit route registration]
  G --> H[http.ListenAndServe]
```

## Health probes

`GET /livez` only reports whether the process is alive. It does not access
databases or Redis.

`GET /readyz` checks required dependencies and returns `503` while the service
is starting, stopping, or a required dependency is unavailable. Redis is
reported as disabled when it is optional and `REDIS_ENABLED=false`.
