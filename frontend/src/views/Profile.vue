<template>
  <main class="profile-page">
    <section class="profile-card">
      <div v-if="isLoading" class="status-message">Loading profile...</div>
      <div v-else-if="errorMessage" class="status-message error">{{ errorMessage }}</div>

      <template v-else-if="profile">
        <div class="profile-body">
        <header class="profile-header">
          <div class="avatar-wrap">
            <img v-if="profile.avatar" :src="avatarSrc" alt="Profile avatar" />
            <div v-else class="avatar-fallback">
              {{ initials }}
            </div>
          </div>

          <div class="identity">
            <h1>{{ fullName }}</h1>
            <p class="nickname" v-if="profile.nickname">@{{ profile.nickname }}</p>
            <p class="privacy-tag" :class="{ private: !profile.isPublic }">
              {{ profile.isPublic ? 'Public profile' : 'Private profile' }}
            </p>
            <div v-if="!isOwnProfile" class="follow-controls">
              <div v-if="incomingFromViewedRequestID" class="inline-request-actions">
                <button
                  type="button"
                  class="request-btn accept"
                  :disabled="isIncomingFromViewedLoading"
                  @click="handleIncomingFromViewed('accepted')"
                >
                  Accept request
                </button>
                <button
                  type="button"
                  class="request-btn decline"
                  :disabled="isIncomingFromViewedLoading"
                  @click="handleIncomingFromViewed('declined')"
                >
                  Decline request
                </button>
              </div>

              <button
                v-if="followState === 'following'"
                type="button"
                class="follow-btn unfollow-btn"
                :disabled="isUnfollowLoading"
                @click="handleUnfollow"
              >
                {{ isUnfollowLoading ? 'Please wait...' : 'Unfollow' }}
              </button>
              <button
                v-else-if="outgoingToViewedRequestID"
                type="button"
                class="follow-btn withdraw-btn"
                :disabled="isOutgoingToViewedLoading"
                @click="handleWithdrawFromViewed"
              >
                {{ isOutgoingToViewedLoading ? 'Please wait...' : 'Withdraw request' }}
              </button>
              <p v-else-if="incomingFromViewedRequestID" class="pending-note">
                Respond to the request above.
              </p>
              <button
                v-else
                type="button"
                class="follow-btn"
                :disabled="!canSendFollow"
                @click="handleFollow"
              >
                {{ followButtonLabel }}
              </button>
            </div>

            <RouterLink
              v-if="canMessage"
              class="message-btn"
              :to="`/chats/${profile!.id}`"
            >
              Message
            </RouterLink>
          </div>

          <div v-if="isOwnProfile" class="header-side">
            <div class="privacy-toggle-row header-privacy">
              <div class="privacy-toggle-copy">
                <p>Account privacy</p>
                <small>{{ isUpdatingPrivacy ? 'Saving...' : profile.isPublic ? 'Public profile' : 'Private profile' }}</small>
              </div>
              <label class="privacy-switch" aria-label="Toggle account privacy">
                <input
                  type="checkbox"
                  :checked="profile.isPublic"
                  :disabled="isUpdatingPrivacy"
                  @change="togglePrivacy"
                />
                <span class="privacy-slider"></span>
              </label>
            </div>
          </div>
        </header>

        <dl class="profile-grid">
          <div v-if="canViewRestrictedSections">
            <dt>Email</dt>
            <dd>{{ profile.email }}</dd>
          </div>
          <div v-if="canViewRestrictedSections">
            <dt>Date of birth</dt>
            <dd>{{ profile.dateOfBirth }}</dd>
          </div>
          <div class="full-width">
            <dt>About</dt>
            <dd>{{ aboutDisplay }}</dd>
          </div>
        </dl>

        <div v-if="canViewRestrictedSections" class="connections">
          <div class="stat">
            <span class="stat-number">{{ followersCount }}</span>
            <RouterLink class="stat-link" :to="followersPath">Followers</RouterLink>
          </div>
          <div class="stat">
            <span class="stat-number">{{ followingCount }}</span>
            <RouterLink class="stat-link" :to="followingPath">Following</RouterLink>
          </div>
        </div>
        <p v-else class="private-lock-note">This account is private. Follow to view posts and connections.</p>

        <section v-if="canViewRestrictedSections" class="connections-panel">
          <div class="connections-grid">
            <div class="connections-column">
              <div class="connections-header">
                <h2>Followers</h2>
                <RouterLink class="mini-link" :to="followersPath">See all</RouterLink>
              </div>
              <p v-if="followersUsers.length === 0" class="connections-empty">No followers yet.</p>
              <ul v-else class="connections-list">
                <li v-for="user in followersUsers" :key="`f-${user.id}`">
                  <RouterLink class="connections-user" :to="`/users/${user.id}`">
                    {{ user.nickname || `${user.firstName} ${user.lastName}` }}
                  </RouterLink>
                </li>
              </ul>
            </div>

            <div class="connections-column">
              <div class="connections-header">
                <h2>Following</h2>
                <RouterLink class="mini-link" :to="followingPath">See all</RouterLink>
              </div>
              <p v-if="followingUsers.length === 0" class="connections-empty">Not following anyone yet.</p>
              <ul v-else class="connections-list">
                <li v-for="user in followingUsers" :key="`g-${user.id}`">
                  <RouterLink class="connections-user" :to="`/users/${user.id}`">
                    {{ user.nickname || `${user.firstName} ${user.lastName}` }}
                  </RouterLink>
                </li>
              </ul>
            </div>
          </div>
        </section>

        <section v-if="canViewRestrictedSections" class="activity-panel">
          <h2>User activity</h2>
          <ul class="activity-list">
            <li>Joined: {{ joinDate }}</li>
            <li>Posts: {{ displayedPostsCount }}</li>
          </ul>
        </section>

        <section v-if="isOwnProfile" class="requests-panel">
          <h2>Follow requests</h2>
          <p v-if="isRequestsLoading" class="requests-status">Loading follow requests...</p>
          <p v-else-if="requestsErrorMessage" class="requests-status error">{{ requestsErrorMessage }}</p>

          <div v-else class="requests-grid">
            <div class="request-column">
              <h3>Incoming</h3>
              <p v-if="incomingRequests.length === 0" class="request-empty">No pending incoming requests.</p>
              <ul v-else class="request-list">
                <li v-for="request in incomingRequests" :key="request.id" class="request-item">
                  <div class="request-main">
                    <RouterLink class="request-user" :to="`/users/${request.senderId}`">
                      {{ formatRequestUser(request.sender, request.senderId) }}
                    </RouterLink>
                    <span class="request-date">{{ formatDate(request.createdAt) }}</span>
                  </div>
                  <div class="request-actions">
                    <button
                      type="button"
                      class="request-btn accept"
                      :disabled="processingIncomingID === request.id"
                      @click="handleIncomingRequest(request.id, 'accepted')"
                    >
                      Accept
                    </button>
                    <button
                      type="button"
                      class="request-btn decline"
                      :disabled="processingIncomingID === request.id"
                      @click="handleIncomingRequest(request.id, 'declined')"
                    >
                      Decline
                    </button>
                  </div>
                </li>
              </ul>
            </div>

            <div class="request-column">
              <h3>Outgoing</h3>
              <p v-if="outgoingRequests.length === 0" class="request-empty">No pending outgoing requests.</p>
              <ul v-else class="request-list">
                <li v-for="request in outgoingRequests" :key="request.id" class="request-item">
                  <div class="request-main">
                    <RouterLink class="request-user" :to="`/users/${request.receiverId}`">
                      {{ formatRequestUser(request.receiver, request.receiverId) }}
                    </RouterLink>
                    <span class="request-date">{{ formatDate(request.createdAt) }}</span>
                  </div>
                  <div class="request-actions">
                    <button
                      type="button"
                      class="request-btn withdraw"
                      :disabled="processingOutgoingID === request.id"
                      @click="withdrawOutgoingRequest(request.id)"
                    >
                      Withdraw
                    </button>
                  </div>
                </li>
              </ul>
            </div>
          </div>
        </section>

        <section v-if="canViewRestrictedSections" class="posts-panel">
          <h2>Posts</h2>
          <p v-if="isPostsLoading" class="posts-status">Loading posts...</p>
          <p v-else-if="postsErrorMessage" class="posts-status error">{{ postsErrorMessage }}</p>
          <p v-else-if="userPosts.length === 0" class="posts-status">No posts yet.</p>
          <ul v-else class="posts-list">
            <li v-for="post in userPosts" :key="post.id" class="post-item">
              <div class="post-header">
                <p class="post-title">{{ post.title }}</p>
                <span class="post-date">{{ formatDate(post.createdAt) }}</span>
              </div>
              <p class="post-content">{{ post.content }}</p>
              <img v-if="post.imagePath" class="post-image" :src="postImageSrc(post.imagePath)" alt="Post image" />
            </li>
          </ul>
        </section>

        <div v-if="isOwnProfile" class="actions">
          <RouterLink class="secondary-btn" to="/profile/edit">Edit profile</RouterLink>
          <button type="button" class="danger-btn" @click="logout" :disabled="isLoggingOut">
            {{ isLoggingOut ? 'Logging out...' : 'Log out' }}
          </button>
        </div>
        </div><!-- end profile-body -->
      </template>
    </section>

    <div v-if="unfollowConfirmOpen" class="confirm-overlay" @click="unfollowConfirmOpen = false">
      <div class="confirm-dialog" @click.stop>
        <h3>Unfollow user</h3>
        <p>Are you sure? If their account is private, you'll need to send a new follow request.</p>
        <div class="confirm-actions">
          <button class="secondary" @click="unfollowConfirmOpen = false">Cancel</button>
          <button class="danger" @click="confirmUnfollow">Unfollow</button>
        </div>
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { API_ROUTES, apiURL } from '@/api/api'

