<template>
  <main class="group-chat-page">
    <section class="group-chat-card">
      <header class="chat-header">
        <RouterLink class="back-link" to="/chats">Back</RouterLink>
        <div class="chat-peer">
          <h1>{{ groupDisplayName }}</h1>
          <p :class="isConnected ? 'status-online' : 'status-offline'">
            {{ isConnected ? 'Connected' : 'Disconnected' }}
          </p>
        </div>
        <button
          class="members-btn"
          type="button"
          aria-controls="group-members-panel"
          :aria-expanded="isMembersOpen"
          @click="isMembersOpen = !isMembersOpen"
        >
          Members
        </button>
        <button
          type="button"
          class="mute-btn"
          :title="isMuted ? 'Unmute notifications' : 'Mute notifications'"
          :aria-pressed="isMuted"
          @click="toggleMute"
        >
          {{ isMuted ? '🔕' : '🔔' }}
        </button>
      </header>

      <aside v-if="isMembersOpen" id="group-members-panel" class="members-panel">
        <h2>Participants In This Thread</h2>
        <p v-if="participantIDs.length === 0">No participants yet.</p>
        <ul v-else>
          <li v-for="participantID in participantIDs" :key="participantID">
            <strong>{{ senderName(participantID) }}</strong>
            <span>{{ participantID === currentUserID ? '(You)' : '' }}</span>
          </li>
        </ul>
      </aside>

      <p v-if="permissionDenied" class="permission-error" role="alert">You are not a group member.</p>
      <p v-else-if="connectionError" class="connection-banner" role="alert">
        {{ connectionError }}
        <button type="button" @click="reconnect">Retry now</button>
      </p>
      <p v-if="!permissionDenied && historyError" class="history-banner" role="alert">
        {{ historyError }}
        <button type="button" @click="retryHistory">Retry history</button>
      </p>

      <section ref="messagesPanelRef" class="messages-panel" aria-label="Group chat messages panel" @scroll="handlePanelScroll">
        <div v-if="isLoadingInitial" class="messages-status" aria-live="polite">Loading group messages...</div>
        <div v-else-if="!permissionDenied && messages.length === 0" class="messages-status">No group messages yet.</div>

        <button v-if="!isLoadingInitial && hasMore && !permissionDenied" class="load-older-btn" type="button" :disabled="isLoadingOlder" @click="loadOlder">
          {{ isLoadingOlder ? 'Loading older messages...' : 'Load older messages' }}
        </button>

        <ul v-if="messages.length > 0" class="message-list">
          <template v-for="(message, idx) in messages" :key="message.id">
            <li v-if="showDaySeparator(idx)" class="day-separator">
              {{ dayLabel(message.createdAt) }}
            </li>
            <li class="message-item" :class="{ mine: message.senderID === currentUserID }">
              <p v-if="message.senderID !== currentUserID" class="sender-name">
                {{ senderName(message.senderID) }}
              </p>
              <p class="message-content">{{ message.content }}</p>
              <p class="message-meta">
                <span>{{ formatTime(message.createdAt) }}</span>

              </p>
            </li>
          </template>
        </ul>
      </section>

      <div class="typing-indicator" aria-live="polite">
        <template v-if="typingUsers.size > 0">
          <span class="typing-dots"><span /><span /><span /></span>
          <span class="typing-label">{{ typingLabel }}</span>
        </template>
      </div>

      <form class="composer" @submit.prevent="submitMessage">
        <label class="sr-only" for="group-chat-composer">Group message input</label>
        <textarea
          id="group-chat-composer"
          ref="textareaRef"
          v-model="draft"
          rows="2"
          maxlength="4096"
          placeholder="Write to the group..."
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
import { getGroup } from '@/api/groups'
import EmojiPicker from '@/components/EmojiPicker.vue'
import { type GroupChatMessage, useGroupChat } from '@/composables/useGroupChat'
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
const draft = ref('')
const isMembersOpen = ref(false)
const isMuted = ref(false)
const messagesPanelRef = ref<HTMLElement | null>(null)
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const stickToBottom = ref(true)
const senderMap = ref<Record<string, UserSummary>>({})
const loadedGroupName = ref('')

const groupID = computed(() => {
  const raw = route.params.groupId
  return typeof raw === 'string' ? raw.trim() : ''
})

const groupName = computed(() => {
  const queryName = route.query.name
  return typeof queryName === 'string' ? queryName.trim() : ''
})

