<template>
  <main class="search-page">
    <section class="search-shell">
      <header class="search-hero">
        <div class="search-hero-copy">
          <span class="search-kicker">Discover people</span>
          <h1>Expand your network </h1>
          <p>Search people by name or nickname and go directly to their profile.</p>
        </div>

        <div class="search-hero-meta">
          <article>
            <strong>{{ activeQuery ? results.length : 0 }}</strong>
            <span>Results shown</span>
          </article>
          <article>
            <strong>{{ activeQuery || 'Ready' }}</strong>
            <span>Current search</span>
          </article>
        </div>
      </header>

      <section class="search-panel">
        <form class="search-form" @submit.prevent="submitSearch">
          <div class="search-input-wrap">
            <input
              v-model.trim="searchTerm"
              type="search"
              class="search-input"
              placeholder="Search by name or nickname"
              maxlength="80"
              autocomplete="off"
            />
          </div>

          <button type="submit" class="search-button" :disabled="isLoading">
            {{ isLoading ? 'Searching...' : 'Search' }}
          </button>
        </form>

        <section v-if="!activeQuery && suggestedUsers.length > 0" class="suggestions-panel">
          <div class="suggestions-head">
            <h2>Suggested people</h2>
            <p>Recent authors from your network</p>
          </div>

          <div class="suggestions-list">
            <div
              v-for="user in suggestedUsers"
              :key="user.id"
              class="suggestion-chip"
            >
              <button type="button" class="suggestion-main suggestion-profile-btn" @click="openSuggestedProfile(user.id)">
                <img v-if="avatarSrc(user.avatar)" :src="avatarSrc(user.avatar)" alt="Suggested user avatar" />
                <div v-else class="avatar-fallback suggestion-avatar">{{ initials(user.firstName, user.lastName) }}</div>
                <span class="suggestion-copy">
                  <strong>{{ user.firstName || user.lastName ? `${user.firstName} ${user.lastName}`.trim() : 'User' }}</strong>
                  <small>{{ user.nickname ? '@' + user.nickname : 'View posts by this user' }}</small>
                </span>
              </button>
              <button
                v-if="getFollowStatus(user.id) === 'following'"
                type="button"
                class="suggestion-action suggestion-action--following"
                :disabled="followActionInProgress[user.id]"
                @click.stop="handleUnfollow(user.id, false)"
              >Following</button>
              <button
                v-else-if="getFollowStatus(user.id) === 'requested'"
                type="button"
                class="suggestion-action suggestion-action--requested"
                :disabled="followActionInProgress[user.id]"
                @click.stop="handleCancelRequest(user.id, false)"
              >Requested</button>
              <button
                v-else
                type="button"
                class="suggestion-action suggestion-action--follow"
                :disabled="followActionInProgress[user.id]"
                @click.stop="handleFollow(user.id, false)"
              >Follow</button>
            </div>
          </div>
        </section>

        <p v-if="errorMessage" class="status-message error">{{ errorMessage }}</p>
        <p v-else-if="isLoading" class="status-message">Loading results...</p>
        <div v-else-if="!activeQuery" class="status-card empty-state">
          <img src="@/assets/empty-states/search-empty.svg" alt="" class="empty-state-img" />
          <p>Enter a name or nickname to start searching.</p>
        </div>
        <div v-else-if="results.length === 0" class="status-card empty-state">
          <img src="@/assets/empty-states/search-empty.svg" alt="" class="empty-state-img" />
          <p>No users matched "{{ activeQuery }}".</p>
        </div>

        <ul v-else class="results-list">
          <li v-for="user in results" :key="user.id" class="result-item">
            <div class="result-row">
              <div class="result-main">
                <img v-if="avatarSrc(user.avatar)" :src="avatarSrc(user.avatar)" alt="User avatar" />
                <div v-else class="avatar-fallback">{{ initials(user.firstName, user.lastName) }}</div>

                <div class="result-meta">
                  <RouterLink class="name" :to="`/users/${user.id}`">
                    {{ user.firstName || user.lastName ? `${user.firstName} ${user.lastName}`.trim() : 'Private account' }}
                  </RouterLink>
                  <p class="nickname">
                    {{ user.nickname ? '@' + user.nickname : user.isPublic ? 'Public account' : 'Private account' }}
                  </p>
                </div>
              </div>

              <button
                v-if="user.followStatus === 'following'"
                type="button"
                class="view-profile-btn view-profile-btn--following"
                :disabled="followActionInProgress[user.id]"
                @click="handleUnfollow(user.id, true)"
              >Following</button>
              <button
                v-else-if="user.followStatus === 'requested'"
                type="button"
                class="view-profile-btn view-profile-btn--requested"
                :disabled="followActionInProgress[user.id]"
                @click="handleCancelRequest(user.id, true)"
              >Requested</button>
              <button
                v-else
                type="button"
                class="view-profile-btn view-profile-btn--follow"
                :disabled="followActionInProgress[user.id]"
                @click="handleFollow(user.id, true)"
              >Follow</button>
            </div>
          </li>
        </ul>
      </section>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { API_ROUTES, apiURL } from '@/api/api'