type UserProfile = {
  id: string
  email: string
  firstName: string
  lastName: string
  dateOfBirth: string
  avatar: string
  nickname: string
  aboutMe: string
  isPublic: boolean
  createdAt: string
}

type ErrorResponse = {
  error?: string
}

type ConnectionUser = {
  id: string
  firstName: string
  lastName: string
  nickname: string
}

type UserPost = {
  id: string
  userId: string
  title: string
  content: string
  imagePath: string
  createdAt: string
}

type FollowRequest = {
  id: string
  status: 'pending' | 'accepted' | 'declined'
}

type FollowRequestUser = {
  id: string
  firstName: string
  lastName: string
  nickname: string
  avatar: string
}

type IncomingFollowRequest = {
  id: string
  senderId: string
  receiverId: string
  status: 'pending' | 'accepted' | 'declined'
  createdAt: string
  sender?: FollowRequestUser
}

type OutgoingFollowRequest = {
  id: string
  senderId: string
  receiverId: string
  status: 'pending' | 'accepted' | 'declined'
  createdAt: string
  receiver?: FollowRequestUser
}

const router = useRouter()
const route = useRoute()

const isLoading = ref(true)
const isLoggingOut = ref(false)
const isUpdatingPrivacy = ref(false)
const errorMessage = ref('')
const profile = ref<UserProfile | null>(null)
const followersCount = ref(0)
const followingCount = ref(0)
const followersUsers = ref<ConnectionUser[]>([])
const followingUsers = ref<ConnectionUser[]>([])
const userPosts = ref<UserPost[]>([])
const isPostsLoading = ref(false)
const postsErrorMessage = ref('')
const isOwnProfile = ref(true)
const currentUserID = ref('')
const followState = ref<'unknown' | 'none' | 'requested' | 'following'>('unknown')
const isFollowActionLoading = ref(false)
const incomingFromViewedRequestID = ref('')
const outgoingToViewedRequestID = ref('')
const isIncomingFromViewedLoading = ref(false)
const isOutgoingToViewedLoading = ref(false)
const isUnfollowLoading = ref(false)
const unfollowConfirmOpen = ref(false)
const incomingRequests = ref<IncomingFollowRequest[]>([])
const outgoingRequests = ref<OutgoingFollowRequest[]>([])
const isRequestsLoading = ref(false)
const requestsErrorMessage = ref('')
const processingIncomingID = ref('')
const processingOutgoingID = ref('')

