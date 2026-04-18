<template>
  <section class="groups-shell">
    <header class="groups-topbar">
      <div class="groups-hero-copy">
        <p class="eyebrow">Community</p>
        <h1>Surround yourself with soldiers.</h1>
        <p class="hero-subtitle">Discover communities, manage invitations, and jump into group chats and events from one place.</p>
        <div class="hero-actions">
          <router-link class="primary-link" to="/groups/new">+ Create group</router-link>
        </div>
      </div>
      <div class="groups-hero-meta">
        <article>
          <strong>{{ groups.length }}</strong>
          <span>Total groups</span>
        </article>
        <article>
          <strong>{{ memberGroupsCount }}</strong>
          <span>Your memberships</span>
        </article>
        <article>
          <strong>{{ pendingWork }}</strong>
          <span>Pending actions</span>
        </article>
      </div>
    </header>

    <section class="groups-search-bar">
      <label class="groups-search-input-wrap" for="groups-search">
        <svg class="groups-search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9" aria-hidden="true">
          <circle cx="11" cy="11" r="7"></circle>
          <path d="m20 20-3.5-3.5"></path>
        </svg>
        <input
          id="groups-search"
          v-model.trim="searchTerm"
          class="groups-search-input"
          type="search"
          placeholder="Search groups, creators, or descriptions"
        />
      </label>
      <button v-if="searchTerm" class="groups-search-clear" type="button" @click="searchTerm = ''">
        Clear
      </button>
    </section>

    <section class="groups-layout">
      <aside class="left-rail">
        <article class="rail-card">
          <div class="rail-head">
            <span class="section-kicker">Overview</span>
            <h2>Workspace snapshot</h2>
          </div>
          <ul class="stat-list">
            <li>
              <span>Total groups</span>
              <strong>{{ groups.length }}</strong>
            </li>
            <li>
              <span>Your memberships</span>
              <strong>{{ memberGroupsCount }}</strong>
            </li>
            <li>
              <span>Pending actions</span>
              <strong>{{ pendingWork }}</strong>
            </li>
          </ul>
        </article>

        <article class="rail-card discover-card">
          <div class="discover-head">
            <div class="rail-head compact">
              <span class="section-kicker">Discover</span>
              <h2>Find your next group</h2>
            </div>
          </div>
          <div v-if="filteredDiscoverGroups.length === 0" class="discover-empty empty-state">
            <img src="@/assets/empty-states/groups-empty.svg" alt="" class="empty-state-img" />
            <p>No discoverable groups right now.</p>
          </div>
          <ul v-else class="discover-list">
            <li v-for="group in filteredDiscoverGroups" :key="group.id">
              <div class="discover-item-body">
                <div class="discover-title-row">
                  <strong>{{ group.title }}</strong>
                  <span v-if="isFollowedCreatorGroup(group)" class="discover-badge">Follows creator</span>
                </div>
                <p>{{ group.membersCount }} members</p>
              </div>
              <button
                v-if="!group.hasPendingRequest && !group.hasPendingInvite"
                class="request-button"
                :disabled="requestingGroupId === group.id"
                @click="handleRequestJoin(group.id)"
              >
                {{ requestingGroupId === group.id ? 'Sending...' : 'Request' }}
              </button>
              <button
                v-else-if="group.hasPendingRequest"
                class="request-button pending-button"
                type="button"
                disabled
              >
                Pending
              </button>
              <button
                v-else-if="group.hasPendingInvite"
                class="request-button pending-button"
                type="button"
                disabled
              >
                Invite sent
              </button>
              <div
                v-else
                class="request-actions"
              >
                <button
                  class="request-button pending-button"
                  type="button"
                  disabled
                  aria-disabled="true"
                >
                  Ask to join
                </button>
              </div>
            </li>
          </ul>
        </article>
      </aside>

      <main class="feed-column">
        <div v-if="loading" class="status-card"><p>Loading groups...</p></div>
        <div v-else-if="error" class="status-card"><p class="message error">{{ error }}</p></div>
        <div v-else-if="feedback" class="status-card compact"><p class="message success">{{ feedback }}</p></div>

        <section v-else class="group-feed">
          <article
            v-for="group in pendingRequestGroups"
            :key="`${group.id}-request-row`"
            class="invite-row-card request-row-card"
          >
            <div class="invite-row-copy">
              <strong>{{ group.title }}</strong>
              <p>Your join request is pending for this group.</p>
            </div>
            <div class="request-actions invite-row-actions">
              <span class="request-row-pill">Pending</span>
            </div>
          </article>

          <article
            v-for="group in invitationGroups"
            :key="`${group.id}-invite-row`"
            class="invite-row-card"
          >
            <div class="invite-row-copy">
              <strong>{{ group.title }}</strong>
              <p>You received an invitation to join this group.</p>
            </div>
            <div class="request-actions invite-row-actions">
              <button
                class="request-button"
                :disabled="respondingInviteGroupId === group.id || !pendingInvitesByGroup[group.id]"
                @click="handleInviteResponse(group.id, 'accepted')"
              >
                {{ respondingInviteGroupId === group.id ? 'Saving...' : 'Accept' }}
              </button>
              <button
                class="request-button decline-button"
                :disabled="respondingInviteGroupId === group.id || !pendingInvitesByGroup[group.id]"
                @click="handleInviteResponse(group.id, 'declined')"
              >
                Decline
              </button>
            </div>
          </article>

          <article v-for="group in filteredMemberGroups" :key="group.id" class="group-item">
            <div class="group-head">
              <div class="avatar-dot" aria-hidden="true">{{ group.title.slice(0, 1).toUpperCase() }}</div>
              <div>
                <h3>{{ group.title }}</h3>
                <p class="meta">Created by {{ group.creatorName }}</p>
              </div>
              <span class="members-pill">{{ group.membersCount }} members</span>
            </div>

            <p class="description">{{ group.description || 'No description yet.' }}</p>

            <div class="tag-row">
              <span :class="['status-chip', group.isMember ? 'member' : 'guest']">
                {{ getGroupMembershipLabel(group) }}
              </span>
              <span v-if="group.hasPendingInvite" class="status-chip warning">Invite pending</span>
            </div>

            <div class="actions-row">
              <router-link class="action-link" :to="`/groups/${group.id}`">Open group</router-link>
              <router-link class="action-link" :to="`/groups/${group.id}/events`">View events</router-link>
              <router-link class="action-link" :to="`/chats/groups/${group.id}?name=${encodeURIComponent(group.title)}`">
                Open chat
              </router-link>
            </div>

            <div
              v-if="creatorPendingRequests(group.id).length > 0"
              class="group-request-panel"
            >
              <div class="group-request-head">
                <strong>Join requests</strong>
                <span class="request-row-pill">{{ creatorPendingRequests(group.id).length }} pending</span>
              </div>
              <ul class="group-request-list">
                <li
                  v-for="request in creatorPendingRequests(group.id)"
                  :key="request.id"
                  class="group-request-item"
                >
                  <span>{{ requestDisplayNames[request.userId] || request.userId }}</span>
                  <div class="request-actions group-request-actions">
                    <button
                      class="request-button"
                      :disabled="respondingJoinRequestId === request.id"
                      @click="handleJoinRequestResponse(request.id, 'accepted')"
                    >
                      {{ respondingJoinRequestId === request.id ? 'Saving...' : 'Accept' }}
                    </button>
                    <button
                      class="request-button decline-button"
                      :disabled="respondingJoinRequestId === request.id"
                      @click="handleJoinRequestResponse(request.id, 'declined')"
                    >
                      Decline
                    </button>
                  </div>
                </li>
              </ul>
            </div>
          </article>
          <div v-if="filteredMemberGroups.length === 0" class="status-card empty-state">
            <img src="@/assets/empty-states/groups-empty.svg" alt="" class="empty-state-img" />
            <p>You are not a member of any group yet. Use Discover to request access.</p>
          </div>
        </section>
      </main>
    </section>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

