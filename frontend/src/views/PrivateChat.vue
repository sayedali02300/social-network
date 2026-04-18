<template>
  <main class="private-chat-page">
    <section class="private-chat-card">
      <header class="chat-header">
        <RouterLink class="back-link" to="/chats">Back</RouterLink>
        <div class="chat-peer">
          <span class="chat-kicker">Private chat</span>
          <h1>{{ peerDisplayName }}</h1>
          <p :class="isConnected ? 'status-online' : 'status-offline'">
            {{ isConnected ? 'Connected' : 'Disconnected' }}
          </p>
        </div>
        <button
          type="button"
          class="mute-btn"
          :title="isMuted ? 'Unmute notifications' : 'Mute notifications'"
          :aria-pressed="isMuted"
          @click="toggleMute"
        >
          {{ isMuted ? 'Muted' : 'Alerts' }}
        </button>
      </header>

      <p v-if="connectionError" class="connection-banner" role="alert">
        {{ connectionError }}
        <button type="button" @click="reconnect">Retry now</button>
      </p>
      <p v-if="historyError" class="history-banner" role="alert">
        {{ historyError }}
        <button type="button" @click="retryHistory">Retry history</button>
      </p>

      <section ref="messagesPanelRef" class="messages-panel" aria-label="Private chat messages panel" @scroll="handlePanelScroll">
        <div v-if="isLoadingInitial" class="messages-status" aria-live="polite">Loading messages...</div>
        <div v-else-if="messages.length === 0" class="messages-status">No messages yet. Say hello.</div>

        <button v-if="!isLoadingInitial && hasMore" class="load-older-btn" type="button" :disabled="isLoadingOlder" @click="loadOlder">
          {{ isLoadingOlder ? 'Loading older messages...' : 'Load older messages' }}
        </button>

        <ul v-if="messages.length > 0" class="message-list">
          <li
            v-for="message in messages"
            :key="message.id"
            class="message-item"
            :class="{ mine: message.senderID === currentUserID }"
          >
            <p class="message-content">{{ message.content }}</p>
            <p class="message-meta">
              <span>{{ formatTime(message.createdAt) }}</span>
            </p>
          </li>
        </ul>
      </section>

      <div class="typing-indicator" aria-live="polite">
        <template v-if="peerIsTyping">
          <span class="typing-dots"><span /><span /><span /></span>
          <span class="typing-label">{{ peerDisplayName }} is typing</span>
        </template>
      </div>

      <form class="composer" @submit.prevent="submitMessage">
        <label class="sr-only" for="private-chat-composer">Message input</label>
        <textarea
          id="private-chat-composer"
          ref="textareaRef"
          v-model="draft"
          rows="2"
          maxlength="4096"
          placeholder="Write a message..."
          @keydown.enter.exact.prevent="submitMessage"
          @keydown.enter.shift.stop
          @input="onDraftInput"
          @blur="stopTyping"
        />
        <div class="composer-footer">
          <EmojiPicker @pick="insertEmoji" />
          <p v-if="sendError" class="send-error" role="alert">{{ sendError }}</p>
          <button type="submit" :disabled="!canSend" aria-label="Send message">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" width="20" height="20" aria-hidden="true">
              <path d="M3.478 2.405a.75.75 0 00-.926.94l2.432 7.905H13.5a.75.75 0 010 1.5H4.984l-2.432 7.905a.75.75 0 00.926.94 60.519 60.519 0 0018.445-8.986.75.75 0 000-1.218A60.517 60.517 0 003.478 2.405z" />
            </svg>
          </button>
        </div>
      </form>
    </section>
  </main>
</template>

