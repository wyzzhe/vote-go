# API 测试文档

## 测试环境

- **服务器地址**: `http://localhost:8080`
- **WebSocket地址**: `ws://localhost:8080/ws/poll`

## 1. 获取投票问卷

### cURL 命令
```bash
curl -X GET http://localhost:8080/api/poll \
  -H "Content-Type: application/json"
```

### 预期响应
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
        "vote_count": 0
      },
      {
        "id": 2,
        "created_at": "2024-01-01T10:00:00Z",
        "updated_at": "2024-01-01T10:00:00Z",
        "poll_id": 1,
        "text": "Python",
        "vote_count": 0
      },
      {
        "id": 3,
        "created_at": "2024-01-01T10:00:00Z",
        "updated_at": "2024-01-01T10:00:00Z",
        "poll_id": 1,
        "text": "JavaScript",
        "vote_count": 0
      },
      {
        "id": 4,
        "created_at": "2024-01-01T10:00:00Z",
        "updated_at": "2024-01-01T10:00:00Z",
        "poll_id": 1,
        "text": "Java",
        "vote_count": 0
      },
      {
        "id": 5,
        "created_at": "2024-01-01T10:00:00Z",
        "updated_at": "2024-01-01T10:00:00Z",
        "poll_id": 1,
        "text": "TypeScript",
        "vote_count": 0
      }
    ]
  },
  "total_votes": 0,
  "user_voted": false,
  "voted_option": null
}
```

## 2. 提交投票

### cURL 命令
```bash
curl -X POST http://localhost:8080/api/poll/vote \
  -H "Content-Type: application/json" \
  -d '{"option_id": 1}'
```

### 预期响应
```json
{
  "message": "Vote submitted successfully"
}
```

### 错误情况测试

#### 重复投票
```bash
# 再次执行相同的投票请求
curl -X POST http://localhost:8080/api/poll/vote \
  -H "Content-Type: application/json" \
  -d '{"option_id": 1}'
```

预期响应：
```json
{
  "error": "You have already voted"
}
```

#### 无效选项ID
```bash
curl -X POST http://localhost:8080/api/poll/vote \
  -H "Content-Type: application/json" \
  -d '{"option_id": 999}'
```

预期响应：
```json
{
  "error": "Invalid option"
}
```

#### 无效JSON格式
```bash
curl -X POST http://localhost:8080/api/poll/vote \
  -H "Content-Type: application/json" \
  -d '{"invalid_json"}'
```

## 3. 清除用户投票

### cURL 命令
```bash
curl -X DELETE http://localhost:8080/api/poll/clear-my-vote \
  -H "Content-Type: application/json"
```

### 预期响应
```json
{
  "message": "Vote cleared successfully"
}
```

## 4. 重置投票

### cURL 命令
```bash
curl -X DELETE http://localhost:8080/api/poll/reset \
  -H "Content-Type: application/json"
```

### 预期响应
```json
{
  "message": "Poll reset successfully"
}
```

## 5. WebSocket 测试

### JavaScript 客户端示例

```javascript
// 连接WebSocket
const ws = new WebSocket('ws://localhost:8080/ws/poll');

// 连接成功
ws.onopen = function(event) {
    console.log('WebSocket连接已建立');
};

// 接收消息
ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    console.log('收到消息:', message);
    
    if (message.type === 'poll_update') {
        console.log('投票数据更新:', message.data);
        // 更新前端界面
        updatePollDisplay(message.data);
    }
};

// 连接关闭
ws.onclose = function(event) {
    console.log('WebSocket连接已关闭');
};

// 连接错误
ws.onerror = function(error) {
    console.error('WebSocket错误:', error);
};

function updatePollDisplay(pollData) {
    // 更新投票显示逻辑
    pollData.options.forEach(option => {
        console.log(`${option.text}: ${option.vote_count} 票`);
    });
}
```

### Node.js 测试脚本

```javascript
const WebSocket = require('ws');

const ws = new WebSocket('ws://localhost:8080/ws/poll');

ws.on('open', function open() {
    console.log('WebSocket连接已建立');
});

ws.on('message', function message(data) {
    const message = JSON.parse(data);
    console.log('收到消息:', JSON.stringify(message, null, 2));
});

ws.on('close', function close() {
    console.log('WebSocket连接已关闭');
});

ws.on('error', function error(err) {
    console.error('WebSocket错误:', err);
});
```

## 6. 完整测试流程

### 测试脚本
```bash
#!/bin/bash

