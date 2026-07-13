<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { showToast } from 'vant'
import { mapApi } from '@/api/driver'
import type { AiChatMessage, AiChatType, UserRole } from '@/types'

const props = withDefaults(
  defineProps<{
    role: UserRole
    embedded?: boolean
  }>(),
  { embedded: false },
)

const chatType = ref<AiChatType>(2)
const input = ref('')
const sending = ref(false)
const listRef = ref<HTMLElement | null>(null)
const messages = ref<AiChatMessage[]>([])

const roleLabel = computed(() => (props.role === 'passenger' ? '乘客' : '司机'))

const quickQuestions = computed(() => {
  if (props.role === 'passenger') {
    return chatType.value === 2
      ? ['我的余额是多少', '查看我的订单', '我有哪些优惠券', '从北京西站到天安门多少钱']
      : ['哪里的景点好看', '火锅哪里出名', '推荐一部好看的电影']
  }
  return chatType.value === 2
    ? ['我的余额是多少', '查看我的订单', '附近有什么单可以抢']
    : ['哪里的景点好看', '火锅哪里出名', '周末有什么放松方式']
})

const welcomeText = computed(() => {
  if (chatType.value === 2) {
    return props.role === 'passenger'
      ? '你好，我是出行助手小高 🚗\n可以帮你查余额、订单、优惠券，还能估算车费。'
      : '你好，我是司机助手小高 🛞\n可以帮你查余额、订单和附近可抢订单。'
  }
  return '你好，我是小高 ✨\n随便聊聊，有什么想了解的都可以问我。'
})

function genId() {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
}

async function scrollToBottom() {
  await nextTick()
  const el = listRef.value
  if (el) el.scrollTop = el.scrollHeight
}

function pushWelcome() {
  messages.value = [
    {
      id: genId(),
      role: 'assistant',
      content: welcomeText.value,
    },
  ]
}

async function sendQuestion(question: string) {
  const text = question.trim()
  if (!text || sending.value) return

  messages.value.push({ id: genId(), role: 'user', content: text })
  const pendingId = genId()
  messages.value.push({ id: pendingId, role: 'assistant', content: '', loading: true })
  input.value = ''
  sending.value = true
  await scrollToBottom()

  try {
    const res = await mapApi.chat(text, chatType.value, props.role)
    const answer = res?.answer?.trim() || '暂无回复'
    const idx = messages.value.findIndex((m) => m.id === pendingId)
    if (idx >= 0) {
      messages.value[idx] = { id: pendingId, role: 'assistant', content: answer }
    }
  } catch (e) {
    const msg = e instanceof Error ? e.message : '发送失败'
    const idx = messages.value.findIndex((m) => m.id === pendingId)
    if (idx >= 0) {
      messages.value[idx] = { id: pendingId, role: 'assistant', content: `抱歉，暂时无法回答：${msg}` }
    }
    showToast(msg)
  } finally {
    sending.value = false
    await scrollToBottom()
  }
}

function onSend() {
  void sendQuestion(input.value)
}

watch(chatType, () => {
  pushWelcome()
})

onMounted(() => {
  pushWelcome()
})
</script>

<template>
  <div class="ai-page" :class="{ embedded }">
    <van-nav-bar title="智能助手" />

    <div class="header-card">
      <div class="header-top primary-gradient" :class="{ 'driver-gradient': role === 'driver' }">
        <van-icon name="service-o" size="40" />
        <div class="profile-info">
          <p class="phone">小高在线</p>
          <p class="hint">{{ roleLabel }}端 · 智能助手</p>
        </div>
      </div>

      <div class="mode-section">
        <div class="mode-segment">
          <button
            type="button"
            class="segment-btn"
            :class="{ active: chatType === 1 }"
            @click="chatType = 1"
          >
            闲聊
          </button>
          <button
            type="button"
            class="segment-btn"
            :class="{ active: chatType === 2 }"
            @click="chatType = 2"
          >
            业务助手
          </button>
        </div>
        <p class="mode-hint">
          {{ chatType === 2 ? '可查询余额、订单、优惠券等业务数据' : '轻松聊天，不访问业务数据' }}
        </p>
      </div>
    </div>

    <div ref="listRef" class="message-list">
      <div
        v-for="msg in messages"
        :key="msg.id"
        class="message-row"
        :class="msg.role === 'user' ? 'is-user' : 'is-assistant'"
      >
        <div v-if="msg.role === 'assistant'" class="msg-avatar assistant">
          <van-icon name="service-o" />
        </div>

        <div class="bubble-wrap">
          <div class="bubble" :class="{ loading: msg.loading, error: msg.content.startsWith('抱歉') }">
            <van-loading v-if="msg.loading" size="18" color="#1677ff">思考中</van-loading>
            <p v-else class="bubble-text">{{ msg.content }}</p>
          </div>
        </div>

        <div v-if="msg.role === 'user'" class="msg-avatar user">
          <van-icon name="user-o" />
        </div>
      </div>
    </div>

    <div class="bottom-panel card">
      <div class="quick-bar">
        <p class="quick-label">试试这样问</p>
        <div class="quick-scroll">
          <button
            v-for="q in quickQuestions"
            :key="q"
            type="button"
            class="quick-chip"
            :disabled="sending"
            @click="sendQuestion(q)"
          >
            {{ q }}
          </button>
        </div>
      </div>

      <div class="input-bar">
        <van-field
          v-model="input"
          rows="1"
          autosize
          type="textarea"
          maxlength="500"
          placeholder="输入你的问题…"
          :disabled="sending"
          :border="false"
          @keyup.enter.exact="onSend"
        />
        <van-button
          type="primary"
          round
          size="small"
          :loading="sending"
          :disabled="!input.trim()"
          @click="onSend"
        >
          发送
        </van-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ai-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: var(--app-bg, #eef2f8);
}
.ai-page.embedded {
  height: 100%;
  min-height: 0;
}

