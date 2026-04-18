<template>
  <nav class="navbar">
    <div class="left-nav">
      <router-link to="/" class="logo-link" @click="scrollToTop">
        <img class="logo-full" src="@/assets/logo.svg" alt="Nexus" />
        <img class="logo-icon" src="@/assets/logo-icon.svg" alt="Nexus" />
      </router-link>
    </div>

    <div class="center-nav">
      <NavLinks :isMobile="false" @scrollToTop="scrollToTop" />
    </div>

    <div class="right-nav">
      <button class="nav-tab nav-action-btn hamburger-btn" :class="{ 'active-link': isOpen }" @click="isOpen = !isOpen">
        <Bars3Icon v-if="!isOpen" class="nav-icon" />
        <XMarkIcon v-else class="nav-icon" />
      </button>

      <div ref="notifWrapperRef" class="notif-wrapper" @click.stop>
        <button type="button" class="nav-tab nav-action-btn" :class="{ 'active-link': isNotificationsOpen }" @click="toggleNotifications">
          <BellSolid v-if="isNotificationsOpen" class="nav-icon" />
          <BellOutline v-else class="nav-icon" />
          <span v-if="unreadNotificationsCount > 0" class="notif-badge">
            {{ unreadNotificationsCount > 99 ? '99+' : unreadNotificationsCount }}
          </span>
        </button>

        <div v-if="isNotificationsOpen" class="notif-dropdown" @click.stop>
          <div class="notif-head">
            <h3>Notifications</h3>
            <div class="notif-actions">
              <RouterLink class="notif-open-center" to="/notifications" @click="isNotificationsOpen = false">
                Open center
              </RouterLink>
              <button
                v-if="unreadNotificationsCount > 0"
                type="button"
                class="notif-mark-all"
                :disabled="isMarkingAll"
                @click="markAllAsRead"
              >
                {{ isMarkingAll ? 'Saving...' : 'Mark all read' }}
              </button>
            </div>
          </div>

          <p v-if="isLoadingNotifications" class="notif-status" aria-live="polite">Loading...</p>
          <div v-else-if="notificationsError" class="notif-status error with-action" role="alert">
            <span>{{ notificationsError }}</span>
            <button type="button" class="notif-retry-btn" @click="retryNotificationsLoad">Retry</button>
          </div>
          <template v-else>
            <p v-if="notificationActionError" class="notif-status error" role="alert">{{ notificationActionError }}</p>
            <p v-if="notifications.length === 0" class="notif-status">No notifications yet.</p>
            <ul v-else class="notif-list">
              <li
                v-for="item in notifications"
                :key="item.id"
                class="notif-item"
                :class="{ unread: !item.isRead, clickable: !item.isRead }"
                :tabindex="item.isRead ? -1 : 0"
                :aria-label="item.isRead ? `Read notification: ${item.message}` : `Unread notification: ${item.message}. Activate to mark as read`"
                :aria-disabled="item.isRead"
                @click="handleRowClick(item.id, item.isRead)"
                @keydown.enter.prevent="handleRowKeyDown($event, item.id, item.isRead)"
                @keydown.space.prevent="handleRowKeyDown($event, item.id, item.isRead)"
              >
                <div class="notif-type-line">
                  <span :class="['notif-type-pill', notificationTone(item.type)]">{{ notificationLabel(item.type) }}</span>
                  <span v-if="notificationTargetSummary(item)" class="notif-target">{{ notificationTargetSummary(item) }}</span>
                </div>
                <p class="notif-message">{{ item.message }}</p>
                <div class="notif-meta">
                  <span>{{ formatNotificationTime(item.createdAt) }}</span>
                  <button
                    v-if="!item.isRead"
                    type="button"
                    class="notif-read-btn"
                    :disabled="markingNotificationID === item.id"
                    @click.stop="markOneAsRead(item.id)"
                  >
                    {{ markingNotificationID === item.id ? '...' : 'Read' }}
                  </button>
                </div>
              </li>
            </ul>
          </template>
        </div>
      </div>

      <div ref="profileWrapperRef" class="profile-wrapper" @click.stop>
        <button
          type="button"
          class="nav-tab nav-action-btn profile-trigger"
          :class="{ 'active-link': isProfileMenuOpen || route.path === '/profile' }"
          @click="toggleProfileMenu"
        >
          <img v-if="currentUserAvatar" class="nav-profile-avatar" :src="currentUserAvatar" alt="Your profile" />
          <span v-else class="nav-profile-fallback">{{ currentUserInitials }}</span>
        </button>

        <div v-if="isProfileMenuOpen" class="profile-dropdown" @click.stop>
          <RouterLink class="profile-menu-btn profile-link" to="/profile" @click="isProfileMenuOpen = false">
            Profile
          </RouterLink>
          <button type="button" class="profile-menu-btn logout-btn" :disabled="isLoggingOut" @click="logout">
            {{ isLoggingOut ? 'Logging out...' : 'Log out' }}
          </button>
        </div>
      </div>
    </div>
  </nav>

  <Teleport to="#app">
    <div v-if="isOpen" ref="menuRef" class="mobile-menu-overlay">
      <div class="menu-links">
        <NavLinks :isMobile="true" @closeMenu="closeMobileMenu" @scrollToTop="scrollToTop" />
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import NavLinks from '@/components/NavLinks.vue'

