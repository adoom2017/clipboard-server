# Clipboard Sync Server API Documentation

## 概述

Clipboard Sync Server 是一个基于 Go 和 Gin 框架构建的剪贴板同步服务器，提供用户认证、剪贴板数据管理和多设备同步功能。

**基础信息**
- 基础URL: `http://localhost:8080`
- API版本: v1
- 认证方式: JWT Bearer Token
- 内容类型: `application/json`

---

## 认证

所有需要认证的端点都需要在请求头中包含有效的 JWT 令牌：

```
Authorization: Bearer <jwt_token>
```

---

## API 端点

### 1. 系统端点

#### 1.1 健康检查
```http
GET /api/v1/system/health
```

**描述**: 检查服务器和数据库健康状态

**响应**:
```json
{
  "database": "ok",
  "service": "clipboard-sync-server", 
  "status": "ok",
  "timestamp": "2025-09-17T21:40:20+08:00",
  "uptime": "1m12.1702842s",
  "version": "1.0.0"
}
```

#### 1.2 系统信息
```http
GET /api/v1/system/info
```

**描述**: 获取服务器系统信息

**响应**:
```json
{
  "service": "clipboard-sync-server",
  "version": "1.0.0",
  "go_version": "go1.21.0",
  "build_time": "2025-09-17T13:34:52Z",
  "git_commit": "latest",
  "environment": "development"
}
```

#### 1.3 系统统计
```http
GET /api/v1/system/stats
```

**描述**: 获取数据库统计信息

**需要认证**: 否

**响应**:
```json
{
  "total_users": 2,
  "total_items": 3,
  "total_size": "67 bytes",
  "uptime": "1m30s"
}
```

---

### 2. 用户认证端点

#### 2.1 用户注册
```http
POST /api/v1/auth/register
```

**描述**: 注册新用户账号

**请求体**:
```json
{
  "username": "testuser",
  "email": "test@example.com", 
  "password": "test123456"
}
```

**字段要求**:
- `username`: 3-50字符，字母数字下划线
- `email`: 有效邮箱地址
- `password`: 最少6字符

**响应** (201 Created):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "19f87f65-7dea-47da-a1a6-e450c6709eb2",
    "username": "testuser",
    "email": "test@example.com",
    "created_at": "2025-09-17T13:41:24.156Z"
  }
}
```

**错误响应** (400 Bad Request):
```json
{
  "error": "用户名已存在"
}
```

#### 2.2 用户登录
```http
POST /api/v1/auth/login
```

**描述**: 用户登录验证

**请求体**:
```json
{
  "username": "testuser",
  "password": "test123456"
}
```

**响应** (200 OK):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "19f87f65-7dea-47da-a1a6-e450c6709eb2",
    "username": "testuser", 
    "email": "test@example.com",
    "last_login": "2025-09-17T13:41:32.156Z"
  }
}
```

**错误响应** (401 Unauthorized):
```json
{
  "error": "用户名或密码错误"
}
```

#### 2.3 刷新令牌
```http
POST /api/v1/auth/refresh
```

**描述**: 刷新 JWT 令牌

**需要认证**: 是

