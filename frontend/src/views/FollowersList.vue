<template>
  <main class="followers-page">
    <section class="followers-shell">
      <header class="followers-hero">
        <div class="followers-hero-copy">
          <span class="followers-kicker">Your network</span>
          <h1>Keep close to the people.</h1>
          <p>{{ subtitle }}</p>
        </div>

        <div class="followers-hero-meta">
          <article>
            <strong>{{ followers.length }}</strong>
            <span>Followers</span>
          </article>
          <article v-if="isOwnList">
            <strong>{{ incomingRequests.length }}</strong>
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
            <h2>Incoming follow requests</h2>
          </div>
        </div>

        <p v-if="isRequestsLoading" class="status-message">Loading requests...</p>
        <p v-else-if="requestsErrorMessage" class="status-message error">{{ requestsErrorMessage }}</p>
        <div v-else-if="incomingRequests.length === 0" class="status-card compact">
          <p>No pending incoming requests.</p>
        </div>
        <ul v-else class="request-list">
          <li v-for="request in incomingRequests" :key="request.id" class="request-item">
            <div class="follower-main">
              <img
                v-if="request.sender && avatarSrc(request.sender.avatar)"
                :src="avatarSrc(request.sender.avatar)"
                alt="Requester avatar"
              />
              <div v-else class="avatar-fallback">
                {{ initials(request.sender?.firstName || '', request.sender?.lastName || '') }}
              </div>

              <div class="follower-meta">
                <RouterLink class="name" :to="`/users/${request.senderId}`">
                  {{ request.sender?.firstName || 'Unknown' }} {{ request.sender?.lastName || 'User' }}
                </RouterLink>
                <p class="nickname" v-if="request.sender?.nickname">@{{ request.sender.nickname }}</p>
              </div>
            </div>

            <div class="request-actions">
              <button
                class="accept-btn"
                type="button"
                :disabled="processingIncomingID === request.id"
                @click="handleIncomingRequest(request.id, 'accepted')"
              >
                {{ processingIncomingID === request.id ? 'Saving...' : 'Accept' }}
              </button>
              <button
                class="decline-btn"
                type="button"
                :disabled="processingIncomingID === request.id"
                @click="handleIncomingRequest(request.id, 'declined')"
              >
                {{ processingIncomingID === request.id ? 'Saving...' : 'Decline' }}
              </button>
            </div>
          </li>
        </ul>
      </section>

      <section class="surface-card followers-panel">
        <div class="section-head">
          <div>
            <span class="section-kicker">Followers</span>
            <h2>People in this circle</h2>
          </div>
        </div>

        <div v-if="isLoading" class="status-card">
          <p>Loading followers...</p>
        </div>
        <div v-else-if="errorMessage" class="status-card">
          <p class="status-message error">{{ errorMessage }}</p>
        </div>
        <div v-else-if="followers.length === 0" class="status-card">
          <p>No followers yet.</p>
        </div>

        <ul v-if="!isLoading && !errorMessage && followers.length > 0" class="followers-list">
          <li v-for="follower in followers" :key="follower.id" class="follower-item">
            <div class="follower-main">
              <img v-if="avatarSrc(follower.avatar)" :src="avatarSrc(follower.avatar)" alt="Follower avatar" />
              <div v-else class="avatar-fallback">{{ initials(follower.firstName, follower.lastName) }}</div>

              <div class="follower-meta">
                <RouterLink class="name" :to="`/users/${follower.id}`">
                  {{ follower.firstName }} {{ follower.lastName }}
                </RouterLink>
                <p class="nickname" v-if="follower.nickname">@{{ follower.nickname }}</p>
                <p class="nickname" v-else>{{ follower.email }}</p>
              </div>
            </div>

            <div class="item-actions">
              <RouterLink v-if="follower.id !== currentUserID" class="message-btn" :to="`/chats/private/${follower.id}`">
                Message
              </RouterLink>
              <button
                v-if="isOwnList"
                class="remove-btn"
                type="button"
                :disabled="removingFollowerID === follower.id"
                @click="removeFollower(follower.id)"
              >
                {{ removingFollowerID === follower.id ? 'Removing...' : 'Remove' }}
              </button>
            </div>
          </li>
        </ul>
      </section>
    </section>

    <div v-if="removeFollowerConfirmId" class="confirm-overlay" @click="removeFollowerConfirmId = ''">
      <div class="confirm-dialog" @click.stop>
        <h3>Remove follower</h3>
        <p>Are you sure you want to remove this follower?</p>
        <div class="confirm-actions">
          <button class="secondary" @click="removeFollowerConfirmId = ''">Cancel</button>
          <button class="danger" @click="confirmRemoveFollower">Remove</button>
        </div>
      </div>
    </div>
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

