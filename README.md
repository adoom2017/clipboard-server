# 剪贴板同步服务器 (Clipboard Sync Server)

基于 Go + Gin 框架开发的剪贴板同步服务端，为 Flutter 客户端提供剪贴板数据的云端存储和同步功能。支持用户认证、数据同步、实时监控和高性能存储。

### 🐳 容器化部署
- **Docker 支持**：完整的容器化部署方案
- **Docker Compose**：一键部署的编排配置
- **Nginx 代理**：反向代理和负载均衡支持
- **SSL 证书**：HTTPS 安全连接支持

## 🏗️ 技术架构

### 技术栈
- **Web 框架**：Gin (Go 1.21+)
- **数据库**：SQLite + GORM ORM
- **认证**：JWT (JSON Web Tokens)

### 项目结构

```
server/
├── main.go                     # 主入口文件和路由设置
├── go.mod                      # Go 模块依赖
├── go.sum                      # 依赖版本锁定
├── Dockerfile                  # Docker 构建文件
├── docker-compose.yml          # Docker Compose 编排
├── .env.example                # 环境配置示例
├── auth/                       # JWT 认证模块
│   └── jwt.go                  # JWT Token 生成和验证
├── config/                     # 配置管理
│   └── config.go               # 配置加载和验证
├── database/                   # 数据库相关
│   └── database.go             # 数据库连接、迁移和操作
├── handlers/                   # HTTP 请求处理器
│   ├── auth_handler.go         # 用户认证处理器
│   └── clipboard_handler.go    # 剪贴板数据处理器
├── middleware/                 # HTTP 中间件
│   └── middleware.go           # CORS、限流、日志等中间件
├── models/                     # 数据模型
│   └── models.go               # 数据结构定义和验证
├── utils/                      # 工具函数
│   └── utils.go                # 通用工具函数
├── data/                       # 数据存储目录
│   └── clipboard.db            # SQLite 数据库文件
├── logs/                       # 日志文件目录
├── ssl/                        # SSL 证书目录
└── nginx/                      # Nginx 配置文件
    └── nginx.conf
```

## 🚀 快速开始

### 环境要求

- **Go**: 1.21 或更高版本
- **SQLite**: 3.35+ (通常包含在 Go SQLite 驱动中)
- **Docker**: 20.10+ (可选，用于容器化部署)
- **操作系统**: Linux, macOS, Windows

### 本地开发部署

1. **克隆项目**
   ```bash
   git clone https://github.com/your-repo/clipboard-auto.git
   cd clipboard-auto/server
   ```

2. **安装依赖**
   ```bash
   go mod download
   ```

3. **配置环境变量**
   ```bash
   cp .env.example .env
   # 编辑 .env 文件设置你的配置
   vim .env
   ```

4. **创建必要目录**
   ```bash
   mkdir -p data logs uploads
   ```

5. **运行服务器**
   ```bash
   # 开发模式运行
   go run main.go
   
   # 编译并运行
   go build -o clipboard-server
   ./clipboard-server
   ```

### Docker 容器化部署

1. **构建镜像**
   ```bash
   docker build -t clipboard-server .
   ```

2. **使用 Docker Compose 一键部署**
   ```bash
   docker-compose up -d
   ```

3. **查看服务状态**
   ```bash
   docker-compose ps
   docker-compose logs -f clipboard-server
   ```

### 生产环境部署

1. **使用 Docker Compose (推荐)**
   ```bash
   # 生产环境配置
   docker-compose -f docker-compose.prod.yml up -d
   ```

2. **直接部署**
   ```bash
   # 构建生产版本
   CGO_ENABLED=1 GOOS=linux go build -a -ldflags="-s -w" -o clipboard-server
   
   # 设置环境变量
   export GO_ENV=production
   export GIN_MODE=release
   
   # 运行服务
   ./clipboard-server
   ```

## ⚙️ 配置说明

### 环境变量配置

创建 `.env` 文件：

```bash
# 服务器配置
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# JWT 配置
JWT_SECRET=your-super-secure-secret-key-change-in-production
JWT_EXPIRE_HOUR=168  # 7天

# 数据库配置
DB_PATH=data/clipboard.db
DB_DEBUG=false

# CORS 配置
CORS_ALLOW_ORIGINS=*
CORS_ALLOW_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOW_HEADERS=Origin,Content-Type,Authorization,X-Requested-With

# 日志配置
LOG_LEVEL=info
LOG_FILE=logs/server.log

# 内容限制
MAX_CONTENT_SIZE=1048576    # 1MB
CLEANUP_DAYS=30
ENABLE_CLEANUP=true
CLEANUP_INTERVAL=24h

# 限流配置
RATE_LIMIT_RPS=100
RATE_LIMIT_BURST=200

# 文件上传
UPLOAD_MAX_SIZE=10485760    # 10MB
UPLOAD_PATH=uploads/

# 生产环境
GO_ENV=development          # development/production
```

