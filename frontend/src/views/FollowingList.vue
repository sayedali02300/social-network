<template>
  <main class="following-page">
    <section class="following-shell">
      <header class="following-hero">
        <div class="following-hero-copy">
          <span class="following-kicker">Your network</span>
          <h1>Stay close to the people.</h1>
          <p>{{ subtitle }}</p>
        </div>

        <div class="following-hero-meta">
          <article>
            <strong>{{ following.length }}</strong>
            <span>Following</span>
          </article>
          <article v-if="isOwnList">
            <strong>{{ outgoingRequests.length }}</strong>
            <span>Pending requests</span>
          </article>
          <article v-else>
            <strong>{{ viewedUserID ? 'Live' : 'Ready' }}</strong>
            <span>Connection view</span>
          </article>
        </div>
      </header>

      <section v-if="isOwnList" class="surface-card requests-panel">
        <div class="section-head">
          <div>
            <span class="section-kicker">Requests</span>
            <h2>Outgoing follow requests</h2>
          </div>
        </div>

        <p v-if="isRequestsLoading" class="status-message">Loading requests...</p>
        <p v-else-if="requestsErrorMessage" class="status-message error">{{ requestsErrorMessage }}</p>
        <div v-else-if="outgoingRequests.length === 0" class="status-card compact">
          <p>No pending outgoing requests.</p>
        </div>
        <ul v-else class="request-list">
          <li v-for="request in outgoingRequests" :key="request.id" class="request-item">
            <div class="following-main">
              <img
                v-if="request.receiver && avatarSrc(request.receiver.avatar)"
                :src="avatarSrc(request.receiver.avatar)"
                alt="Requested user avatar"
              />
              <div v-else class="avatar-fallback">
                {{ initials(request.receiver?.firstName || '', request.receiver?.lastName || '') }}
              </div>

              <div class="following-meta">
                <RouterLink class="name" :to="`/users/${request.receiverId}`">
                  {{ request.receiver?.firstName || 'Unknown' }} {{ request.receiver?.lastName || 'User' }}
                </RouterLink>
                <p class="nickname" v-if="request.receiver?.nickname">@{{ request.receiver.nickname }}</p>
              </div>
            </div>
            <button
              class="withdraw-btn"
              type="button"
              :disabled="processingOutgoingID === request.id"
              @click="withdrawRequest(request.id)"
            >
              {{ processingOutgoingID === request.id ? 'Withdrawing...' : 'Withdraw' }}
            </button>
          </li>
        </ul>
      </section>

      <section class="surface-card following-panel">
        <div class="section-head">
          <div>
            <span class="section-kicker">Following</span>
            <h2>People in your orbit</h2>
          </div>
        </div>

        <div v-if="isLoading" class="status-card">
          <p>Loading following list...</p>
        </div>
        <div v-else-if="errorMessage" class="status-card">
          <p class="status-message error">{{ errorMessage }}</p>
        </div>
        <div v-else-if="following.length === 0" class="status-card">
          <p>You are not following anyone yet.</p>
        </div>

        <ul v-if="!isLoading && !errorMessage && following.length > 0" class="following-list">
          <li v-for="user in following" :key="user.id" class="following-item">
            <div class="following-main">
              <img v-if="avatarSrc(user.avatar)" :src="avatarSrc(user.avatar)" alt="User avatar" />
              <div v-else class="avatar-fallback">{{ initials(user.firstName, user.lastName) }}</div>

              <div class="following-meta">
                <RouterLink class="name" :to="`/users/${user.id}`">
                  {{ user.firstName }} {{ user.lastName }}
                </RouterLink>
                <p class="nickname" v-if="user.nickname">@{{ user.nickname }}</p>
                <p class="nickname" v-else>{{ user.email }}</p>
              </div>
            </div>
            <RouterLink class="message-btn" :to="`/chats/private/${user.id}`">Message</RouterLink>
          </li>
        </ul>
      </section>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { API_ROUTES, apiURL } from '@/api/api'

type User = {
  id: string
  email: string
  firstName: string
  lastName: string
  avatar: string
  nickname: string
}

type ErrorResponse = {
  error?: string
}

type OutgoingFollowRequest = {
  id: string
  senderId: string
  receiverId: string
  status: 'pending' | 'accepted' | 'declined'
  createdAt: string
  receiver?: User
}

const isLoading = ref(true)
const errorMessage = ref('')
const following = ref<User[]>([])
const outgoingRequests = ref<OutgoingFollowRequest[]>([])
const isRequestsLoading = ref(false)
const requestsErrorMessage = ref('')
const processingOutgoingID = ref('')
const currentUserID = ref('')
const viewedUserID = ref('')
const route = useRoute()

