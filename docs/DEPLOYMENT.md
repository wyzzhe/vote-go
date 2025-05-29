# 投票系统部署指南

## 1. 本地开发环境

### 1.1 环境要求

- **Go**: 1.24+ 
- **MySQL**: 8.0+
- **Git**: 最新版本

### 1.2 安装步骤

#### 步骤1: 克隆项目
```bash
git clone <repository-url>
cd vote-go/backend
```

#### 步骤2: 安装依赖
```bash
go mod download
go mod tidy
```

#### 步骤3: 配置数据库

创建MySQL数据库：
```sql
CREATE DATABASE vote_system CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'vote_user'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON vote_system.* TO 'vote_user'@'localhost';
FLUSH PRIVILEGES;
```

#### 步骤4: 配置环境变量

创建 `.env` 文件：
```bash
# .env
PORT=8080
DATABASE_URL=vote_user:your_password@tcp(localhost:3306)/vote_system?charset=utf8mb4&parseTime=True&loc=Local
```

或者直接设置环境变量：
```bash
export PORT=8080
export DATABASE_URL="vote_user:your_password@tcp(localhost:3306)/vote_system?charset=utf8mb4&parseTime=True&loc=Local"
```

#### 步骤5: 运行应用
```bash
go run main.go
```

应用将在 `http://localhost:8080` 启动。

### 1.3 开发工具

#### 热重载 (推荐使用 Air)
```bash
# 安装Air
go install github.com/cosmtrek/air@latest

# 创建配置文件 .air.toml
cat > .air.toml << EOF
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
EOF

# 启动热重载
air
```

## 2. Docker 部署

### 2.1 单容器部署

#### Dockerfile
```dockerfile
# 多阶段构建
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 运行阶段
FROM alpine:latest

# 安装ca证书
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./main"]
```

#### 构建和运行
```bash
# 构建镜像
docker build -t vote-system-backend .

# 运行容器
docker run -d \
  --name vote-backend \
  -p 8080:8080 \
  -e DATABASE_URL="root:password@tcp(host.docker.internal:3306)/vote_system?charset=utf8mb4&parseTime=True&loc=Local" \
  vote-system-backend
```

### 2.2 Docker Compose 部署

#### docker-compose.yml
```yaml
version: '3.8'

services:
  # MySQL数据库
  mysql:
    image: mysql:8.0
    container_name: vote-mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: vote_system
      MYSQL_USER: vote_user
      MYSQL_PASSWORD: vote_password
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    networks:
      - vote-network

  # Go后端应用
  backend:
    build: .
    container_name: vote-backend
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      PORT: 8080
      DATABASE_URL: "vote_user:vote_password@tcp(mysql:3306)/vote_system?charset=utf8mb4&parseTime=True&loc=Local"
    depends_on:
      - mysql
    networks:
      - vote-network

volumes:
  mysql_data:

networks:
  vote-network:
    driver: bridge
```

#### 启动服务
```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 停止并删除数据
docker-compose down -v
```

### 2.3 数据库初始化脚本

#### init.sql
```sql
-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS vote_system CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE vote_system;

-- 创建用户（如果不存在）
CREATE USER IF NOT EXISTS 'vote_user'@'%' IDENTIFIED BY 'vote_password';
GRANT ALL PRIVILEGES ON vote_system.* TO 'vote_user'@'%';
FLUSH PRIVILEGES;

-- 表结构将由GORM自动创建
```

## 3. 生产环境部署

### 3.1 服务器要求

- **操作系统**: Ubuntu 20.04+ / CentOS 8+ / RHEL 8+
- **内存**: 最少2GB，推荐4GB+
- **CPU**: 最少2核，推荐4核+
- **存储**: 最少20GB，推荐50GB+
- **网络**: 稳定的互联网连接

### 3.2 使用systemd管理服务

#### 步骤1: 编译生产版本
```bash
# 在开发机器上编译
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o vote-system main.go

# 上传到服务器
scp vote-system user@server:/opt/vote-system/
```

#### 步骤2: 创建系统用户
```bash
sudo useradd --system --shell /bin/false vote-system
sudo mkdir -p /opt/vote-system
sudo chown vote-system:vote-system /opt/vote-system
```

#### 步骤3: 创建systemd服务文件
```bash
sudo tee /etc/systemd/system/vote-system.service > /dev/null << EOF
[Unit]
Description=Vote System Backend
After=network.target mysql.service

[Service]
Type=simple
User=vote-system
Group=vote-system
WorkingDirectory=/opt/vote-system
ExecStart=/opt/vote-system/vote-system
Restart=always
RestartSec=5
Environment=PORT=8080
Environment=DATABASE_URL=vote_user:vote_password@tcp(localhost:3306)/vote_system?charset=utf8mb4&parseTime=True&loc=Local

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/vote-system

[Install]
WantedBy=multi-user.target
EOF
```

#### 步骤4: 启动服务
```bash
# 重新加载systemd
sudo systemctl daemon-reload

# 启用服务
sudo systemctl enable vote-system

# 启动服务
sudo systemctl start vote-system

# 查看状态
sudo systemctl status vote-system

# 查看日志
sudo journalctl -u vote-system -f
```

### 3.3 Nginx反向代理

#### 安装Nginx
```bash
sudo apt update
sudo apt install nginx
```

