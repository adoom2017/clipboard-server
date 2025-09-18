# Clipboard Sync Server

基于 Go + Gin 框架开发的剪贴板同步服务端，为 Flutter 客户端提供剪贴板数据的云端存储和同步功能。

## 功能特性

- 🔐 **用户认证**：JWT 令牌认证，支持注册、登录、令牌刷新
- 📋 **剪贴板管理**：支持文本、图片、文件等多种类型的剪贴板内容
- 🔄 **批量同步**：支持批量上传和同步剪贴板项目
- 📊 **数据统计**：提供用户剪贴板使用统计和分析
- 🛡️ **安全防护**：限流、CORS、内容大小限制等安全措施
- 🗄️ **SQLite 数据库**：轻量级数据库，支持 WAL 模式提升性能
- 📖 **API 文档**：完整的 RESTful API 接口

## 快速开始

### 环境要求

- Go 1.21 或更高版本
- SQLite3

### 安装部署

1. **克隆项目**
```bash
git clone <repository-url>
cd clipboard-sync-server
```

2. **安装依赖**
```bash
go mod download
```

3. **配置环境**
```bash
cp .env.example .env
# 编辑 .env 文件，修改相关配置
```

4. **运行服务**
```bash
go run main.go
```

### 配置说明

主要配置项说明：

| 配置项 | 默认值 | 说明 |
|-------|-------|------|
| SERVER_HOST | localhost | 服务器监听地址 |
| SERVER_PORT | 8080 | 服务器监听端口 |
| JWT_SECRET | - | JWT 签名密钥（生产环境必须修改）|
| JWT_EXPIRE_HOUR | 168 | JWT 令牌过期时间（小时）|
| DB_PATH | data/clipboard.db | SQLite 数据库文件路径 |
| MAX_CONTENT_SIZE | 1048576 | 剪贴板内容最大大小（字节）|
| RATE_LIMIT_RPS | 100 | 限流：每秒最大请求数 |

更多配置项请参考 `.env.example` 文件。

### 构建部署

#### 本地构建
```bash
# 构建可执行文件
go build -o clipboard-sync-server

# 运行
./clipboard-sync-server
```

#### Docker 部署

##### 简单部署（仅后端服务）
```bash
# 构建镜像
docker build -t clipboard-sync-server .

# 运行容器
docker run -p 8080:8080 -v $(pwd)/data:/app/data clipboard-sync-server
```

##### Docker Compose 部署（推荐）

本项目提供了包含 Nginx 反向代理的完整 Docker Compose 配置：

```bash
# 1. 配置 Nginx（自动选择配置）
./configure-nginx.sh    # Linux/Mac
configure-nginx.bat     # Windows

# 2. 启动所有服务
docker compose up -d --build

# 3. 查看服务状态
docker compose ps

# 4. 查看日志
docker compose logs -f
```

**服务架构**：
- `clipboard-sync-server`: 后端 Go 服务（端口 8080）
- `nginx`: Nginx 反向代理
  - HTTP 模式：端口 80, 8081（调试）
  - HTTPS 模式：端口 80（重定向）, 443, 8080（开发）

**配置选择**：
- **HTTP 模式**：适用于开发环境和内网部署，无需 SSL 证书
- **HTTPS 模式**：适用于生产环境，需要 SSL 证书

**SSL 证书管理**：
```bash
# 生成自签名证书（用于测试）
./generate-ssl.sh       # Linux/Mac
generate-ssl.bat        # Windows

# 或放置您的证书文件到 ssl/ 目录：
# ssl/server.crt (证书文件)
# ssl/server.key (私钥文件)
```

**服务访问**：
- HTTP 模式：`http://localhost/api/v1/`
- HTTPS 模式：`https://localhost/api/v1/`
- 健康检查：`http://localhost/health`

详细的 Nginx 配置说明请参考：[nginx.conf/README.md](nginx.conf/README.md)

## API 接口

### 认证相关

#### 用户注册
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "user123",
  "email": "user@example.com",
  "password": "password123"
}
```

#### 用户登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "user123",
  "password": "password123"
}
```

### 剪贴板相关

#### 获取剪贴板列表
```http
GET /api/v1/clipboard/items?page=1&page_size=20
Authorization: Bearer <token>
```

#### 创建剪贴板项目
```http
POST /api/v1/clipboard/items
Authorization: Bearer <token>
Content-Type: application/json

{
  "content": "Hello World",
  "type": "text",
  "timestamp": "2023-12-01T10:00:00Z"
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
      "content": "Item 1",
      "type": "text"
    },
    {
      "content": "Item 2",
      "type": "text"
    }
  ]
}
```