**响应** (200 OK):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-09-18T13:41:32Z"
}
```

---

### 3. 用户管理端点

#### 3.1 获取用户资料
```http
GET /api/v1/user/profile
```

**描述**: 获取当前用户资料信息

**需要认证**: 是

**响应** (200 OK):
```json
{
  "id": "19f87f65-7dea-47da-a1a6-e450c6709eb2",
  "username": "testuser",
  "email": "test@example.com", 
  "created_at": "2025-09-17T13:41:24.156Z",
  "last_login": "2025-09-17T13:41:32.156Z",
  "total_items": 3
}
```

#### 3.2 用户登出
```http
POST /api/v1/user/logout
```

**描述**: 用户登出（令牌加入黑名单）

**需要认证**: 是

**响应** (200 OK):
```json
{
  "message": "登出成功"
}
```

---

### 4. 剪贴板管理端点

#### 4.1 获取剪贴板项目列表
```http
GET /api/v1/clipboard/items
```

**描述**: 获取当前用户的剪贴板项目列表

**需要认证**: 是

**查询参数**:
- `page`: 页码（默认: 1）
- `page_size`: 每页数量（默认: 20，最大: 100）
- `type`: 内容类型过滤（text/image/file）
- `device_id`: 设备ID过滤
- `search`: 内容搜索关键词

**响应** (200 OK):
```json
{
  "items": [
    {
      "id": "441dd0d1-7576-46f4-ad0a-7ecb17e602db",
      "content": "Hello World - Test Clipboard Item",
      "type": "text",
      "is_synced": true,
      "synced_at": "2025-09-17T13:41:32.130Z",
      "timestamp": "2025-09-17T13:41:32.130Z",
      "created_at": "2025-09-17T13:41:32.130Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20,
  "total_pages": 1,
  "has_next": false,
  "has_prev": false
}
```

#### 4.2 创建剪贴板项目
```http
POST /api/v1/clipboard/items
```

**描述**: 创建新的剪贴板项目

**需要认证**: 是

**请求体**:
```json
{
  "type": "text",
  "content": "Hello World - Test Clipboard Item",
  "device_id": "test-device-001"
}
```

**字段说明**:
- `type`: 内容类型 (text/image/file)
- `content`: 剪贴板内容（最大1MB）
- `device_id`: 设备标识符

**响应** (201 Created):
```json
{
  "id": "441dd0d1-7576-46f4-ad0a-7ecb17e602db",
  "content": "Hello World - Test Clipboard Item",
  "type": "text",
  "is_synced": false,
  "synced_at": null,
  "timestamp": "2025-09-17T13:41:32.130Z",
  "created_at": "2025-09-17T13:41:32.130Z"
}
```

#### 4.3 获取单个剪贴板项目
```http
GET /api/v1/clipboard/items/:id
```

**描述**: 根据ID获取特定剪贴板项目

**需要认证**: 是

**路径参数**:
- `id`: 剪贴板项目ID

**响应** (200 OK):
```json
{
  "id": "441dd0d1-7576-46f4-ad0a-7ecb17e602db",
  "content": "Hello World - Test Clipboard Item",
  "type": "text",
  "is_synced": true,
  "synced_at": "2025-09-17T13:41:32.130Z",
  "timestamp": "2025-09-17T13:41:32.130Z",
  "created_at": "2025-09-17T13:41:32.130Z"
}
```

#### 4.4 更新剪贴板项目
```http
PUT /api/v1/clipboard/items/:id
```

**描述**: 更新现有剪贴板项目

**需要认证**: 是

**路径参数**:
- `id`: 剪贴板项目ID

**请求体**:
```json
{
  "content": "Updated clipboard content",
  "type": "text"
}
```

**响应** (200 OK):
```json
{
  "id": "441dd0d1-7576-46f4-ad0a-7ecb17e602db",
  "content": "Updated clipboard content",
  "type": "text",
  "is_synced": false,
  "synced_at": null,
  "timestamp": "2025-09-17T13:45:32.130Z",
  "created_at": "2025-09-17T13:41:32.130Z"
}
```

#### 4.5 删除剪贴板项目
```http
DELETE /api/v1/clipboard/items/:id
```

**描述**: 删除指定的剪贴板项目

**需要认证**: 是

**路径参数**:
- `id`: 剪贴板项目ID

**响应** (200 OK):
```json
{
  "message": "剪贴板项目删除成功"
}
```

#### 4.6 批量同步
```http
POST /api/v1/clipboard/sync
```

**描述**: 批量同步剪贴板项目

**需要认证**: 是

**请求体**:
```json
{
  "device_id": "test-device-001",
  "items": [
    {
      "type": "text",
      "content": "Batch sync item 1",
      "timestamp": "2025-09-17T13:41:00Z"
    },
    {
      "type": "text", 
      "content": "Batch sync item 2",
      "timestamp": "2025-09-17T13:42:00Z"
    }
  ]
}
```

**响应** (200 OK):
```json
{
  "synced": [
    {
      "id": "0a423099-038a-4eca-94dd-dd0006a905b2",
      "content": "Batch sync item 1",
      "type": "text",
      "is_synced": true,
      "synced_at": "2025-09-17T13:42:15.095Z",
      "timestamp": "2025-09-17T13:41:00Z",
      "created_at": "2025-09-17T13:42:15.095Z"
    }
  ],
  "failed": [],
  "total": 2
}
```

#### 4.7 获取统计信息
```http
GET /api/v1/clipboard/statistics
```

**描述**: 获取当前用户的剪贴板统计信息

**需要认证**: 是

**响应** (200 OK):
```json
{
  "total_items": 3,
  "synced_items": 2,
  "unsynced_items": 1,
  "total_content_size": 67,
  "type_distribution": {
    "text": 3
  },
  "recent_activity": [
    {
      "date": "2025-09-17",
      "count": 3
    }
  ]
}
```

---

## 错误码说明

| HTTP状态码 | 说明 |
|-----------|------|
| 200 | 请求成功 |
| 201 | 资源创建成功 |
| 400 | 请求参数错误 |
| 401 | 认证失败或令牌无效 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 409 | 资源冲突（如用户名已存在）|
| 413 | 请求体过大 |
| 429 | 请求频率过高 |
| 500 | 服务器内部错误 |

### 错误响应格式

```json
{
  "error": "错误描述信息",
  "code": "ERROR_CODE", 
  "timestamp": "2025-09-17T13:41:32Z"
}
```

---

## 安全特性

### 安全头
- `Content-Security-Policy`: `default-src 'self'`
- `X-Frame-Options`: `DENY`
- `X-Content-Type-Options`: `nosniff`  
- `Referrer-Policy`: `strict-origin-when-cross-origin`

### 限流
- 默认限制: 100 RPS，突发200请求
- 基于IP地址限流
- 超出限制返回 429 状态码

### 认证
- JWT令牌有效期: 24小时
- 密码使用 bcrypt 加密存储
- 令牌黑名单机制

---

## SDK 示例

### JavaScript/Node.js

```javascript
class ClipboardSyncClient {
  constructor(baseUrl = 'http://localhost:8080') {
    this.baseUrl = baseUrl;
    this.token = null;
  }

  async register(username, email, password) {
    const response = await fetch(`${this.baseUrl}/api/v1/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, email, password })
    });
    const data = await response.json();
    if (response.ok) {
      this.token = data.token;
    }
    return data;
  }

  async login(username, password) {
    const response = await fetch(`${this.baseUrl}/api/v1/auth/login`, {
      method: 'POST', 
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    });
    const data = await response.json();
    if (response.ok) {
      this.token = data.token;
    }
    return data;
  }

  async createClipboardItem(type, content, deviceId) {
    const response = await fetch(`${this.baseUrl}/api/v1/clipboard/items`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.token}`
      },
      body: JSON.stringify({ type, content, device_id: deviceId })
    });
    return response.json();
  }

  async getClipboardItems(page = 1, pageSize = 20) {
    const response = await fetch(`${this.baseUrl}/api/v1/clipboard/items?page=${page}&page_size=${pageSize}`, {
      headers: { 'Authorization': `Bearer ${this.token}` }
    });
    return response.json();
  }
}
```

### Python

```python
import requests
import json

