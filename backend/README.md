# 投票系统技术文档

## 1. 系统架构图

```
                                系统架构图
    ┌─────────────────────────────────────────────────────────────────┐
    │                                                                 │
    │  ┌─────────────────┐         ┌─────────────────┐                │
    │  │                 │   HTTP  │                 │                │
    │  │   前端应用       │ ◄─────► │   Go后端服务    │                │
    │  │  (React/Vue)    │   API   │   (Gin框架)     │                │
    │  │                 │         │                 │                │
    │  └─────────────────┘         └─────────────────┘                │
    │           │                           │                         │
    │           │                           │                         │
    │           │ WebSocket连接             │ SQL查询                  │
    │           │                           │                         │
    │           ▼                           ▼                         │
    │  ┌─────────────────┐         ┌─────────────────┐                │
    │  │  WebSocket Hub  │         │   MySQL数据库   │                │
    │  │   (实时推送)     │         │                 │                │
    │  │  多客户端管理    │         │  - polls表      │                │
    │  │                 │         │  - options表    │                │
    │  └─────────────────┘         │  - votes表      │                │
    │                              └─────────────────┘                │
    └─────────────────────────────────────────────────────────────────┘

                              数据流向说明
    ┌─────────────────────────────────────────────────────────────────┐
    │                                                                 │
    │  1. 用户访问前端 → HTTP API请求 → Go后端处理                     │
    │  2. 后端查询数据库 → 返回JSON响应 → 前端展示                     │
    │  3. 用户投票 → WebSocket实时推送 → 所有客户端同步更新             │
    │  4. 数据持久化 → MySQL事务保证 → 数据一致性                      │
    │                                                                 │
    └─────────────────────────────────────────────────────────────────┘
```

### 架构说明

- **前端层**: 使用现代前端框架，通过HTTP API和WebSocket与后端通信
- **后端层**: Go语言 + Gin框架，提供RESTful API和WebSocket服务
- **数据层**: MySQL数据库，存储投票问卷、选项和投票记录
- **实时通信**: WebSocket Hub管理多客户端连接，实现实时数据推送

## 2. API 接口说明

### 2.1 获取投票问卷

**接口**: `GET /api/poll`

**描述**: 获取当前活跃的投票问卷及统计信息

**请求参数**: 无

**响应格式**:
```json
{
  "poll": {
    "id": 1,
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T10:00:00Z",
    "title": "您最喜欢的编程语言是什么？",
    "description": "请选择您最喜欢的编程语言",
    "is_active": true,
    "options": [
      {
        "id": 1,
        "created_at": "2024-01-01T10:00:00Z",
        "updated_at": "2024-01-01T10:00:00Z",
        "poll_id": 1,
        "text": "Go",
        "vote_count": 15
      },
      {
        "id": 2,
        "created_at": "2024-01-01T10:00:00Z",
        "updated_at": "2024-01-01T10:00:00Z",
        "poll_id": 1,
        "text": "Python",
        "vote_count": 12
      }
    ]
  },
  "total_votes": 27,
  "user_voted": false,
  "voted_option": null
}
```

**状态码**:
- `200`: 成功获取投票问卷
- `404`: 没有找到活跃的投票问卷

### 2.2 提交投票

**接口**: `POST /api/poll/vote`

**描述**: 用户提交投票

**请求头**:
```
Content-Type: application/json
```

**请求参数**:
```json
{
  "option_id": 1
}
```

**参数说明**:
- `option_id` (uint, 必填): 选择的选项ID

**成功响应**:
```json
{
  "message": "Vote submitted successfully"
}
```

**错误响应**:
```json
{
  "error": "You have already voted"
}
```

**状态码**:
- `200`: 投票成功
- `400`: 请求参数错误或用户已投票
- `404`: 投票问卷不存在
- `500`: 服务器内部错误

### 2.3 清除用户投票

**接口**: `DELETE /api/poll/clear-my-vote`

**描述**: 清除当前用户的投票记录（开发模式功能）

**请求参数**: 无

**成功响应**:
```json
{
  "message": "Vote cleared successfully"
}
```

**状态码**:
- `200`: 清除成功
- `404`: 没有找到投票记录
- `500`: 服务器内部错误

### 2.4 重置投票

**接口**: `DELETE /api/poll/reset`

**描述**: 重置投票问卷，清除所有投票记录

**请求参数**: 无

**成功响应**:
```json
{
  "message": "Poll reset successfully"
}
```

**状态码**:
- `200`: 重置成功
- `404`: 没有找到活跃的投票问卷
- `500`: 服务器内部错误

## 3. 实时推送机制说明

### 3.1 WebSocket连接

**连接地址**: `ws://localhost:8080/ws/poll`