import type { Post } from '@/types/post'
import { fetchSessionData } from '@/router'

type SearchUser = {
  id: string
  email: string
  firstName: string
  lastName: string
  avatar: string
  nickname: string
  isPublic: boolean
  followStatus: 'following' | 'requested' | ''
}

type ErrorResponse = {
  error?: string
}

const route = useRoute()
const router = useRouter()
const searchTerm = ref('')
const activeQuery = ref('')
const results = ref<SearchUser[]>([])
const suggestedUsers = ref<SearchUser[]>([])
const isLoading = ref(false)
const errorMessage = ref('')
const currentUserId = ref('')
// Map userId → followStatus for suggested users (search results carry status from API)
const suggestedFollowStatus = ref<Record<string, 'following' | 'requested' | ''>>({})
// Map userId → pending follow-request ID (for cancelling)
const pendingRequestIds = ref<Record<string, string>>({})
const followActionInProgress = ref<Record<string, boolean>>({})
let requestSequence = 0
let debounceTimer: number | null = null

const routeQuery = computed(() => {
  const raw = route.query.q
  return typeof raw === 'string' ? raw.trim() : ''
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

const clearResults = () => {
  activeQuery.value = ''
  results.value = []
  errorMessage.value = ''
  isLoading.value = false
}

const loadSuggestedUsers = async () => {
  try {
    const [feedRes, outgoingRes, followingRes] = await Promise.all([
      fetch(apiURL(API_ROUTES.FEED), { credentials: 'include' }),
      fetch(apiURL('/api/follow-requests/outgoing'), { credentials: 'include' }),
      fetch(apiURL(`/api/users/${currentUserId.value}/following`), { credentials: 'include' }),
    ])

    if (!feedRes.ok) return

    const posts = (await feedRes.json()) as Post[]
    const authorMap = new Map<string, SearchUser>()

    for (const post of posts) {
      if (post.author?.id === currentUserId.value) continue
      if (!post.author?.id || authorMap.has(post.author.id)) continue
      authorMap.set(post.author.id, {
        id: post.author.id,
        email: '',
        firstName: post.author.firstName || '',
        lastName: post.author.lastName || '',
        avatar: post.author.avatar || '',
        nickname: post.author.nickname || '',
        isPublic: true,
        followStatus: '',
      })
      if (authorMap.size >= 5) break
    }

    suggestedUsers.value = Array.from(authorMap.values())

    // Build follow status map for suggested users
    const statusMap: Record<string, 'following' | 'requested' | ''> = {}
    const reqIdMap: Record<string, string> = {}

    if (followingRes.ok) {
      const following = (await followingRes.json()) as { id: string }[]
      for (const u of following) statusMap[u.id] = 'following'
    }
    if (outgoingRes.ok) {
      const outgoing = (await outgoingRes.json()) as { id: string; receiverId: string; status: string }[]
      for (const req of outgoing) {
        if (req.status === 'pending' && !statusMap[req.receiverId]) {
          statusMap[req.receiverId] = 'requested'
          reqIdMap[req.receiverId] = req.id
        }
      }
    }

    suggestedFollowStatus.value = statusMap
    pendingRequestIds.value = reqIdMap
  } catch {
    suggestedUsers.value = []
  }
}

const getFollowStatus = (userId: string, fromSearch = false, searchStatus = ''): 'following' | 'requested' | '' => {
  if (fromSearch) return (searchStatus as 'following' | 'requested' | '') || ''
  return suggestedFollowStatus.value[userId] || ''
}

const handleFollow = async (userId: string, fromSearch = false) => {
  if (followActionInProgress.value[userId]) return
  followActionInProgress.value[userId] = true
  try {
    const res = await fetch(apiURL('/api/follow-requests'), {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ receiverId: userId }),
    })
    if (!res.ok) return
    const data = (await res.json()) as { id: string; status: string }
    const newStatus = data.status === 'accepted' ? 'following' : 'requested'
    if (newStatus === 'requested') {
      pendingRequestIds.value = { ...pendingRequestIds.value, [userId]: data.id }
    }
    if (fromSearch) {
      const idx = results.value.findIndex(u => u.id === userId)
      if (idx !== -1) results.value[idx] = { ...results.value[idx], followStatus: newStatus } as SearchUser
    } else {
      suggestedFollowStatus.value = { ...suggestedFollowStatus.value, [userId]: newStatus }
    }
  } finally {
    followActionInProgress.value[userId] = false
  }
}

