<template>
  <div class="app-shell" :class="{ 'with-navbar': showNavbar }">
    <Navbar v-if="showNavbar"></Navbar>

    <main class="app-content">
      <RouterView v-slot="{ Component }">
        <Transition name="page">
          <component :is="Component" :key="route.fullPath" />
        </Transition>
      </RouterView>
    </main>
  </div>

  <div v-if="liveToasts.length > 0" class="notification-toasts">
    <TransitionGroup name="toast" tag="div" class="toasts-inner">
      <div v-for="toast in liveToasts" :key="toast.id" class="toast-item">
        <p>{{ toast.message }}</p>
        <span>{{ toast.timeLabel }}</span>
      </div>
    </TransitionGroup>
  </div>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import Navbar from './views/AppNavbar.vue'
import { API_BASE_URL } from '@/api/api'
import { type AppNotification, useNotifications } from '@/composables/useNotifications'

type NotificationToast = {
  id: string
  message: string
  timeLabel: string
}

type NotificationPayload = AppNotification

type NotificationUnreadCountPayload = {
  unread: number
}

const route = useRoute()
const showNavbar = computed(() => route.meta.hideNavbar !== true)
const liveToasts = ref<NotificationToast[]>([])
const { applyRealtimeNotification, applyRealtimeUnreadCount, ensureNotificationStateLoaded } =
  useNotifications()

let socket: WebSocket | null = null
let reconnectTimerID: number | null = null
let closedManually = false

const buildWebSocketURL = () => {
  const url = new URL(API_BASE_URL)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.pathname = '/ws'
  url.search = ''
  url.hash = ''
  return url.toString()
}

const formatNotificationTime = (value: string) => {
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  return parsed.toLocaleString([], {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const isActiveChat = (payload: NotificationPayload): boolean => {
  if (payload.type === 'new_private_message') {
    return route.name === 'private-chat' && route.params.userId === payload.actor?.id
  }
  if (payload.type === 'new_group_message') {
    return route.name === 'group-chat' && route.params.groupId === payload.target?.groupId
  }
  return false
}

const pushToast = (payload: NotificationPayload) => {
  if (!payload.id || !payload.message) return
  if (isActiveChat(payload)) return

  applyRealtimeNotification(payload as AppNotification)

  liveToasts.value = [
    { id: payload.id, message: payload.message, timeLabel: formatNotificationTime(payload.createdAt) },
    ...liveToasts.value,
  ].slice(0, 5)

  window.setTimeout(() => {
    liveToasts.value = liveToasts.value.filter((item) => item.id !== payload.id)
  }, 7000)
}

const isNotificationPayload = (payload: unknown): payload is NotificationPayload => {
  if (!payload || typeof payload !== 'object') return false
  const item = payload as Partial<NotificationPayload>
  return (
    typeof item.id === 'string' &&
    typeof item.message === 'string' &&
    typeof item.createdAt === 'string'
  )
}

const isUnreadCountPayload = (payload: unknown): payload is NotificationUnreadCountPayload => {
  if (!payload || typeof payload !== 'object') return false
  const item = payload as Partial<NotificationUnreadCountPayload>
  return typeof item.unread === 'number'
}

const scheduleReconnect = () => {
  if (reconnectTimerID !== null || closedManually || !showNavbar.value) return
  reconnectTimerID = window.setTimeout(() => {
    reconnectTimerID = null
    connectWebSocket()
  }, 1500)
}

const connectWebSocket = () => {
  if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
    return
  }

  closedManually = false
  socket = new WebSocket(buildWebSocketURL())

  socket.onmessage = (event) => {
    try {
      const raw = JSON.parse(event.data) as {
        type?: string
        payload?: unknown
      }

      if (raw.type === 'notification' && isNotificationPayload(raw.payload)) {
        pushToast(raw.payload)
        return
      }

      if (raw.type === 'notification_unread_count' && isUnreadCountPayload(raw.payload)) {
        applyRealtimeUnreadCount(raw.payload)
      }
    } catch {
      // Ignore non-JSON websocket events.
    }
  }

  socket.onclose = () => {
    socket = null
    scheduleReconnect()
  }

  socket.onerror = () => {
    socket?.close()
  }
}

const disconnectWebSocket = () => {
  closedManually = true
  if (reconnectTimerID !== null) {
    window.clearTimeout(reconnectTimerID)
    reconnectTimerID = null
  }
  if (socket) {
    socket.close()
    socket = null
  }
}

watch(
  () => showNavbar.value,
  (enabled) => {
    if (enabled) {
      void ensureNotificationStateLoaded()
      connectWebSocket()
    } else {
      disconnectWebSocket()
    }
  },
  { immediate: true },
)

onUnmounted(() => {
  disconnectWebSocket()
})
</script>

<style>
.page-enter-active,
.page-leave-active {
  transition: opacity 180ms ease;
}

/* Pull the leaving page out of flow immediately so it doesn't
   push/shift the entering page or any sticky/fixed elements,
   and so its own transform doesn't create a rogue stacking
   context that repositions position:fixed descendants. */
.page-leave-active {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.page-enter-from,
.page-leave-to {
  opacity: 0;
}
</style>

<style scoped>
.app-shell {
  height: 100%;
  display: flex;
  flex-direction: column;
  background:
    radial-gradient(circle at top left, rgba(59, 130, 246, 0.18), transparent 22%),
    radial-gradient(circle at top right, rgba(14, 165, 233, 0.12), transparent 24%),
    linear-gradient(180deg, #f8fbff 0%, #f2f7ff 24%, #f7f9fc 100%);
}

.app-shell.with-navbar {
  padding: 0.9rem 1rem 1.2rem;
}

.app-content {
  position: relative;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
}

.notification-toasts {
  position: fixed;
  right: 1rem;
  bottom: 1rem;
  z-index: 1000;
  width: min(360px, calc(100vw - 2rem));
}

.toasts-inner {
  display: grid;
  gap: 0.55rem;
}

.toast-enter-active,
.toast-leave-active {
  transition: opacity 200ms ease, transform 200ms ease;
}

.toast-enter-from {
  opacity: 0;
  transform: translateY(-12px);
}

.toast-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

.toast-item {
  background: rgba(15, 23, 42, 0.95);
  color: #fff;
  border-radius: 14px;
  padding: 0.78rem 0.9rem;
  border: 1px solid rgba(148, 163, 184, 0.35);
  box-shadow: 0 18px 34px rgba(15, 23, 42, 0.32);
}

.toast-item p {
  font-weight: 700;
  font-size: 0.9rem;
}

.toast-item span {
  display: inline-block;
  margin-top: 0.2rem;
  font-size: 0.78rem;
  color: rgba(226, 232, 240, 0.9);
}

@media (max-width: 768px) {
  .app-shell.with-navbar {
    padding: 0.7rem 0.75rem 1rem;
  }
}
</style>