import { apiURL } from '@/api/api'
import { buildWebSocketURL } from '@/api/websocket'
import {
  createJoinRequest,
  listGroupInvites,
  listGroups,
  listJoinRequests,
  respondToInvite,
  respondToJoinRequest,
} from '@/api/groups'
import { fetchSessionData } from '@/router'
import type { GroupInvite, GroupJoinRequest, GroupSummary } from '@/types/groups'
import type { SessionData, User } from '@/types/User'

const groups = ref<GroupSummary[]>([])
const loading = ref(false)
const error = ref('')
const feedback = ref('')
const requestingGroupId = ref('')
const respondingInviteGroupId = ref('')
const searchTerm = ref('')
const currentUserId = ref('')
const pendingJoinRequestCounts = ref<Record<string, number>>({})
const pendingInvitesByGroup = ref<Record<string, GroupInvite>>({})
const pendingJoinRequestsByGroup = ref<Record<string, GroupJoinRequest[]>>({})
const requestDisplayNames = ref<Record<string, string>>({})
const respondingJoinRequestId = ref('')
const followingIds = ref<string[]>([])
let socket: WebSocket | null = null
let reconnectTimerId: number | null = null
let socketClosedManually = false
let groupReloadInFlight = false

const memberGroups = computed(() => groups.value.filter((group) => group.isMember))
const memberGroupsCount = computed(() => memberGroups.value.length)
const discoverGroups = computed(() =>
  groups.value.filter((group) => !group.isMember && group.creatorId !== currentUserId.value),
)
const invitationGroups = computed(() =>
  groups.value.filter(
    (group) =>
      !group.isMember &&
      group.hasPendingInvite &&
      group.creatorId !== currentUserId.value &&
      Boolean(pendingInvitesByGroup.value[group.id]),
  ),
)
const pendingRequestGroups = computed(() =>
  groups.value.filter(
    (group) => !group.isMember && group.hasPendingRequest && group.creatorId !== currentUserId.value,
  ),
)
const pendingWork = computed(
  () =>
    groups.value.filter((group) => group.hasPendingInvite).length +
    Object.values(pendingJoinRequestCounts.value).reduce((total, count) => total + count, 0),
)
const normalizedSearch = computed(() => searchTerm.value.toLowerCase())
const matchesSearch = (group: GroupSummary) =>
  group.title.toLowerCase().includes(normalizedSearch.value) ||
  group.description.toLowerCase().includes(normalizedSearch.value) ||
  group.creatorName.toLowerCase().includes(normalizedSearch.value)
