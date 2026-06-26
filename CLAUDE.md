# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

基于 go-zero v1.10.2 的微服务项目，按照 go-zero 推荐的项目布局组织：

- **`api/user/`** — user-api REST 服务，对外提供用户注册/登录/查询等 HTTP 接口
- **`rpc/user/`** — user-rpc gRPC 服务，提供用户 CRUD 接口，操作数据库
- **`proto/`** — 共享的 protobuf 定义
- **`internal/database/`** — 共享的 GORM 数据库配置和初始化

## 构建与运行

```bash
# 启动基础设施（etcd + Jaeger + etcd-keeper）
docker compose up -d

# 启动 user-rpc（gRPC，先启动）
go run rpc/user/service.go -f rpc/user/etc/user-rpc.yaml

# 启动 user-api（REST）
go run api/user/user.go -f api/user/etc/user-api.yaml

# 从 proto 重新生成 gRPC 代码（修改 proto/user.proto 后）
protoc --proto_path=. --go_out=. --go-grpc_out=. proto/user.proto
# 生成后移动文件: mv gozero-demo/rpc/user/pb/*.go rpc/user/pb/ && rm -rf gozero-demo

# 跨平台打包
./scripts/build.sh v1.0.0
```

## 基础设施

| 服务 | 地址 | 说明 |
|------|------|------|
| etcd | :2379 | 服务注册中心 |
| Jaeger UI | :16686 | 链路追踪 |
| Jaeger OTLP | :4317 | 追踪数据上报 |
| etcd-keeper | :18080 | etcd 可视化 |

## 接口列表

| 接口 | 方法 | 鉴权 | 权限 | 说明 |
|------|------|------|------|------|
| `/user/register` | POST | 无 | — | 注册，bcrypt 哈希密码 |
| `/user/login` | POST | 无 | — | 登录，校验 DB → 签发 JWT（含 roles+permissions） |
| `/user/sms/send` | POST | 无 | — | 发送验证码到手机号（60 秒限频） |
| `/user/sms/register` | POST | 无 | — | 短信验证码注册，自动生成用户名 |
| `/user/sms/login` | POST | 无 | — | 短信验证码登录，校验成功签发 JWT |
| `/user/info` | POST | JWT | — | 查当前用户（userId 从 JWT 取） |
| `/user/update` | PUT | JWT | — | 修改邮箱 |
| `/user/password` | PUT | JWT | — | 修改密码（需旧密码验证） |
| `/user/list` | GET | JWT | — | 用户列表（分页） |
| `/role/list` | GET | JWT | — | 角色列表 |
| `/permission/list` | GET | JWT | — | 权限列表 |
| `/role/assign` | POST | JWT | role:assign | 分配角色 |
| `/role/remove` | POST | JWT | role:assign | 移除角色 |
| `/role/permission/assign` | POST | JWT | role:assign | 给角色分配权限 |
| `/role/permission/remove` | POST | JWT | role:assign | 移除角色权限 |

API 文档: `docs/openapi-v3.json`（OpenAPI 3.0，可导入 Apifox/Postman/Swagger UI）

## RBAC 权限模型

```
users ──→ user_roles ──→ roles ──→ role_permissions ──→ permissions
                                          │
登录时 JWT claims 自动包含:                  │
  { userId, username, roles: [codes],        │
    permissions: [codes] }                   │
                              ┌──────────────┘
PermissionMiddleware 校验:      │
  需 "role:assign" ─ 对比 JWT claims 中的 permissions 数组
  无权限 → 403 { Code: 40300, Message: "无权限访问" }
```

### 种子数据（user-rpc 启动时自动初始化）

| 角色 | 权限 |
|------|------|
| admin | user:read, user:write, user:delete, role:read, role:assign |
| user  | user:read |

## 架构

### REST 服务分层（api/user）

- **`user.go`** — 主入口，加载配置、初始化服务上下文、注册路由和中间件、启动服务
- **`internal/config/config.go`** — 配置结构体，内嵌 `rest.RestConf`，含 Auth、RateLimit、UserRpc、DB 配置
- **`internal/svc/servicecontext.go`** — 服务上下文，持有 DB、RPC 客户端、限流器等
- **`internal/handler/routes.go`** — 路由注册，按鉴权分组（公开: register/login/sms, JWT: info/update/password/list）
- **`internal/handler/xxxhandler.go`** — HTTP handler 层，负责解析请求、调用 logic、写回响应
- **`internal/logic/xxxlogic.go`** — 业务逻辑层。loginlogic 直查 DB + bcrypt；userinfologic 调 rpc + 降级 DB
- **`internal/types/types.go`** — 请求/响应结构体
- **`internal/middleware/`** — 自定义中间件（日志、限流、JWT 辅助）

### gRPC 服务分层（rpc/user）

- **`service.go`** — 主入口，实现 `pb.UserServer` 接口，通过 `zrpc.MustNewServer` 自动注册到 etcd
- **`internal/config/config.go`** — 配置结构体，内嵌 `zrpc.RpcServerConf`
- **`internal/svc/servicecontext.go`** — 服务上下文，启动时 `AutoMigrate` 建表
- **`internal/logic/`** — 8 个 RPC 的业务逻辑（GetUser, ListUsers, CreateUser, UpdateUser, UpdatePassword, SendSms, SmsRegister, SmsLogin），全部操作 DB
- **`internal/model/user.go`** — GORM 模型定义

### 中间件执行顺序

```
内置中间件 → JWT 鉴权(per-route, rest.WithJwt) → 自定义日志中间件(server.Use) → 限流中间件(server.Use) → Handler
```

### 跨服务调用链

```
user-api (REST) → [etcd 发现] → user-rpc (gRPC) → DB
                      ↓ 失败时
                 降级: 直查 DB（部分接口）
```

用户密码使用 bcrypt 哈希存储，明文不落库。

## 配置说明

所有配置文件均使用 `conf.MustLoad` 加载，结构体中有 `json:",default=..."` 标签的字段可省略。

### user-api 配置（`api/user/etc/user-api.yaml`）

| 配置项 | 说明 |
|--------|------|
| `Telemetry` | OpenTelemetry 导出配置 |
| `Auth` | JWT 密钥和过期时间 |
| `RateLimit` | 限流开关/频率/突发数 |
| `UserRpc` | user-rpc 客户端，含 etcd 服务发现 |
| `DB` | GORM 数据库连接配置 |

### user-rpc 配置（`rpc/user/etc/user-rpc.yaml`）

| 配置项 | 说明 |
|--------|------|
| `Etcd` | 注册到 etcd 的 hosts+key |
| `Telemetry` | 链路追踪 |
| `DB` | GORM 数据库连接 |