**连接流程**:
1. 客户端发起WebSocket连接请求
2. 服务器验证连接并注册客户端
3. 客户端加入WebSocket Hub管理池
4. 服务器推送实时数据更新

### 3.2 消息格式

**推送消息结构**:
```json
{
  "type": "poll_update",
  "data": {
    "id": 1,
    "title": "您最喜欢的编程语言是什么？",
    "description": "请选择您最喜欢的编程语言",
    "is_active": true,
    "options": [
      {
        "id": 1,
        "poll_id": 1,
        "text": "Go",
        "vote_count": 16
      },
      {
        "id": 2,
        "poll_id": 1,
        "text": "Python",
        "vote_count": 12
      }
    ]
  }
}
```

### 3.3 触发场景

实时推送在以下场景触发：
- 用户提交投票
- 用户清除投票
- 管理员重置投票

### 3.4 连接管理

- **自动重连**: 客户端应实现断线重连机制
- **心跳检测**: 服务器定期检测连接状态
- **优雅断开**: 客户端离开时自动清理连接

## 4. 技术选型说明

### 4.1 后端技术栈

| 技术 | 版本 | 选择理由 |
|------|------|----------|
| **Go** | 1.24+ | 高性能、并发友好、编译型语言，适合高并发场景 |
| **Gin** | v1.10+ | 轻量级Web框架，性能优秀，中间件丰富 |
| **GORM** | v1.30+ | 功能强大的ORM框架，支持自动迁移和关联查询 |
| **Gorilla WebSocket** | v1.5+ | 成熟的WebSocket库，支持并发连接管理 |
| **MySQL** | 8.0+ | 成熟稳定的关系型数据库，支持事务和复杂查询 |

### 4.2 架构设计原则

#### 4.2.1 分层架构
```
┌─────────────────┐
│   Handler层     │  ← HTTP路由和请求处理
├─────────────────┤
│   Service层     │  ← 业务逻辑处理（可扩展）
├─────────────────┤
│   Model层       │  ← 数据模型定义
├─────────────────┤
│   Database层    │  ← 数据库操作
└─────────────────┘
```

#### 4.2.2 并发处理
- **WebSocket Hub**: 使用Go协程管理多客户端连接
- **数据库事务**: 确保投票操作的原子性
- **IP限制**: 基于用户IP防止重复投票

#### 4.2.3 配置管理
- 环境变量配置数据库连接
- 支持开发/生产环境切换
- CORS跨域配置

### 4.3 数据库设计

#### 4.3.1 表结构

**polls表** (投票问卷):
```sql
CREATE TABLE polls (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    INDEX idx_deleted_at (deleted_at)
);
```

**options表** (选项):
```sql
CREATE TABLE options (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    poll_id BIGINT NOT NULL,
    text VARCHAR(255) NOT NULL,
    vote_count INT DEFAULT 0,
    FOREIGN KEY (poll_id) REFERENCES polls(id),
    INDEX idx_deleted_at (deleted_at)
);
```

**votes表** (投票记录):
```sql
CREATE TABLE votes (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    poll_id BIGINT NOT NULL,
    option_id BIGINT NOT NULL,
    user_ip VARCHAR(45),
    FOREIGN KEY (poll_id) REFERENCES polls(id),
    FOREIGN KEY (option_id) REFERENCES options(id),
    INDEX idx_deleted_at (deleted_at),
    UNIQUE KEY unique_user_poll (poll_id, user_ip)
);
```

### 4.4 性能优化

#### 4.4.1 数据库优化
- 使用索引优化查询性能
- 软删除机制保留历史数据
- 事务确保数据一致性

#### 4.4.2 并发优化
- WebSocket连接池管理
- Go协程处理并发请求
- 数据库连接池复用

#### 4.4.3 缓存策略
- 可扩展Redis缓存热点数据
- 内存缓存活跃投票问卷
- 客户端缓存减少请求

### 4.5 安全考虑

- **CORS配置**: 限制跨域访问来源
- **IP限制**: 防止同一IP重复投票
- **输入验证**: 严格验证请求参数
- **SQL注入防护**: 使用ORM参数化查询

### 4.6 部署方案

#### 4.6.1 Docker部署
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

#### 4.6.2 环境变量
```bash
PORT=8080
DATABASE_URL=root:password@tcp(localhost:3306)/vote_system?charset=utf8mb4&parseTime=True&loc=Local
```

## 5. 扩展性考虑

### 5.1 水平扩展
- 支持多实例部署
- 使用Redis共享WebSocket连接状态
- 数据库读写分离

### 5.2 功能扩展
- 多投票问卷支持
- 用户认证系统
- 投票结果统计分析
- 管理后台界面

### 5.3 监控告警
- 应用性能监控
- 数据库性能监控
- WebSocket连接监控
- 错误日志收集 