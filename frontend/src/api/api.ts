export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const API_ROUTES = {
  AUTH_REGISTER: '/api/auth/register',
  AUTH_LOGIN: '/api/auth/login',
  AUTH_LOGOUT: '/api/auth/logout',
  AUTH_SESSION: '/api/auth/session',
  USERS_ME: '/api/users/me',
  USERS_ME_PRIVACY: '/api/users/me/privacy',
  USERS_SEARCH: '/api/users/search',
  NOTIFICATIONS: '/api/notifications',
  NOTIFICATIONS_UNREAD_COUNT: '/api/notifications/unread-count',
  POSTS: '/api/posts',
  FEED: '/api/posts/feed',
  COMMENTS: '/api/comments',
}

export const apiURL = (path: string) => `${API_BASE_URL}${path}`
