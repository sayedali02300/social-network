import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import SearchView from '@/views/SearchView.vue'
import FollowersListView from '@/views/FollowersList.vue'
import ProfileView from '@/views/Profile.vue'
import EditProfileView from '@/views/EditProfile.vue'
import FollowingListView from '@/views/FollowingList.vue'
import GroupsView from '@/views/GroupsView.vue'
import CreateGroupView from '@/views/CreateGroupView.vue'
import GroupDetailsView from '@/views/GroupDetailsView.vue'
import GroupMembersView from '@/views/GroupMembersView.vue'
import EventsView from '@/views/EventsView.vue'
import LoginView from '@/views/Login.vue'
import RegisterView from '@/views/Register.vue'
import PrivateChatView from '@/views/PrivateChat.vue'
import GroupChatView from '@/views/GroupChat.vue'
import NotificationsView from '@/views/NotificationsView.vue'
import ChatListView from '@/views/ChatListView.vue'
import { API_ROUTES, apiURL } from '@/api/api'
import type { SessionData } from '@/types/User.ts'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/search',
      name: 'search',
      component: SearchView,
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/profile',
      name: 'profile',
      component: ProfileView,
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/users/:userId',
      name: 'user-profile',
      component: ProfileView,
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/profile/edit',
      name: 'edit-profile',
      component: EditProfileView,
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/friends',
      name: 'friends',
      component: FollowersListView,
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/users/:userId/followers',
      name: 'user-followers',
      component: FollowersListView,
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/following',
      name: 'following',
      component: FollowingListView,
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/users/:userId/following',
      name: 'user-following',
      component: FollowingListView,
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/groups',
      name: 'groups',
      component: GroupsView,
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/groups/new',
      name: 'create-group',
      component: CreateGroupView,
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/groups/:groupId',
      name: 'group-details',
      component: GroupDetailsView,
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/groups/:groupId/members',
      name: 'group-members',
      component: GroupMembersView,
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/groups/:groupId/events',
      name: 'group-events',
      component: EventsView,
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/chats',
      name: 'chat-list',
      component: ChatListView,
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/notifications',
      name: 'notifications',
      component: NotificationsView,
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/chats/private/:userId',
      name: 'private-chat',
      component: PrivateChatView,
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/chats/groups/:groupId',
      name: 'group-chat',
      component: GroupChatView,
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView,
      meta: {
        publicOnly: true,
        hideNavbar: true,
      },
    },
    {
      path: '/register',
      name: 'register',
      component: RegisterView,
      meta: {
        publicOnly: true,
        hideNavbar: true,
      },
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/',
    },
    {
      path: '/posts/:id',
      name: 'post',
      component: () => import('@/views/PostPage.vue')
    },
  ],
})

const resolveSession = async (): Promise<boolean> => {
  try {
    const response = await fetch(apiURL(API_ROUTES.AUTH_SESSION), {
      method: 'GET',
      credentials: 'include',
      cache: 'no-store',
    })
    if (!response.ok) return false

    const contentType = response.headers.get('content-type') || ''
    if (!contentType.includes('application/json')) return false

    const payload = (await response.json().catch(() => null)) as
      | { session?: { id?: string }; user?: { id?: string } }
      | null

    return Boolean(payload?.session?.id && payload?.user?.id)
  } catch {
    return false
  }
}

export const fetchSessionData = async (): Promise<SessionData | null> => {
  try {
    const response = await fetch(apiURL(API_ROUTES.AUTH_SESSION), {
      method: 'GET',
      credentials: 'include',
      cache: 'no-store',
    })
    if (!response.ok) return null

    const payload = (await response.json().catch(() => null)) as SessionData | null

    if (payload?.session?.id && payload?.user?.id) {
      return payload
    }

    return null
  } catch {
    return null
  }
}

router.beforeEach(async (to) => {
  const authenticated = await resolveSession()
  const requiresAuth = to.meta.requiresAuth === true
  const isPublicOnlyRoute = to.meta.publicOnly === true

  if (!authenticated && requiresAuth) {
    return {
      name: 'login',
      query: {
        redirect: to.fullPath,
      },
    }
  }

  if (authenticated && isPublicOnlyRoute) {
    return {
      path: '/',
    }
  }

  return true
})

export default router