const handleUnfollow = async (userId: string, fromSearch = false) => {
  if (followActionInProgress.value[userId]) return
  followActionInProgress.value[userId] = true
  try {
    const res = await fetch(apiURL(`/api/following/${userId}`), {
      method: 'DELETE',
      credentials: 'include',
    })
    if (!res.ok) return
    if (fromSearch) {
      const idx = results.value.findIndex(u => u.id === userId)
      if (idx !== -1) results.value[idx] = { ...results.value[idx], followStatus: '' } as SearchUser
    } else {
      suggestedFollowStatus.value = { ...suggestedFollowStatus.value, [userId]: '' }
    }
  } finally {
    followActionInProgress.value[userId] = false
  }
}

const handleCancelRequest = async (userId: string, fromSearch = false) => {
  if (followActionInProgress.value[userId]) return
  const requestId = pendingRequestIds.value[userId]
  if (!requestId) return
  followActionInProgress.value[userId] = true
  try {
    const res = await fetch(apiURL(`/api/follow-requests/${requestId}`), {
      method: 'DELETE',
      credentials: 'include',
    })
    if (!res.ok) return
    if (fromSearch) {
      const idx = results.value.findIndex(u => u.id === userId)
      if (idx !== -1) results.value[idx] = { ...results.value[idx], followStatus: '' } as SearchUser
    } else {
      suggestedFollowStatus.value = { ...suggestedFollowStatus.value, [userId]: '' }
    }
    const updated = { ...pendingRequestIds.value }
    delete updated[userId]
    pendingRequestIds.value = updated
  } finally {
    followActionInProgress.value[userId] = false
  }
}

const runSearch = async (query: string) => {
  if (!query) {
    clearResults()
    return
  }

  const currentRequest = ++requestSequence
  isLoading.value = true
  errorMessage.value = ''
  activeQuery.value = query

  try {
    const response = await fetch(apiURL(`${API_ROUTES.USERS_SEARCH}?q=${encodeURIComponent(query)}`), {
      method: 'GET',
      credentials: 'include',
    })

    if (!response.ok) {
      let message = 'Could not search users.'
      const payload = (await response.json().catch(() => null)) as ErrorResponse | null
      if (payload?.error) message = payload.error
      if (currentRequest !== requestSequence) return
      errorMessage.value = message
      results.value = []
      return
    }

    const users = (await response.json()) as SearchUser[]
    if (currentRequest !== requestSequence) return
    results.value = users
  } catch {
    if (currentRequest !== requestSequence) return
    errorMessage.value = 'Network error while searching users.'
    results.value = []
  } finally {
    if (currentRequest === requestSequence) {
      isLoading.value = false
    }
  }
}