<script setup lang="ts">
import { API_ROUTES, apiURL } from '@/api/api'
import EmojiPicker from '@/components/EmojiPicker.vue'
import { usePrivateChat } from '@/composables/usePrivateChat'
import { computed, nextTick, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

type UserSummary = {
  id: string
  firstName: string
  lastName: string
  nickname: string
}

const route = useRoute()

const currentUserID = ref('')
const peer = ref<UserSummary | null>(null)
const draft = ref('')
const isMuted = ref(false)
const messagesPanelRef = ref<HTMLElement | null>(null)
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const stickToBottom = ref(true)

const peerID = computed(() => {
  const raw = route.params.userId
  return typeof raw === 'string' ? raw.trim() : ''
})

const peerDisplayName = computed(() => {
  if (!peer.value) return 'Private chat'
  if (peer.value.nickname) return `@${peer.value.nickname}`
  return `${peer.value.firstName} ${peer.value.lastName}`.trim() || peer.value.id
})

const {
  isLoadingInitial,
  isLoadingOlder,
  isConnected,
  isReadyToSend,
  connectionError,
  historyError,
  sendError,
  messages,
  hasMore,
  peerIsTyping,
  loadOlderHistory,
  reloadHistory,
  sendMessage,
  sendTyping,
  reconnect,
} = usePrivateChat(
  () => currentUserID.value,
  () => peerID.value,
)

const canSend = computed(() => {
  return isReadyToSend.value && draft.value.trim().length > 0
})

const formatTime = (value: string) => {
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  return parsed.toLocaleString([], { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

const scrollToBottom = async () => {
  await nextTick()
  const panel = messagesPanelRef.value
  if (!panel) return
  panel.scrollTop = panel.scrollHeight
}

const muteURL = () => apiURL(`/api/chats/private/${encodeURIComponent(peerID.value)}/mute`)

const loadMutePreference = async () => {
  if (!peerID.value) return
  try {
    const res = await fetch(muteURL(), { credentials: 'include' })
    if (res.ok) {
      const data = (await res.json()) as { muted: boolean }
      isMuted.value = data.muted
    }
  } catch {
    // ignore
  }
}

const toggleMute = async () => {
  const next = !isMuted.value
  try {
    const res = await fetch(muteURL(), {
      method: 'PUT',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ muted: next }),
    })
    if (res.ok) isMuted.value = next
  } catch {
    // ignore
  }
}

const loadViewerContext = async () => {
  const [meResponse, peerResponse] = await Promise.all([
    fetch(apiURL(API_ROUTES.USERS_ME), { credentials: 'include' }),
    fetch(apiURL(`/api/users/${encodeURIComponent(peerID.value)}`), { credentials: 'include' }),
  ])

  if (!meResponse.ok || !peerResponse.ok) {
    return
  }

  const me = (await meResponse.json()) as { id: string }
  const peerPayload = (await peerResponse.json()) as UserSummary
  currentUserID.value = me.id
  peer.value = peerPayload
}

const insertEmoji = (emoji: string) => {
  const el = textareaRef.value
  if (!el) {
    draft.value += emoji
    return
  }
  const start = el.selectionStart ?? draft.value.length
  const end = el.selectionEnd ?? draft.value.length
  draft.value = draft.value.slice(0, start) + emoji + draft.value.slice(end)
  void nextTick(() => {
    el.focus()
    el.setSelectionRange(start + emoji.length, start + emoji.length)
  })
}

let typingStopTimer: ReturnType<typeof setTimeout> | null = null
let isCurrentlyTyping = false

const stopTyping = () => {
  if (typingStopTimer !== null) { clearTimeout(typingStopTimer); typingStopTimer = null }
  if (isCurrentlyTyping) { isCurrentlyTyping = false; sendTyping(false) }
}

const onDraftInput = () => {
  if (!isCurrentlyTyping) { isCurrentlyTyping = true; sendTyping(true) }
  if (typingStopTimer !== null) clearTimeout(typingStopTimer)
  typingStopTimer = setTimeout(stopTyping, 2000)
}

const submitMessage = () => {
  const text = draft.value.trim()
  if (!text) return
  stopTyping()
  sendMessage(text)
  draft.value = ''
  void scrollToBottom()
}

const loadOlder = async () => {
  const panel = messagesPanelRef.value
  const previousHeight = panel?.scrollHeight ?? 0
  const loaded = await loadOlderHistory()
  if (!loaded) return
  await nextTick()
  if (!panel) return
  panel.scrollTop = panel.scrollHeight - previousHeight + panel.scrollTop
}

const retryHistory = async () => {
  await reloadHistory()
}

let scrollThrottleTimer: ReturnType<typeof setTimeout> | null = null
const handlePanelScroll = () => {
  if (scrollThrottleTimer !== null) return
  scrollThrottleTimer = setTimeout(() => {
    scrollThrottleTimer = null
    const panel = messagesPanelRef.value
    if (!panel) return
    const distanceToBottom = panel.scrollHeight - panel.scrollTop - panel.clientHeight
    stickToBottom.value = distanceToBottom < 60
    if (panel.scrollTop <= 120 && hasMore.value && !isLoadingOlder.value) {
      void loadOlder()
    }
  }, 120)
}

watch(
  () => peerID.value,
  async () => {
    currentUserID.value = ''
    peer.value = null
    isMuted.value = false
    await loadViewerContext()
    await loadMutePreference()
  },
  { immediate: true },
)

watch(
  () => messages.value.length,
  async () => {
    if (!stickToBottom.value) return
    await scrollToBottom()
  },
)
</script>

<style>
.private-chat-page {
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  padding: 0.25rem 0 1.75rem;
}

.private-chat-card {
  width: min(100%, 1320px);
  flex: 1;
  min-height: 0;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-radius: 32px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(248, 250, 252, 0.86));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.08);
}