const profileUserID = computed(() => {
  const raw = route.params.userId
  return typeof raw === 'string' ? raw.trim() : ''
})

const fullName = computed(() => {
  if (!profile.value) return ''
  return `${profile.value.firstName} ${profile.value.lastName}`
})

const initials = computed(() => {
  if (!profile.value) return '?'
  return `${profile.value.firstName[0] || ''}${profile.value.lastName[0] || ''}`.toUpperCase()
})

const avatarSrc = computed(() => {
  if (!profile.value?.avatar) return ''
  if (profile.value.avatar.startsWith('http://') || profile.value.avatar.startsWith('https://')) {
    return profile.value.avatar
  }
  return apiURL(profile.value.avatar.startsWith('/') ? profile.value.avatar : `/${profile.value.avatar}`)
})

const followersPath = computed(() => {
  if (!profile.value) return '/friends'
  return `/users/${profile.value.id}/followers`
})

const followingPath = computed(() => {
  if (!profile.value) return '/following'
  return `/users/${profile.value.id}/following`
})

const joinDate = computed(() => {
  if (!profile.value?.createdAt) return 'Unknown'
  return formatDate(profile.value.createdAt)
})

const canViewRestrictedSections = computed(() => {
  if (!profile.value) return false
  if (isOwnProfile.value) return true
  if (profile.value.isPublic) return true
  return followState.value === 'following'
})

const displayedPostsCount = computed(() => {
  return canViewRestrictedSections.value ? String(userPosts.value.length) : 'Hidden'
})

const aboutDisplay = computed(() => {
  if (!profile.value) return ''
  if (!isOwnProfile.value && !profile.value.isPublic && followState.value !== 'following') {
    return 'Follow this user to view full profile details.'
  }
  return profile.value.aboutMe || 'No bio yet.'
})

const followButtonLabel = computed(() => {
  if (isFollowActionLoading.value) return 'Please wait...'
  if (followState.value === 'following') return 'Following'
  if (followState.value === 'requested') return 'Request sent'
  return 'Follow'
})

const canSendFollow = computed(() => {
  return !isFollowActionLoading.value && followState.value === 'none'
})

// Backend allows messaging when EITHER user follows the other (one-way is enough).
// I follow them → followState === 'following'
// They follow me → their id appears in followingUsers (people the viewed user follows)
const canMessage = computed(() => {
  if (isOwnProfile.value || !profile.value) return false
  if (followState.value === 'following') return true
  return followingUsers.value.some((u) => u.id === currentUserID.value)
})

const clearRestrictedSectionsData = () => {
  followersCount.value = 0
  followingCount.value = 0
  followersUsers.value = []
  followingUsers.value = []
  userPosts.value = []
  postsErrorMessage.value = ''
}

const refreshRestrictedSectionsForViewedProfile = async () => {
  if (!profile.value) return

  if (isOwnProfile.value || profile.value.isPublic || followState.value === 'following') {
    await Promise.all([loadConnectionsCount(profile.value.id), loadPostsForUser(profile.value.id)])
    return
  }

  clearRestrictedSectionsData()
}