const submitSearch = async () => {
  const query = searchTerm.value.trim()
  await router.replace({
    path: '/search',
    query: query ? { q: query } : {},
  })
}

const openSuggestedProfile = async (userId: string) => {
  await router.push(`/users/${userId}`)
}

// Realtime: debounce input → search directly without waiting for form submit
watch(searchTerm, (value) => {
  const query = value.trim()
  if (debounceTimer !== null) window.clearTimeout(debounceTimer)
  if (!query) {
    clearResults()
    return
  }
  debounceTimer = window.setTimeout(() => {
    debounceTimer = null
    void runSearch(query)
  }, 300)
})

// Keep route-query in sync (supports deep-linking / browser back)
watch(
  routeQuery,
  (value) => {
    if (value && value !== searchTerm.value) {
      searchTerm.value = value
    }
    if (!value) {
      clearResults()
      return
    }
    void runSearch(value)
  },
  { immediate: true },
)

onMounted(() => {
  void (async () => {
    const sessionData = await fetchSessionData()
    currentUserId.value = sessionData?.user?.id || ''
    await loadSuggestedUsers()
  })()
})
</script>

<style>
.search-page {
  min-height: calc(100dvh - var(--navbar-height, 60px));
}

.search-shell {
  width: min(100%, 1320px);
  margin: 0 auto;
  display: grid;
  gap: 1.1rem;
}

.search-hero {
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
  width: 100%;
}

.search-hero::before,
.search-hero::after {
  content: '';
  position: absolute;
  border-radius: 999px;
  pointer-events: none;
}

.search-hero::before {
  width: 220px;
  height: 220px;
  right: -40px;
  top: -60px;
  background: rgba(255, 255, 255, 0.08);
}

.search-hero::after {
  width: 160px;
  height: 160px;
  left: -40px;
  bottom: -60px;
  background: rgba(125, 211, 252, 0.12);
}

.search-hero-copy,
.search-hero-meta {
  position: relative;
  z-index: 1;
}

.search-kicker {
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

.search-hero h1 {
  margin-top: 1rem;
  max-width: 11ch;
  font-size: clamp(2.2rem, 5vw, 3.5rem);
  line-height: 0.96;
  font-weight: 900;
  color: var(--white);
}

.search-hero p {
  margin-top: 0.9rem;
  max-width: 44ch;
  color: rgba(255, 255, 255, 0.82);
  font-size: 1rem;
}

.search-hero-meta {
  display: grid;
  gap: 0.85rem;
  align-content: end;
}

.search-hero-meta article {
  padding: 1rem 1.05rem;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.12);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.search-hero-meta strong {
  display: block;
  color: var(--white);
  font-size: 1.55rem;
  font-weight: 900;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.search-hero-meta span {
  color: rgba(255, 255, 255, 0.76);
  font-size: 0.86rem;
  font-weight: 700;
}

.search-panel {
  display: grid;
  gap: 1rem;
  width: 100%;
}

.search-form {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 0.75rem;
  align-items: center;
  padding: 0.75rem;
  border-radius: 24px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(239, 246, 255, 0.76));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 12px 28px rgba(15, 23, 42, 0.05);
  width: 100%;
}

.search-input-wrap {
  position: relative;
  min-width: 0;
  display: flex;
  align-items: center;
  padding: 0.95rem 1rem;
  border-radius: 18px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 8px 22px rgba(15, 23, 42, 0.04);
}

.search-input {
  display: block;
  width: 100%;
  min-width: 0;
  box-sizing: border-box;
  min-height: 0;
  border: none;
  border-radius: 0;
  padding: 0;
  background: transparent;
  color: var(--gray-800);
  font: inherit;
  font-size: 0.95rem;
  font-weight: 600;
  box-shadow: none;
}

.search-input::placeholder {
  color: #94a3b8;
}

.search-input:focus {
  outline: none;
}

.search-input-wrap:focus-within {
  border-color: rgba(37, 99, 235, 0.55);
  box-shadow:
    0 0 0 5px rgba(37, 99, 235, 0.1),
    0 16px 34px rgba(37, 99, 235, 0.12);
}

.search-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 18px;
  padding: 0.88rem 1.25rem;
  font-weight: 800;
  font-size: 0.92rem;
  color: var(--white);
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  cursor: pointer;
  box-shadow: 0 16px 32px rgba(37, 99, 235, 0.24);
}