#### 获取统计信息
```http
GET /api/v1/clipboard/statistics
Authorization: Bearer <token>
```

### 系统相关

#### 健康检查
```http
GET /api/v1/system/health
```

#### 系统信息
```http
GET /api/v1/system/info
```

## 项目结构

```
server/
├── main.go                 # 主入口文件
├── go.mod                  # Go模块依赖
├── .env.example            # 环境配置示例
├── auth/                   # JWT认证模块
│   └── jwt.go
├── config/                 # 配置管理
│   └── config.go
├── database/              # 数据库相关
│   └── database.go
├── handlers/              # HTTP处理器
│   ├── auth_handler.go    # 认证处理器
│   └── clipboard_handler.go # 剪贴板处理器
├── middleware/            # 中间件
│   └── middleware.go
├── models/               # 数据模型
│   └── models.go
├── utils/                # 工具函数
│   └── utils.go
├── data/                 # 数据目录（运行时创建）
└── logs/                 # 日志目录（运行时创建）
```

## 数据库设计

### 用户表 (users)
| 字段 | 类型 | 说明 |
|-----|------|-----|
| id | VARCHAR | 主键，UUID |
| username | VARCHAR | 用户名，唯一 |
| email | VARCHAR | 邮箱，唯一 |
| password | VARCHAR | 加密后的密码 |
| token | VARCHAR | JWT令牌 |
| is_active | BOOLEAN | 是否激活 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 剪贴板项目表 (clipboard_items)
| 字段 | 类型 | 说明 |
|-----|------|-----|
| id | VARCHAR | 主键，UUID |
| user_id | VARCHAR | 用户ID，外键 |
| content | TEXT | 剪贴板内容 |
| type | VARCHAR | 内容类型 |
| is_synced | BOOLEAN | 是否已同步 |
| synced_at | DATETIME | 同步时间 |
| timestamp | DATETIME | 剪贴板创建时间 |
| created_at | DATETIME | 记录创建时间 |
| updated_at | DATETIME | 记录更新时间 |

## 安全考虑

1. **密码加密**：使用 bcrypt 加密存储密码
2. **JWT 安全**：令牌过期机制，生产环境必须修改默认密钥
3. **限流保护**：防止暴力请求攻击
4. **CORS 配置**：限制跨域访问来源
5. **内容过滤**：自动检测和隐藏敏感内容
6. **大小限制**：限制请求和内容大小

## 性能优化

1. **数据库优化**
   - 启用 SQLite WAL 模式
   - 创建必要的索引
   - 定期清理过期数据

2. **内存优化**
   - 分页查询大量数据
   - 内容截断和压缩
   - 连接池管理

3. **缓存策略**
   - JWT 令牌缓存
   - 统计数据缓存

## 监控和日志

- **请求日志**：记录所有 API 请求
- **错误日志**：记录应用程序错误
- **性能指标**：响应时间、请求量等
- **健康检查**：数据库连接状态检查

## 开发指南

### 添加新的 API 接口

1. 在 `models/` 目录添加请求/响应模型
2. 在 `handlers/` 目录添加处理器函数
3. 在 `main.go` 中注册路由
4. 更新 API 文档

### 数据库迁移

项目使用 GORM 的自动迁移功能，新增字段时：

1. 修改 `models/` 中的结构体
2. 重启应用，GORM 会自动更新表结构

### 测试

```bash
# 运行所有测试
go test ./...

# 运行特定测试
go test ./handlers -v

# 生成测试覆盖率报告
go test -cover ./...
```

## 故障排查

### 常见问题

1. **数据库锁定错误**
   - 检查是否有多个程序同时访问数据库文件
   - 确保数据目录有写权限

2. **JWT 认证失败**
   - 检查 JWT_SECRET 配置
   - 确认令牌未过期

3. **CORS 错误**
   - 检查 CORS_ALLOW_ORIGINS 配置
   - 确认客户端域名在允许列表中

### 日志查看

```bash
# 查看应用日志
tail -f logs/app.log

# 查看错误日志
grep ERROR logs/app.log
```

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目基于 MIT 许可证开源。详见 [LICENSE](LICENSE) 文件。

## 联系方式

- 项目地址：[GitHub Repository]
- 问题反馈：[Issues]
- 邮箱：[email@example.com]

---

**注意**：生产环境部署前，请务必修改默认的 JWT_SECRET 和其他安全相关配置。