import {
  UserCircleIcon as ProfileSolid,
  BellIcon as BellSolid,
} from '@heroicons/vue/24/solid'
import {
  UserCircleIcon as ProfileOutline,
  BellIcon as BellOutline,
  Bars3Icon,
  XMarkIcon,
} from '@heroicons/vue/24/outline'

import { useRoute, useRouter } from 'vue-router'
import { computed, ref, onMounted, onUnmounted, watch } from 'vue'
import { API_ROUTES, apiURL } from '@/api/api'
import { type AppNotification, useNotifications } from '@/composables/useNotifications'
import { fetchSessionData } from '@/router'
import type { SessionData, User } from '@/types/User'

const route = useRoute()
const router = useRouter()
const isOpen = ref(false)
const isMobile = ref(false)
const isNotificationsOpen = ref(false)
const isProfileMenuOpen = ref(false)
const isLoggingOut = ref(false)
const notifWrapperRef = ref<HTMLElement | null>(null)
const profileWrapperRef = ref<HTMLElement | null>(null)
const currentUser = ref<User | null>(null)

const currentUserInitials = computed(() => {
  const first = currentUser.value?.firstName?.[0] || ''
  const last = currentUser.value?.lastName?.[0] || ''
  return `${first}${last}`.trim().toUpperCase() || 'U'
})

const currentUserAvatar = computed(() => {
  const value = currentUser.value?.avatar || ''
  if (!value) return ''
  if (value.startsWith('http://') || value.startsWith('https://')) return value
  return apiURL(value.startsWith('/') ? value : `/${value}`)
})

const loadCurrentUser = async () => {
  const sessionData: SessionData | null = await fetchSessionData()
  if (sessionData?.user) {
    currentUser.value = sessionData.user
  }
}

const {
  notifications,
  isLoadingNotifications,
  notificationsError,
  notificationActionError,
  isMarkingAll,
  markingNotificationID,
  unreadNotificationsCount,
  ensureNotificationStateLoaded,
  loadNotifications,
  markOneAsRead,
  markAllAsRead,
} = useNotifications()

const checkWidth = () => {
  isMobile.value = window.innerWidth <= 768
  if (!isMobile.value) isOpen.value = false
}

const closeMobileMenu = () => {
  isOpen.value = false
}