.search-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  box-shadow: none;
}

.status-message {
  color: var(--text-secondary);
  font-weight: 700;
}

.suggestions-panel {
  display: grid;
  gap: 0.85rem;
  padding: 1rem;
  border-radius: 28px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(248, 250, 252, 0.86));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.06);
}

.suggestions-head {
  display: grid;
  gap: 0.2rem;
}

.suggestions-head h2 {
  margin: 0;
  color: var(--gray-900);
  font-size: 1.05rem;
  font-weight: 800;
}

.suggestions-head p {
  margin: 0;
  color: var(--gray-500);
  font-size: 0.88rem;
  font-weight: 600;
}

.suggestions-list {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 0.75rem;
}

.suggestion-chip {
  display: grid;
  gap: 0.7rem;
  width: 100%;
  min-width: 0;
  padding: 0.88rem 0.8rem;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 20px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  color: var(--gray-800);
  text-align: left;
  cursor: pointer;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 8px 22px rgba(15, 23, 42, 0.04);
  transition:
    transform var(--dur-base) var(--ease-standard),
    box-shadow var(--dur-base) var(--ease-standard),
    border-color var(--dur-base) var(--ease-standard);
}

.suggestion-main {
  display: grid;
  justify-items: center;
  text-align: center;
  gap: 0.55rem;
  min-width: 0;
}

.suggestion-chip:hover {
  transform: translateY(-1px);
  border-color: rgba(96, 165, 250, 0.24);
  box-shadow: 0 16px 32px rgba(37, 99, 235, 0.1);
}

.suggestion-chip img,
.suggestion-avatar {
  width: 44px;
  height: 44px;
  flex-shrink: 0;
  border-radius: 999px;
  object-fit: cover;
}

.suggestion-copy {
  display: grid;
  gap: 0.12rem;
  min-width: 0;
  width: 100%;
}

.suggestion-copy strong,
.suggestion-copy small {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.suggestion-copy strong {
  color: var(--gray-900);
  font-size: 0.9rem;
  font-weight: 800;
}

.suggestion-copy small {
  color: var(--gray-500);
  font-size: 0.78rem;
  font-weight: 600;
}

.suggestion-profile-btn {
  background: none;
  border: none;
  padding: 0;
  cursor: pointer;
  text-align: inherit;
  width: 100%;
}

.suggestion-action {
  display: inline-flex;
  justify-content: center;
  align-items: center;
  border: none;
  border-radius: 16px;
  width: 100%;
  padding: 0.62rem 0.8rem;
  font-size: 0.8rem;
  font-weight: 800;
  cursor: pointer;
  white-space: nowrap;
  transition: opacity 0.15s;
}

.suggestion-action:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.suggestion-action--follow {
  color: var(--white);
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  box-shadow: 0 12px 24px rgba(37, 99, 235, 0.18);
}

.suggestion-action--following {
  color: var(--brand-700, #1d4ed8);
  background: rgba(37, 99, 235, 0.08);
  box-shadow: none;
}

.suggestion-action--following:hover:not(:disabled) {
  color: #dc2626;
  background: rgba(220, 38, 38, 0.08);
}

.suggestion-action--requested {
  color: var(--gray-600);
  background: rgba(148, 163, 184, 0.12);
  box-shadow: none;
}

.suggestion-action--requested:hover:not(:disabled) {
  color: #dc2626;
  background: rgba(220, 38, 38, 0.08);
}

.status-message.error {
  color: var(--status-error);
}

.status-card {
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
  width: 100%;
}

.status-card p {
  margin: 0;
  color: var(--gray-600);
  font-weight: 700;
}

.empty-state-img {
  width: clamp(84px, 10vw, 120px);
  height: auto;
  opacity: 0.92;
}

.results-list {
  margin: 0;
  padding: 0;
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  width: 100%;
}

.result-item {
  list-style: none;
  display: grid;
  gap: 0.75rem;
  min-height: 0;
  border-radius: 20px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.88));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 8px 22px rgba(15, 23, 42, 0.04);
  padding: 0.88rem 0.95rem;
  transition:
    transform var(--dur-base) var(--ease-standard),
    box-shadow var(--dur-base) var(--ease-standard),
    border-color var(--dur-base) var(--ease-standard);
}