const filteredMemberGroups = computed(() =>
  (
    !normalizedSearch.value ? memberGroups.value : memberGroups.value.filter(matchesSearch)
  )
    .slice()
    .sort((left, right) => {
      const leftHasPending = (pendingJoinRequestCounts.value[left.id] ?? 0) > 0
      const rightHasPending = (pendingJoinRequestCounts.value[right.id] ?? 0) > 0
      if (leftHasPending === rightHasPending) return 0
      return leftHasPending ? -1 : 1
    }),
)
const filteredDiscoverGroups = computed(() =>
  (
    !normalizedSearch.value ? discoverGroups.value : discoverGroups.value.filter(matchesSearch)
  )
    .slice()
    .sort((left, right) => {
      const leftFollowed = followingIds.value.includes(left.creatorId)
      const rightFollowed = followingIds.value.includes(right.creatorId)
      if (leftFollowed !== rightFollowed) {
        return leftFollowed ? -1 : 1
      }
      if (left.hasPendingInvite === right.hasPendingInvite) return 0
      return left.hasPendingInvite ? -1 : 1
    }),
)

const creatorPendingRequests = (groupId: string) => pendingJoinRequestsByGroup.value[groupId] ?? []
const isFollowedCreatorGroup = (group: GroupSummary) => followingIds.value.includes(group.creatorId)
const getGroupMembershipLabel = (group: GroupSummary) => {
  if (!group.isMember) return 'Not joined'
  return group.creatorId === currentUserId.value ? 'Creator' : 'Member'
}

function isGroupSummaryPayload(payload: unknown): payload is GroupSummary {
  if (!payload || typeof payload !== 'object') return false
  const group = payload as Partial<GroupSummary>
  return typeof group.id === 'string' && typeof group.creatorId === 'string' && typeof group.title === 'string'
}

function isGroupStateRefreshPayload(payload: unknown): payload is { groupId: string } {
  if (!payload || typeof payload !== 'object') return false
  const item = payload as Partial<{ groupId: string }>
  return typeof item.groupId === 'string'
}

function formatUserDisplayName(user: Pick<User, 'id' | 'nickname' | 'firstName' | 'lastName'>) {
  if (user.nickname?.trim()) return `@${user.nickname.trim()}`
  const fullName = `${user.firstName || ''} ${user.lastName || ''}`.trim()
  return fullName || user.id
}

async function loadUserDisplayNames(userIDs: string[]) {
  const entries = await Promise.allSettled(
    userIDs.map(async (userID) => {
      const response = await fetch(apiURL(`/api/users/${encodeURIComponent(userID)}`), {
        credentials: 'include',
      })
      if (!response.ok) {
        throw new Error(`Failed to load user ${userID}`)
      }

      const user = (await response.json()) as User
      return [userID, formatUserDisplayName(user)] as const
    }),
  )

  const names: Record<string, string> = {}
  entries.forEach((entry) => {
    if (entry.status === 'fulfilled') {
      const [userID, label] = entry.value
      names[userID] = label
    }
  })

  return names
}

