<template>
  <main class="chat-list-page">
    <section class="chat-list-shell">
      <header class="chat-list-header">
        <div class="chat-hero-copy">
          <span class="chat-kicker">Your inbox</span>
          <h1>Keep every conversation close.</h1>
          <p>Open private and group conversations from one inbox.</p>
          <div class="header-actions">
            <RouterLink class="quick-link" to="/following">New private</RouterLink>
            <RouterLink class="quick-link alt" to="/groups">New group</RouterLink>
          </div>
        </div>
        <div class="chat-hero-meta">
          <article>
            <strong>{{ allConversations.length }}</strong>
            <span>Total conversations</span>
          </article>
          <article>
            <strong>{{ Object.values(unreadCounts).reduce((sum, count) => sum + count, 0) }}</strong>
            <span>Unread updates</span>
          </article>
        </div>
      </header>

      <section class="chat-controls">
        <div class="tabs" role="tablist" aria-label="Conversation type tabs">
          <button type="button" role="tab" :aria-selected="activeTab === 'all'" :class="{ active: activeTab === 'all' }" @click="activeTab = 'all'">All</button>
          <button type="button" role="tab" :aria-selected="activeTab === 'private'" :class="{ active: activeTab === 'private' }" @click="activeTab = 'private'">Private</button>
          <button type="button" role="tab" :aria-selected="activeTab === 'groups'" :class="{ active: activeTab === 'groups' }" @click="activeTab = 'groups'">Groups</button>
        </div>
        <div class="chat-search-bar">
          <label class="chat-search-wrap" for="chat-search">
            <svg class="chat-search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9" aria-hidden="true">
              <circle cx="11" cy="11" r="7"></circle>
              <path d="m20 20-3.5-3.5"></path>
            </svg>
            <input id="chat-search" v-model.trim="searchText" aria-label="Search conversations" placeholder="Search users or group names" maxlength="120" />
          </label>
        </div>
      </section>

      <div v-if="isLoading" class="skeleton-list" aria-hidden="true">
        <div v-for="n in 8" :key="n" class="skeleton-row"></div>
      </div>
      <div v-else-if="errorMessage" class="status error with-action" role="alert">
        <span>{{ errorMessage }}</span>
        <button type="button" class="retry-btn" @click="loadConversations">Retry</button>
      </div>
      <div v-else-if="filteredConversations.length === 0" class="status empty-state">
        <img src="@/assets/empty-states/chats-empty.svg" alt="" class="empty-state-img" />
        <p>No conversations yet.</p>
      </div>

      <ul v-else class="conversation-list">
        <li
          v-for="conversation in filteredConversations"
          :key="conversation.id"
          class="conversation-item"
          :class="{ active: isActiveRoute(conversation) }"
        >
          <RouterLink class="conversation-link" :to="conversation.route">
            <div class="avatar">
              <img v-if="conversation.avatar" :src="conversation.avatar" alt="Conversation avatar" />
              <span v-else>{{ conversation.initials }}</span>
            </div>
            <div class="conversation-copy">
              <p class="title">
                <span class="title-text">{{ conversation.name }}</span>
                <small>{{ conversation.kind === 'group' ? 'Group' : 'Private' }}</small>
              </p>
              <p class="preview">{{ conversation.preview }}</p>
            </div>
            <span v-if="conversation.unread > 0" class="unread-badge">
              {{ conversation.unread > 99 ? '99+' : conversation.unread }}
            </span>
          </RouterLink>
        </li>
      </ul>
    </section>
  </main>
</template>

<script setup lang="ts">
import { API_ROUTES, apiURL } from '@/api/api'
import { listGroups } from '@/api/groups'
import { buildWebSocketURL } from '@/api/websocket'
import type { GroupSummary } from '@/types/groups'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'

type UserItem = {
  id: string
  firstName: string
  lastName: string
  nickname: string
  avatar: string
}

