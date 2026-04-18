<template>
  <main class="notifications-page">
    <header class="notifications-hero">
      <div class="hero-copy">
        <span class="hero-eyebrow">Notification center</span>
        <h1>Stay on top of your updates.</h1>
        <p>Review invites, messages, replies, and activity from one clean inbox without changing how notifications work.</p>
      </div>
      <div class="hero-meta">
        <article class="hero-card">
          <strong>{{ unreadNotificationsCount > 99 ? '99+' : unreadNotificationsCount }}</strong>
          <span>Unread now</span>
        </article>
        <article class="hero-card">
          <strong>{{ notifications.length }}</strong>
          <span>Total notifications</span>
        </article>
      </div>
    </header>

    <section class="notifications-card">
      <header class="notifications-header">
        <h1>Notifications</h1>
        <div class="header-actions">
          <span v-if="unreadNotificationsCount > 0" class="unread-pill">
            {{ unreadNotificationsCount > 99 ? '99+' : unreadNotificationsCount }} unread
          </span>
          <button
            type="button"
            class="mark-all-btn"
            :disabled="isMarkingAll || unreadNotificationsCount === 0"
            @click="markAllAsRead"
          >
            {{ isMarkingAll ? 'Saving...' : 'Mark all read' }}
          </button>
        </div>
      </header>

      <p v-if="isLoadingNotifications" class="status" aria-live="polite">Loading notifications...</p>
      <div v-else-if="notificationsError" class="status error with-action" role="alert">
        <span>{{ notificationsError }}</span>
        <button type="button" class="retry-btn" @click="ensureNotificationStateLoaded">Retry</button>
      </div>
      <template v-else>
        <p v-if="notificationActionError" class="status error" role="alert">{{ notificationActionError }}</p>
        <div v-if="notifications.length === 0" class="status empty-state">
          <img src="@/assets/empty-states/notifications-empty.svg" alt="" class="empty-state-img" />
          <p>You are all caught up.</p>
        </div>

        <ul v-else class="notif-list">
        <li
          v-for="item in notifications"
          :key="item.id"
          class="notif-item"
          :class="{ unread: !item.isRead, clickable: !item.isRead }"
          @click="handleRowClick(item.id, item.isRead)"
          @keydown.enter.prevent="handleRowKeyDown($event, item.id, item.isRead)"
          @keydown.space.prevent="handleRowKeyDown($event, item.id, item.isRead)"
          :tabindex="item.isRead ? -1 : 0"
          :aria-label="item.isRead ? `Read notification: ${item.message}` : `Unread notification: ${item.message}. Activate to mark as read`"
          :aria-disabled="item.isRead"
        >
            <div class="notif-copy">
              <div class="notif-meta">
                <span :class="['type-pill', notificationTone(item.type)]">{{ notificationLabel(item.type) }}</span>
                <span v-if="notificationActor(item)" class="actor-copy">{{ notificationActor(item) }}</span>
              </div>
              <p v-if="notificationTargetSummary(item)" class="target-copy">{{ notificationTargetSummary(item) }}</p>
            <p class="message">{{ item.message }}</p>
            <div class="notif-foot">
              <span class="date">{{ formatNotificationTime(item.createdAt) }}</span>
              <RouterLink
                v-if="notificationTargetLink(item)"
                class="open-link"
                :to="notificationTargetLink(item)!"
                @click.stop
              >
                {{ notificationTargetLinkLabel(item) }}
              </RouterLink>
            </div>
          </div>
          <button
            v-if="!item.isRead"
            type="button"
            class="mark-one-btn"
            :disabled="markingNotificationID === item.id"
            @click.stop="markOneAsRead(item.id)"
          >
            {{ markingNotificationID === item.id ? '...' : 'Mark read' }}
          </button>
        </li>
      </ul>
      </template>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { type AppNotification, useNotifications } from '@/composables/useNotifications'

const {
  notifications,
  isLoadingNotifications,
  notificationsError,
  notificationActionError,
  isMarkingAll,
  markingNotificationID,
  unreadNotificationsCount,
  ensureNotificationStateLoaded,
  markOneAsRead,
  markAllAsRead,
} = useNotifications()

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

const handleRowClick = (notificationID: string, isRead: boolean) => {
  if (isRead) return
  void markOneAsRead(notificationID)
}

const handleRowKeyDown = (_event: KeyboardEvent, notificationID: string, isRead: boolean) => {
  if (isRead) return
  void markOneAsRead(notificationID)
}