async function loadPendingJoinRequestCounts(items: GroupSummary[]) {
  const creatorGroups = items.filter(
    (group) => group.isMember && group.creatorId === currentUserId.value,
  )

  if (creatorGroups.length === 0) {
    pendingJoinRequestCounts.value = {}
    pendingJoinRequestsByGroup.value = {}
    requestDisplayNames.value = {}
    return
  }

  const counts: Record<string, number> = {}
  const requestsByGroup: Record<string, GroupJoinRequest[]> = {}
  const results = await Promise.allSettled(
    creatorGroups.map(async (group) => {
      const requests = await listJoinRequests(group.id)
      const pendingRequests = requests.filter((request) => request.status === 'pending')
      return {
        groupId: group.id,
        requests: pendingRequests,
        count: pendingRequests.length,
      }
    }),
  )

  const requestUserIDs = new Set<string>()
  results.forEach((result) => {
    if (result.status === 'fulfilled' && result.value.count > 0) {
      counts[result.value.groupId] = result.value.count
      requestsByGroup[result.value.groupId] = result.value.requests
      result.value.requests.forEach((request) => requestUserIDs.add(request.userId))
    }
  })

  pendingJoinRequestCounts.value = counts
  pendingJoinRequestsByGroup.value = requestsByGroup
  requestDisplayNames.value =
    requestUserIDs.size > 0 ? await loadUserDisplayNames([...requestUserIDs]) : {}
}

async function loadPendingInvites(items: GroupSummary[]) {
  const invitedGroups = items.filter((group) => group.hasPendingInvite)

  if (invitedGroups.length === 0) {
    pendingInvitesByGroup.value = {}
    return
  }

  const inviteEntries = await Promise.allSettled(
    invitedGroups.map(async (group) => {
      const invites = await listGroupInvites(group.id)
      const mine = invites.find(
        (invite) => invite.receiverId === currentUserId.value && invite.status === 'pending',
      )
      return mine ? ([group.id, mine] as const) : null
    }),
  )

  const mapped: Record<string, GroupInvite> = {}
  inviteEntries.forEach((entry) => {
    if (entry.status === 'fulfilled' && entry.value) {
      const [groupId, invite] = entry.value
      mapped[groupId] = invite
    }
  })

  pendingInvitesByGroup.value = mapped
}

async function loadFollowing() {
  if (!currentUserId.value) {
    followingIds.value = []
    return
  }

  try {
    const response = await fetch(apiURL(`/api/users/${encodeURIComponent(currentUserId.value)}/following`), {
      credentials: 'include',
    })
    if (!response.ok) {
      followingIds.value = []
      return
    }

    const following = (await response.json()) as Array<{ id: string }>
    followingIds.value = following.map((user) => user.id)
  } catch {
    followingIds.value = []
  }
}

async function loadGroups() {
  loading.value = true
  error.value = ''
  feedback.value = ''

  try {
    const loadedGroups = await listGroups()
    groups.value = loadedGroups
    await Promise.all([loadPendingJoinRequestCounts(loadedGroups), loadPendingInvites(loadedGroups)])
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load groups.'
  } finally {
    loading.value = false
  }
}

async function reloadGroupsWorkspace() {
  if (groupReloadInFlight) {
    return
  }

  groupReloadInFlight = true
  try {
    await loadGroups()
  } finally {
    groupReloadInFlight = false
  }
}

async function handleInviteResponse(groupId: string, status: 'accepted' | 'declined') {
  const invite = pendingInvitesByGroup.value[groupId]
  if (!invite) {
    error.value = 'Pending invite not found.'
    return
  }

  respondingInviteGroupId.value = groupId
  error.value = ''
  feedback.value = ''

  try {
    await respondToInvite(invite.id, status)
    feedback.value = status === 'accepted' ? 'Invite accepted.' : 'Invite declined.'
    await loadGroups()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to respond to invite.'
  } finally {
    respondingInviteGroupId.value = ''
  }
}

async function handleJoinRequestResponse(requestId: string, status: 'accepted' | 'declined') {
  respondingJoinRequestId.value = requestId
  error.value = ''
  feedback.value = ''

  try {
    await respondToJoinRequest(requestId, status)
    feedback.value = status === 'accepted' ? 'Join request accepted.' : 'Join request declined.'
    await loadGroups()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to respond to join request.'
  } finally {
    respondingJoinRequestId.value = ''
  }
}

async function handleRequestJoin(groupId: string) {
  requestingGroupId.value = groupId
  error.value = ''
  feedback.value = ''

  try {
    await createJoinRequest(groupId)
    feedback.value = 'Join request sent.'
    await loadGroups()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to send join request.'
  } finally {
    requestingGroupId.value = ''
  }
}

function scheduleSocketReconnect() {
  if (socketClosedManually || reconnectTimerId !== null) {
    return
  }

  reconnectTimerId = window.setTimeout(() => {
    reconnectTimerId = null
    connectRealtimeSocket()
  }, 1500)
}

function handleRealtimeMessage(raw: { type?: string; payload?: unknown }) {
  switch (raw.type) {
    case 'group_summary_event': {
      const payload = raw.payload
      if (!isGroupSummaryPayload(payload)) {
        return
      }

      groups.value = groups.value.map((group) => (group.id === payload.id ? payload : group))
      return
    }
    case 'group_state_refresh':
      if (isGroupStateRefreshPayload(raw.payload)) {
        void reloadGroupsWorkspace()
      }
      return
  }
}

