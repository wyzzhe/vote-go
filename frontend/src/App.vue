<template>
  <div id="app">
    <!-- 连接状态指示器 -->
    <div 
      class="connection-status"
      :class="{ connected: isConnected, disconnected: !isConnected }"
    >
      {{ isConnected ? '已连接' : '未连接' }}
    </div>

    <div class="poll-container">
      <h1>实时投票系统</h1>
      
      <div v-if="loading" class="card">
        <p>加载中...</p>
      </div>

      <div v-else-if="error" class="card">
        <p style="color: red;">{{ error }}</p>
        <button @click="fetchPoll">重新加载</button>
      </div>

      <div v-else-if="poll">
        <!-- 开发模式按钮 -->
        <div class="dev-controls" v-if="isDev">
          <button @click="clearMyVote" class="dev-btn">清除我的投票</button>
          <button @click="resetPoll" class="dev-btn reset-btn">重置投票</button>
        </div>
        
        <!-- 标题和描述 -->
        <div class="card poll-header">
          <h2 class="poll-title">{{ poll.title }}</h2>
          <p class="poll-description">{{ poll.description }}</p>
          <div class="stats">
            <span>总票数: {{ totalVotes }}</span>
            <span v-if="userVoted">您已投票</span>
          </div>
        </div>

        <!-- 投票区域和结果区域横向排列 -->
        <div class="poll-content">
          <!-- 投票问卷 -->
          <div class="card voting-section">
            <!-- 选项列表 -->
            <form @submit.prevent="submitVote" v-if="!userVoted">
              <div 
                v-for="option in poll.options" 
                :key="option.id"
                class="option"
                :class="{ selected: selectedOption === option.id }"
                @click="selectedOption = option.id"
              >
                <input 
                  type="radio" 
                  :id="`option-${option.id}`"
                  :value="option.id"
                  v-model="selectedOption"
                />
                <label :for="`option-${option.id}`">
                  {{ option.text }} ({{ option.vote_count }} 票)
                </label>
              </div>

              <button 
                type="submit" 
                class="submit-btn"
                :disabled="!selectedOption || submitting"
              >
                {{ submitting ? '提交中...' : '提交投票' }}
              </button>
            </form>

            <!-- 已投票用户显示结果 -->
            <div v-else>
              <div 
                v-for="option in poll.options" 
                :key="option.id"
                class="option disabled"
                :class="{ selected: votedOption === option.id }"
              >
                <span>{{ option.text }}: {{ option.vote_count }} 票</span>
                <span v-if="votedOption === option.id"> ✓ 您的选择</span>
              </div>
            </div>
          </div>

          <!-- 投票结果图表 -->
          <div class="card chart-section">
            <h3>投票结果</h3>
            <div class="chart-container">
              <PollChart :poll-data="poll" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import PollChart from './components/PollChart.vue'

interface Option {
  id: number
  text: string
  vote_count: number
  poll_id: number
  created_at: string
  updated_at: string
}

interface Poll {
  id: number
  title: string
  description: string
  is_active: boolean
  options: Option[]
  created_at: string
  updated_at: string
}

interface PollResponse {
  poll: Poll
  total_votes: number
  user_voted: boolean
  voted_option?: number
}

const API_BASE = 'http://localhost:8080/api'
const WS_URL = 'ws://localhost:8080/ws/poll'

const poll = ref<Poll | null>(null)
const totalVotes = ref(0)
const userVoted = ref(false)
const votedOption = ref<number | null>(null)
const selectedOption = ref<number | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)
const submitting = ref(false)
const isConnected = ref(false)
const isDev = import.meta.env.DEV // 仅在开发模式显示

let websocket: WebSocket | null = null

// 生成或获取会话ID
const getSessionId = () => {
  let sessionId = localStorage.getItem('vote-session-id')
  if (!sessionId) {
    sessionId = 'session-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9)
    localStorage.setItem('vote-session-id', sessionId)
  }
  return sessionId
}

const sessionId = getSessionId()