type ConversationItem = {
  id: string
  kind: 'private' | 'group'
  name: string
  avatar: string
  initials: string
  preview: string
  timeLabel: string
  route: string
  unread: number
}

const route = useRoute()
const isLoading = ref(true)
const errorMessage = ref('')
const searchText = ref('')
const activeTab = ref<'all' | 'private' | 'groups'>('all')
const privateConversations = ref<ConversationItem[]>([])
const groupConversations = ref<ConversationItem[]>([])
const unreadCounts = ref<Record<string, number>>({})

const toInitials = (firstName: string, lastName: string) => {
  const value = `${firstName?.[0] || ''}${lastName?.[0] || ''}`.toUpperCase().trim()
  return value || '?'
}

const toAvatarSrc = (value: string) => {
  if (!value) return ''
  if (value.startsWith('http://') || value.startsWith('https://')) return value
  return apiURL(value.startsWith('/') ? value : `/${value}`)
}

const privateName = (user: UserItem) => {
  if (user.nickname) return `@${user.nickname}`
  return `${user.firstName} ${user.lastName}`.trim() || user.id
}

const toGroupConversation = (group: Pick<GroupSummary, 'id' | 'title' | 'membersCount'>): ConversationItem => {
  return {
    id: `group-${group.id}`,
    kind: 'group',
    name: group.title || `Group ${group.id}`,
    avatar: '',
    initials: '#',
    preview: `${group.membersCount} members`,
    timeLabel: 'Group',
    route: `/chats/groups/${encodeURIComponent(group.id)}?name=${encodeURIComponent(group.title || `Group ${group.id}`)}`,
    unread: 0,
  }
}

const loadConversations = async () => {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const meResponse = await fetch(apiURL(API_ROUTES.USERS_ME), {
      method: 'GET',
      credentials: 'include',
    })
    if (!meResponse.ok) {
      errorMessage.value = 'Could not load account.'
      return
    }
    const me = (await meResponse.json()) as { id: string }
    const userID = me.id

    const [followersResponse, followingResponse, groups] = await Promise.all([
      fetch(apiURL(`/api/users/${encodeURIComponent(userID)}/followers`), { credentials: 'include' }),
      fetch(apiURL(`/api/users/${encodeURIComponent(userID)}/following`), { credentials: 'include' }),
      listGroups(),
    ])
    if (!followersResponse.ok || !followingResponse.ok) {
      errorMessage.value = 'Could not load conversations.'
      return
    }

    const followers = (await followersResponse.json()) as UserItem[]
    const following = (await followingResponse.json()) as UserItem[]
    const mergedMap = new Map<string, UserItem>()
    for (const user of [...followers, ...following]) {
      if (!user?.id) continue
      mergedMap.set(user.id, user)
    }

    const unreadResponse = await fetch(apiURL('/api/chats/unread-counts'), { credentials: 'include' })
    if (unreadResponse.ok) {
      unreadCounts.value = (await unreadResponse.json()) as Record<string, number>
    }

    privateConversations.value = [...mergedMap.values()].map((user) => ({
      id: `private-${user.id}`,
      kind: 'private',
      name: privateName(user),
      avatar: toAvatarSrc(user.avatar),
      initials: toInitials(user.firstName, user.lastName),
      preview: 'Open private conversation',
      timeLabel: '',
      route: `/chats/private/${encodeURIComponent(user.id)}`,
      unread: unreadCounts.value[`private:${user.id}`] ?? 0,
    }))

    groupConversations.value = groups
      .filter((group) => group.isMember)
      .map((group) => ({ ...toGroupConversation(group), unread: unreadCounts.value[`group:${group.id}`] ?? 0 }))
  } catch {
    errorMessage.value = 'Network error while loading conversations.'
  } finally {
    isLoading.value = false
  }
}

const allConversations = computed(() => {
  return [...privateConversations.value, ...groupConversations.value]
})