.header-card {
  margin: 12px 16px 0;
  padding: 0;
  overflow: hidden;
  flex-shrink: 0;
  background: var(--app-card, #fff);
  border-radius: var(--app-radius-md, 14px);
  border: 1px solid var(--app-border, #e8edf5);
  box-shadow: var(--app-shadow-sm, 0 2px 10px rgba(15, 23, 42, 0.06));
}
.header-top {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  color: #fff;
}
.profile-info .phone {
  font-size: 17px;
  font-weight: 600;
}
.profile-info .hint {
  font-size: 12px;
  opacity: 0.88;
  margin-top: 2px;
}

.mode-section {
  padding: 12px 14px 14px;
}
.mode-segment {
  display: flex;
  gap: 6px;
  padding: 4px;
  background: var(--app-bg, #eef2f8);
  border-radius: 10px;
}
.segment-btn {
  flex: 1;
  border: none;
  background: transparent;
  padding: 9px 0;
  font-size: 14px;
  color: var(--app-text-secondary, #6b7280);
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s, color 0.2s;
}
.segment-btn.active {
  background: #fff;
  color: var(--app-primary, #1677ff);
  font-weight: 600;
  box-shadow: var(--app-shadow-sm, 0 2px 10px rgba(15, 23, 42, 0.06));
}
.mode-hint {
  margin-top: 10px;
  text-align: center;
  font-size: 12px;
  line-height: 1.5;
  color: var(--app-text-muted, #9ca3af);
}

.message-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 12px 16px 8px;
}
.message-row {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  margin-bottom: 12px;
}
.message-row.is-user {
  justify-content: flex-end;
}
.message-row.is-assistant {
  justify-content: flex-start;
}
.msg-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  flex-shrink: 0;
}
.msg-avatar.assistant {
  background: #fff;
  color: #1677ff;
  border: 1px solid var(--app-border, #e8edf5);
}
.msg-avatar.user {
  background: #1677ff;
  color: #fff;
}
.bubble-wrap {
  max-width: calc(100% - 88px);
}
.bubble {
  padding: 10px 14px;
  border-radius: 12px;
  font-size: 14px;
  line-height: 1.6;
  word-break: break-word;
}
.is-user .bubble {
  background: #1677ff;
  color: #fff;
}
.is-assistant .bubble {
  background: #fff;
  color: #262626;
  border: 1px solid var(--app-border, #e8edf5);
  box-shadow: var(--app-shadow-sm, 0 2px 10px rgba(15, 23, 42, 0.06));
}
.bubble.loading {
  min-width: 96px;
  padding: 12px 14px;
}
.bubble.error {
  border-color: #ffccc7;
  background: #fff2f0;
}
.bubble-text {
  white-space: pre-wrap;
}

.bottom-panel {
  flex-shrink: 0;
  margin: 0 16px 12px;
  padding: 12px;
  border-radius: var(--app-radius-md, 14px);
}
.quick-bar {
  margin-bottom: 8px;
}
.quick-label {
  font-size: 12px;
  color: var(--app-text-muted, #9ca3af);
  margin-bottom: 8px;
}
.quick-scroll {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 4px;
  scrollbar-width: none;
}
.quick-scroll::-webkit-scrollbar {
  display: none;
}
.quick-chip {
  flex-shrink: 0;
  border: 1px solid #bfdaff;
  background: var(--app-primary-bg, #e6f4ff);
  color: #1677ff;
  font-size: 12px;
  padding: 6px 12px;
  border-radius: 999px;
  cursor: pointer;
}
.quick-chip:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.input-bar {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  padding-top: 8px;
  border-top: 1px solid var(--app-border, #e8edf5);
}
.input-bar :deep(.van-field) {
  flex: 1;
  background: var(--app-bg, #eef2f8);
  border-radius: 8px;
  padding: 4px 10px;
}
</style>