const fetchProfile = async () => {
  isLoading.value = true
  errorMessage.value = ''
  profile.value = null
  clearRestrictedSectionsData()
  incomingRequests.value = []
  outgoingRequests.value = []
  requestsErrorMessage.value = ''
  processingIncomingID.value = ''
  processingOutgoingID.value = ''
  incomingFromViewedRequestID.value = ''
  outgoingToViewedRequestID.value = ''
  isIncomingFromViewedLoading.value = false
  isOutgoingToViewedLoading.value = false
  isUnfollowLoading.value = false
  isOwnProfile.value = profileUserID.value === ''
  followState.value = 'unknown'

  try {
    const targetUserID = profileUserID.value
    const path = targetUserID ? `/api/users/${encodeURIComponent(targetUserID)}` : API_ROUTES.USERS_ME

    const response = await fetch(apiURL(path), {
      method: 'GET',
      credentials: 'include',
    })

    if (!response.ok) {
      let message = 'Could not load profile.'
      const payload = (await response.json().catch(() => null)) as ErrorResponse | null
      if (payload?.error) message = payload.error
      errorMessage.value = message
      return
    }

    const user = (await response.json()) as UserProfile
    profile.value = user

    if (!targetUserID) {
      currentUserID.value = user.id
      isOwnProfile.value = true
    } else if (currentUserID.value) {
      isOwnProfile.value = currentUserID.value === user.id
    } else {
      const meResponse = await fetch(apiURL(API_ROUTES.USERS_ME), {
        method: 'GET',
        credentials: 'include',
      })
      if (meResponse.ok) {
        const me = (await meResponse.json()) as { id: string }
        currentUserID.value = me.id
        isOwnProfile.value = me.id === user.id
      }
    }

    if (isOwnProfile.value) {
      followState.value = 'following'
      await refreshRestrictedSectionsForViewedProfile()
      await loadFollowRequests()
    } else {
      await loadFollowState(user.id)
      await refreshRestrictedSectionsForViewedProfile()
    }
  } catch {
    errorMessage.value = 'Network error while loading profile.'
  } finally {
    isLoading.value = false
  }
}

const loadConnectionsCount = async (userID: string) => {
  try {
    const [followersResponse, followingResponse] = await Promise.all([
      fetch(apiURL(`/api/users/${userID}/followers`), {
        method: 'GET',
        credentials: 'include',
      }),
      fetch(apiURL(`/api/users/${userID}/following`), {
        method: 'GET',
        credentials: 'include',
      }),
    ])

    if (followersResponse.ok) {
      const followers = (await followersResponse.json()) as ConnectionUser[]
      followersCount.value = followers.length
      followersUsers.value = followers
    }

    if (followingResponse.ok) {
      const following = (await followingResponse.json()) as ConnectionUser[]
      followingCount.value = following.length
      followingUsers.value = following
    }
  } catch {
    followersCount.value = 0
    followingCount.value = 0
    followersUsers.value = []
    followingUsers.value = []
  }
}

const loadPostsForUser = async (userID: string) => {
  isPostsLoading.value = true
  postsErrorMessage.value = ''

  try {
    const response = await fetch(apiURL(API_ROUTES.FEED), {
      method: 'GET',
      credentials: 'include',
    })

    if (!response.ok) {
      let message = 'Could not load user posts.'
      const payload = (await response.json().catch(() => null)) as ErrorResponse | null
      if (payload?.error) message = payload.error
      postsErrorMessage.value = message
      return
    }

    const allPosts = (await response.json()) as UserPost[]
    userPosts.value = allPosts.filter((post) => post.userId === userID)
  } catch {
    postsErrorMessage.value = 'Network error while loading posts.'
  } finally {
    isPostsLoading.value = false
  }
}

const loadFollowState = async (targetUserID: string) => {
  if (isOwnProfile.value) {
    followState.value = 'following'
    incomingFromViewedRequestID.value = ''
    outgoingToViewedRequestID.value = ''
    return
  }
  if (!currentUserID.value) {
    followState.value = 'none'
    incomingFromViewedRequestID.value = ''
    outgoingToViewedRequestID.value = ''
    return
  }

  followState.value = 'none'
  incomingFromViewedRequestID.value = ''
  outgoingToViewedRequestID.value = ''
  try {
    const [followersResponse, outgoingResponse, incomingResponse] = await Promise.all([
      fetch(apiURL(`/api/users/${targetUserID}/followers`), {
        method: 'GET',
        credentials: 'include',
      }),
      fetch(apiURL('/api/follow-requests/outgoing'), {
        method: 'GET',
        credentials: 'include',
      }),
      fetch(apiURL('/api/follow-requests/incoming'), {
        method: 'GET',
        credentials: 'include',
      }),
    ])

    if (followersResponse.ok) {
      const followers = (await followersResponse.json()) as Array<{ id: string }>
      if (followers.some((item) => item.id === currentUserID.value)) {
        followState.value = 'following'
      }
    }

    if (outgoingResponse.ok) {
      const outgoing = (await outgoingResponse.json()) as OutgoingFollowRequest[]
      const pendingOutgoing = outgoing.find((item) => item.receiverId === targetUserID && item.status === 'pending')
      if (pendingOutgoing && followState.value !== 'following') {
        outgoingToViewedRequestID.value = pendingOutgoing.id
        followState.value = 'requested'
      }
    }

    if (incomingResponse.ok) {
      const incoming = (await incomingResponse.json()) as IncomingFollowRequest[]
      const pendingIncoming = incoming.find((item) => item.senderId === targetUserID && item.status === 'pending')
      if (pendingIncoming) {
        incomingFromViewedRequestID.value = pendingIncoming.id
      }
    }
  } catch {
    followState.value = 'none'
    incomingFromViewedRequestID.value = ''
    outgoingToViewedRequestID.value = ''
  }
}