type IncomingFollowRequest = {
  id: string
  senderId: string
  receiverId: string
  status: 'pending' | 'accepted' | 'declined'
  createdAt: string
  sender?: User
}

const isLoading = ref(true)
const errorMessage = ref('')
const followers = ref<User[]>([])
const removingFollowerID = ref('')
const removeFollowerConfirmId = ref('')
const incomingRequests = ref<IncomingFollowRequest[]>([])
const isRequestsLoading = ref(false)
const requestsErrorMessage = ref('')
const processingIncomingID = ref('')
const currentUserID = ref('')
const viewedUserID = ref('')
const route = useRoute()

const routeUserID = computed(() => {
  const raw = route.params.userId
  return typeof raw === 'string' ? raw.trim() : ''
})

const isOwnList = computed(() => viewedUserID.value !== '' && viewedUserID.value === currentUserID.value)

const subtitle = computed(() => {
  return isOwnList.value ? 'People following your account.' : 'People following this account.'
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

const loadIncomingRequests = async () => {
  if (!isOwnList.value) {
    incomingRequests.value = []
    return
  }

  isRequestsLoading.value = true
  requestsErrorMessage.value = ''
  try {
    const response = await fetch(apiURL('/api/follow-requests/incoming'), {
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

    const requests = (await response.json()) as IncomingFollowRequest[]
    incomingRequests.value = requests.filter((item) => item.status === 'pending')
  } catch {
    requestsErrorMessage.value = 'Network error while loading follow requests.'
  } finally {
    isRequestsLoading.value = false
  }
}

const fetchFollowersList = async () => {
  const followersResponse = await fetch(apiURL(`/api/users/${viewedUserID.value}/followers`), {
    method: 'GET',
    credentials: 'include',
  })
  if (!followersResponse.ok) {
    let message = 'Could not load followers.'
    const payload = (await followersResponse.json().catch(() => null)) as ErrorResponse | null
    if (payload?.error) message = payload.error
    errorMessage.value = message
    return
  }

  followers.value = (await followersResponse.json()) as User[]
}

const loadFollowers = async () => {
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

    await fetchFollowersList()
    if (isOwnList.value) {
      await loadIncomingRequests()
    } else {
      incomingRequests.value = []
      requestsErrorMessage.value = ''
    }
  } catch {
    errorMessage.value = 'Network error while loading followers.'
  } finally {
    isLoading.value = false
  }
}

const removeFollower = (followerID: string) => {
  if (removingFollowerID.value) return
  removeFollowerConfirmId.value = followerID
}

const confirmRemoveFollower = async () => {
  const followerID = removeFollowerConfirmId.value
  removeFollowerConfirmId.value = ''
  if (!followerID) return

  removingFollowerID.value = followerID
  try {
    const response = await fetch(apiURL(`/api/followers/${followerID}`), {
      method: 'DELETE',
      credentials: 'include',
    })
    if (!response.ok) {
      let message = 'Could not remove follower.'
      const payload = (await response.json().catch(() => null)) as ErrorResponse | null
      if (payload?.error) message = payload.error
      errorMessage.value = message
      return
    }

    followers.value = followers.value.filter((item) => item.id !== followerID)
  } catch {
    errorMessage.value = 'Network error while removing follower.'
  } finally {
    removingFollowerID.value = ''
  }
}

const handleIncomingRequest = async (requestID: string, status: 'accepted' | 'declined') => {
  if (processingIncomingID.value) return

  processingIncomingID.value = requestID
  requestsErrorMessage.value = ''
  try {
    const response = await fetch(apiURL(`/api/follow-requests/${requestID}`), {
      method: 'PATCH',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ status }),
    })
    if (!response.ok) {
      let message = 'Could not update follow request.'
      const payload = (await response.json().catch(() => null)) as ErrorResponse | null
      if (payload?.error) message = payload.error
      requestsErrorMessage.value = message
      return
    }

    incomingRequests.value = incomingRequests.value.filter((item) => item.id !== requestID)
    if (status === 'accepted') {
      await fetchFollowersList()
    }
  } catch {
    requestsErrorMessage.value = 'Network error while updating follow request.'
  } finally {
    processingIncomingID.value = ''
  }
}

onMounted(loadFollowers)

watch(
  () => route.params.userId,
  () => {
    void loadFollowers()
  },
)
</script>

<style>
.followers-page {
  min-height: calc(100dvh - var(--navbar-height, 60px));
  padding: 0.25rem 0 1.75rem;
}

.followers-shell {
  width: min(100%, 1320px);
  margin: 0 auto;
  display: grid;
  gap: 1rem;
}

.followers-hero {
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

.followers-hero::before,
.followers-hero::after {
  content: '';
  position: absolute;
  border-radius: 999px;
  pointer-events: none;
}

.followers-hero::before {
  width: 220px;
  height: 220px;
  right: -40px;
  top: -60px;
  background: rgba(255, 255, 255, 0.08);
}

.followers-hero::after {
  width: 160px;
  height: 160px;
  left: -40px;
  bottom: -60px;
  background: rgba(125, 211, 252, 0.12);
}

.followers-hero-copy,
.followers-hero-meta {
  position: relative;
  z-index: 1;
}

.followers-kicker,
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

.followers-hero h1 {
  margin-top: 1rem;
  max-width: 10ch;
  font-size: clamp(2.2rem, 5vw, 3.5rem);
  line-height: 0.96;
  font-weight: 900;
  color: var(--white);
}

.followers-hero p {
  margin-top: 0.9rem;
  max-width: 44ch;
  color: rgba(255, 255, 255, 0.82);
  font-size: 1rem;
}

.followers-hero-meta {
  display: grid;
  gap: 0.85rem;
  align-content: end;
}

.followers-hero-meta article {
  padding: 1rem 1.05rem;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.12);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.followers-hero-meta strong {
  display: block;
  color: var(--white);
  font-size: 1.55rem;
  font-weight: 900;
}

.followers-hero-meta span {
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

.followers-list {
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
.follower-item {
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
.follower-item:hover {
  transform: translateY(-2px);
  border-color: rgba(96, 165, 250, 0.22);
  box-shadow: 0 26px 58px rgba(15, 23, 42, 0.12);
}

.follower-main {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  min-width: 0;
}

.follower-main img,
.avatar-fallback {
  width: 56px;
  height: 56px;
  border-radius: 999px;
}

.follower-main img {
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

.follower-meta {
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

.item-actions {
  display: flex;
  gap: 0.6rem;
  align-items: center;
  flex-wrap: wrap;
}

.message-btn,
.remove-btn,
.accept-btn,
.decline-btn,
.secondary,
.danger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  text-decoration: none;
  border: none;
  border-radius: 18px;
  padding: 0.78rem 0.95rem;
  font-weight: 800;
  font-size: 0.88rem;
  cursor: pointer;
  transition:
    transform var(--dur-fast) var(--ease-standard),
    box-shadow var(--dur-fast) var(--ease-standard),
    opacity var(--dur-fast) var(--ease-standard);
}

.message-btn {
  color: var(--white);
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  box-shadow: 0 14px 28px rgba(37, 99, 235, 0.18);
}

.request-actions {
  display: flex;
  gap: 0.6rem;
  flex-wrap: wrap;
}

.accept-btn {
  color: var(--white);
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  box-shadow: 0 14px 28px rgba(37, 99, 235, 0.18);
}

.decline-btn {
  color: #8a3b12;
  background: linear-gradient(180deg, rgba(255, 247, 237, 0.98), rgba(255, 237, 213, 0.9));
  border: 1px solid rgba(251, 146, 60, 0.25);
}

.remove-btn,
.danger {
  color: #fff;
  background: linear-gradient(135deg, #ef4444, #b91c1c);
  box-shadow: 0 14px 28px rgba(239, 68, 68, 0.22);
}

.secondary {
  color: var(--gray-700);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  border: 1px solid rgba(148, 163, 184, 0.18);
}

.message-btn:hover,
.remove-btn:hover,
.accept-btn:hover,
.decline-btn:hover,
.secondary:hover,
.danger:hover {
  transform: translateY(-1px);
}

.remove-btn:disabled,
.accept-btn:disabled,
.decline-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.confirm-overlay {
  position: fixed;
  inset: 0;
  background: rgba(17, 24, 39, 0.45);
  backdrop-filter: blur(10px);
  display: grid;
  place-items: center;
  z-index: 50;
  padding: 1rem;
}

.confirm-dialog {
  width: min(420px, 100%);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 28px;
  padding: 1.2rem;
  display: grid;
  gap: 0.9rem;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 28px 70px rgba(15, 23, 42, 0.2);
}

.confirm-dialog h3 {
  color: var(--gray-900);
  font-size: 1.2rem;
  font-weight: 800;
}

.confirm-dialog p {
  color: var(--gray-600);
}

.confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.65rem;
}

@media (max-width: 899px) {
  .followers-hero {
    grid-template-columns: 1fr;
  }

  .followers-hero h1 {
    max-width: 13ch;
  }
}

@media (max-width: 650px) {
  .request-item,
  .follower-item {
    align-items: flex-start;
    flex-direction: column;
  }

  .item-actions,
  .request-actions,
  .confirm-actions {
    width: 100%;
  }

  .message-btn,
  .remove-btn,
  .accept-btn,
  .decline-btn,
  .secondary,
  .danger {
    flex: 1 1 0;
  }
}
</style>