const routeUserID = computed(() => {
  const raw = route.params.userId
  return typeof raw === 'string' ? raw.trim() : ''
})

const isOwnList = computed(() => viewedUserID.value !== '' && viewedUserID.value === currentUserID.value)

const subtitle = computed(() => {
  return isOwnList.value ? 'People you are currently following.' : 'People this account follows.'
})

const initials = (firstName: string, lastName: string) => {
  const first = firstName?.[0] || ''
  const last = lastName?.[0] || ''
  const value = `${first}${last}`.trim().toUpperCase()
  return value || '?'
}

const avatarSrc = (value: string) => {
  if (!value) return ''
  if (value.startsWith('http://') || value.startsWith('https://')) return value
  return apiURL(value.startsWith('/') ? value : `/${value}`)
}

const loadOutgoingRequests = async () => {
  if (!isOwnList.value) {
    outgoingRequests.value = []
    return
  }

  isRequestsLoading.value = true
  requestsErrorMessage.value = ''
  try {
    const response = await fetch(apiURL('/api/follow-requests/outgoing'), {
      method: 'GET',
      credentials: 'include',
    })
    if (!response.ok) {
      let message = 'Could not load follow requests.'
      const payload = (await response.json().catch(() => null)) as ErrorResponse | null
      if (payload?.error) message = payload.error
      requestsErrorMessage.value = message
      return
    }

    const requests = (await response.json()) as OutgoingFollowRequest[]
    outgoingRequests.value = requests.filter((item) => item.status === 'pending')
  } catch {
    requestsErrorMessage.value = 'Network error while loading follow requests.'
  } finally {
    isRequestsLoading.value = false
  }
}

const loadFollowing = async () => {
  isLoading.value = true
  errorMessage.value = ''

  try {
    const meResponse = await fetch(apiURL(API_ROUTES.USERS_ME), {
      method: 'GET',
      credentials: 'include',
    })
    if (!meResponse.ok) {
      let message = 'Could not load account.'
      const payload = (await meResponse.json().catch(() => null)) as ErrorResponse | null
      if (payload?.error) message = payload.error
      errorMessage.value = message
      return
    }

    const me = (await meResponse.json()) as { id: string }
    currentUserID.value = me.id
    viewedUserID.value = routeUserID.value || me.id

    const followingResponse = await fetch(apiURL(`/api/users/${viewedUserID.value}/following`), {
      method: 'GET',
      credentials: 'include',
    })
    if (!followingResponse.ok) {
      let message = 'Could not load following list.'
      const payload = (await followingResponse.json().catch(() => null)) as ErrorResponse | null
      if (payload?.error) message = payload.error
      errorMessage.value = message
      return
    }

    following.value = (await followingResponse.json()) as User[]
    if (isOwnList.value) {
      await loadOutgoingRequests()
    } else {
      outgoingRequests.value = []
      requestsErrorMessage.value = ''
    }
  } catch {
    errorMessage.value = 'Network error while loading following list.'
  } finally {
    isLoading.value = false
  }
}

const withdrawRequest = async (requestID: string) => {
  if (processingOutgoingID.value) return

  processingOutgoingID.value = requestID
  requestsErrorMessage.value = ''
  try {
    const response = await fetch(apiURL(`/api/follow-requests/${requestID}`), {
      method: 'DELETE',
      credentials: 'include',
    })
    if (!response.ok) {
      let message = 'Could not withdraw follow request.'
      const payload = (await response.json().catch(() => null)) as ErrorResponse | null
      if (payload?.error) message = payload.error
      requestsErrorMessage.value = message
      return
    }

    outgoingRequests.value = outgoingRequests.value.filter((item) => item.id !== requestID)
  } catch {
    requestsErrorMessage.value = 'Network error while withdrawing follow request.'
  } finally {
    processingOutgoingID.value = ''
  }
}

onMounted(loadFollowing)

watch(
  () => route.params.userId,
  () => {
    void loadFollowing()
  },
)
</script>

<style>
.following-page {
  min-height: calc(100dvh - var(--navbar-height, 60px));
  padding: 0.25rem 0 1.75rem;
}

.following-shell {
  width: min(100%, 1320px);
  margin: 0 auto;
  display: grid;
  gap: 1rem;
}

