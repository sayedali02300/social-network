import { apiURL } from '@/api/api'
import type {
  EventItem,
  EventResponse,
  Group,
  GroupInvite,
  GroupJoinRequest,
  GroupMember,
  GroupSummary,
} from '@/types/groups'
import type { Post } from '@/types/post'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers ?? {})
  if (init?.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const response = await fetch(apiURL(path), {
    ...init,
    headers,
    credentials: 'include',
  })

  if (!response.ok) {
    const text = await response.text()
    throw new Error(text || `Request failed with status ${response.status}`)
  }

  return (await response.json()) as T
}

function serializeEventTimeInput(eventTime: string) {
  const parsed = new Date(eventTime)
  if (Number.isNaN(parsed.getTime())) {
    return eventTime
  }

  return parsed.toISOString()
}

export function listGroups() {
  return request<GroupSummary[]>('/api/groups')
}

export function createGroup(payload: { title: string; description: string }) {
  return request<Group>('/api/groups', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function getGroup(groupId: string) {
  return request<GroupSummary>(`/api/groups/${groupId}`)
}

export function updateGroup(groupId: string, payload: { title: string; description: string }) {
  return request<Group>(`/api/groups/${groupId}`, {
    method: 'PATCH',
    body: JSON.stringify(payload),
  })
}

export function deleteGroup(groupId: string) {
  return request<{ message: string }>(`/api/groups/${groupId}`, {
    method: 'DELETE',
  })
}

export function listGroupMembers(groupId: string) {
  return request<GroupMember[]>(`/api/groups/${groupId}/members`)
}

export function removeGroupMember(groupId: string, userId: string) {
  return request<{ message: string }>(`/api/groups/${groupId}/members/${userId}`, {
    method: 'DELETE',
  })
}

export function createGroupInvite(groupId: string, receiverId: string) {
  return request<GroupInvite>(`/api/groups/${groupId}/invites`, {
    method: 'POST',
    body: JSON.stringify({ receiverId }),
  })
}

export function listGroupInvites(groupId: string) {
  return request<GroupInvite[]>(`/api/groups/${groupId}/invites`)
}

export function respondToInvite(inviteId: string, status: 'accepted' | 'declined') {
  return request<GroupInvite>(`/api/group-invites/${inviteId}`, {
    method: 'PATCH',
    body: JSON.stringify({ status }),
  })
}

export function createJoinRequest(groupId: string) {
  return request<GroupJoinRequest>(`/api/groups/${groupId}/requests`, {
    method: 'POST',
  })
}

export function listJoinRequests(groupId: string) {
  return request<GroupJoinRequest[]>(`/api/groups/${groupId}/requests`)
}

export function respondToJoinRequest(requestId: string, status: 'accepted' | 'declined') {
  return request<GroupJoinRequest>(`/api/group-requests/${requestId}`, {
    method: 'PATCH',
    body: JSON.stringify({ status }),
  })
}

export function listGroupPosts(groupId: string) {
  return request<Post[]>(`/api/groups/${groupId}/posts`)
}

export async function createGroupPost(
  groupId: string,
  payload: { title: string; body: string; image?: File | null },
) {
  const formData = new FormData()
  formData.append('title', payload.title)
  formData.append('body', payload.body)
  if (payload.image) {
    formData.append('image', payload.image)
  }

  const response = await fetch(apiURL(`/api/groups/${groupId}/posts`), {
    method: 'POST',
    body: formData,
    credentials: 'include',
  })

  if (!response.ok) {
    const text = await response.text()
    throw new Error(text || `Request failed with status ${response.status}`)
  }

  return (await response.json()) as Post
}

export function updateGroupPost(postId: string, payload: { title: string; content: string }) {
  return request<Post>(`/api/posts/${postId}`, {
    method: 'PATCH',
    body: JSON.stringify(payload),
  })
}

export function deleteGroupPost(postId: string) {
  return request<{ message: string }>(`/api/posts/${postId}`, {
    method: 'DELETE',
  })
}

export function updateComment(commentId: string, payload: { content: string }) {
  return request<{ id: string; content: string; createdAt: string; imagePath?: string; author: unknown }>(
    `/api/comments/${commentId}`,
    {
      method: 'PATCH',
      body: JSON.stringify(payload),
    },
  )
}

export function deleteComment(commentId: string) {
  return request<{ message: string }>(`/api/comments/${commentId}`, {
    method: 'DELETE',
  })
}

export function listEvents(groupId: string) {
  return request<EventItem[]>(`/api/groups/${groupId}/events`)
}

export function createEvent(
  groupId: string,
  payload: { title: string; description: string; eventTime: string },
) {
  return request<EventItem>(`/api/groups/${groupId}/events`, {
    method: 'POST',
    body: JSON.stringify({
      ...payload,
      eventTime: serializeEventTimeInput(payload.eventTime),
    }),
  })
}

export function updateEvent(
  eventId: string,
  payload: { title: string; description: string; eventTime: string },
) {
  return request<EventItem>(`/api/events/${eventId}`, {
    method: 'PATCH',
    body: JSON.stringify({
      ...payload,
      eventTime: serializeEventTimeInput(payload.eventTime),
    }),
  })
}

export function deleteEvent(eventId: string) {
  return request<{ message: string }>(`/api/events/${eventId}`, {
    method: 'DELETE',
  })
}

export function respondToEvent(eventId: string, response: 'going' | 'not_going') {
  return request<EventResponse>(`/api/events/${eventId}/responses`, {
    method: 'POST',
    body: JSON.stringify({ response }),
  })
}
