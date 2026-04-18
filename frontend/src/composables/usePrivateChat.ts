import { computed, onUnmounted, ref, watch } from 'vue'
import { API_BASE_URL, apiURL } from '@/api/api'

type HistoryMessage = {
  id: string
  sender_id: string
  receiver_id: string
  content: string
  created_at: string
}

type PrivateMessageEvent = {
  id: string
  sender_id: string
  receiver_id: string
  content: string
  created_at: string
}

type HistoryResponse = {
  items: HistoryMessage[]
  next_before?: string
}

type IncomingEnvelope = {
  type?: string
  payload?: unknown
}

type ChatMessage = {
  id: string
  senderID: string
  receiverID: string
  content: string
  createdAt: string
}

const messageLimit = 30
const maxClientMessageBytes = 65536

const buildWebSocketURL = () => {
  const url = new URL(API_BASE_URL)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.pathname = '/ws'
  url.search = ''
  url.hash = ''
  return url.toString()
}

const normalizeMessage = (value: HistoryMessage | PrivateMessageEvent): ChatMessage => ({
  id: value.id,
  senderID: value.sender_id,
  receiverID: value.receiver_id,
  content: value.content,
  createdAt: value.created_at,
})

export const usePrivateChat = (currentUserID: () => string, otherUserID: () => string) => {
  const isLoadingInitial = ref(true)
  const isLoadingOlder = ref(false)
  const isConnected = ref(false)
  const connectionError = ref('')
  const historyError = ref('')
  const sendError = ref('')
  const messages = ref<ChatMessage[]>([])
  const nextBefore = ref('')
  const hasMore = ref(true)

  const peerIsTyping = ref(false)

  let socket: WebSocket | null = null
  let reconnectTimer: number | null = null
  let manuallyClosed = false
  let peerTypingTimer: ReturnType<typeof setTimeout> | null = null
  const pendingOptimisticIDs = ref<string[]>([])

  const isReadyToSend = computed(() => {
    return isConnected.value && currentUserID() !== '' && otherUserID() !== ''
  })

  const privateHistoryURL = (before?: string) => {
    const params = new URLSearchParams()
    params.set('limit', String(messageLimit))
    if (before) params.set('before', before)
    const query = params.toString()
    return apiURL(`/api/chats/private/${encodeURIComponent(otherUserID())}/messages${query ? `?${query}` : ''}`)
  }

  const clearReconnect = () => {
    if (reconnectTimer !== null) {
      window.clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
  }

  const scheduleReconnect = () => {
    if (manuallyClosed || reconnectTimer !== null) return
    reconnectTimer = window.setTimeout(() => {
      reconnectTimer = null
      connect()
    }, 1500)
  }

  const isConversationMessage = (message: PrivateMessageEvent) => {
    const me = currentUserID()
    const other = otherUserID()
    if (!me || !other) return false
    const firstDirection = message.sender_id === me && message.receiver_id === other
    const secondDirection = message.sender_id === other && message.receiver_id === me
    return firstDirection || secondDirection
  }

  const mergeInitialHistory = (items: HistoryMessage[]) => {
    messages.value = items.map(normalizeMessage)
  }

  const prependOlderHistory = (items: HistoryMessage[]) => {
    const knownIDs = new Set(messages.value.map((msg) => msg.id))
    const older = items.map(normalizeMessage).filter((msg) => !knownIDs.has(msg.id))
    messages.value = [...older, ...messages.value]
  }

  const reconcileOptimisticMessage = (payload: PrivateMessageEvent) => {
    if (payload.sender_id !== currentUserID()) return false

    const optimisticID = pendingOptimisticIDs.value[0]
    if (!optimisticID) return false

    const nextMessage = normalizeMessage(payload)
    const matchIndex = messages.value.findIndex((message) => message.id === optimisticID)
    if (matchIndex === -1) return false

    messages.value = messages.value.map((message, index) =>
      index === matchIndex ? nextMessage : message,
    )
    pendingOptimisticIDs.value.shift()
    return true
  }

  const loadInitialHistory = async () => {
    if (!currentUserID() || !otherUserID()) return

    isLoadingInitial.value = true
    historyError.value = ''
    try {
      const response = await fetch(privateHistoryURL(), {
        method: 'GET',
        credentials: 'include',
      })
      if (!response.ok) {
        const payload = (await response.json().catch(() => null)) as { error?: string } | null
        historyError.value = payload?.error || 'Could not load private chat history.'
        messages.value = []
        hasMore.value = false
        nextBefore.value = ''
        return
      }

      const payload = (await response.json()) as HistoryResponse
      mergeInitialHistory(payload.items || [])
      nextBefore.value = payload.next_before || ''
      hasMore.value = Boolean(payload.next_before)
    } catch {
      historyError.value = 'Network error while loading private chat history.'
      messages.value = []
      hasMore.value = false
      nextBefore.value = ''
    } finally {
      isLoadingInitial.value = false
    }
  }

  const loadOlderHistory = async () => {
    if (!hasMore.value || !nextBefore.value || isLoadingOlder.value || !otherUserID()) return false

    isLoadingOlder.value = true
    historyError.value = ''
    try {
      const response = await fetch(privateHistoryURL(nextBefore.value), {
        method: 'GET',
        credentials: 'include',
      })
      if (!response.ok) {
        const payload = (await response.json().catch(() => null)) as { error?: string } | null
        historyError.value = payload?.error || 'Could not load older private chat history.'
        return false
      }

      const payload = (await response.json()) as HistoryResponse
      const previousCount = messages.value.length
      prependOlderHistory(payload.items || [])
      nextBefore.value = payload.next_before || ''
      hasMore.value = Boolean(payload.next_before)
      return messages.value.length > previousCount
    } catch {
      historyError.value = 'Network error while loading older private chat history.'
      return false
    } finally {
      isLoadingOlder.value = false
    }
  }

  const closeSocket = () => {
    clearReconnect()
    if (socket) {
      socket.close()
      socket = null
    }
    isConnected.value = false
  }

  const connect = () => {
    if (!currentUserID() || !otherUserID()) return
    if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) return

    manuallyClosed = false
    socket = new WebSocket(buildWebSocketURL())

    socket.onopen = () => {
      isConnected.value = true
      connectionError.value = ''
    }

    socket.onclose = () => {
      isConnected.value = false
      socket = null
      if (!manuallyClosed) {
        connectionError.value = 'Realtime connection lost. Reconnecting...'
      }
      scheduleReconnect()
    }

    socket.onerror = () => {
      socket?.close()
    }

    socket.onmessage = (event) => {
      let envelope: IncomingEnvelope
      try {
        envelope = JSON.parse(event.data) as IncomingEnvelope
      } catch {
        return
      }

      if (envelope.type === 'ack') {
        return
      }

      if (envelope.type === 'error') {
        const payload = envelope.payload as { message?: string } | undefined
        sendError.value = payload?.message || 'Could not send message.'
        return
      }

      if (envelope.type === 'typing_event') {
        const p = envelope.payload as { sender_id: string; context_type: string; context_id: string; is_typing: boolean } | undefined
        if (p?.context_type === 'private' && p.sender_id === otherUserID()) {
          peerIsTyping.value = p.is_typing
          if (peerTypingTimer !== null) clearTimeout(peerTypingTimer)
          if (p.is_typing) {
            // Auto-clear after 4s in case the stop event is lost
            peerTypingTimer = setTimeout(() => { peerIsTyping.value = false }, 4000)
          }
        }
        return
      }

      if (envelope.type !== 'private_message') return
      const payload = envelope.payload as PrivateMessageEvent | undefined
      if (!payload || !payload.id || !isConversationMessage(payload)) return

      if (messages.value.some((message) => message.id === payload.id)) return
      if (reconcileOptimisticMessage(payload)) return
      messages.value = [...messages.value, normalizeMessage(payload)]
    }
  }

  const sendMessage = (rawContent: string) => {
    sendError.value = ''
    const content = rawContent.trim()
    if (!content) return
    if (new TextEncoder().encode(content).length > maxClientMessageBytes) {
      sendError.value = 'Message is too long.'
      return
    }
    if (!socket || socket.readyState !== WebSocket.OPEN || !isReadyToSend.value) {
      sendError.value = 'Realtime connection is not ready yet.'
      return
    }

    const optimisticID = `temp-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
    const now = new Date().toISOString()
    messages.value = [
      ...messages.value,
      {
        id: optimisticID,
        senderID: currentUserID(),
        receiverID: otherUserID(),
        content,
        createdAt: now,
      },
    ]
    pendingOptimisticIDs.value.push(optimisticID)

    socket.send(
      JSON.stringify({
        type: 'private_message_send',
        payload: {
          to_user_id: otherUserID(),
          content,
        },
      }),
    )
  }

  const sendTyping = (isTyping: boolean) => {
    if (!socket || socket.readyState !== WebSocket.OPEN || !otherUserID()) return
    socket.send(JSON.stringify({
      type: 'private_typing',
      payload: { to_user_id: otherUserID(), is_typing: isTyping },
    }))
  }

  const disconnect = () => {
    manuallyClosed = true
    closeSocket()
  }

  const reloadHistory = async () => {
    await loadInitialHistory()
  }

  watch(
    () => `${currentUserID()}::${otherUserID()}`,
    async () => {
      peerIsTyping.value = false
      disconnect()
      messages.value = []
      hasMore.value = true
      nextBefore.value = ''
      pendingOptimisticIDs.value = []
      historyError.value = ''
      await loadInitialHistory()
      connect()
    },
    { immediate: true },
  )

  onUnmounted(() => {
    disconnect()
  })

  return {
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
    reconnect: connect,
  }
}