const loadFollowRequests = async () => {
  if (!isOwnProfile.value) return

  isRequestsLoading.value = true
  requestsErrorMessage.value = ''
  try {
    const [incomingResponse, outgoingResponse] = await Promise.all([
      fetch(apiURL('/api/follow-requests/incoming'), {
        method: 'GET',
        credentials: 'include',
      }),
      fetch(apiURL('/api/follow-requests/outgoing'), {
        method: 'GET',
        credentials: 'include',
      }),
    ])

    if (!incomingResponse.ok || !outgoingResponse.ok) {
      let message = 'Could not load follow requests.'
      if (!incomingResponse.ok) {
        const payload = (await incomingResponse.json().catch(() => null)) as ErrorResponse | null
        if (payload?.error) message = payload.error
      } else if (!outgoingResponse.ok) {
        const payload = (await outgoingResponse.json().catch(() => null)) as ErrorResponse | null
        if (payload?.error) message = payload.error
      }
      requestsErrorMessage.value = message
      return
    }

    const incoming = (await incomingResponse.json()) as IncomingFollowRequest[]
    const outgoing = (await outgoingResponse.json()) as OutgoingFollowRequest[]

    incomingRequests.value = incoming.filter((request) => request.status === 'pending')
    outgoingRequests.value = outgoing.filter((request) => request.status === 'pending')
  } catch {
    requestsErrorMessage.value = 'Network error while loading follow requests.'
  } finally {
    isRequestsLoading.value = false
  }
}

const formatRequestUser = (user: FollowRequestUser | undefined, fallbackID: string) => {
  if (!user) return fallbackID
  if (user.nickname) return `@${user.nickname}`
  const fullName = `${user.firstName} ${user.lastName}`.trim()
  return fullName || fallbackID
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
    if (status === 'accepted' && profile.value) {
      await loadConnectionsCount(profile.value.id)
    }
  } catch {
    requestsErrorMessage.value = 'Network error while updating follow request.'
  } finally {
    processingIncomingID.value = ''
  }
}

const withdrawOutgoingRequest = async (requestID: string) => {
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

const handleIncomingFromViewed = async (status: 'accepted' | 'declined') => {
  if (!profile.value || !incomingFromViewedRequestID.value || isIncomingFromViewedLoading.value) return

  isIncomingFromViewedLoading.value = true
  errorMessage.value = ''
  try {
    const requestID = incomingFromViewedRequestID.value
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
      errorMessage.value = message
      return
    }

    incomingFromViewedRequestID.value = ''
    await loadFollowState(profile.value.id)
    await refreshRestrictedSectionsForViewedProfile()
  } catch {
    errorMessage.value = 'Network error while updating follow request.'
  } finally {
    isIncomingFromViewedLoading.value = false
  }
}

const handleWithdrawFromViewed = async () => {
  if (!profile.value || !outgoingToViewedRequestID.value || isOutgoingToViewedLoading.value) return

  isOutgoingToViewedLoading.value = true
  errorMessage.value = ''
  try {
    const requestID = outgoingToViewedRequestID.value
    const response = await fetch(apiURL(`/api/follow-requests/${requestID}`), {
      method: 'DELETE',
      credentials: 'include',
    })

    if (!response.ok) {
      let message = 'Could not withdraw follow request.'
      const payload = (await response.json().catch(() => null)) as ErrorResponse | null
      if (payload?.error) message = payload.error
      errorMessage.value = message
      return
    }

    outgoingToViewedRequestID.value = ''
    followState.value = 'none'
    await loadFollowState(profile.value.id)
    await refreshRestrictedSectionsForViewedProfile()
  } catch {
    errorMessage.value = 'Network error while withdrawing follow request.'
  } finally {
    isOutgoingToViewedLoading.value = false
  }
}

const handleUnfollow = () => {
  if (!profile.value || isUnfollowLoading.value) return
  unfollowConfirmOpen.value = true
}

const confirmUnfollow = async () => {
  unfollowConfirmOpen.value = false
  if (!profile.value) return

  isUnfollowLoading.value = true
  errorMessage.value = ''
  try {
    const response = await fetch(apiURL(`/api/following/${profile.value.id}`), {
      method: 'DELETE',
      credentials: 'include',
    })

    if (!response.ok) {
      let message = 'Could not unfollow this user.'
      const payload = (await response.json().catch(() => null)) as ErrorResponse | null
      if (payload?.error) message = payload.error
      errorMessage.value = message
      return
    }

    followState.value = 'none'
    await loadFollowState(profile.value.id)
    await refreshRestrictedSectionsForViewedProfile()
  } catch {
    errorMessage.value = 'Network error while unfollowing user.'
  } finally {
    isUnfollowLoading.value = false
  }
}