const scrollToTop = () => {
  if (route.path === '/') {
    const postsContainer = document.querySelector('.PostsDiv')
    if (postsContainer) {
      postsContainer.scrollTo({
        top: 0,
        behavior: 'smooth',
      })
    }
  }
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

const toggleNotifications = async () => {
  isNotificationsOpen.value = !isNotificationsOpen.value
  if (isNotificationsOpen.value) {
    isProfileMenuOpen.value = false
    await loadNotifications()
  }
}

const handleRowClick = (notificationID: string, isRead: boolean) => {
  if (isRead) return
  void markOneAsRead(notificationID)
}

const retryNotificationsLoad = () => {
  void loadNotifications()
}

const handleRowKeyDown = (_event: KeyboardEvent, notificationID: string, isRead: boolean) => {
  if (isRead) return
  void markOneAsRead(notificationID)
}

const notificationLabel = (type: string) => {
  switch (type) {
    case 'group_invitation_received':
      return 'Invite'
    case 'group_join_request_received':
      return 'Join Request'
    case 'group_event_created':
      return 'Event'
    case 'follow_request_received':
      return 'Follow Request'
    case 'follow_request_accepted':
      return 'Accepted'
    case 'new_follower':
      return 'Follower'
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
    case 'group_invitation_received':
      return 'invite'
    case 'group_join_request_received':
      return 'request'
    case 'group_event_created':
      return 'event'
    default:
      return 'default'
  }
}

const notificationTargetSummary = (item: { type: string; target?: AppNotification['target'] }) => {
  const target = item.target
  if (!target) return ''
  if (item.type === 'group_event_created') {
    if (target.eventTitle && target.groupTitle) return `${target.eventTitle} in ${target.groupTitle}`
    if (target.eventTitle) return target.eventTitle
  }
  return target.groupTitle || ''
}

const toggleProfileMenu = () => {
  isProfileMenuOpen.value = !isProfileMenuOpen.value
  if (isProfileMenuOpen.value) {
    isNotificationsOpen.value = false
  }
}

const logout = async () => {
  if (isLoggingOut.value) return

  isLoggingOut.value = true
  try {
    await fetch(apiURL(API_ROUTES.AUTH_LOGOUT), {
      method: 'POST',
      credentials: 'include',
    })
  } finally {
    isLoggingOut.value = false
    isProfileMenuOpen.value = false
    isNotificationsOpen.value = false
    await router.replace('/login')
  }
}

const handleDocumentClick = (event: MouseEvent) => {
  const target = event.target as Node | null
  if (!target) return

  if (isNotificationsOpen.value && notifWrapperRef.value && !notifWrapperRef.value.contains(target)) {
    isNotificationsOpen.value = false
  }

  if (isProfileMenuOpen.value && profileWrapperRef.value && !profileWrapperRef.value.contains(target)) {
    isProfileMenuOpen.value = false
  }
}

onMounted(() => {
  checkWidth()
  void ensureNotificationStateLoaded()
  void loadCurrentUser()
  window.addEventListener('resize', checkWidth)
  document.addEventListener('mousedown', handleDocumentClick)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkWidth)
  document.removeEventListener('mousedown', handleDocumentClick)
})

watch(
  () => route.fullPath,
  () => {
    isNotificationsOpen.value = false
    isProfileMenuOpen.value = false
    void loadCurrentUser()
  },
)
</script>

<style scoped>
.hamburger-btn {
  display: none;
}

@media (max-width: 768px) {
  .hamburger-btn {
    display: inline-flex;
  }
}

.notif-wrapper,
.profile-wrapper {
  position: relative;
}

.nav-action-btn {
  position: relative;
  min-width: 52px;
  padding-inline: 0.9rem;
}

.profile-trigger {
  padding-inline: 0.55rem;
}

.nav-profile-avatar,
.nav-profile-fallback {
  width: 2.35rem;
  height: 2.35rem;
  border-radius: 999px;
}

.nav-profile-avatar {
  object-fit: cover;
  border: 2px solid rgba(255, 255, 255, 0.96);
  box-shadow: 0 10px 22px rgba(37, 99, 235, 0.12);
}

.nav-profile-fallback {
  display: grid;
  place-items: center;
  background: linear-gradient(135deg, rgba(191, 219, 254, 0.95), rgba(219, 234, 254, 0.9));
  color: var(--brand-700);
  font-size: 0.95rem;
  font-weight: 900;
  border: 2px solid rgba(255, 255, 255, 0.96);
  box-shadow: 0 10px 22px rgba(37, 99, 235, 0.12);
}

.notif-badge {
  position: absolute;
  top: 0.4rem;
  right: 0.45rem;
  min-width: 1.25rem;
  height: 1.25rem;
  border-radius: 999px;
  display: grid;
  place-items: center;
  background: linear-gradient(135deg, #ef4444, #dc2626);
  color: #fff;
  font-size: 0.7rem;
  font-weight: 800;
  padding: 0 0.28rem;
  box-shadow: 0 8px 16px rgba(220, 38, 38, 0.3);
}

.notif-dropdown {
  position: absolute;
  right: 0;
  top: calc(100% + 0.55rem);
  width: min(400px, calc(100vw - 2rem));
  max-height: 72vh;
  overflow: auto;
  z-index: 1000;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.94));
  border: 1px solid rgba(148, 163, 184, 0.18);
  border-radius: 22px;
  box-shadow: 0 24px 56px rgba(15, 23, 42, 0.16);
  padding: 0.8rem;
  backdrop-filter: blur(18px);
  animation: dropdownOpen 180ms var(--ease-enter) both;
  box-sizing: border-box;
  overflow-x: hidden;
  overflow-y: auto;
}

@keyframes dropdownOpen {
  from { opacity: 0; transform: translateY(-6px); }
  to { opacity: 1; transform: translateY(0); }
}

.notif-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.6rem;
  margin-bottom: 0.55rem;
    min-width: 0;

}