.chat-header {
  padding: 1.2rem 1.3rem 1rem;
  border-bottom: 1px solid rgba(148, 163, 184, 0.16);
  display: flex;
  gap: 1rem;
  align-items: center;
  background:
    radial-gradient(circle at top right, rgba(59, 130, 246, 0.12), transparent 32%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.9), rgba(248, 250, 252, 0.8));
}

.back-link {
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 44px;
  padding: 0.72rem 0.95rem;
  border-radius: 18px;
  font-weight: 800;
  color: var(--brand-700);
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.12), rgba(14, 165, 233, 0.08));
}

.mute-btn {
  margin-left: auto;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.9);
  padding: 0.72rem 0.9rem;
  font-size: 0.88rem;
  font-weight: 800;
  cursor: pointer;
  line-height: 1;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
}

.chat-peer {
  min-width: 0;
}

.chat-kicker {
  display: inline-flex;
  align-items: center;
  padding: 0.34rem 0.72rem;
  border-radius: 999px;
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.14), rgba(14, 165, 233, 0.08));
  color: var(--brand-700);
  font-size: 0.74rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.chat-peer h1 {
  margin-top: 0.7rem;
  color: var(--gray-900);
  font-size: clamp(1.35rem, 2.8vw, 2rem);
  line-height: 1.2;
  font-weight: 900;
}

.chat-peer p {
  margin-top: 0.35rem;
  font-size: 0.84rem;
  font-weight: 800;
}

.status-online {
  color: #15803d;
}

.status-offline {
  color: #b91c1c;
}

.connection-banner,
.history-banner {
  margin: 0.9rem 1rem 0;
  border-radius: 22px;
  padding: 0.75rem 0.9rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.7rem;
  font-weight: 700;
  font-size: 0.86rem;
}

.connection-banner {
  background: linear-gradient(180deg, rgba(255, 247, 237, 0.98), rgba(255, 237, 213, 0.9));
  color: #b45309;
  border: 1px solid rgba(251, 146, 60, 0.28);
}

.history-banner {
  margin-top: 0.65rem;
  background: linear-gradient(180deg, rgba(254, 242, 242, 0.98), rgba(254, 226, 226, 0.92));
  color: #991b1b;
  border: 1px solid rgba(239, 68, 68, 0.28);
}

.connection-banner button,
.history-banner button {
  border: none;
  border-radius: 16px;
  color: #fff;
  padding: 0.62rem 0.8rem;
  font-weight: 800;
}

.connection-banner button {
  background: #9a3412;
}

.history-banner button {
  background: #991b1b;
}

.connection-banner button:focus-visible,
.history-banner button:focus-visible {
  outline: 2px solid var(--brand-500);
  outline-offset: 2px;
}

.messages-panel {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
  padding: 1rem;
  display: grid;
  align-content: start;
  gap: 0.85rem;
  background:
    linear-gradient(180deg, rgba(248, 250, 252, 0.28), rgba(255, 255, 255, 0.18));
}

.messages-status {
  min-height: 180px;
  display: grid;
  place-items: center;
  text-align: center;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 28px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.92), rgba(248, 250, 252, 0.84));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.06);
  color: var(--gray-600);
  font-weight: 700;
  padding: 1.5rem;
}