const formatDate = (value: string) => {
  if (!value) return ''
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  return parsed.toLocaleDateString()
}

const postImageSrc = (value: string) => {
  if (!value) return ''
  if (value.startsWith('http://') || value.startsWith('https://')) return value
  return apiURL(value.startsWith('/') ? value : `/${value}`)
}

const handleFollow = async () => {
  if (!profile.value || isOwnProfile.value || !canSendFollow.value) return

  isFollowActionLoading.value = true
  errorMessage.value = ''
  try {
    const response = await fetch(apiURL('/api/follow-requests'), {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        receiverId: profile.value.id,
      }),
    })

    if (!response.ok) {
      const payload = (await response.json().catch(() => null)) as ErrorResponse | null
      const message = payload?.error || 'Could not follow this user.'

      if (response.status === 409) {
        if (message.toLowerCase().includes('already following')) {
          followState.value = 'following'
          await loadFollowState(profile.value.id)
          await refreshRestrictedSectionsForViewedProfile()
          return
        }
        if (message.toLowerCase().includes('already exists')) {
          await loadFollowState(profile.value.id)
          await refreshRestrictedSectionsForViewedProfile()
          return
        }
      }

      errorMessage.value = message
      return
    }

    const created = (await response.json()) as FollowRequest
    if (created.status === 'accepted') {
      followState.value = 'following'
      outgoingToViewedRequestID.value = ''
    } else {
      followState.value = 'requested'
      outgoingToViewedRequestID.value = created.id
    }
    await loadFollowState(profile.value.id)
    await refreshRestrictedSectionsForViewedProfile()
  } catch {
    errorMessage.value = 'Network error while following user.'
  } finally {
    isFollowActionLoading.value = false
  }
}

const togglePrivacy = async () => {
  if (!profile.value || !isOwnProfile.value || isUpdatingPrivacy.value) return

  isUpdatingPrivacy.value = true
  try {
    const response = await fetch(apiURL(API_ROUTES.USERS_ME_PRIVACY), {
      method: 'PATCH',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        isPublic: !profile.value.isPublic,
      }),
    })

    if (!response.ok) {
      let message = 'Could not update profile privacy.'
      const payload = (await response.json().catch(() => null)) as ErrorResponse | null
      if (payload?.error) message = payload.error
      errorMessage.value = message
      return
    }

    const updatedProfile = (await response.json()) as UserProfile
    profile.value = updatedProfile
  } catch {
    errorMessage.value = 'Network error while updating profile privacy.'
  } finally {
    isUpdatingPrivacy.value = false
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
    await router.replace('/login')
  }
}

onMounted(fetchProfile)

watch(
  () => route.params.userId,
  () => {
    void fetchProfile()
  },
)
</script>

<style scoped>
.profile-page {
  min-height: calc(100dvh - var(--navbar-height, 60px));
  padding: 0.25rem 0 1.75rem;
}

.profile-card {
  width: min(100%, 1320px);
  margin: 0 auto;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(248, 250, 252, 0.86));
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 32px;
  overflow: hidden;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.08);
}

/* ── Profile banner ── */
.profile-card::before {
  content: '';
  display: block;
  height: 220px;
  background:
    radial-gradient(circle at top right, rgba(255, 255, 255, 0.08), transparent 24%),
    radial-gradient(circle at bottom left, rgba(125, 211, 252, 0.12), transparent 26%),
    linear-gradient(135deg, rgba(8, 18, 37, 0.96), rgba(15, 70, 180, 0.96));
}

/* Body wrapper provides padding for all content below banner */
.profile-body {
  padding: 0 1.5rem 1.5rem;
}

/* Avatar overlaps banner by 50% */
.profile-header {
  display: flex;
  gap: 1.25rem;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.5rem;
  margin-top: -58px;
}

.status-message {
  min-height: 220px;
  display: grid;
  place-items: center;
  text-align: center;
  color: var(--gray-600);
  font-weight: 700;
  padding: 2rem 1rem;
}

.status-message.error {
  color: var(--status-error);
}

.avatar-wrap img,
.avatar-fallback {
  width: 100px;
  height: 100px;
  border-radius: var(--radius-full);
  border: 4px solid rgba(255, 255, 255, 0.96);
  box-shadow: 0 18px 34px rgba(15, 23, 42, 0.14);
}

.avatar-wrap img {
  object-fit: cover;
}

.avatar-fallback {
  display: grid;
  place-items: center;
  background: linear-gradient(135deg, rgba(191, 219, 254, 0.95), rgba(219, 234, 254, 0.9));
  color: var(--brand-700);
  font-weight: 900;
  font-size: 1.8rem;
}

.identity {
  display: grid;
  align-content: end;
  gap: 0.25rem;
  min-width: 0;
  padding-top: 0.4rem;
}

.identity h1 {
  font-size: clamp(1.8rem, 3vw, 2.5rem);
  font-weight: 900;
  color: #fff;
  margin: 0;
  line-height: 1.05;
  text-shadow: 0 2px 12px rgba(15, 23, 42, 0.18);
}

.nickname {
  color: var(--gray-900);
  font-weight: 800;
  font-size: 0.98rem;
  margin-top: 0.2rem;
}