const notificationLabel = (type: string) => {
  switch (type) {
    case 'new_private_message':
      return 'New Message'
    case 'new_group_message':
      return 'Group Message'
    case 'group_invitation_accepted':
      return 'Invite Accepted'
    case 'group_invitation_declined':
      return 'Invite Declined'
    case 'group_join_request_accepted':
      return 'Request Accepted'
    case 'group_join_request_declined':
      return 'Request Declined'
    case 'group_invitation_received':
      return 'Invite Received'
    case 'group_join_request_received':
      return 'Join Request'
    case 'group_event_created':
      return 'Event Created'
    case 'group_event_updated':
      return 'Event Updated'
    case 'group_event_deleted':
      return 'Event Cancelled'
    case 'follow_request_received':
      return 'Follow Request'
    case 'follow_request_accepted':
      return 'Request Accepted'
    case 'new_follower':
      return 'New Follower'
    case 'new_comment':
      return 'New Comment'
    case 'new_comment_reply':
      return 'Comment Reply'
    default:
      return 'Notification'
  }
}

const notificationTone = (type: string) => {
  switch (type) {
    case 'group_invitation_accepted':
    case 'group_join_request_accepted':
      return 'event'
    case 'group_invitation_declined':
    case 'group_join_request_declined':
      return 'request'
    case 'group_invitation_received':
      return 'invite'
    case 'group_join_request_received':
      return 'request'
    case 'group_event_created':
    case 'group_event_updated':
      return 'event'
    case 'group_event_deleted':
      return 'request'
    default:
      return 'default'
  }
}

const notificationActor = (item: AppNotification) => {
  const actor = item.actor
  if (!actor) return ''
  if (actor.nickname) return `From @${actor.nickname}`
  const fullName = `${actor.firstName || ''} ${actor.lastName || ''}`.trim()
  return fullName ? `From ${fullName}` : ''
}

const notificationTargetSummary = (item: AppNotification) => {
  const target = item.target
  if (!target) return ''

  if (item.type === 'group_event_created' || item.type === 'group_event_updated') {
    if (target.eventTitle && target.groupTitle) return `Event: ${target.eventTitle} in ${target.groupTitle}`
    if (target.eventTitle) return `Event: ${target.eventTitle}`
  }
  if (item.type === 'group_event_deleted') {
    if (target.eventTitle && target.groupTitle) return `Cancelled: ${target.eventTitle} in ${target.groupTitle}`
    if (target.eventTitle) return `Cancelled: ${target.eventTitle}`
  }

  if (target.groupTitle) return `Group: ${target.groupTitle}`
  return ''
}

const notificationTargetLink = (item: AppNotification) => {
  const target = item.target
  if ((item.type === 'new_comment' || item.type === 'new_comment_reply') && target?.postId) {
    return `/posts/${target.postId}`
  }
  if (item.type === 'new_private_message' && item.actor?.id) {
    return `/chats/private/${item.actor.id}`
  }
  if (item.type === 'new_group_message' && target?.groupId) {
    return `/chats/groups/${target.groupId}`
  }
  if (item.type === 'follow_request_received' && item.actor?.id) {
    return `/users/${item.actor.id}`
  }
  if (!target?.groupId) return ''
  if (item.type === 'group_event_created' || item.type === 'group_event_updated') {
    return `/groups/${target.groupId}/events`
  }
  if (item.type === 'group_event_deleted') {
    return `/groups/${target.groupId}/events`
  }
  if (item.type === 'group_invitation_received') {
    return '/groups'
  }
  if (
    item.type === 'group_invitation_accepted' ||
    item.type === 'group_invitation_declined' ||
    item.type === 'group_join_request_accepted' ||
    item.type === 'group_join_request_declined'
  ) {
    return target?.groupId ? `/groups/${target.groupId}` : '/groups'
  }
  return `/groups/${target.groupId}`
}

const notificationTargetLinkLabel = (item: AppNotification) => {
  if (item.type === 'new_comment' || item.type === 'new_comment_reply') return 'View post'
  if (item.type === 'new_private_message') return 'Open chat'
  if (item.type === 'new_group_message') return 'Open group chat'
  if (item.type === 'follow_request_received') return 'View profile'
  if (item.type === 'group_event_created' || item.type === 'group_event_updated') return 'Open events'
  if (item.type === 'group_event_deleted') return 'View group events'
  if (item.type === 'group_invitation_received') return 'View invite in Groups'
  if (item.type === 'group_invitation_accepted' || item.type === 'group_join_request_accepted') return 'Open group'
  if (item.type === 'group_invitation_declined' || item.type === 'group_join_request_declined') return 'View group'
  return 'Open group'
}