function connectRealtimeSocket() {
  if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
    return
  }

  socketClosedManually = false
  socket = new WebSocket(buildWebSocketURL('/ws'))

  socket.onmessage = (event) => {
    try {
      const raw = JSON.parse(event.data) as { type?: string; payload?: unknown }
      handleRealtimeMessage(raw)
    } catch {
      // Ignore invalid websocket payloads.
    }
  }

  socket.onclose = () => {
    socket = null
    scheduleSocketReconnect()
  }

  socket.onerror = () => {
    socket?.close()
  }
}

function disconnectRealtimeSocket() {
  socketClosedManually = true
  if (reconnectTimerId !== null) {
    window.clearTimeout(reconnectTimerId)
    reconnectTimerId = null
  }
  if (socket) {
    socket.close()
    socket = null
  }
}

onMounted(async () => {
  const sessionData: SessionData | null = await fetchSessionData()
  currentUserId.value = sessionData?.user.id ?? ''
  connectRealtimeSocket()
  await Promise.all([loadFollowing(), loadGroups()])
})

onBeforeUnmount(() => {
  disconnectRealtimeSocket()
})
</script>

<style>
.groups-shell {
  padding: 0.25rem 0 1.75rem;
  width: min(100%, 1320px);
  margin: 0 auto;
  display: grid;
  gap: 1rem;
  overflow: hidden;
}

.groups-topbar {
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

.groups-topbar::before,
.groups-topbar::after {
  content: '';
  position: absolute;
  border-radius: 999px;
  pointer-events: none;
}

.groups-topbar::before {
  width: 220px;
  height: 220px;
  right: -40px;
  top: -60px;
  background: rgba(255, 255, 255, 0.08);
}

.groups-topbar::after {
  width: 160px;
  height: 160px;
  left: -40px;
  bottom: -60px;
  background: rgba(125, 211, 252, 0.12);
}

.groups-hero-copy,
.groups-hero-meta,
.hero-actions {
  position: relative;
  z-index: 1;
}

.groups-hero-copy {
  min-width: 0;
  overflow: hidden;
  display: grid;
  align-content: start;
  gap: 0.75rem;
}

.hero-actions {
  display: flex;
  justify-content: flex-start;
  align-items: flex-start;
  flex-wrap: wrap;
}

.eyebrow {
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
  text-transform: uppercase;
  letter-spacing: 0.12em;
}

.groups-topbar h1 {
  margin-top: 1rem;
  max-width: 11ch;
  color: var(--white);
  font-size: clamp(2.2rem, 5vw, 3.5rem);
  line-height: 0.96;
  font-weight: 900;
  letter-spacing: -0.04em;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  overflow-wrap: break-word;
  word-break: break-word;
  min-width: 0;
}

.hero-subtitle {
  margin-top: 0.9rem;
  max-width: 44ch;
  color: rgba(255, 255, 255, 0.82);
  font-size: 1rem;
  line-height: 1.75;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  overflow-wrap: break-word;
  word-break: break-word;
}

.primary-link {
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 46px;
  font-weight: 800;
  font-size: 0.92rem;
  border-radius: 18px;
  padding: 0.88rem 1.15rem;
  color: var(--white);
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.95), #1e40af);
  box-shadow: 0 16px 32px rgba(37, 99, 235, 0.24);
  white-space: nowrap;
}

.groups-hero-meta {
  display: grid;
  gap: 0.85rem;
  align-content: end;
  min-width: 0;
}

.groups-hero-meta article {
  padding: 1rem 1.05rem;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.12);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.08);
  min-width: 0;
  overflow: hidden;
}

.groups-hero-meta strong {
  display: block;
  color: var(--white);
  font-size: 1.55rem;
  font-weight: 900;
}

.groups-hero-meta span {
  color: rgba(255, 255, 255, 0.76);
  font-size: 0.86rem;
  font-weight: 700;
}

.groups-layout {
  display: grid;
  grid-template-columns: minmax(0, 240px) minmax(0, 1fr);
  gap: 0.85rem;
  align-items: start;
}

.groups-search-bar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 0.75rem;
  align-items: center;
  padding: 0.75rem;
  border-radius: 24px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(239, 246, 255, 0.76));
  border: 1px solid rgba(148, 163, 184, 0.16);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 12px 28px rgba(15, 23, 42, 0.05);
}

.groups-search-input-wrap {
  display: flex;
  align-items: center;
  gap: 0.72rem;
  min-width: 0;
  padding: 0.95rem 1rem;
  border-radius: 18px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 8px 22px rgba(15, 23, 42, 0.04);
}