class ClipboardSyncClient:
    def __init__(self, base_url='http://localhost:8080'):
        self.base_url = base_url
        self.token = None
    
    def register(self, username, email, password):
        response = requests.post(f'{self.base_url}/api/v1/auth/register', 
                               json={'username': username, 'email': email, 'password': password})
        data = response.json()
        if response.status_code == 201:
            self.token = data['token']
        return data
    
    def login(self, username, password):
        response = requests.post(f'{self.base_url}/api/v1/auth/login',
                               json={'username': username, 'password': password})
        data = response.json() 
        if response.status_code == 200:
            self.token = data['token']
        return data
    
    def create_clipboard_item(self, item_type, content, device_id):
        headers = {'Authorization': f'Bearer {self.token}'}
        data = {'type': item_type, 'content': content, 'device_id': device_id}
        response = requests.post(f'{self.base_url}/api/v1/clipboard/items',
                               json=data, headers=headers)
        return response.json()
    
    def get_clipboard_items(self, page=1, page_size=20):
        headers = {'Authorization': f'Bearer {self.token}'}
        params = {'page': page, 'page_size': page_size}
        response = requests.get(f'{self.base_url}/api/v1/clipboard/items',
                              params=params, headers=headers)
        return response.json()
```

---

## 版本历史

### v1.0.0 (2025-09-17)
- 初始版本发布
- 用户认证系统
- 剪贴板CRUD操作
- 批量同步功能
- 统计信息接口
- JWT认证机制
- 限流和安全防护