.load-older-btn {
  justify-self: center;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 999px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  color: var(--gray-700);
  font-size: 0.82rem;
  font-weight: 800;
  padding: 0.52rem 0.9rem;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
}

.load-older-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.message-list {
  display: grid;
  gap: 0.7rem;
  margin: 0;
  padding: 0;
}

.message-item {
  list-style: none;
  justify-self: start;
  max-width: 70%;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(241, 245, 249, 0.92));
  color: var(--gray-900);
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 22px 22px 22px 6px;
  padding: 0.72rem 0.92rem;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 10px 24px rgba(15, 23, 42, 0.05);
  animation: msgEnter var(--dur-base) var(--ease-enter);
}

@keyframes msgEnter {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

.message-item.mine {
  justify-self: end;
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  color: #fff;
  border: none;
  border-radius: 22px 22px 6px 22px;
  box-shadow: 0 16px 30px rgba(37, 99, 235, 0.18);
}

.message-content {
  margin: 0;
  font-size: var(--text-base);
  line-height: 1.55;
  white-space: pre-wrap;
  word-break: break-word;
}

.message-meta {
  margin-top: 0.25rem;
  font-size: var(--text-xs);
  display: flex;
  justify-content: flex-end;
  gap: 0.4rem;
  color: rgba(255, 255, 255, 0.7);
}

.message-item:not(.mine) .message-meta {
  color: var(--gray-400);
}

.typing-indicator {
  flex-shrink: 0;
  min-height: 1.5rem;
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.2rem 1rem 0;
}

.typing-label {
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--gray-500);
}

.typing-dots {
  display: inline-flex;
  align-items: center;
  gap: 3px;
}

.typing-dots span {
  display: block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--brand-400, #60a5fa);
  animation: typingBounce 1.2s ease-in-out infinite;
}

.typing-dots span:nth-child(2) { animation-delay: 0.2s; }
.typing-dots span:nth-child(3) { animation-delay: 0.4s; }

@keyframes typingBounce {
  0%, 60%, 100% { transform: translateY(0); opacity: 0.5; }
  30% { transform: translateY(-5px); opacity: 1; }
}

.composer {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.9));
  border-top: 1px solid rgba(148, 163, 184, 0.16);
  padding: 1rem;
}

.composer textarea {
  width: 100%;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 24px;
  padding: 0.85rem 1rem;
  resize: none;
  font: inherit;
  font-size: var(--text-base);
  line-height: 1.4;
  min-height: 44px;
  max-height: 120px;
  transition: background var(--dur-fast), border-color var(--dur-fast), box-shadow var(--dur-fast);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 8px 22px rgba(15, 23, 42, 0.04);
}

.composer textarea:focus {
  outline: none;
  border-color: rgba(37, 99, 235, 0.55);
  box-shadow:
    0 0 0 5px rgba(37, 99, 235, 0.1),
    0 16px 34px rgba(37, 99, 235, 0.12);
}

.composer-footer {
  display: flex;
  justify-content: space-between;
  gap: 0.8rem;
  align-items: center;
}

.send-error {
  color: #b91c1c;
  font-size: var(--text-sm);
  font-weight: 700;
}

.composer-footer button[type="submit"] {
  border: none;
  border-radius: 50%;
  width: 44px;
  height: 44px;
  min-width: 44px;
  min-height: 44px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  color: #fff;
  font-weight: 800;
  font-size: 0.9rem;
  transition: background var(--dur-base) var(--ease-standard),
              box-shadow var(--dur-base) var(--ease-standard);
  box-shadow: 0 16px 30px rgba(37, 99, 235, 0.24);
}

.composer-footer button[type="submit"]:not(:disabled):hover {
  background: linear-gradient(135deg, var(--brand-600), #1d4ed8);
  box-shadow: 0 4px 14px rgba(37, 99, 235, 0.35);
}

.composer-footer button[type="submit"]:disabled {
  background: var(--gray-300);
  cursor: not-allowed;
  box-shadow: none;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@media (max-width: 760px) {
  .message-item {
    max-width: 85%;
  }

  .chat-header {
    align-items: flex-start;
    flex-direction: column;
  }

  .private-chat-card {
    border-radius: 24px;
  }
}
</style>