#### 配置Nginx
```bash
sudo tee /etc/nginx/sites-available/vote-system > /dev/null << EOF
server {
    listen 80;
    server_name your-domain.com;

    # 重定向到HTTPS
    return 301 https://\$server_name\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    # SSL证书配置
    ssl_certificate /path/to/your/certificate.crt;
    ssl_certificate_key /path/to/your/private.key;
    
    # SSL安全配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;

    # 安全头
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload";

    # API代理
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
    }

    # WebSocket代理
    location /ws/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_read_timeout 86400;
    }

    # 静态文件（如果有前端）
    location / {
        root /var/www/vote-system;
        try_files \$uri \$uri/ /index.html;
    }
}
EOF

# 启用站点
sudo ln -s /etc/nginx/sites-available/vote-system /etc/nginx/sites-enabled/

# 测试配置
sudo nginx -t

# 重启Nginx
sudo systemctl restart nginx
```

### 3.4 SSL证书配置

#### 使用Let's Encrypt
```bash
# 安装Certbot
sudo apt install certbot python3-certbot-nginx

# 获取证书
sudo certbot --nginx -d your-domain.com

# 自动续期
sudo crontab -e
# 添加以下行
0 12 * * * /usr/bin/certbot renew --quiet
```

### 3.5 防火墙配置

```bash
# 启用UFW
sudo ufw enable

# 允许SSH
sudo ufw allow ssh

# 允许HTTP和HTTPS
sudo ufw allow 80
sudo ufw allow 443

# 允许MySQL（仅本地）
sudo ufw allow from 127.0.0.1 to any port 3306

# 查看状态
sudo ufw status
```

### 3.6 监控和日志

#### 日志轮转
```bash
sudo tee /etc/logrotate.d/vote-system > /dev/null << EOF
/var/log/vote-system/*.log {
    daily
    missingok
    rotate 52
    compress
    delaycompress
    notifempty
    create 644 vote-system vote-system
    postrotate
        systemctl reload vote-system
    endscript
}
EOF
```

#### 系统监控
```bash
# 安装htop
sudo apt install htop

# 监控系统资源
htop

# 监控磁盘使用
df -h

# 监控内存使用
free -h

# 监控网络连接
ss -tulpn
```

## 4. 备份和恢复

### 4.1 数据库备份

#### 自动备份脚本
```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/opt/backups"
DATE=$(date +%Y%m%d_%H%M%S)
DB_NAME="vote_system"
DB_USER="vote_user"
DB_PASS="vote_password"

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份数据库
mysqldump -u$DB_USER -p$DB_PASS $DB_NAME > $BACKUP_DIR/vote_system_$DATE.sql

# 压缩备份文件
gzip $BACKUP_DIR/vote_system_$DATE.sql

# 删除7天前的备份
find $BACKUP_DIR -name "vote_system_*.sql.gz" -mtime +7 -delete

echo "Backup completed: vote_system_$DATE.sql.gz"
```

#### 设置定时备份
```bash
# 添加到crontab
sudo crontab -e

# 每天凌晨2点备份
0 2 * * * /opt/scripts/backup.sh
```

### 4.2 数据恢复

```bash
# 恢复数据库
gunzip -c /opt/backups/vote_system_20240101_020000.sql.gz | mysql -uvote_user -p vote_system
```

## 5. 性能优化

### 5.1 数据库优化

#### MySQL配置优化
```ini
# /etc/mysql/mysql.conf.d/mysqld.cnf

[mysqld]
# 基本设置
max_connections = 200
innodb_buffer_pool_size = 1G
innodb_log_file_size = 256M
innodb_flush_log_at_trx_commit = 2

# 查询缓存
query_cache_type = 1
query_cache_size = 64M

# 慢查询日志
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 2
```

### 5.2 应用优化

#### 连接池配置
```go
// 在database/database.go中添加
func Init(databaseURL string) (*gorm.DB, error) {
    db, err := gorm.Open(mysql.Open(databaseURL), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    // 获取底层sql.DB
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }

    // 设置连接池
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)

    // 其他初始化代码...
    return db, nil
}
```

## 6. 故障排除

### 6.1 常见问题

#### 问题1: 数据库连接失败
```bash
# 检查MySQL服务状态
sudo systemctl status mysql

# 检查端口是否开放
netstat -tulpn | grep 3306

# 测试数据库连接
mysql -uvote_user -p -h localhost vote_system
```

#### 问题2: 应用无法启动
```bash
# 查看应用日志
sudo journalctl -u vote-system -f

# 检查端口占用
sudo netstat -tulpn | grep 8080

# 检查文件权限
ls -la /opt/vote-system/
```

#### 问题3: WebSocket连接失败
```bash
# 检查Nginx配置
sudo nginx -t

# 查看Nginx错误日志
sudo tail -f /var/log/nginx/error.log

# 测试WebSocket连接
wscat -c ws://localhost:8080/ws/poll
```

### 6.2 性能问题诊断

```bash
# 查看系统负载
top
htop

# 查看内存使用
free -h

# 查看磁盘I/O
iotop

# 查看网络连接
ss -tulpn

# 查看MySQL进程
mysqladmin -uvote_user -p processlist
```

这份部署指南涵盖了从本地开发到生产环境的完整部署流程，包括安全配置、监控、备份和故障排除等重要方面。 