const filteredConversations = computed(() => {
  const source =
    activeTab.value === 'private'
      ? privateConversations.value
      : activeTab.value === 'groups'
        ? groupConversations.value
        : allConversations.value

  if (!searchText.value) return source
  const q = searchText.value.toLowerCase()
  return source.filter((item) => `${item.name} ${item.preview}`.toLowerCase().includes(q))
})

const isActiveRoute = (conversation: ConversationItem) => {
  if (conversation.kind === 'private') {
    return route.path === conversation.route
  }
  const base = conversation.route.split('?')[0]
  return route.path === base
}

let socket: WebSocket | null = null
let reconnectTimer: number | null = null
let closedManually = false

function handleRealtimeMessage(raw: { type?: string; payload?: unknown }) {
  if (raw.type === 'private_message') {
    const p = raw.payload as { sender_id?: string }
    if (!p?.sender_id) return
    const conv = privateConversations.value.find((c) => c.id === `private-${p.sender_id}`)
    if (conv) conv.unread += 1
  } else if (raw.type === 'group_message') {
    const p = raw.payload as { group_id?: string }
    if (!p?.group_id) return
    const conv = groupConversations.value.find((c) => c.id === `group-${p.group_id}`)
    if (conv) conv.unread += 1
  }
}

function connectSocket() {
  if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) return
  closedManually = false
  socket = new WebSocket(buildWebSocketURL('/ws'))
  socket.onmessage = (event) => {
    try {
      const raw = JSON.parse(event.data) as { type?: string; payload?: unknown }
      handleRealtimeMessage(raw)
    } catch {
      // ignore malformed frames
    }
  }
  socket.onclose = () => {
    socket = null
    if (!closedManually) {
      reconnectTimer = window.setTimeout(() => {
        reconnectTimer = null
        connectSocket()
      }, 1500)
    }
  }
  socket.onerror = () => socket?.close()
}

function disconnectSocket() {
  closedManually = true
  if (reconnectTimer !== null) {
    window.clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
  if (socket) {
    socket.close()
    socket = null
  }
}

onMounted(async () => {
  await loadConversations()
  connectSocket()
})

onBeforeUnmount(() => {
  disconnectSocket()
})
</script>

<style>
.chat-list-page {
  min-height: calc(100dvh - var(--navbar-height, 60px));
  padding: 0.25rem 0 1.75rem;
}

.chat-list-shell {
  width: min(100%, 1320px);
  margin: 0 auto;
  display: grid;
  gap: 1rem;
}

.chat-list-header {
  display: grid;
  grid-template-columns: minmax(0, 1.15fr) minmax(240px, 0.85fr) auto;
  gap: 1rem;
  padding: 1.5rem;
  border-radius: 32px;
  background:
    radial-gradient(circle at top left, rgba(59, 130, 246, 0.2), transparent 28%),
    linear-gradient(135deg, rgba(8, 18, 37, 0.96), rgba(15, 70, 180, 0.96));
  box-shadow: 0 28px 70px rgba(15, 23, 42, 0.18);
  color: var(--white);
  overflow: hidden;
  position: relative;
}

.chat-list-header::before,
.chat-list-header::after {
  content: '';
  position: absolute;
  border-radius: 999px;
  pointer-events: none;
}

.chat-list-header::before {
  width: 220px;
  height: 220px;
  right: -40px;
  top: -60px;
  background: rgba(255, 255, 255, 0.08);
}

.chat-list-header::after {
  width: 160px;
  height: 160px;
  left: -40px;
  bottom: -60px;
  background: rgba(125, 211, 252, 0.12);
}
.title-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}
.chat-hero-copy,
.chat-hero-meta,
.header-actions {
  position: relative;
  z-index: 1;
}

.chat-hero-copy {
  display: grid;
  align-content: start;
  gap: 0.9rem;
}