### 安全配置

#### JWT 安全设置
```bash
# 生产环境必须修改
JWT_SECRET=your-256-bit-secret-key-here
JWT_EXPIRE_HOUR=168  # Token有效期
```

#### CORS 安全设置
```bash
# 生产环境应指定具体域名
CORS_ALLOW_ORIGINS=https://yourdomain.com,https://app.yourdomain.com
```

## 📡 API 接口文档

### 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **认证方式**: Bearer Token (JWT)
- **Content-Type**: `application/json`
- **API 版本**: v1

### 认证接口 (Authentication)

#### 用户注册
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

**响应**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "testuser",
  "email": "test@example.com",
  "is_active": true,
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### 用户登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "testuser",     # 支持用户名或邮箱
  "password": "password123"
}
```

**响应**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "testuser",
  "email": "test@example.com",
  "is_active": true,
  "last_login": "2024-01-01T12:00:00Z",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Token 刷新
```http
POST /api/v1/auth/refresh
Authorization: Bearer <current_token>
```

**响应**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-01-08T12:00:00Z"
}
```

### 用户接口 (User)

#### 获取用户资料
```http
GET /api/v1/user/profile
Authorization: Bearer <token>
```

**响应**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "testuser",
  "email": "test@example.com",
  "is_active": true,
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z",
  "last_login": "2024-01-01T12:00:00Z"
}
```

#### 用户登出
```http
POST /api/v1/user/logout
Authorization: Bearer <token>
```

**响应**:
```json
{
  "message": "logout successful"
}
```

### 剪贴板接口 (Clipboard)

#### 创建剪贴板项目
```http
POST /api/v1/clipboard/items
Authorization: Bearer <token>
Content-Type: application/json

{
  "id": "client-generated-uuid",
  "client_id": "device-unique-id",
  "content": "Hello, World!",
  "type": "text",
  "timestamp": "2024-01-01T12:00:00.000000Z"
}
```

**响应**:
```json
{
  "id": "client-generated-uuid",
  "content": "Hello, World!",
  "type": "text",
  "timestamp": "2024-01-01T12:00:00Z",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

#### 获取剪贴板列表
```http
GET /api/v1/clipboard/items?page=1&page_size=20&type=text&since=2024-01-01T00:00:00Z&search=keyword
Authorization: Bearer <token>
```

**查询参数**:
- `page`: 页码 (默认: 1)
- `page_size`: 每页数量 (默认: 20, 最大: 100)
- `type`: 类型过滤 (`text`, `image`, `file`)
- `since`: 时间过滤，获取指定时间后的数据
- `search`: 内容搜索关键词

**响应**:
```json
{
  "items": [
    {
      "id": "uuid-1",
      "content": "Hello, World!",
      "type": "text",
      "timestamp": "2024-01-01T12:00:00Z",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "page_size": 20,
  "total_pages": 5,
  "has_next": true,
  "has_prev": false
}
```

#### 获取单个剪贴板项目
```http
GET /api/v1/clipboard/items/{id}
Authorization: Bearer <token>
```

**响应**:
```json
{
  "id": "uuid-1",
  "content": "Hello, World!",
  "type": "text",
  "timestamp": "2024-01-01T12:00:00Z",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

#### 更新剪贴板项目
```http
PUT /api/v1/clipboard/items/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "content": "Updated content",
  "type": "text"
}
```

**响应**:
```json
{
  "id": "uuid-1",
  "content": "Updated content",
  "type": "text",
  "timestamp": "2024-01-01T12:00:00Z",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:01:00Z"
}
```

#### 删除剪贴板项目
```http
DELETE /api/v1/clipboard/items/{id}
Authorization: Bearer <token>
```

**响应**:
```json
{
  "message": "item deleted successfully"
}
```

#### 批量同步
```http
POST /api/v1/clipboard/sync
Authorization: Bearer <token>
Content-Type: application/json

{
  "items": [
    {
      "id": "client-uuid-1",
      "client_id": "device-id",
      "content": "Content 1",
      "type": "text",
      "timestamp": "2024-01-01T12:00:00.000000Z"
    },
    {
      "id": "client-uuid-2",
      "client_id": "device-id",
      "content": "Content 2",
      "type": "text",
      "timestamp": "2024-01-01T12:01:00.000000Z"
    }
  ]
}
```

**响应**:
```json
{
  "message": "sync completed",
  "synchronized_count": 2,
  "skipped_count": 0,
  "failed_items": []
}
```

#### 单项同步
```http
POST /api/v1/clipboard/sync-single
Authorization: Bearer <token>
Content-Type: application/json