echo "=== 投票系统API测试 ==="

echo "1. 获取初始投票状态"
curl -s -X GET http://localhost:8080/api/poll | jq .

echo -e "\n2. 提交投票 (选择Go)"
curl -s -X POST http://localhost:8080/api/poll/vote \
  -H "Content-Type: application/json" \
  -d '{"option_id": 1}' | jq .

echo -e "\n3. 再次获取投票状态 (应该看到投票数增加)"
curl -s -X GET http://localhost:8080/api/poll | jq .

echo -e "\n4. 尝试重复投票 (应该失败)"
curl -s -X POST http://localhost:8080/api/poll/vote \
  -H "Content-Type: application/json" \
  -d '{"option_id": 2}' | jq .

echo -e "\n5. 清除投票"
curl -s -X DELETE http://localhost:8080/api/poll/clear-my-vote | jq .

echo -e "\n6. 确认投票已清除"
curl -s -X GET http://localhost:8080/api/poll | jq .

echo -e "\n7. 重新投票 (选择Python)"
curl -s -X POST http://localhost:8080/api/poll/vote \
  -H "Content-Type: application/json" \
  -d '{"option_id": 2}' | jq .

echo -e "\n8. 最终状态"
curl -s -X GET http://localhost:8080/api/poll | jq .

echo -e "\n=== 测试完成 ==="
```

### 保存为test.sh并运行
```bash
chmod +x test.sh
./test.sh
```

## 7. Postman 测试集合

### 导入以下JSON到Postman

```json
{
  "info": {
    "name": "投票系统API测试",
    "description": "投票系统的完整API测试集合",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "获取投票问卷",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "{{baseUrl}}/api/poll",
          "host": ["{{baseUrl}}"],
          "path": ["api", "poll"]
        }
      }
    },
    {
      "name": "提交投票",
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"option_id\": 1\n}"
        },
        "url": {
          "raw": "{{baseUrl}}/api/poll/vote",
          "host": ["{{baseUrl}}"],
          "path": ["api", "poll", "vote"]
        }
      }
    },
    {
      "name": "清除投票",
      "request": {
        "method": "DELETE",
        "header": [],
        "url": {
          "raw": "{{baseUrl}}/api/poll/clear-my-vote",
          "host": ["{{baseUrl}}"],
          "path": ["api", "poll", "clear-my-vote"]
        }
      }
    },
    {
      "name": "重置投票",
      "request": {
        "method": "DELETE",
        "header": [],
        "url": {
          "raw": "{{baseUrl}}/api/poll/reset",
          "host": ["{{baseUrl}}"],
          "path": ["api", "poll", "reset"]
        }
      }
    }
  ],
  "variable": [
    {
      "key": "baseUrl",
      "value": "http://localhost:8080"
    }
  ]
}
```

## 8. 性能测试

### 使用Apache Bench (ab)

```bash
# 测试获取投票接口的并发性能
ab -n 1000 -c 10 http://localhost:8080/api/poll

# 测试投票接口的并发性能
ab -n 100 -c 5 -p vote_data.json -T application/json http://localhost:8080/api/poll/vote
```

### vote_data.json 文件内容
```json
{"option_id": 1}
```

### 使用wrk进行压力测试

```bash
# 安装wrk (Ubuntu/Debian)
sudo apt-get install wrk

# 测试获取投票接口
wrk -t12 -c400 -d30s http://localhost:8080/api/poll

# 测试投票接口
wrk -t12 -c400 -d30s -s vote.lua http://localhost:8080/api/poll/vote
```

### vote.lua 脚本
```lua
wrk.method = "POST"
wrk.body   = '{"option_id": 1}'
wrk.headers["Content-Type"] = "application/json"
```

## 9. 错误处理测试

### 测试各种错误情况

```bash
# 1. 测试无效的HTTP方法
curl -X PUT http://localhost:8080/api/poll

# 2. 测试不存在的端点
curl -X GET http://localhost:8080/api/nonexistent

# 3. 测试无效的JSON
curl -X POST http://localhost:8080/api/poll/vote \
  -H "Content-Type: application/json" \
  -d '{"invalid": json}'

# 4. 测试缺少必需参数
curl -X POST http://localhost:8080/api/poll/vote \
  -H "Content-Type: application/json" \
  -d '{}'

# 5. 测试超大的option_id
curl -X POST http://localhost:8080/api/poll/vote \
  -H "Content-Type: application/json" \
  -d '{"option_id": 999999999}'
``` 