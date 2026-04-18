export interface GroupSummary {
  id: string
  creatorId: string
  creatorName: string
  title: string
  description: string
  membersCount: number
  isMember: boolean
  hasPendingInvite: boolean
  hasPendingRequest: boolean
  createdAt: string
}

export interface Group {
  id: string
  creatorId: string
  title: string
  description: string
  createdAt: string
}

export interface GroupMember {
  userId: string
  nickname: string
  avatar: string
  role: string
  joinedAt: string
}

export interface GroupInvite {
  id: string
  groupId: string
  senderId: string
  receiverId: string
  status: 'pending' | 'accepted' | 'declined'
  createdAt: string
}

export interface GroupJoinRequest {
  id: string
  groupId: string
  userId: string
  status: 'pending' | 'accepted' | 'declined'
  createdAt: string
}

export interface EventItem {
  id: string
  groupId: string
  creatorId: string
  title: string
  description: string
  eventTime: string
  createdAt: string
  goingCount: number
  notGoingCount: number
  myResponse: '' | 'going' | 'not_going'
}

export interface EventResponse {
  eventId: string
  userId: string
  response: 'going' | 'not_going'
  createdAt: string
}
