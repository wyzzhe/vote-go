# 实时投票系统

一个基于Go+Vue3+TypeScript的实时投票系统，支持WebSocket实时通信。

## 技术栈

### 后端
- Go 1.19+
- Gin Web框架
- GORM (ORM)
- MySQL数据库
- WebSocket实时通信
- CORS跨域支持

### 前端
- Vue 3
- TypeScript
- Vite
- Chart.js图表
- 响应式设计

## 功能特性

- ✅ 实时投票：用户可以选择选项并立即提交投票
- ✅ 实时更新：投票结果通过WebSocket实时推送给所有用户
- ✅ 数据可视化：使用Chart.js展示投票结果统计图表
- ✅ IP限制：每个IP地址只能投票一次
- ✅ 响应式设计：支持移动端和桌面端
- ✅ 连接状态：显示WebSocket连接状态指示器

## 项目结构

```
vote-go/
├── backend/                 # Go后端
│   ├── config/             # 配置管理
│   ├── database/           # 数据库连接和初始化
│   ├── handlers/           # HTTP处理器
│   ├── models/             # 数据模型
│   ├── websocket/          # WebSocket处理
│   ├── main.go             # 程序入口
│   └── go.mod              # Go模块文件
├── frontend/               # Vue3前端
│   ├── src/
│   │   ├── components/     # Vue组件
│   │   ├── App.vue         # 主应用组件
│   │   ├── main.ts         # 应用入口
│   │   └── style.css       # 全局样式
│   ├── index.html          # HTML模板
│   ├── package.json        # 前端依赖
│   └── vite.config.ts      # Vite配置
└── README.md               # 项目文档
```

## 环境要求

- Go 1.19 或更高版本
- Node.js 16+ 
- MySQL 5.7+ 或 8.0+

## 安装运行

### 1. 克隆项目

```bash
git clone <repository-url>
cd vote-go
```

### 2. 数据库设置

创建MySQL数据库：

```sql
CREATE DATABASE vote_system CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 3. 后端设置

```bash
cd backend

# 安装依赖
go mod tidy

# 设置环境变量（可选）
export DATABASE_URL="root:password@tcp(localhost:3306)/vote_system?charset=utf8mb4&parseTime=True&loc=Local"
export PORT="8080"

# 运行后端服务
go run main.go
```

### 4. 前端设置

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

### 5. 访问应用

- 前端地址: http://localhost:3000
- 后端API: http://localhost:8080/api

## API接口

### 获取投票信息
```
GET /api/poll
```

### 提交投票
```
POST /api/poll/vote
Content-Type: application/json

{
  "option_id": 1
}
```

### WebSocket连接
```
ws://localhost:8080/ws/poll
```

## 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| PORT | 8080 | 后端服务端口 |
| DATABASE_URL | root:password@tcp(localhost:3306)/vote_system?charset=utf8mb4&parseTime=True&loc=Local | MySQL连接字符串 |

## 开发模式

### 后端热重载
```bash
# 安装air (如果还没有安装)
go install github.com/air-verse/air@latest

# 运行热重载
air
```

### 前端开发
```bash
npm run dev
```

## 构建部署

### 前端构建
```bash
cd frontend
npm run build
```

### 后端构建
```bash
cd backend
go build -o vote-system main.go
```

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！ 