.result-item:hover {
  transform: translateY(-1px);
  border-color: rgba(96, 165, 250, 0.24);
  box-shadow: 0 16px 32px rgba(37, 99, 235, 0.1);
}

.result-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 0.75rem;
  width: 100%;
}

.result-main {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  min-width: 0;
  flex: 1 1 auto;
}

.result-main img,
.avatar-fallback {
  width: 44px;
  height: 44px;
  border-radius: 999px;
}

.result-main img {
  object-fit: cover;
  border: 1px solid rgba(148, 163, 184, 0.16);
}

.avatar-fallback {
  display: grid;
  place-items: center;
  background: var(--brand-100);
  color: var(--brand-600);
  font-weight: 800;
  flex-shrink: 0;
}

.result-meta {
  min-width: 0;
  flex: 1 1 auto;
}

.name {
  display: inline-block;
  color: var(--gray-900);
  font-weight: 800;
  font-size: 0.9rem;
  text-decoration: none;
}

.name:hover {
  color: var(--brand-600);
}

.nickname {
  margin-top: 0.1rem;
  color: var(--gray-500);
  font-size: 0.78rem;
  font-weight: 600;
  overflow-wrap: anywhere;
}

.view-profile-btn {
  display: inline-flex;
  justify-content: center;
  align-items: center;
  border: none;
  border-radius: 16px;
  padding: 0.72rem 0.95rem;
  font-weight: 800;
  font-size: 0.84rem;
  white-space: nowrap;
  flex-shrink: 0;
  margin-left: auto;
  cursor: pointer;
  transition: opacity 0.15s;
}

.view-profile-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.view-profile-btn--follow {
  color: var(--white);
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  box-shadow: 0 12px 24px rgba(37, 99, 235, 0.18);
}

.view-profile-btn--following {
  color: var(--brand-700, #1d4ed8);
  background: rgba(37, 99, 235, 0.08);
  box-shadow: none;
}

.view-profile-btn--following:hover:not(:disabled) {
  color: #dc2626;
  background: rgba(220, 38, 38, 0.08);
}

.view-profile-btn--requested {
  color: var(--gray-600);
  background: rgba(148, 163, 184, 0.12);
  box-shadow: none;
}

.view-profile-btn--requested:hover:not(:disabled) {
  color: #dc2626;
  background: rgba(220, 38, 38, 0.08);
}

.search-page .search-hero-copy,
.search-page .search-hero-meta,
.search-page .search-hero-meta article,
.search-page .search-form,
.search-page .status-card,
.search-page .result-item {
  box-sizing: border-box;
}

@media (max-width: 1100px) {
  .suggestions-list {
    grid-template-columns: 1fr 1fr;
  }

  .results-list {
    grid-template-columns: 1fr 1fr;
  }
}

@media (max-width: 899px) {
  .search-hero {
    grid-template-columns: 1fr;
  }

  .search-hero h1 {
    max-width: 13ch;
  }
}

@media (max-width: 640px) {
  .search-form {
    grid-template-columns: 1fr;
  }

  .suggestions-list {
    grid-template-columns: 1fr;
  }

  .results-list {
    grid-template-columns: 1fr;
  }

  .search-hero,
  .search-form,
  .status-card,
  .result-item {
    border-radius: 24px;
  }
}
</style>