// 获取投票数据
const fetchPoll = async () => {
  try {
    loading.value = true
    error.value = null
    
    const response = await fetch(`${API_BASE}/poll`, {
      headers: {
        'X-Session-ID': sessionId
      }
    })
    if (!response.ok) {
      throw new Error('获取投票数据失败')
    }
    
    const data: PollResponse = await response.json()
    poll.value = data.poll
    totalVotes.value = data.total_votes
    userVoted.value = data.user_voted
    votedOption.value = data.voted_option || null
    
  } catch (err) {
    error.value = err instanceof Error ? err.message : '未知错误'
  } finally {
    loading.value = false
  }
}

// 提交投票
const submitVote = async () => {
  if (!selectedOption.value) return
  
  try {
    submitting.value = true
    
    const response = await fetch(`${API_BASE}/poll/vote`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Session-ID': sessionId
      },
      body: JSON.stringify({
        option_id: selectedOption.value
      })
    })
    
    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.error || '投票失败')
    }
    
    // 投票成功后重新获取数据
    await fetchPoll()
    
  } catch (err) {
    error.value = err instanceof Error ? err.message : '投票失败'
  } finally {
    submitting.value = false
  }
}

// WebSocket连接
const connectWebSocket = () => {
  try {
    websocket = new WebSocket(WS_URL)
    
    websocket.onopen = () => {
      console.log('WebSocket连接已建立')
      isConnected.value = true
    }
    
    websocket.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data)
        if (message.type === 'poll_update' && message.data) {
          // 更新投票数据
          poll.value = message.data
          // 重新计算总票数
          totalVotes.value = message.data.options.reduce(
            (sum: number, option: Option) => sum + option.vote_count, 
            0
          )
        }
      } catch (err) {
        console.error('解析WebSocket消息失败:', err)
      }
    }
    
    websocket.onclose = () => {
      console.log('WebSocket连接已断开')
      isConnected.value = false
      // 尝试重连
      setTimeout(connectWebSocket, 3000)
    }
    
    websocket.onerror = (error) => {
      console.error('WebSocket错误:', error)
      isConnected.value = false
    }
    
  } catch (err) {
    console.error('WebSocket连接失败:', err)
    isConnected.value = false
  }
}

// 清除当前用户投票
const clearMyVote = async () => {
  try {
    const response = await fetch(`${API_BASE}/poll/clear-my-vote`, {
      method: 'DELETE',
      headers: {
        'X-Session-ID': sessionId
      }
    })
    
    if (response.ok) {
      await fetchPoll() // 重新获取数据
    } else {
      const errorData = await response.json()
      error.value = errorData.error || '清除投票失败'
    }
  } catch (err) {
    error.value = '清除投票失败'
  }
}

// 重置投票
const resetPoll = async () => {
  if (!confirm('确定要重置所有投票吗？')) return
  
  try {
    const response = await fetch(`${API_BASE}/poll/reset`, {
      method: 'DELETE'
    })
    
    if (response.ok) {
      await fetchPoll() // 重新获取数据
    } else {
      const errorData = await response.json()
      error.value = errorData.error || '重置投票失败'
    }
  } catch (err) {
    error.value = '重置投票失败'
  }
}

onMounted(() => {
  fetchPoll()
  connectWebSocket()
})

onUnmounted(() => {
  if (websocket) {
    websocket.close()
  }
})
</script>

<style>
/* ... existing styles ... */

.dev-controls {
  background: #fff3cd;
  border: 1px solid #ffeaa7;
  padding: 10px;
  margin: 10px 0;
  border-radius: 6px;
  text-align: center;
}

.dev-btn {
  background: #ffc107;
  color: #212529;
  border: none;
  padding: 8px 16px;
  margin: 0 5px;
  border-radius: 4px;
  cursor: pointer;
}

.dev-btn:hover {
  background: #e0a800;
}

.reset-btn {
  background: #dc3545;
  color: white;
}

.reset-btn:hover {
  background: #c82333;
}
</style> 