{
  "id": "client-uuid-1",
  "client_id": "device-id",
  "content": "Single item content",
  "type": "text",
  "timestamp": "2024-01-01T12:00:00.000000Z"
}
```

**响应**:
```json
{
  "message": "item synchronized successfully",
  "item": {
    "id": "client-uuid-1",
    "content": "Single item content",
    "type": "text",
    "timestamp": "2024-01-01T12:00:00Z",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

#### 获取统计信息
```http
GET /api/v1/clipboard/statistics
Authorization: Bearer <token>
```

**响应**:
```json
{
  "total_items": 1500,
  "synced_items": 1450,
  "unsynced_items": 50,
  "total_content_size": 2048576,
  "type_distribution": {
    "text": 1200,
    "image": 250,
    "file": 50
  },
  "recent_activity": [
    {
      "date": "2024-01-01",
      "count": 150
    },
    {
      "date": "2023-12-31",
      "count": 120
    }
  ]
}
```

#### 获取最近同步项目
```http
GET /api/v1/clipboard/recent?limit=10
Authorization: Bearer <token>
```

**响应**:
```json
{
  "items": [
    {
      "id": "uuid-1",
      "content": "Recent content",
      "type": "text",
      "timestamp": "2024-01-01T12:00:00Z",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ]
}
```

#### 获取最新单条记录
```http
GET /api/v1/clipboard/latest
Authorization: Bearer <token>
```

**响应**:
```json
{
  "id": "uuid-latest",
  "content": "Latest clipboard content",
  "type": "text",
  "timestamp": "2024-01-01T12:00:00Z",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### 系统接口 (System)

#### 健康检查
```http
GET /api/v1/system/health
```

**响应**:
```json
{
  "status": "ok",
  "timestamp": "2024-01-01T12:00:00Z",
  "service": "clipboard-sync-server",
  "version": "1.0.0",
  "database": "ok",
  "uptime": "72h15m30s"
}
```

#### 系统信息
```http
GET /api/v1/system/info
```

**响应**:
```json
{
  "service": "clipboard-sync-server",
  "version": "1.0.0",
  "environment": "production",
  "config": {
    "max_content_size": 1048576,
    "cleanup_days": 30,
    "rate_limit_rps": 100,
    "rate_limit_burst": 200,
    "upload_max_size": 10485760
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "uptime": "72h15m30s"
}
```

#### 系统统计
```http
GET /api/v1/system/stats
```

**响应**:
```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "uptime": "72h15m30s",
  "database": {
    "status": "connected",
    "open_connections": 5,
    "in_use": 2,
    "idle": 3,
    "user_count": 150,
    "clipboard_item_count": 1500
  }
}
```

### 通用响应格式

#### 成功响应
```json
{
  "message": "operation successful",
  "data": { /* 具体数据 */ }
}
```

#### 错误响应
```json
{
  "error": "error_code",
  "message": "详细错误描述",
  "details": { /* 可选的额外信息 */ }
}
```

#### HTTP 状态码
- `200` OK - 请求成功
- `201` Created - 资源创建成功
- `400` Bad Request - 请求参数错误
- `401` Unauthorized - 未授权或Token无效
- `403` Forbidden - 权限不足
- `404` Not Found - 资源不存在
- `409` Conflict - 资源冲突
- `429` Too Many Requests - 请求过于频繁
- `500` Internal Server Error - 服务器内部错误


## 📄 许可证

本项目采用 MIT 许可证。详细信息请查看 [LICENSE](LICENSE) 文件。

## 🤝 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 本项目
2. 创建功能分支：`git checkout -b feature/amazing-feature`
3. 提交更改：`git commit -m 'Add amazing feature'`
4. 推送到分支：`git push origin feature/amazing-feature`
5. 打开 Pull Request

### 贡献规范

- 遵循 Go 官方代码规范
- 编写单元测试覆盖新功能
- 更新相关文档
- 提交前运行完整测试套件

## 🙏 致谢

- [Gin](https://github.com/gin-gonic/gin) - 高性能 Go Web 框架
- [GORM](https://gorm.io/) - Go 对象关系映射库
- [JWT-Go](https://github.com/dgrijalva/jwt-go) - JWT 实现库
- [SQLite](https://www.sqlite.org/) - 嵌入式数据库