.chat-kicker {
  display: inline-flex;
  align-items: center;
  justify-self: start;
  width: fit-content;
  padding: 0.38rem 0.78rem;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.12);
  color: rgba(255, 255, 255, 0.84);
  font-size: 0.76rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.chat-list-header h1 {
  margin-top: 1rem;
  max-width: 10ch;
  color: var(--white);
  font-size: clamp(2.2rem, 5vw, 3.5rem);
  line-height: 0.96;
  font-weight: 900;
}

.chat-list-header p {
  margin-top: 0.9rem;
  max-width: 42ch;
  color: rgba(255, 255, 255, 0.82);
  font-size: 1rem;
}

.chat-hero-meta {
  display: grid;
  gap: 0.85rem;
  align-content: end;
}

.chat-hero-meta article {
  padding: 1rem 1.05rem;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.12);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.chat-hero-meta strong {
  display: block;
  color: var(--white);
  font-size: 1.55rem;
  font-weight: 900;
}

.chat-hero-meta span {
  color: rgba(255, 255, 255, 0.76);
  font-size: 0.86rem;
  font-weight: 700;
}

.header-actions {
  display: flex;
  gap: 0.6rem;
  align-items: flex-start;
  justify-content: flex-start;
  flex-wrap: wrap;
}

.quick-link {
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 52px;
  border-radius: 20px;
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.95), #1e40af);
  color: #fff;
  padding: 0.8rem 1rem;
  font-weight: 800;
  font-size: 0.9rem;
  box-shadow: 0 16px 30px rgba(37, 99, 235, 0.24);
}


.quick-link.alt {
  background: linear-gradient(135deg, rgba(14, 165, 233, 0.92), #2563eb);
}

.chat-controls {
  display: grid;
  gap: 0.8rem;
}

.tabs {
  display: flex;
  gap: 0.55rem;
  flex-wrap: wrap;
  padding: 0.45rem;
  border-radius: 24px;
  justify-self: start;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.9), rgba(248, 250, 252, 0.82));
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.96);
  width: fit-content;
}

.tabs button {
  border: 1px solid transparent;
  background: transparent;
  border-radius: 999px;
  padding: 0.82rem 1rem;
  color: var(--gray-500);
  font-size: 0.9rem;
  font-weight: 800;
  cursor: pointer;
}

.tabs button.active {
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.14), rgba(14, 165, 233, 0.08));
  border-color: rgba(37, 99, 235, 0.18);
  color: var(--brand-700);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.9),
    0 14px 28px rgba(37, 99, 235, 0.12);
}

.tabs button:focus-visible {
  outline: 2px solid var(--brand-500);
  outline-offset: 2px;
}

.chat-search-bar {
  width: 100%;
  padding: 0.75rem;
  border-radius: 24px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(239, 246, 255, 0.76));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 12px 28px rgba(15, 23, 42, 0.05);
}

.chat-search-wrap {
  display: flex;
  align-items: center;
  gap: 0.72rem;
  width: 100%;
  min-width: 0;
  padding: 1rem;
  border-radius: 18px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 8px 22px rgba(15, 23, 42, 0.04);
}

.chat-search-icon {
  width: 1.1rem;
  height: 1.1rem;
  flex-shrink: 0;
  color: var(--gray-500);
}

.chat-controls input {
  width: 100%;
  min-width: 0;
  box-sizing: border-box;
  border: none;
  border-radius: 0;
  padding: 0;
  font: inherit;
  background: transparent;
  color: var(--gray-800);
  font-size: 0.95rem;
  font-weight: 600;
  cursor: text;
  box-shadow: none;
}

.chat-controls input:focus {
  outline: none;
}

.chat-search-wrap:focus-within {
  border-color: rgba(37, 99, 235, 0.55);
  box-shadow:
    0 0 0 5px rgba(37, 99, 235, 0.1),
    0 16px 34px rgba(37, 99, 235, 0.12);
}

.chat-controls input:focus {
  outline: none;
}

.status {
  color: var(--gray-600);
  font-weight: 700;
}

.status.with-action {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.7rem;
  min-height: 140px;
  border-radius: 28px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.92), rgba(248, 250, 252, 0.84));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.06);
  padding: 1.25rem;
}