.groups-search-icon {
  width: 1.1rem;
  height: 1.1rem;
  flex-shrink: 0;
  color: var(--gray-500);
}

.groups-search-input {
  width: 100%;
  min-width: 0;
  border: none;
  outline: none;
  background: transparent;
  color: var(--gray-800);
  font: inherit;
  font-size: 0.95rem;
  font-weight: 600;
}

.groups-search-input-wrap:focus-within {
  border-color: rgba(37, 99, 235, 0.55);
  box-shadow:
    0 0 0 5px rgba(37, 99, 235, 0.1),
    0 16px 34px rgba(37, 99, 235, 0.12);
}

.groups-search-clear {
  border: 1px solid rgba(148, 163, 184, 0.16);
  padding: 0.88rem 1.15rem;
  border-radius: 18px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(239, 246, 255, 0.92));
  color: var(--gray-700);
  font-size: 0.9rem;
  font-weight: 800;
  cursor: pointer;
}

.left-rail {
  position: sticky;
  top: 1rem;
  min-width: 0;
  overflow: hidden;
}

.rail-card {
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 1.25rem;
  padding: 0.85rem;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(248, 250, 252, 0.86));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 10px 28px rgba(15, 23, 42, 0.06);
  min-width: 0;
  overflow: hidden;
}

.rail-head {
  display: grid;
  gap: 0.35rem;
  margin-bottom: 0.6rem;
}

.rail-head.compact {
  margin-bottom: 0;
}

.rail-head h2 {
  color: var(--gray-900);
  font-size: 1.05rem;
  line-height: 1.1;
  font-weight: 900;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.section-kicker {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  padding: 0.28rem 0.6rem;
  border-radius: 999px;
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.14), rgba(14, 165, 233, 0.08));
  color: var(--brand-700);
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.stat-list {
  list-style: none;
  display: grid;
  gap: 0.45rem;
  padding: 0;
  margin: 0;
}

.stat-list li {
  display: grid;
  grid-template-columns: 1fr auto;
  align-items: center;
  gap: 0.4rem;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 12px;
  padding: 0.6rem 0.7rem;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.92));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 6px 16px rgba(15, 23, 42, 0.04);
  min-width: 0;
}

.stat-list strong {
  color: var(--gray-900);
  font-size: 1rem;
  font-weight: 900;
  justify-self: end;
  text-align: right;
}

.stat-list span {
  color: var(--gray-500);
  font-size: 0.76rem;
  font-weight: 700;
  justify-self: start;
  text-align: left;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.feed-column {
  display: grid;
  gap: 0.75rem;
  min-width: 0;
}

.discover-card {
  margin-top: 0.55rem;
}

.discover-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.45rem;
  margin-bottom: 0.55rem;
}

.search-toggle {
  border: 1px solid rgba(148, 163, 184, 0.16);
  background: rgba(255, 255, 255, 0.9);
  color: var(--brand-700);
  border-radius: 12px;
  padding: 0.3rem 0.5rem;
  font-size: 0.72rem;
  font-weight: 800;
  cursor: pointer;
  box-shadow: 0 6px 16px rgba(15, 23, 42, 0.04);
  flex-shrink: 0;
}

.search-toggle.icon-only {
  width: 34px;
  height: 34px;
  padding: 0;
  display: grid;
  place-items: center;
}

.search-toggle.icon-only :deep(svg) {
  width: 14px;
  height: 14px;
  stroke-width: 2.2;
}

.search-wrap {
  max-height: 0;
  opacity: 0;
  overflow: hidden;
  transition: max-height 220ms ease, opacity 200ms ease, margin-bottom 220ms ease;
  margin-bottom: 0;
}

.search-wrap.open {
  max-height: 60px;
  opacity: 1;
  margin-bottom: 0.5rem;
}

.search-input {
  width: 100%;
  border: 1px solid rgba(148, 163, 184, 0.2);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  border-radius: 12px;
  padding: 0.55rem 0.7rem;
  font-size: 0.82rem;
  color: var(--gray-900);
  box-sizing: border-box;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 4px 14px rgba(15, 23, 42, 0.03);
}

.search-input:focus {
  outline: none;
  border-color: rgba(37, 99, 235, 0.55);
  box-shadow:
    0 0 0 3px rgba(37, 99, 235, 0.1),
    0 8px 20px rgba(37, 99, 235, 0.08);
}

.discover-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 0.5rem;
  max-height: calc((72px * 3) + (0.5rem * 2));
  overflow-y: auto;
  padding-right: 0.15rem;
}