/* nickname text already includes @ from the template */

.privacy-tag {
  display: inline-block;
  margin-top: 0.35rem;
  font-size: 0.78rem;
  background: #dcfce7;
  color: #15803d;
  border-radius: 999px;
  padding: 0.3rem 0.8rem;
  font-weight: 800;
  width: fit-content;
}

.privacy-tag.private {
  background: #ffedd5;
  color: #b45309;
}

.follow-btn {
  border: none;
  border-radius: 18px;
  padding: 0.78rem 0.95rem;
  font-weight: 800;
  font-size: 0.88rem;
  color: var(--white);
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  cursor: pointer;
  box-shadow: 0 14px 28px rgba(37, 99, 235, 0.18);
  transition:
    transform var(--dur-fast) var(--ease-standard),
    box-shadow var(--dur-fast) var(--ease-standard),
    opacity var(--dur-fast) var(--ease-standard);
}

.follow-btn:not(.unfollow-btn):not(.withdraw-btn):hover {
  transform: translateY(-1px);
}

.follow-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.follow-controls {
  margin-top: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  align-items: flex-start;
}

.message-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  margin-top: 0.5rem;
  border: none;
  border-radius: 18px;
  padding: 0.72rem 1rem;
  font-weight: 800;
  font-size: 0.9rem;
  color: var(--white);
  background: linear-gradient(135deg, #059669, #047857);
  box-shadow: 0 12px 24px rgba(5, 150, 105, 0.22);
  text-decoration: none;
  cursor: pointer;
  transition:
    transform var(--dur-fast) var(--ease-standard),
    box-shadow var(--dur-fast) var(--ease-standard);
}

.message-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 16px 30px rgba(5, 150, 105, 0.3);
}

.header-side {
  margin-left: auto;
  align-self: center;
  padding-top: 0;
}

.inline-request-actions {
  display: flex;
  gap: 0.6rem;
  flex-wrap: wrap;
}

/* Following state: ghost button, danger on hover */
.unfollow-btn {
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  border: 1px solid rgba(148, 163, 184, 0.18);
  color: var(--gray-700);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
}

.unfollow-btn:hover {
  border-color: var(--danger);
  color: var(--danger);
}

.withdraw-btn {
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  border: 1px solid rgba(148, 163, 184, 0.18);
  color: var(--gray-700);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
}

.pending-note {
  color: var(--text-secondary);
  font-weight: 600;
  font-size: 0.9rem;
}

.profile-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem 1.2rem;
}

.profile-grid dt {
  color: var(--gray-500);
  font-size: 0.82rem;
  margin-bottom: 0.28rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.profile-grid dd {
  color: var(--gray-900);
  font-weight: 700;
  word-break: break-word;
}

.full-width {
  grid-column: 1 / -1;
}

.actions {
  margin-top: 1.5rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.privacy-toggle-row {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--gray-200);
  border-radius: var(--radius-md);
  background: var(--brand-50);
}

.header-privacy {
  min-width: 220px;
  justify-content: center;
}

.privacy-toggle-copy p {
  color: var(--text-primary);
  font-size: 0.86rem;
  font-weight: 700;
  line-height: 1.1;
}

.privacy-toggle-copy small {
  display: none;
}

.privacy-switch {
  position: relative;
  display: inline-flex;
  width: 46px;
  height: 26px;
}

.privacy-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.privacy-slider {
  position: absolute;
  inset: 0;
  border-radius: var(--radius-full);
  background: var(--gray-300);
  transition: background 200ms ease;
  cursor: pointer;
}

.privacy-slider::before {
  content: '';
  position: absolute;
  left: 3px;
  top: 3px;
  width: 20px;
  height: 20px;
  border-radius: var(--radius-full);
  background: var(--white);
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.25);
  transition: transform 200ms ease;
}

.privacy-switch input:checked + .privacy-slider {
  background: var(--brand-500);
}

.privacy-switch input:checked + .privacy-slider::before {
  transform: translateX(20px);
}

.privacy-switch input:disabled + .privacy-slider {
  opacity: 0.65;
  cursor: not-allowed;
}

.connections {
  margin-top: 1rem;
  display: flex;
  gap: 1.25rem;
  flex-wrap: wrap;
}

.private-lock-note {
  margin-top: 0.9rem;
  color: var(--text-secondary);
  font-weight: 600;
}

.stat {
  display: flex;
  align-items: baseline;
  gap: 0.3rem;
}

.stat-number {
  font-size: 1.7rem;
  font-weight: 900;
  color: var(--gray-900);
}

.stat-link {
  color: var(--brand-600);
  text-decoration: none;
  font-weight: 800;
  font-size: 0.95rem;
}

.stat-link:hover {
  text-decoration: underline;
}

.secondary-btn,
.danger-btn {
  border: none;
  border-radius: 18px;
  padding: 0.78rem 0.95rem;
  font-weight: 800;
  cursor: pointer;
  text-decoration: none;
  font-size: 0.88rem;
}

.secondary-btn {
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  color: var(--text-white);
  box-shadow: 0 14px 28px rgba(37, 99, 235, 0.18);
}

.secondary-btn:hover {
  transform: translateY(-1px);
}