.retry-btn {
  border: none;
  border-radius: 18px;
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  color: #fff;
  padding: 0.78rem 0.95rem;
  font-size: 0.84rem;
  font-weight: 800;
  box-shadow: 0 14px 28px rgba(37, 99, 235, 0.18);
}


.skeleton-list {
  display: grid;
  gap: 1rem;
}

.skeleton-row {
  height: 88px;
  border-radius: 24px;
  background: linear-gradient(90deg, #e2e8f0 25%, #f1f5f9 37%, #e2e8f0 63%);
  background-size: 400% 100%;
  animation: shimmer 1.2s infinite;
}

.status.error {
  color: var(--status-error);
}

.empty-state {
  min-height: 260px;
  border-radius: 32px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.9), rgba(248, 250, 252, 0.82));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.06);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  text-align: center;
  padding: 2rem 1rem;
}

.empty-state-img {
  width: clamp(84px, 10vw, 120px);
  height: auto;
  opacity: 0.92;
}

.conversation-list {
  display: grid;
  gap: 1rem;
  margin: 0;
  padding: 0;
}

.conversation-item {
  list-style: none;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 28px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.88));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.08);
  transition:
    transform var(--dur-base) var(--ease-standard),
    border-color var(--dur-base) var(--ease-standard),
    box-shadow var(--dur-base) var(--ease-standard);
}

.conversation-item:hover {
  transform: translateY(-2px);
  border-color: rgba(96, 165, 250, 0.22);
  box-shadow: 0 26px 58px rgba(15, 23, 42, 0.12);
}

.conversation-item.active {
  border-color: rgba(37, 99, 235, 0.24);
  background:
    linear-gradient(180deg, rgba(239, 246, 255, 0.96), rgba(255, 255, 255, 0.9));
}

.conversation-link {
  text-decoration: none;
  padding: 1rem;
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 0.9rem;
  align-items: center;
}

.conversation-link:focus-visible {
  outline: 2px solid var(--brand-500);
  outline-offset: -2px;
  border-radius: 28px;
}

.avatar {
  width: 56px;
  height: 56px;
  border-radius: 999px;
  overflow: hidden;
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.16), rgba(14, 165, 233, 0.08));
  color: var(--brand-700);
  display: grid;
  place-items: center;
  font-weight: 800;
}

.avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.conversation-copy {
  min-width: 0;
}

.title {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  color: var(--gray-900);
  font-weight: 800;
  margin: 0;
}

.title small {
  color: var(--gray-500);
  font-size: 0.72rem;
}

.preview {
  margin-top: 0.2rem;
  color: var(--gray-500);
  font-size: 0.84rem;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.time {
  color: var(--text-secondary);
  font-size: 0.74rem;
  font-weight: 700;
}

.unread-badge {
  min-width: 24px;
  height: 24px;
  border-radius: 999px;
  background: var(--brand-500, #2563eb);
  color: #fff;
  font-size: 0.76rem;
  font-weight: 800;
  display: grid;
  place-items: center;
  padding: 0 0.45rem;
  animation: pulse-ring 1.8s ease-out 1;
  box-shadow: 0 12px 24px rgba(37, 99, 235, 0.2);
}

@keyframes pulse-ring {
  0%   { box-shadow: 0 0 0 0 rgba(37, 99, 235, 0.45); }
  70%  { box-shadow: 0 0 0 7px rgba(37, 99, 235, 0); }
  100% { box-shadow: 0 0 0 0 rgba(37, 99, 235, 0); }
}

@keyframes shimmer {
  0% {
    background-position: 100% 50%;
  }
  100% {
    background-position: 0 50%;
  }
}

@media (max-width: 700px) {
  .chat-list-header {
    grid-template-columns: 1fr;
  }

  .chat-controls,
  .empty-state,
  .conversation-item {
    border-radius: 24px;
  }
}
</style>