.discover-empty {
  min-height: 140px;
  border: 1px dashed rgba(148, 163, 184, 0.28);
  border-radius: 1rem;
  padding: 0.85rem;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.9), rgba(248, 250, 252, 0.82));
  color: var(--gray-500);
  font-size: 0.82rem;
  display: grid;
  place-items: center;
  text-align: center;
}

.discover-list li {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.5rem;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 14px;
  padding: 0.6rem 0.7rem;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.88));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 8px 20px rgba(15, 23, 42, 0.04);
  min-width: 0;
  transition:
    transform var(--dur-base) var(--ease-standard),
    border-color var(--dur-base) var(--ease-standard),
    box-shadow var(--dur-base) var(--ease-standard);
}

.discover-list li:hover {
  transform: translateY(-1px);
  border-color: rgba(96, 165, 250, 0.22);
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.1);
}

.discover-item-body {
  display: grid;
  gap: 0.2rem;
  min-width: 0;
  overflow: hidden;
}

.discover-title-row {
  display: flex;
  gap: 0.35rem;
  align-items: center;
  flex-wrap: wrap;
  min-width: 0;
}

.discover-badge {
  border-radius: 999px;
  padding: 0.15rem 0.4rem;
  background: rgba(219, 234, 254, 0.9);
  color: #1d4ed8;
  font-size: 0.6rem;
  font-weight: 800;
  flex-shrink: 0;
}

.discover-list::-webkit-scrollbar {
  width: 6px;
}

.discover-list::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 999px;
}

.discover-list strong {
  color: var(--gray-900);
  font-size: 0.84rem;
  font-weight: 800;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.discover-list p {
  color: var(--gray-500);
  font-size: 0.72rem;
  font-weight: 600;
}

.request-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  color: var(--white);
  border-radius: 12px;
  padding: 0.45rem 0.7rem;
  font-size: 0.76rem;
  font-weight: 800;
  cursor: pointer;
  box-shadow: 0 8px 18px rgba(37, 99, 235, 0.14);
  white-space: nowrap;
  flex-shrink: 0;
}

.request-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  box-shadow: none;
}

.request-actions {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  flex-wrap: wrap;
  margin-left: auto;
  justify-content: flex-end;
  align-self: center;
}

.pending-button {
  color: var(--gray-500);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  cursor: default;
  border: 1px solid rgba(148, 163, 184, 0.16);
}

.pending-button:disabled {
  cursor: default;
}

.decline-button {
  color: #8a3b12;
  background: linear-gradient(180deg, rgba(255, 247, 237, 0.98), rgba(255, 237, 213, 0.9));
  border: 1px solid rgba(251, 146, 60, 0.25);
  box-shadow: none;
}

.group-feed {
  display: grid;
  gap: 0.75rem;
}

.invite-row-card {
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.88));
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 1.25rem;
  padding: 0.85rem;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 10px 28px rgba(15, 23, 42, 0.06);
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.65rem;
  min-width: 0;
}

.request-row-card {
  border-color: rgba(251, 146, 60, 0.24);
  background:
    linear-gradient(180deg, rgba(255, 247, 237, 0.96), rgba(255, 255, 255, 0.92));
}

.invite-row-copy {
  display: grid;
  gap: 0.15rem;
  min-width: 0;
  overflow: hidden;
}

.invite-row-copy strong {
  color: var(--gray-900);
  font-size: 0.88rem;
  font-weight: 800;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.invite-row-copy p {
  color: var(--gray-500);
  font-size: 0.78rem;
  font-weight: 600;
}

.invite-row-actions {
  margin-left: 0;
  flex-shrink: 0;
}

.request-row-pill {
  border-radius: 999px;
  padding: 0.25rem 0.55rem;
  background: #fff1e8;
  color: #c2410c;
  border: 1px solid #fdba74;
  font-size: 0.72rem;
  font-weight: 800;
  white-space: nowrap;
}

.group-item {
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.88));
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 1.25rem;
  padding: 0.85rem;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 10px 28px rgba(15, 23, 42, 0.06);
  display: grid;
  gap: 0.6rem;
  min-width: 0;
  overflow: hidden;
  transition:
    transform var(--dur-base) var(--ease-standard),
    border-color var(--dur-base) var(--ease-standard),
    box-shadow var(--dur-base) var(--ease-standard);
}

.group-item:hover {
  transform: translateY(-1px);
  border-color: rgba(96, 165, 250, 0.22);
  box-shadow: 0 14px 36px rgba(15, 23, 42, 0.1);
}

.group-request-panel {
  border-top: 1px solid rgba(148, 163, 184, 0.16);
  padding-top: 0.65rem;
  display: grid;
  gap: 0.5rem;
}

.group-request-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.45rem;
  flex-wrap: wrap;
}

.group-request-head strong {
  color: var(--gray-900);
  font-size: 0.85rem;
  font-weight: 800;
}

