export interface CommentAuthor {
  user_id? : string
  avatar?: string
  nickname?: string
  firstName?: string
  lastName?: string
}

export interface Comment {
  id: string
  postId: string
  content: string
  imagePath?: string
  parentId?: string
  createdAt: string
  author: CommentAuthor
}