const groupDisplayName = computed(() => {
  return loadedGroupName.value || groupName.value || `Group ${groupID.value}`
})

const {
  isLoadingInitial,
  isLoadingOlder,
  isConnected,
  isReadyToSend,
  connectionError,
  historyError,
  sendError,
  permissionDenied,
  messages,
  hasMore,
  loadOlderHistory,
  reloadHistory,
  sendMessage,
  sendTyping,
  typingUsers,
  reconnect,
} = useGroupChat(
  () => currentUserID.value,
  () => groupID.value,
)

const canSend = computed(() => {
  return isReadyToSend.value && draft.value.trim().length > 0
})

const participantIDs = computed(() => {
  const ids = new Set<string>()
  for (const message of messages.value) {
    ids.add(message.senderID)
  }
  return [...ids]
})

const senderName = (senderID: string) => {
  const entry = senderMap.value[senderID]
  if (!entry) return senderID
  if (entry.nickname) return `@${entry.nickname}`
  const full = `${entry.firstName} ${entry.lastName}`.trim()
  return full || senderID
}

const typingLabel = computed(() => {
  const ids = [...typingUsers.value]
  if (ids.length === 0) return ''
  if (ids.length === 1) return `${senderName(ids[0] ?? '')} is typing`
  if (ids.length === 2) return `${senderName(ids[0] ?? '')} and ${senderName(ids[1] ?? '')} are typing`
  return 'Several people are typing'
})

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