.group-request-list {
  list-style: none;
  display: grid;
  gap: 0.5rem;
  padding: 0;
  margin: 0;
}

.group-request-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.5rem;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 14px;
  padding: 0.55rem 0.65rem;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.92));
  min-width: 0;
}

.group-request-item span {
  color: var(--gray-900);
  font-size: 0.8rem;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.group-request-actions {
  margin-left: auto;
  flex-shrink: 0;
}

.group-head,
.actions-row,
.tag-row {
  display: flex;
  justify-content: space-between;
  gap: 0.5rem;
  align-items: center;
  flex-wrap: wrap;
  min-width: 0;
}

.group-head > div {
  min-width: 0;
  overflow: hidden;
  flex: 1;
}

.group-head h3 {
  color: var(--gray-900);
  font-size: 0.95rem;
  line-height: 1.2;
  font-weight: 800;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.avatar-dot {
  max-width: 38px;
  height: 38px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  font-weight: 800;
  font-size: 0.88rem;
  color: var(--brand-700);
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.16), rgba(14, 165, 233, 0.08));
  flex-shrink: 0;
}

.meta {
  color: var(--gray-500);
  font-size: 0.76rem;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.members-pill {
  border-radius: 999px;
  padding: 0.22rem 0.55rem;
  background: rgba(241, 245, 249, 0.92);
  color: var(--gray-700);
  font-size: 0.7rem;
  font-weight: 800;
  white-space: nowrap;
  flex-shrink: 0;
}

.description {
  color: var(--gray-700);
  min-height: 1.8rem;
  font-size: 0.85rem;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  overflow-wrap: break-word;
  word-break: break-word;
  min-width: 0;
}

.status-chip {
  border-radius: 999px;
  padding: 0.2rem 0.55rem;
  font-size: 0.68rem;
  font-weight: 800;
  background: #edf2f7;
  color: #334155;
  white-space: nowrap;
}

.status-chip.member {
  background: #e8f8ed;
  color: #166534;
}

.status-chip.guest {
  background: #eef4fb;
  color: #32506f;
}

.status-chip.warning {
  background: var(--warn-soft);
  color: #92400e;
}

.action-link {
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 800;
  border-radius: 12px;
  padding: 0.5rem 0.7rem;
  font-size: 0.8rem;
  border: 1px solid rgba(148, 163, 184, 0.16);
  color: var(--gray-700);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  box-shadow: 0 6px 16px rgba(15, 23, 42, 0.04);
  white-space: nowrap;
}

.status-card {
  min-height: 160px;
  border-radius: 1.25rem;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.9), rgba(248, 250, 252, 0.82));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 10px 28px rgba(15, 23, 42, 0.05);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  text-align: center;
  padding: 1.5rem 0.85rem;
}

.status-card.compact {
  min-height: 90px;
}

.message {
  margin: 0;
  color: var(--gray-600);
  font-weight: 700;
  font-size: 0.88rem;
}

.message.error {
  color: #9f1239;
}

.message.success {
  color: #166534;
}

.empty-state-img {
  width: clamp(64px, 10vw, 100px);
  height: auto;
  opacity: 0.92;
}

@media (max-width: 900px) {
  .groups-topbar {
    grid-template-columns: 1fr;
    padding: 1rem;
  }

  .groups-layout {
    grid-template-columns: 1fr;
  }

  .left-rail {
    position: static;
  }

  .hero-actions {
    justify-content: flex-start;
  }
}

@media (max-width: 640px) {
  .groups-shell {
    gap: 0.65rem;
  }

  .groups-search-bar {
    grid-template-columns: 1fr;
  }

  .groups-topbar {
    padding: 0.85rem;
    border-radius: 1rem;
  }

  .invite-row-card {
    flex-direction: column;
    align-items: flex-start;
  }

  .group-request-item {
    flex-direction: column;
    align-items: flex-start;
  }

  .actions-row {
    gap: 0.35rem;
  }

  .action-link {
    padding: 0.4rem 0.55rem;
    font-size: 0.74rem;
  }

  .rail-card,
  .status-card,
  .invite-row-card,
  .group-item {
    border-radius: 1rem;
  }
}

@media (max-width: 400px) {
  .groups-topbar {
    padding: 0.7rem;
  }

  .groups-topbar h1 {
    font-size: 1.05rem;
  }

  .hero-subtitle {
    font-size: 0.74rem;
  }

  .groups-hero-meta strong {
    font-size: 1rem;
  }

  .groups-hero-meta span {
    font-size: 0.68rem;
  }

  .actions-row {
    flex-direction: column;
    align-items: stretch;
  }

  .action-link {
    justify-content: center;
  }
}
</style>