.danger-btn {
  background: linear-gradient(135deg, #ef4444, #b91c1c);
  color: #fff;
  box-shadow: 0 14px 28px rgba(239, 68, 68, 0.22);
}

.danger-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.activity-panel,
.connections-panel,
.requests-panel,
.posts-panel {
  margin-top: 1.2rem;
}

.activity-panel h2,
.connections-panel h2,
.requests-panel h2,
.posts-panel h2 {
  font-size: 1.35rem;
  color: var(--gray-900);
  margin-bottom: 0.9rem;
  font-weight: 900;
}

.connections-grid {
  display: grid;
  gap: 0.9rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.connections-column {
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 24px;
  padding: 1rem;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.88));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.08);
}

.connections-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.45rem;
}

.mini-link {
  text-decoration: none;
  color: var(--brand-600);
  font-weight: 800;
  font-size: 0.88rem;
}

.mini-link:hover {
  text-decoration: underline;
}

.connections-empty {
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.connections-list {
  display: grid;
  gap: 0.45rem;
}

.connections-list li {
  list-style: none;
}

.connections-user {
  text-decoration: none;
  color: var(--text-primary);
  font-weight: 600;
}

.connections-user:hover {
  color: var(--primary-blue);
}

.requests-status {
  color: var(--text-secondary);
  font-weight: 600;
}

.requests-status.error {
  color: var(--status-error);
}

.requests-grid {
  display: grid;
  gap: 0.9rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.request-column {
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 24px;
  padding: 1rem;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.88));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.08);
}

.request-column h3 {
  color: var(--text-primary);
  font-size: 0.95rem;
  margin-bottom: 0.55rem;
}

.request-empty {
  color: var(--text-secondary);
}

.request-list {
  display: grid;
  gap: 0.55rem;
}

.request-item {
  list-style: none;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 20px;
  padding: 0.9rem;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.92));
}

.request-main {
  display: flex;
  justify-content: space-between;
  gap: 0.6rem;
  align-items: baseline;
}

.request-user {
  color: var(--text-primary);
  font-weight: 700;
  text-decoration: none;
}

.request-user:hover {
  color: var(--primary-blue);
}

.request-date {
  color: var(--text-secondary);
  font-size: 0.8rem;
}

.request-actions {
  margin-top: 0.5rem;
  display: flex;
  gap: 0.5rem;
}

.request-btn {
  border: none;
  border-radius: 18px;
  padding: 0.78rem 0.95rem;
  font-weight: 800;
  font-size: 0.84rem;
  color: #fff;
  cursor: pointer;
}

.request-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.request-btn.accept {
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  box-shadow: 0 14px 28px rgba(37, 99, 235, 0.18);
}

.request-btn.decline {
  color: #8a3b12;
  background: linear-gradient(180deg, rgba(255, 247, 237, 0.98), rgba(255, 237, 213, 0.9));
  border: 1px solid rgba(251, 146, 60, 0.25);
}

.request-btn.withdraw {
  color: var(--gray-700);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  border: 1px solid rgba(148, 163, 184, 0.18);
}

.activity-list {
  display: flex;
  gap: 0.8rem;
  flex-wrap: wrap;
  color: var(--text-secondary);
  font-weight: 600;
}

.activity-list li {
  list-style: none;
}

.posts-status {
  color: var(--text-secondary);
  font-weight: 600;
}

.posts-status.error {
  color: var(--status-error);
}

.posts-list {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.8rem;
}

.post-item {
  list-style: none;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 24px;
  padding: 1rem;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.88));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.08);
}

.post-header {
  display: flex;
  justify-content: space-between;
  gap: 0.8rem;
  align-items: baseline;
  margin-bottom: 0.4rem;
}

.post-title {
  color: var(--gray-900);
  font-weight: 800;
  overflow-wrap: break-word;
  min-width: 0;
}

.post-date {
  color: var(--gray-500);
  font-size: 0.82rem;
}

.post-content {
  color: var(--gray-900);
  white-space: pre-wrap;
  overflow-wrap: break-word;
  line-height: 1.6;
}

.post-image {
  width: 100%;
  max-height: 360px;
  object-fit: cover;
  border-radius: 18px;
  margin-top: 0.8rem;
  border: 1px solid rgba(148, 163, 184, 0.16);
}

@media (max-width: 650px) {
  .profile-header {
    align-items: flex-start;
    flex-direction: column;
    margin-top: -58px;
  }

  .header-side {
    margin-left: 0;
    width: 100%;
    padding-top: 0;
  }

  .profile-grid {
    grid-template-columns: 1fr;
  }

  .requests-grid {
    grid-template-columns: 1fr;
  }

  .connections-grid {
    grid-template-columns: 1fr;
  }

  .actions {
    flex-direction: column;
  }

  .secondary-btn,
  .danger-btn,
  .privacy-toggle-row {
    width: 100%;
  }

  .profile-card,
  .connections-column,
  .request-column,
  .post-item,
  .confirm-dialog {
    border-radius: 24px;
  }
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

.confirm-dialog h3 { color: var(--gray-900); }
.confirm-dialog p  { color: var(--gray-600); }

.confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.65rem;
}
</style>