.following-hero {
  display: grid;
  grid-template-columns: minmax(0, 1.15fr) minmax(260px, 0.85fr);
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

.following-hero::before,
.following-hero::after {
  content: '';
  position: absolute;
  border-radius: 999px;
  pointer-events: none;
}

.following-hero::before {
  width: 220px;
  height: 220px;
  right: -40px;
  top: -60px;
  background: rgba(255, 255, 255, 0.08);
}

.following-hero::after {
  width: 160px;
  height: 160px;
  left: -40px;
  bottom: -60px;
  background: rgba(125, 211, 252, 0.12);
}

.following-hero-copy,
.following-hero-meta {
  position: relative;
  z-index: 1;
}

.following-kicker,
.section-kicker {
  display: inline-flex;
  align-items: center;
  padding: 0.38rem 0.78rem;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.12);
  color: rgba(255, 255, 255, 0.84);
  font-size: 0.76rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.following-hero h1 {
  margin-top: 1rem;
  max-width: 10ch;
  font-size: clamp(2.2rem, 5vw, 3.5rem);
  line-height: 0.96;
  font-weight: 900;
  color: var(--white);
}

.following-hero p {
  margin-top: 0.9rem;
  max-width: 44ch;
  color: rgba(255, 255, 255, 0.82);
  font-size: 1rem;
}

.following-hero-meta {
  display: grid;
  gap: 0.85rem;
  align-content: end;
}

.following-hero-meta article {
  padding: 1rem 1.05rem;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.12);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.following-hero-meta strong {
  display: block;
  color: var(--white);
  font-size: 1.55rem;
  font-weight: 900;
}

.following-hero-meta span {
  color: rgba(255, 255, 255, 0.76);
  font-size: 0.86rem;
  font-weight: 700;
}

.surface-card {
  display: grid;
  gap: 1rem;
  padding: 1rem;
  border-radius: 32px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(248, 250, 252, 0.86));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.08);
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.section-head h2 {
  margin-top: 0.7rem;
  color: var(--gray-900);
  font-size: 1.5rem;
  line-height: 1.02;
  font-weight: 900;
}

.section-kicker {
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.14), rgba(14, 165, 233, 0.08));
  color: var(--brand-700);
}

.status-message {
  color: var(--text-secondary);
  font-weight: 700;
}

.status-message.error {
  color: var(--status-error);
}

.status-card {
  min-height: 220px;
  border-radius: 28px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.92), rgba(248, 250, 252, 0.84));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.06);
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 1.5rem;
}

.status-card.compact {
  min-height: 120px;
}

.status-card p {
  margin: 0;
  color: var(--gray-600);
  font-weight: 700;
}

.following-list {
  display: grid;
  gap: 1rem;
  margin: 0;
  padding: 0;
}

.requests-panel {
  margin: 0;
}

.request-list {
  display: grid;
  gap: 1rem;
  margin: 0;
  padding: 0;
}

.request-item,
.following-item {
  list-style: none;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 28px;
  padding: 1rem;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.88));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.08);
  transition:
    transform var(--dur-base) var(--ease-standard),
    box-shadow var(--dur-base) var(--ease-standard),
    border-color var(--dur-base) var(--ease-standard);
}

.request-item:hover,
.following-item:hover {
  transform: translateY(-2px);
  border-color: rgba(96, 165, 250, 0.22);
  box-shadow: 0 26px 58px rgba(15, 23, 42, 0.12);
}

.following-main {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  min-width: 0;
}

.following-main img,
.avatar-fallback {
  width: 56px;
  height: 56px;
  border-radius: 999px;
}

.following-main img {
  object-fit: cover;
  border: 1px solid rgba(148, 163, 184, 0.16);
}

.avatar-fallback {
  display: grid;
  place-items: center;
  flex-shrink: 0;
  background: var(--brand-100);
  color: var(--brand-600);
  font-weight: 800;
}

.following-meta {
  min-width: 0;
}

.name {
  display: inline-block;
  color: var(--gray-900);
  font-weight: 800;
  text-decoration: none;
}

.name:hover {
  color: var(--brand-600);
}

.nickname {
  margin-top: 0.2rem;
  color: var(--gray-500);
  font-size: 0.88rem;
  font-weight: 600;
  overflow-wrap: anywhere;
}

.withdraw-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  text-decoration: none;
  border: 1px solid rgba(148, 163, 184, 0.18);
  border-radius: 18px;
  padding: 0.78rem 0.95rem;
  font-weight: 800;
  font-size: 0.88rem;
  color: var(--gray-700);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  cursor: pointer;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
}

.message-btn {
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 18px;
  padding: 0.78rem 0.95rem;
  font-weight: 800;
  font-size: 0.88rem;
  color: #fff;
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  box-shadow: 0 14px 28px rgba(37, 99, 235, 0.18);
}

.withdraw-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

@media (max-width: 899px) {
  .following-hero {
    grid-template-columns: 1fr;
  }

  .following-hero h1 {
    max-width: 13ch;
  }
}

@media (max-width: 650px) {
  .request-item,
  .following-item {
    align-items: flex-start;
    flex-direction: column;
  }

  .message-btn,
  .withdraw-btn {
    width: 100%;
  }
}
</style>