const formatTime = (value: string) => {
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  return parsed.toLocaleString([], { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

const dayLabel = (value: string) => {
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  return parsed.toLocaleDateString([], {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

const isSameDay = (a: string, b: string) => {
  const da = new Date(a)
  const db = new Date(b)
  if (Number.isNaN(da.getTime()) || Number.isNaN(db.getTime())) return false
  return da.toDateString() === db.toDateString()
}

const showDaySeparator = (index: number) => {
  if (index === 0) return true
  const previous = messages.value[index - 1]
  const current = messages.value[index]
  if (!previous || !current) return true
  return !isSameDay(previous.createdAt, current.createdAt)
}

const muteURL = () => apiURL(`/api/chats/group/${encodeURIComponent(groupID.value)}/mute`)

const loadMutePreference = async () => {
  if (!groupID.value) return
  try {
    const res = await fetch(muteURL(), { credentials: 'include' })
    if (res.ok) {
      const data = (await res.json()) as { muted: boolean }
      isMuted.value = data.muted
    }
  } catch { /* ignore */ }
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
  } catch { /* ignore */ }
}

const loadViewerContext = async () => {
  const meResponse = await fetch(apiURL(API_ROUTES.USERS_ME), { credentials: 'include' })
  if (!meResponse.ok) return
  const me = (await meResponse.json()) as { id: string }
  currentUserID.value = me.id
}

const loadGroupContext = async () => {
  if (!groupID.value) {
    loadedGroupName.value = ''
    return
  }

  try {
    const group = await getGroup(groupID.value)
    loadedGroupName.value = group.title?.trim() || ''
  } catch {
    loadedGroupName.value = ''
  }
}

const loadUnknownSenders = async (inputMessages: GroupChatMessage[]) => {
  const unknownIDs = [...new Set(inputMessages.map((m) => m.senderID))]
    .filter((senderID) => senderID && !senderMap.value[senderID])
  if (unknownIDs.length === 0) return

  const pairs = await Promise.all(
    unknownIDs.map(async (senderID) => {
      try {
        const response = await fetch(apiURL(`/api/users/${encodeURIComponent(senderID)}`), {
          credentials: 'include',
        })
        if (!response.ok) return null
        const payload = (await response.json()) as UserSummary
        return { senderID, payload }
      } catch {
        return null
      }
    }),
  )

  const nextMap = { ...senderMap.value }
  for (const item of pairs) {
    if (!item) continue
    nextMap[item.senderID] = item.payload
  }
  senderMap.value = nextMap
}

const scrollToBottom = async () => {
  await nextTick()
  const panel = messagesPanelRef.value
  if (!panel) return
  panel.scrollTop = panel.scrollHeight
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
    if (panel.scrollTop <= 120 && hasMore.value && !isLoadingOlder.value && !permissionDenied.value) {
      void loadOlder()
    }
  }, 120)
}

watch(
  () => groupID.value,
  async () => {
    senderMap.value = {}
    currentUserID.value = ''
    loadedGroupName.value = ''
    isMuted.value = false
    await loadViewerContext()
    await loadGroupContext()
    await loadMutePreference()
  },
  { immediate: true },
)

watch(
  () => messages.value,
  async (items) => {
    await loadUnknownSenders(items)
    if (!stickToBottom.value) return
    await scrollToBottom()
  },
  { deep: true },
)
</script>

<style scoped>
.group-chat-page {
  width: min(100%, 1320px);
  height: 100%;
  margin: 0 auto;
  padding: 0.25rem 0 1rem;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background:
    radial-gradient(circle at 10% 5%, rgba(37, 99, 235, 0.12), transparent 40%),
    radial-gradient(circle at 90% 95%, rgba(30, 64, 175, 0.1), transparent 45%),
    var(--bg-main);
}

.group-chat-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;            /* prevents flex children from overflowing */
  border-radius: 2rem;
  background: linear-gradient(180deg, rgba(255,255,255,0.985), rgba(241,247,255,0.94));
  border: 1px solid rgba(191, 219, 254, 0.72);
  box-shadow: 0 26px 60px rgba(148, 163, 184, 0.2);
}

.chat-header {
  padding: 0.85rem 1rem;
  border-bottom: 1px solid var(--border-color);
  display: grid;
  grid-template-columns: auto 1fr auto auto;
  gap: 0.8rem;
  align-items: center;
}

.back-link {
  text-decoration: none;
  font-weight: 700;
  color: var(--primary-blue);
}


.chat-peer h1 {
  color: var(--text-primary);
  font-size: 1.05rem;
  line-height: 1.2;
}

.chat-peer p {
  font-size: 0.82rem;
  font-weight: 700;
}

.members-btn {
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 0.42rem 0.72rem;
  background: #fff;
  color: var(--text-primary);
  font-weight: 700;
}


.mute-btn {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: transparent;
  padding: 0.3rem 0.5rem;
  font-size: 1rem;
  cursor: pointer;
  line-height: 1;
}


.members-panel {
  margin: 0.55rem 0.9rem;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  background: #f8fafc;
  padding: 0.7rem;
}

.members-panel h2 {
  font-size: 0.9rem;
  color: var(--text-primary);
  margin-bottom: 0.4rem;
}

.members-panel ul {
  display: grid;
  gap: 0.3rem;
}

.members-panel li {
  list-style: none;
  color: var(--text-secondary);
  font-size: 0.84rem;
  display: flex;
  gap: 0.35rem;
}

.status-online {
  color: var(--success, #16A34A);
}

.status-offline {
  color: var(--danger, #DC2626);
}

.permission-error {
  margin: 0.55rem 0.9rem;
  background: #fee2e2;
  color: #991b1b;
  border: 1px solid #ef4444;
  border-radius: 10px;
  padding: 0.6rem 0.7rem;
  font-weight: 700;
}

.connection-banner {
  margin: 0.55rem 0.9rem;
  background: var(--warning-bg, #FEF3C7);
  color: var(--warning, #D97706);
  border: 1px solid rgba(217, 119, 6, 0.4);
  border-radius: 10px;
  padding: 0.5rem 0.7rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.7rem;
  font-weight: 600;
  font-size: 0.84rem;
}

.connection-banner button {
  border: none;
  border-radius: 8px;
  background: #92400e;
  color: #fff;
  padding: 0.36rem 0.62rem;
  font-weight: 700;
}

.connection-banner button:focus-visible {
  outline: 2px solid var(--brand-500);
  outline-offset: 2px;
}

.history-banner {
  margin: 0.2rem 0.9rem 0.55rem;
  background: #fee2e2;
  color: #991b1b;
  border: 1px solid #ef4444;
  border-radius: 10px;
  padding: 0.45rem 0.7rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.7rem;
  font-weight: 600;
  font-size: 0.82rem;
}

.history-banner button {
  border: none;
  border-radius: 8px;
  background: #991b1b;
  color: #fff;
  padding: 0.34rem 0.58rem;
  font-weight: 700;
}

.history-banner button:focus-visible {
  outline: 2px solid var(--brand-500);
  outline-offset: 2px;
}

.messages-panel {
  flex: 1;
  overflow-y: auto;
  min-height: 0;            /* required for nested flex scroll */
  padding: 1rem;
  display: grid;
  align-content: start;
  gap: 0.7rem;
}

.messages-status {
  color: var(--text-secondary);
  font-weight: 700;
}

.load-older-btn {
  justify-self: center;
  border: 1px solid var(--border-color);
  border-radius: 999px;
  background: #fff;
  color: var(--text-primary);
  font-size: 0.8rem;
  font-weight: 700;
  padding: 0.35rem 0.7rem;
}

.load-older-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}


.message-list {
  display: grid;
  gap: 0.5rem;
}

.day-separator {
  list-style: none;
  justify-self: center;
  font-size: 0.74rem;
  font-weight: 700;
  color: var(--text-secondary);
  background: #eef2ff;
  border: 1px solid #c7d2fe;
  border-radius: 999px;
  padding: 0.2rem 0.55rem;
}

.message-item {
  list-style: none;
  justify-self: start;
  max-width: 70%;
  background: var(--gray-100);
  color: var(--gray-900);
  border-radius: 18px 18px 18px 4px;
  padding: 0.55rem 0.78rem;
  animation: msgEnter var(--dur-base) var(--ease-enter);
}

@keyframes msgEnter {
  from { opacity: 0; transform: translateY(8px); }
  to   { opacity: 1; transform: translateY(0); }
}

.message-item.mine {
  justify-self: end;
  background: var(--brand-500);
  color: #fff;
  border-radius: 18px 18px 4px 18px;
}

.sender-name {
  font-size: var(--text-xs);
  font-weight: 800;
  margin-bottom: 0.2rem;
  color: var(--brand-600);
}

.message-item.mine .sender-name {
  color: rgba(255,255,255,0.8);
}

.message-content {
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
  color: rgba(255,255,255,0.7);
}

.message-item:not(.mine) .message-meta {
  color: var(--gray-400);
}

/* ── Composer ── */
.composer {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  background: var(--white);
  border-top: 1px solid var(--gray-200);
  padding: 0.75rem 1rem;
}

.composer textarea {
  width: 100%;
  background: var(--gray-100);
  border: 1px solid transparent;
  border-radius: 22px;
  padding: 0.6rem 1rem;
  resize: none;
  font: inherit;
  font-size: var(--text-base);
  line-height: 1.4;
  min-height: 44px;
  max-height: 120px;
  transition: background var(--dur-fast), border-color var(--dur-fast), box-shadow var(--dur-fast);
}

.composer textarea:focus {
  outline: none;
  background: var(--white);
  border-color: var(--brand-500);
  box-shadow: 0 0 0 3px rgba(37,99,235,0.12);
}

.composer-footer {
  display: flex;
  justify-content: space-between;
  gap: 0.8rem;
  align-items: center;
}

.send-error {
  color: var(--danger);
  font-size: var(--text-sm);
  font-weight: 600;
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
  background: var(--brand-500);
  color: #fff;
  font-weight: 700;
  font-size: 0.9rem;
  transition: background var(--dur-base) var(--ease-standard),
              box-shadow var(--dur-base) var(--ease-standard);
}

.composer-footer button[type="submit"]:not(:disabled):hover {
  background: var(--brand-600);
  box-shadow: 0 4px 14px rgba(37,99,235,0.35);
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


.chat-header {
  position: relative;
  overflow: hidden;
  padding: 1.05rem 1.15rem;
  border-bottom: 1px solid rgba(191, 219, 254, 0.72);
  background:
    radial-gradient(circle at top right, rgba(96, 165, 250, 0.2), transparent 28%),
    linear-gradient(135deg, #eff6ff 0%, #ffffff 48%, #eaf3ff 100%);
}

.chat-header::after {
  content: "";
  position: absolute;
  right: -2rem;
  top: -2rem;
  width: 10rem;
  height: 10rem;
  border-radius: 999px;
  background: rgba(191, 219, 254, 0.28);
}

.back-link,
.chat-peer,
.members-btn,
.mute-btn {
  position: relative;
  z-index: 1;
}

.back-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 3rem;
  padding: 0.72rem 1rem;
  border-radius: 999px;
  border: 1px solid rgba(191, 219, 254, 0.84);
  background: linear-gradient(180deg, rgba(255,255,255,0.96), rgba(239,246,255,0.94));
  font-weight: 800;
  box-shadow: 0 14px 28px rgba(148, 163, 184, 0.14);
}

.chat-peer h1 {
  font-size: 1.3rem;
  font-weight: 800;
  letter-spacing: -0.03em;
}

.chat-peer p {
  font-size: 0.84rem;
}

.members-btn,
.mute-btn {
  border: 1px solid rgba(191, 219, 254, 0.84);
  border-radius: 999px;
  background: linear-gradient(180deg, rgba(255,255,255,0.96), rgba(239,246,255,0.94));
  font-weight: 800;
  box-shadow: 0 14px 28px rgba(148, 163, 184, 0.14);
}

.members-btn {
  padding: 0.68rem 1rem;
}

.mute-btn {
  padding: 0.62rem 0.82rem;
  font-size: 0.88rem;
  color: #0f172a;
}

.members-panel,
.permission-error,
.connection-banner,
.history-banner {
  margin-left: 1rem;
  margin-right: 1rem;
  border-radius: 1.2rem;
}

.members-panel {
  margin-top: 0.85rem;
  border: 1px solid rgba(191, 219, 254, 0.72);
  background: linear-gradient(180deg, rgba(255,255,255,0.97), rgba(239,246,255,0.92));
  padding: 0.95rem 1rem;
  box-shadow: 0 18px 34px rgba(148, 163, 184, 0.12);
}

.permission-error,
.connection-banner {
  margin-top: 0.85rem;
  padding: 0.8rem 0.9rem;
}

.history-banner {
  margin-top: 0.2rem;
  padding: 0.7rem 0.9rem;
}

.messages-panel {
  margin: 0.85rem 1rem 1rem;
  padding: 1rem;
  border: 1px solid rgba(191, 219, 254, 0.62);
  border-radius: 1.8rem;
  background: linear-gradient(180deg, rgba(255,255,255,0.96), rgba(248,250,252,0.96));
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.88);
}

.messages-status {
  color: #64748b;
  text-align: center;
  padding: 4rem 1rem;
}

.load-older-btn {
  border: 1px solid rgba(191, 219, 254, 0.84);
  background: linear-gradient(180deg, rgba(255,255,255,0.96), rgba(239,246,255,0.94));
  font-weight: 800;
  padding: 0.58rem 0.9rem;
}

.day-separator {
  background: rgba(219, 234, 254, 0.92);
  border: 1px solid rgba(147, 197, 253, 0.92);
}

.message-item {
  background: linear-gradient(180deg, rgba(255,255,255,0.98), rgba(239,246,255,0.92));
  border: 1px solid rgba(191, 219, 254, 0.62);
  border-radius: 22px 22px 22px 8px;
  padding: 0.72rem 0.92rem;
  box-shadow: 0 14px 28px rgba(148, 163, 184, 0.12);
}

.message-item.mine {
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  border: 1px solid transparent;
  border-radius: 22px 22px 8px 22px;
  box-shadow: 0 18px 32px rgba(37, 99, 235, 0.24);
}

.composer {
  margin: 0 1rem 1rem;
  background: linear-gradient(180deg, rgba(255,255,255,0.985), rgba(239,246,255,0.94));
  border: 1px solid rgba(191, 219, 254, 0.7);
  border-radius: 1.6rem;
  padding: 0.95rem 1rem;
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.88);
}

.composer textarea {
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid rgba(203, 213, 225, 0.92);
  padding: 0.8rem 1rem;
}

.composer textarea:focus {
  box-shadow: 0 0 0 4px rgba(191,219,254,0.62);
}

.composer-footer button[type="submit"] {
  border: 1px solid transparent;
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
}

/* ── Typing indicator ── */
.typing-indicator {
  min-height: 1.6rem;
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.15rem 2rem 0;
}

.typing-dots {
  display: inline-flex;
  align-items: center;
  gap: 3px;
}

.typing-dots span {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--brand-500);
  animation: typingBounce 1.2s ease-in-out infinite;
}

.typing-dots span:nth-child(2) { animation-delay: 0.2s; }
.typing-dots span:nth-child(3) { animation-delay: 0.4s; }

@keyframes typingBounce {
  0%, 60%, 100% { transform: translateY(0); opacity: 0.5; }
  30%            { transform: translateY(-5px); opacity: 1; }
}

.typing-label {
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--text-secondary);
}

@media (max-width: 760px) {
  .message-item {
    max-width: 85%;
  }

  .chat-header {
    grid-template-columns: auto 1fr auto auto;
    align-items: center;
  }

  .group-chat-page {
    width: 100%;
    padding: 0 0 0.75rem;
  }

  .group-chat-card {
    border-radius: 1.5rem;
  }
}
</style>
