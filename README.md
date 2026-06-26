# gozero-demo — go-zero 微服务实战项目

基于 [go-zero](https://github.com/zeromicro/go-zero) v1.10.2 的微服务项目，涵盖中间件、服务注册发现、熔断降级、链路追踪、RBAC 权限等核心能力。

## 服务架构

```
                    ┌──────────────┐
                    │   etcd:2379  │  ← 服务注册中心
                    └──────┬───────┘
                           │ 注册 / 发现
    ┌──────────────────────┼──────────────────────┐
    │                      │                      │
    ▼                      ▼                      ▼
┌──────────┐  gRPC调用  ┌──────────┐         ┌──────────┐
│ user-api │ ─────────→ │ user-rpc │         │  Jaeger  │
│  :8888   │            │  :8080   │  OTLP   │  :16686  │
│  REST    │            │  gRPC    │ ──────→ │  UI       │
└──────────┘            └──────────┘         └──────────┘
    │                        │
    ▼                        ▼
┌──────────┐            ┌──────────┐
│  MySQL   │            │  MySQL   │
│ business │            │ business │
└──────────┘            └──────────┘
```

| 服务 | 类型 | 端口 | 说明 |
|------|------|------|------|
| **user-api** | REST | 8888 | HTTP 接口，JWT 鉴权 + 权限中间件 + 限流，通过 etcd 发现并调用 user-rpc |
| **user-rpc** | gRPC | 8080 | 用户/角色/权限 CRUD，注册到 etcd，操作数据库 |
| **etcd** | 基础设施 | 2379 | 服务注册与发现中心 |
| **Jaeger** | 基础设施 | 16686 (UI) / 4317 (OTLP) | 分布式链路追踪 |
| **etcd-keeper** | 基础设施 | 18080 | etcd 可视化管理 |

## 功能一览

| 功能 | 实现方式 |
|------|---------|
| **统一响应** | `{ Code, Message, Data[] }`，PascalCase 大驼峰 |
| **国际化 i18n** | Accept-Language 自动中英文切换 |
| **JWT 鉴权** | `rest.WithJwt()` 按路由启用，claims 注入 context |
| **RBAC 权限** | 用户 → 角色 → 权限 模型，PermissionMiddleware 中间件 |
| **限流** | 每 IP 令牌桶，超限返回 429 |
| **日志** | 自定义 LoggingMiddleware，结构化日志 |
| **服务发现** | user-rpc 注册到 etcd，user-api 自动发现（p2c_ewma 负载均衡） |
| **熔断降级** | Google SRE breaker + 应用层 fallback（直查 DB） |
| **链路追踪** | OpenTelemetry OTLP 导出到 Jaeger，W3C Trace Context 传播 |
| **GORM 数据库** | 共享 database 包，AutoMigrate 自动建表 |

## 接口列表

### 认证（公开）

| 接口 | 方法 | 说明 |
|------|------|------|
| `/user/register` | POST | 注册，bcrypt 哈希密码 |
| `/user/login` | POST | 登录，校验 DB → 签发 JWT（含 roles+permissions） |

### 用户（需 JWT）

| 接口 | 方法 | 说明 |
|------|------|------|
| `/user/info` | POST | 查询当前用户（userId 从 JWT 取） |
| `/user/update` | PUT | 修改邮箱 |
| `/user/password` | PUT | 修改密码（需旧密码验证） |
| `/user/list` | GET | 用户列表（分页：?Page=1&PageSize=20） |

### 管理（需 JWT + `role:assign` 权限）

| 接口 | 方法 | 说明 |
|------|------|------|
| `/role/list` | GET | 角色列表 |
| `/permission/list` | GET | 权限列表 |
| `/role/assign` | POST | 分配角色 `{ UserId, RoleId }` |
| `/role/remove` | POST | 移除角色 `{ UserId, RoleId }` |
| `/role/permission/assign` | POST | 给角色分配权限 `{ RoleId, PermissionId }` |
| `/role/permission/remove` | POST | 移除角色权限 `{ RoleId, PermissionId }` |

## RBAC 权限模型

```
users ──→ user_roles ──→ roles ──→ role_permissions ──→ permissions
                                          │
JWT claims 自动包含:                        │
  { roles: ["admin"],                       │
    permissions: ["user:read","role:assign"] } ── PermissionMiddleware 校验
```

### 种子数据（user-rpc 启动时自动初始化）

| 角色 | 权限 |
|------|------|
| **admin**（管理员） | user:read, user:write, user:delete, role:read, role:assign |
| **user**（普通用户） | user:read |

## 快速开始

### 1. 启动基础设施

```bash
docker compose up -d
```

### 2. 启动 user-rpc

```bash
go run rpc/user/service.go -f rpc/user/etc/user-rpc.yaml
```

### 3. 启动 user-api

```bash
go run api/user/user.go -f api/user/etc/user-api.yaml
```

### 4. 测试

```bash
# 注册
curl -X POST :8888/user/register \
  -H 'Content-Type: application/json' \
  -d '{"Username":"admin","Password":"123456","Email":"admin@test.com"}'

# 登录
curl -X POST :8888/user/login \
  -H 'Content-Type: application/json' \
  -d '{"Username":"admin","Password":"123456"}'

# 查询自己
curl -X POST :8888/user/info \
  -H 'Authorization: Bearer <TOKEN>'

# 分配角色（需 admin 账号 + role:assign 权限）
curl -X POST :8888/role/assign \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{"UserId":2,"RoleId":1}'

# 用户列表（GET 分页）
curl 'localhost:8888/user/list?Page=1&PageSize=10' \
  -H 'Authorization: Bearer <TOKEN>'
```

## 项目结构

```
gozero-demo/
├── api/user/                         ← user-api REST 服务
│   ├── user.go                       # 入口
│   ├── etc/user-api.yaml             # 配置文件
│   └── internal/
│       ├── config/config.go
│       ├── types/types.go
│       ├── svc/servicecontext.go
│       ├── handler/                  # Handler 层
│       │   ├── routes.go             # 路由（公开/JWT/权限 三组）
│       │   ├── loginhandler.go
│       │   ├── registerhandler.go
│       │   ├── userinfohandler.go
│       │   ├── updateuserhandler.go
│       │   ├── updatepasswordhandler.go
│       │   ├── listusershandler.go
│       │   ├── rbac_handler.go       # 角色/权限管理
│       │   └── error_helper.go       # 错误码映射
│       ├── logic/                    # 业务逻辑层
│       │   ├── loginlogic.go         # 登录(查DB+bcrypt+JWT含RBAC)
│       │   ├── registerlogic.go
│       │   ├── userinfologic.go      # 查用户(rpc+降级DB)
│       │   ├── updateuserlogic.go
│       │   ├── updatepasswordlogic.go
│       │   ├── listuserslogic.go
│       │   └── rbaclogic.go          # 角色/权限逻辑
│       └── middleware/               # 中间件
│           ├── loggingmiddleware.go
│           ├── ratelimitmiddleware.go
│           ├── authhelper.go         # JWT claims 提取
│           └── permissionmiddleware.go # RBAC 权限校验
├── rpc/user/                         ← user-rpc gRPC 服务
│   ├── service.go                    # 入口 (含 RBAC 种子数据)
│   ├── etc/user-rpc.yaml
│   ├── pb/                           # proto 生成代码
│   └── internal/
│       ├── config/config.go
│       ├── svc/servicecontext.go     # AutoMigrate 7 张表
│       ├── model/
│       │   ├── user.go               # users 表
│       │   └── rbac.go               # roles, permissions, user_roles, role_permissions
│       └── logic/
│           ├── getuserlogic.go
│           ├── createuserlogic.go
│           ├── updateuserlogic.go
│           ├── updatepasswordlogic.go
│           ├── listuserslogic.go
│           └── rbaclogic.go          # RBAC CRUD
├── proto/user.proto                  # gRPC 服务定义 (10 个 RPC)
├── internal/
│   ├── database/                     # 共享 GORM 模块
│   ├── response/                     # 统一响应 {Code, Message, Data[]}
│   └── i18n/                         # 中英文国际化
├── docs/openapi-v3.json              # API 文档 (OpenAPI 3.0)
├── scripts/                          # 构建/部署脚本
│   ├── build.sh                      # 跨平台打包
│   ├── install.sh                    # systemd 安装
│   ├── user-api.service
│   └── user-rpc.service
└── docker-compose.yml                # etcd + Jaeger + etcd-keeper
```

## 中间件执行顺序

```
HTTP 请求
  → TraceHandler（链路追踪 root span）
  → PrometheusHandler（监控指标）
  → BreakerHandler（REST 层熔断）
  → TimeoutHandler（超时控制）
  → RecoverHandler（panic 恢复）
  → JWT Auth（per-route，rest.WithJwt）
  → PermissionMiddleware（per-route，RBAC 权限校验）
  → LoggingMiddleware（自定义日志）
  → RateLimitMiddleware（自定义限流）
  → Handler
```

## 统一响应格式

```json
// 成功（单个对象）
{ "Code": 0, "Message": "注册成功", "Data": [{ "Id": 1, "Username": "admin", "Email": "admin@test.com" }] }

// 成功（分页列表）
{ "Code": 0, "Message": "成功", "Data": [{...},{...}], "Total": 100 }

// 成功（无数据）
{ "Code": 0, "Message": "密码修改成功", "Data": [] }

// 错误
{ "Code": 40300, "Message": "无权限访问", "Data": [] }
```

### 业务状态码

| Code | 含义 |
|------|------|
| 0 | 成功 |
| 10001~10004 | 业务错误（用户不存在/已存在/密码/旧密码） |
| 40000 | 参数错误 |
| 40100 | 未授权 |
| 40300 | 无权限 |
| 40400 | 资源不存在 |
| 42900 | 限流 |
| 50000 | 内部错误 |
| 50300 | 服务不可用 |

## 配置说明

### user-api（`api/user/etc/user-api.yaml`）

| 配置项 | 说明 |
|--------|------|
| `Telemetry.Endpoint` | Jaeger OTLP 地址 |
| `Auth.AccessSecret` | JWT 签名密钥 |
| `RateLimit.Enabled/RPS/Burst` | 限流配置 |
| `UserRpc.Etcd.Hosts/Key` | etcd 服务发现 |
| `DB` | GORM 数据库连接 |

### user-rpc（`rpc/user/etc/user-rpc.yaml`）

| 配置项 | 说明 |
|--------|------|
| `ListenOn` | 监听地址 |
| `Etcd.Hosts/Key` | etcd 注册 |
| `Telemetry` | 链路追踪 |
| `DB` | GORM 数据库连接 |