.notif-head h3 {
  color: var(--text-primary);
  font-size: 1rem;
  font-weight: 800;
    min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notif-actions {
  display: flex;
  align-items: center;
  gap: 0.45rem;
    flex-shrink: 0;

}

.notif-open-center {
  text-decoration: none;
  border-radius: 999px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: rgba(255, 255, 255, 0.88);
  color: var(--text-primary);
  font-size: 0.78rem;
  font-weight: 700;
  padding: 0.4rem 0.7rem;
}

.notif-mark-all {
  border: none;
  border-radius: 999px;
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  color: #fff;
  font-size: 0.78rem;
  font-weight: 700;
  padding: 0.4rem 0.72rem;
}

.notif-mark-all:disabled,
.notif-read-btn:disabled,
.logout-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.notif-status {
  color: var(--text-secondary);
  font-weight: 600;
  font-size: 0.85rem;
  overflow-wrap: break-word;
  word-break: break-word;
}

.notif-status.error {
  color: var(--status-error);
}

.notif-status.with-action {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.6rem;
}

.notif-status.with-action span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}

.notif-retry-btn {
  border: none;
  border-radius: 999px;
  background: #334155;
  color: #fff;
  font-size: 0.76rem;
  font-weight: 700;
  padding: 0.32rem 0.58rem;
  flex-shrink: 0;
  white-space: nowrap;
}


.notif-list {
  display: grid;
  gap: 0.6rem;
}

.notif-item {
  list-style: none;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 18px;
  padding: 0.75rem;
  background: rgba(255, 255, 255, 0.76);
  transition:
    transform var(--dur-fast) var(--ease-standard),
    box-shadow var(--dur-fast) var(--ease-standard),
    border-color var(--dur-fast) var(--ease-standard);

  min-width: 0;
  overflow: hidden;
}

.notif-item.unread {
  border-color: rgba(96, 165, 250, 0.34);
  background: rgba(239, 246, 255, 0.96);
  box-shadow: 0 14px 28px rgba(37, 99, 235, 0.08);
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

.notif-type-line {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  flex-wrap: nowrap;
  min-width: 0;
}

.notif-type-pill {
  border-radius: 999px;
  padding: 0.18rem 0.55rem;
  font-size: 0.68rem;
  font-weight: 800;
  background: #e2e8f0;
  color: #334155;
    flex-shrink: 0;
  white-space: nowrap;
}

.notif-type-pill.invite {
  background: #dbeafe;
  color: #1d4ed8;
}

.notif-type-pill.request {
  background: #ffedd5;
  color: #c2410c;
}

.notif-type-pill.event {
  background: #dcfce7;
  color: #166534;
}

.notif-target {
  color: var(--text-secondary);
  font-size: 0.74rem;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.notif-message {
  margin-top: 0.35rem;
  color: var(--text-primary);
  font-size: 0.88rem;
  font-weight: 600;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  overflow-wrap: break-word;
  word-break: break-word;
}

.notif-meta {
  margin-top: 0.4rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.45rem;
}

.notif-meta span {
  color: var(--text-secondary);
  font-size: 0.78rem;
}

.notif-read-btn {
  border: none;
  border-radius: 999px;
  background: #334155;
  color: #fff;
  font-size: 0.72rem;
  font-weight: 700;
  padding: 0.35rem 0.62rem;
   flex-shrink: 0;
  white-space: nowrap;
}

.profile-dropdown {
  position: absolute;
  right: 0;
  top: calc(100% + 0.55rem);
  width: 172px;
  z-index: 1000;
  display: grid;
  gap: 0.45rem;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.94));
  border: 1px solid rgba(148, 163, 184, 0.18);
  border-radius: 22px;
  box-shadow: 0 24px 56px rgba(15, 23, 42, 0.16);
  padding: 0.65rem;
  backdrop-filter: blur(18px);
  animation: dropdownOpen 180ms var(--ease-enter) both;
}

.profile-menu-btn {
  width: 100%;
  border: none;
  border-radius: 16px;
  padding: 0.72rem 0.75rem;
  font-size: 0.86rem;
  font-weight: 700;
  text-align: center;
  text-decoration: none;
  cursor: pointer;
}

.profile-link {
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  color: #fff;
}

.logout-btn {
  background: linear-gradient(135deg, #ef4444, #dc2626);
  color: #fff;
}

@media (max-width: 768px) {
  .notif-dropdown,
  .profile-dropdown {
    width: min(360px, calc(100vw - 2rem));
  }
}
</style>