onMounted(async () => {
  await ensureNotificationStateLoaded()
})
</script>

<style scoped>
.notifications-page {
  min-height: 100vh;
  width: min(100%, 1320px);
  margin: 0 auto;
  padding: 0.25rem 0 2rem;
  display: grid;
  gap: 1.25rem;
  align-content: start;
  background: transparent;
}

.notifications-hero {
  position: relative;
  overflow: hidden;
  align-self: start;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0.85rem 1.4rem;
  border-radius: 1.5rem;
  border: 1px solid rgba(96, 165, 250, 0.36);
  background:
    radial-gradient(circle at top right, rgba(96, 165, 250, 0.28), transparent 32%),
    radial-gradient(circle at bottom left, rgba(147, 197, 253, 0.22), transparent 24%),
    linear-gradient(135deg, #17306b 0%, #1f4eb6 100%);
  box-shadow: 0 20px 48px rgba(37, 99, 235, 0.22);
}

.notifications-hero::after {
  content: "";
  position: absolute;
  top: -3rem;
  right: -2rem;
  width: 14rem;
  height: 14rem;
  border-radius: 999px;
  background: rgba(147, 197, 253, 0.14);
}

.hero-copy,
.hero-meta {
  position: relative;
  z-index: 1;
}

.hero-copy {
  display: grid;
  gap: 0.35rem;
  min-width: 0;
  overflow: hidden;
}

.hero-eyebrow {
  display: inline-flex;
  align-items: center;
  align-self: start;
  justify-self: start;
  width: fit-content;
  border-radius: 999px;
  padding: 0.18rem 0.75rem;
  background: rgba(219, 234, 254, 0.18);
  border: 1px solid rgba(191, 219, 254, 0.16);
  color: #eff6ff;
  font-size: 0.82rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  backdrop-filter: blur(12px);
}

.hero-copy h1 {
  margin: 0;
  font-size: clamp(1rem, 2.4vw, 1.6rem);
  line-height: 1.15;
  letter-spacing: -0.03em;
  color: #fff;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  overflow-wrap: break-word;
  word-break: break-word;
  min-width: 0;
}

.hero-copy p {
  max-width: 44rem;
  margin: 0;
  font-size: 0.8rem;
  line-height: 1.5;
  color: rgba(239, 246, 255, 0.78);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  overflow-wrap: break-word;
  word-break: break-word;
}
.hero-meta {
  display: flex;
  flex-direction: row;
  gap: 0.6rem;
  min-width: 0;
}

.hero-card {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.5rem 0.85rem;
  border-radius: 0.85rem;
  border: 1px solid rgba(147, 197, 253, 0.26);
  background: rgba(255, 255, 255, 0.1);
  color: #eff6ff;
  backdrop-filter: blur(14px);
  min-width: 0;
  overflow: hidden;
}

.hero-card strong {
  font-size: 1.4rem;
  font-weight: 800;
  line-height: 1;
  flex-shrink: 0;
}

.hero-card span {
  font-size: 0.8rem;
  font-weight: 600;
  color: rgba(219, 234, 254, 0.85);
}

.notifications-card {
  width: 100%;
  border-radius: 2rem;
  border: 1px solid rgba(191, 219, 254, 0.72);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.985), rgba(241, 247, 255, 0.94));
  box-shadow: 0 26px 60px rgba(148, 163, 184, 0.2);
  padding: 1.35rem;
   min-width: 0;
  overflow: hidden;
}

.notifications-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.9rem;
}

.notifications-header h1 {
  color: var(--text-primary);
  font-size: 1.75rem;
  font-weight: 800;
  letter-spacing: -0.03em;
  margin: 0;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.55rem;
}

.unread-pill {
  border-radius: 999px;
  padding: 0.45rem 0.8rem;
  background: rgba(219, 234, 254, 0.92);
  color: #1d4ed8;
  font-size: 0.78rem;
  font-weight: 800;
  white-space: nowrap;
  flex-shrink: 0;
}

.mark-all-btn {
  border: 1px solid transparent;
  border-radius: 999px;
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  color: #fff;
  padding: 0.72rem 1.1rem;
  font-weight: 700;
  box-shadow: 0 16px 32px rgba(37, 99, 235, 0.22);
}

.mark-all-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}


.status {
  margin-top: 1rem;
  color: #64748b;
  font-weight: 600;
}

