export interface Session {
  id: string
  userId: string
  createdAt: string
  expiresAt: string
}

export interface User {
  id: string
  email: string
  firstName: string
  lastName: string
  dateOfBirth: string
  avatar: string
  nickname: string
  aboutMe?: string
  isPublic: boolean
  createdAt: string
}

export interface SessionData {
  session: Session
  user: User
}