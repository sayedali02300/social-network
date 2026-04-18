import { computed, ref } from 'vue'
import { API_ROUTES, apiURL } from '@/api/api'

export type AppNotification = {
  id: string
  type: string
  message: string
  createdAt: string
  isRead: boolean
  target?: {
    groupId?: string
    groupTitle?: string
    eventId?: string
    eventTitle?: string
    postId?: string
    commentId?: string
  }
  actor?: {
    id: string
    firstName: string
    lastName: string
    nickname: string
    avatar: string
  }
}

type UnreadCountPayload = {
  unread: number
}

const notifications = ref<AppNotification[]>([])
const isLoadingNotifications = ref(false)
const notificationsError = ref('')
const notificationActionError = ref('')
const isMarkingAll = ref(false)
const markingNotificationID = ref('')
const hasLoadedNotifications = ref(false)

const unreadCount = ref(0)
const unreadCountLoaded = ref(false)

const sortNotifications = (items: AppNotification[]) => {
  return [...items].sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
}

const recalculateUnreadFromList = () => {
  unreadCount.value = notifications.value.filter((item) => !item.isRead).length
}

const loadNotifications = async (limit = 25) => {
  isLoadingNotifications.value = true
  notificationsError.value = ''
  try {
    const response = await fetch(apiURL(`${API_ROUTES.NOTIFICATIONS}?limit=${limit}`), {
      method: 'GET',
      credentials: 'include',
    })

    if (!response.ok) {
      let message = 'Could not load notifications.'
      const payload = (await response.json().catch(() => null)) as { error?: string } | null
      if (payload?.error) message = payload.error
      notificationsError.value = message
      return
    }

    const payload = (await response.json()) as AppNotification[]
    notifications.value = sortNotifications(payload)
    hasLoadedNotifications.value = true
    recalculateUnreadFromList()
  } catch {
    notificationsError.value = 'Network error while loading notifications.'
  } finally {
    isLoadingNotifications.value = false
  }
}

const loadUnreadCount = async () => {
  try {
    const response = await fetch(apiURL(API_ROUTES.NOTIFICATIONS_UNREAD_COUNT), {
      method: 'GET',
      credentials: 'include',
    })
    if (!response.ok) return

    const payload = (await response.json()) as UnreadCountPayload
    unreadCount.value = Number.isFinite(payload.unread) ? payload.unread : 0
    unreadCountLoaded.value = true
  } catch {
    // Keep existing unread state if request fails.
  }
}

const markOneAsRead = async (notificationID: string) => {
  if (markingNotificationID.value) return false

  markingNotificationID.value = notificationID
  notificationActionError.value = ''

  const target = notifications.value.find((item) => item.id === notificationID)
  const wasRead = target?.isRead === true
  if (!wasRead) {
    notifications.value = notifications.value.map((item) =>
      item.id === notificationID ? { ...item, isRead: true } : item,
    )
    recalculateUnreadFromList()
  }

  try {
    const response = await fetch(apiURL(`/api/notifications/${notificationID}/read`), {
      method: 'PATCH',
      credentials: 'include',
    })
    if (!response.ok) {
      if (!wasRead) {
        notifications.value = notifications.value.map((item) =>
          item.id === notificationID ? { ...item, isRead: false } : item,
        )
        recalculateUnreadFromList()
      }
      const payload = (await response.json().catch(() => null)) as { error?: string } | null
      notificationActionError.value = payload?.error || 'Could not mark notification as read.'
      return false
    }
    return true
  } catch {
    if (!wasRead) {
      notifications.value = notifications.value.map((item) =>
        item.id === notificationID ? { ...item, isRead: false } : item,
      )
      recalculateUnreadFromList()
    }
    notificationActionError.value = 'Network error while marking notification as read.'
    return false
  } finally {
    markingNotificationID.value = ''
  }
}

const markAllAsRead = async () => {
  if (isMarkingAll.value || unreadCount.value === 0) return false

  isMarkingAll.value = true
  notificationActionError.value = ''
  const previousNotifications = notifications.value.map((item) => ({ ...item }))
  const previousUnreadCount = unreadCount.value

  notifications.value = notifications.value.map((item) => ({ ...item, isRead: true }))
  unreadCount.value = 0

  try {
    const response = await fetch(apiURL('/api/notifications/read-all'), {
      method: 'PATCH',
      credentials: 'include',
    })
    if (!response.ok) {
      notifications.value = previousNotifications
      unreadCount.value = previousUnreadCount
      const payload = (await response.json().catch(() => null)) as { error?: string } | null
      notificationActionError.value = payload?.error || 'Could not mark all notifications as read.'
      return false
    }
    return true
  } catch {
    notifications.value = previousNotifications
    unreadCount.value = previousUnreadCount
    notificationActionError.value = 'Network error while marking all notifications as read.'
    return false
  } finally {
    isMarkingAll.value = false
  }
}

const applyRealtimeNotification = (payload: AppNotification) => {
  if (!payload?.id) return

  notifications.value = sortNotifications([
    payload,
    ...notifications.value.filter((item) => item.id !== payload.id),
  ])

  if (!payload.isRead) {
    unreadCount.value = unreadCount.value + 1
  }
}

const applyRealtimeUnreadCount = (payload: UnreadCountPayload) => {
  if (!payload || !Number.isFinite(payload.unread)) return
  unreadCount.value = payload.unread
  unreadCountLoaded.value = true
}

const ensureNotificationStateLoaded = async () => {
  if (!hasLoadedNotifications.value) {
    await loadNotifications()
  }
  if (!unreadCountLoaded.value) {
    await loadUnreadCount()
  }
}

export const useNotifications = () => {
  return {
    notifications,
    isLoadingNotifications,
    notificationsError,
    notificationActionError,
    isMarkingAll,
    markingNotificationID,
    unreadNotificationsCount: computed(() => unreadCount.value),
    ensureNotificationStateLoaded,
    loadNotifications,
    loadUnreadCount,
    markOneAsRead,
    markAllAsRead,
    applyRealtimeNotification,
    applyRealtimeUnreadCount,
  }
}