.status.error {
  color: var(--status-error);
}

.status.with-action {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.7rem;
}

.retry-btn {
  border: 1px solid transparent;
  border-radius: 999px;
  background: linear-gradient(135deg, #334155, #1e293b);
  color: #fff;
  padding: 0.55rem 0.9rem;
  font-size: 0.78rem;
  font-weight: 700;
}

.empty-state {
  min-height: 320px;
  display: grid;
  place-items: center;
  gap: 0.9rem;
  padding: 1.4rem;
  border: 1px solid rgba(191, 219, 254, 0.72);
  border-radius: 1.75rem;
  background: linear-gradient(180deg, rgba(255,255,255,0.96), rgba(248,250,252,0.96));
  text-align: center;
}

.empty-state-img {
  width: 90px;
  height: 90px;
  opacity: 0.92;
}

.notif-list {
  margin-top: 1rem;
  display: grid;
  gap: 0.8rem;
  padding: 0;
}

.notif-item {
  list-style: none;
  border: 1px solid rgba(191, 219, 254, 0.58);
  border-left: 4px solid transparent;
  border-radius: 1.4rem;
  padding: 1rem 1.05rem;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  background: linear-gradient(180deg, rgba(255,255,255,0.98), rgba(239,246,255,0.9));
  box-shadow: 0 18px 34px rgba(148, 163, 184, 0.12);
  transition: background var(--dur-base) ease, border-left-color var(--dur-base) ease, transform 0.15s ease;
  min-width: 0;
  overflow: hidden;
}

.notif-item.unread {
  border-color: rgba(96, 165, 250, 0.4);
  border-left-color: var(--brand-500);
  background: linear-gradient(180deg, rgba(239,246,255,0.98), rgba(219,234,254,0.84));
}

.notif-item.clickable {
  cursor: pointer;
}

.notif-item.clickable:hover {
  transform: translateY(-1px);
}

.notif-item.clickable:focus-visible {
  outline: 2px solid var(--brand-500);
  outline-offset: 2px;
}


.notif-copy {
  display: grid;
  gap: 0.35rem;
  min-width: 0;
  overflow: hidden;
}

.notif-meta {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  flex-wrap: wrap;
}

.type-pill {
  border-radius: 999px;
  padding: 0.3rem 0.62rem;
  font-size: 0.72rem;
  font-weight: 800;
  background: #e2e8f0;
  color: #334155;
}

.type-pill.invite {
  background: #dbeafe;
  color: #1d4ed8;
}

.type-pill.request {
  background: #ffedd5;
  color: #c2410c;
}

.type-pill.event {
  background: #dcfce7;
  color: #166534;
}

.actor-copy {
  color: #64748b;
  font-size: 0.8rem;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.target-copy {
  color: #475569;
  font-size: 0.84rem;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.message {
  color: var(--text-primary);
  font-size: 0.98rem;
  font-weight: 700;
  margin: 0;
}
.notif-item .message {
  color: var(--text-primary);
  font-size: 0.98rem;
  font-weight: 700;
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  overflow-wrap: break-word;
  word-break: break-word;
}
.notif-foot {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  flex-wrap: wrap;
}

.date {
  color: #64748b;
  font-size: 0.8rem;
}

.open-link {
  color: #1d4ed8;
  font-size: 0.8rem;
  font-weight: 800;
  text-decoration: none;
}

.open-link:focus-visible,
.open-link:hover {
  text-decoration: underline;
}

.mark-one-btn {
  border: 1px solid transparent;
  border-radius: 999px;
  background: linear-gradient(135deg, #334155, #1e293b);
  color: #fff;
  padding: 0.6rem 0.9rem;
  font-size: 0.78rem;
  font-weight: 700;
   flex-shrink: 0;
  white-space: nowrap;
}


@media (max-width: 680px) {
  .notifications-page {
    width: 100%;
    padding: 0 0 1.5rem;
    overflow: hidden;
  }

  .notifications-hero {
    grid-template-columns: minmax(0, 1.2fr) minmax(0, 0.8fr);
    border-radius: 1.7rem;
    padding: 1.45rem 1.2rem;
  }

  .notifications-card {
    border-radius: 1.6rem;
    padding: 1rem;
  }

  .notifications-header {
    align-items: flex-start;
    flex-direction: column;
  }

  .header-actions {
    width: 100%;
    justify-content: space-between;
  }

  .notif-item {
    flex-direction: column;
  }
}
</style>
