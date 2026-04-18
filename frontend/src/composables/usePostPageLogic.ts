import { ref, onMounted, onUnmounted, watch } from 'vue'
import { getImgURL, formatPostDate } from '@/utils/helpers'
import { usePost } from '@/composables/usePostPage'
import { nextTick } from "vue"
import { API_BASE_URL } from '@/api/api'
import { API_ROUTES } from '@/api/api'
import { buildWebSocketURL } from '@/api/websocket'
import type { CommentAuthor, Comment} from '@/types/comment'

export function usePostPageLogic(postId: string) {
  const { post, loading, error } = usePost(postId)

  const comments        = ref<Comment[]>([])
  const commentsLoading = ref(true)
  const newComment      = ref('')
  const imageFile       = ref<File | null>(null)
  const submitting      = ref(false)
  const commentError    = ref('')
  const lightboxSrc     = ref<string | null>(null)

  // ── Helpers ──────────────────────────────────────────────────────────────

  const initials = (author?: CommentAuthor) =>
    ((author?.firstName?.[0] ?? '') + (author?.lastName?.[0] ?? '')).toUpperCase() || '?'

  // ── Lightbox ─────────────────────────────────────────────────────────────

  const openLightbox = (src: string) => {
    lightboxSrc.value = src
    document.body.style.overflow = 'hidden'
  }

  const closeLightbox = () => {
    lightboxSrc.value = null
    document.body.style.overflow = ''
  }

  const onKeydown = (e: KeyboardEvent) => {
    if (e.key === 'Escape') closeLightbox()
  }

  // ── Comments ─────────────────────────────────────────────────────────────
  
  const fetchComments = async () => {
    commentsLoading.value = true
    try {
      const res = await fetch(`${API_BASE_URL}${API_ROUTES.POSTS}/${postId}/comments`, { method: 'GET', credentials: 'include' })
      if (res.ok) {
        comments.value = await res.json()
      } else {
        commentError.value = 'Failed to load comments.'
      }
    } catch {
      commentError.value = 'Failed to load comments.'
    } finally {
      commentsLoading.value = false
    }
  }

  const onImagePicked = (e: Event) => {
    const file = (e.target as HTMLInputElement).files?.[0] ?? null
    if (file) {
      const validTypes = ['image/jpeg', 'image/png', 'image/gif']
      if (!validTypes.includes(file.type)) {
        commentError.value = 'Invalid file type. Only JPEG, PNG, and GIF allowed.'
        imageFile.value = null
        const fileInput = (e.target as HTMLInputElement); if (fileInput) fileInput.value = ''
        return
      }
      if (file.size > 10 * 1024 * 1024) {
        commentError.value = 'File size must be under 10MB.'
        imageFile.value = null
        const fileInput = (e.target as HTMLInputElement); if (fileInput) fileInput.value = ''
        return
      }
      commentError.value = ''
    }
    imageFile.value = file
  }
  
  // Reply state
  const replyToId    = ref<string | null>(null)
  const replyContent = ref('')
  const replyError   = ref('')
  const replySubmitting = ref(false)

  const openReply = (commentId: string) => {
    replyToId.value    = commentId
    replyContent.value = ''
    replyError.value   = ''
  }
  const closeReply = () => { replyToId.value = null }

  const submitReply = async () => {
    if (!replyContent.value.trim() || !replyToId.value) return
    replySubmitting.value = true
    replyError.value = ''
    try {
      const body = new FormData()
      body.append('content', replyContent.value.trim())
      body.append('parent_id', replyToId.value)
      const res = await fetch(`${API_BASE_URL}${API_ROUTES.POSTS}/${postId}/comments`, {
        method: 'POST',
        credentials: 'include',
        body,
      })
      if (!res.ok) {
        const errData = await res.json().catch(() => null)
        replyError.value = errData?.error || 'Failed to post reply.'
        return
      }
      replyToId.value    = null
      replyContent.value = ''
      await fetchComments()
    } catch {
      replyError.value = 'Failed to post reply.'
    } finally {
      replySubmitting.value = false
    }
  }

  const submitComment = async () => {
  if (!newComment.value.trim() && !imageFile.value) return
  submitting.value = true
  commentError.value = ''

  if (!postId) return alert('Post ID missing')

  try {
    const body = new FormData()
    body.append('postId', postId)
    body.append('content', newComment.value.trim())
    if (imageFile.value) body.append('image', imageFile.value)

    const res = await fetch(`${API_BASE_URL}${API_ROUTES.POSTS}/${postId}/comments`, {
      method: 'POST',
      credentials: 'include',
      body,
    })

    if (!res.ok) {
      try {
        const errData = await res.json()
        commentError.value = errData.error || errData.message || 'Failed to post comment.'
      } catch {
        commentError.value = 'Failed to post comment.'
      }
      return
    }

    newComment.value = ''
    imageFile.value = null
    await fetchComments()
    await nextTick()
    const container = document.querySelector(".comments-scroll")
    if (container) {
      container.scrollTop = 0
    }
  } catch {
    commentError.value = 'Failed to post comment.'
  } finally {
    submitting.value = false
  }
}

const deleteComment = async (commentId: string): Promise<boolean> => {
  try {
    const response = await fetch(`${API_BASE_URL}${API_ROUTES.COMMENTS}/${commentId}`, {
      method: 'DELETE',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      const data = await response.json().catch(() => null)
      console.error('Failed to delete comment:', data?.message)
      return false
    }

    return true
  } catch (err) {
    console.error('Error deleting comment:', err)
    return false
  }
}

  // ── Add Recipients (private posts) ──────────────────────────────────────
  const showAddRecipients    = ref(false)
  const followers            = ref<{ id: string; firstName: string; lastName: string; nickname: string; avatar: string }[]>([])
  const allowedUserIDs       = ref<string[]>([])
  const selectedNewRecipients = ref<string[]>([])
  const recipientsLoading    = ref(false)
  const recipientsError      = ref('')

  const openAddRecipients = async (authorId: string) => {
    showAddRecipients.value = true
    recipientsLoading.value = true
    recipientsError.value = ''
    selectedNewRecipients.value = []

    try {
      const [followersRes, allowedRes] = await Promise.all([
        fetch(`${API_BASE_URL}/api/users/${authorId}/followers`, { credentials: 'include' }),
        fetch(`${API_BASE_URL}${API_ROUTES.POSTS}/${postId}/allowed-users`, { credentials: 'include' }),
      ])

      if (followersRes.ok) {
        const data = await followersRes.json()
        followers.value = Array.isArray(data) ? data : (data.followers || [])
      }

      if (allowedRes.ok) {
        const data = await allowedRes.json()
        allowedUserIDs.value = data.user_ids || []
      }
    } catch {
      recipientsError.value = 'Failed to load data.'
    } finally {
      recipientsLoading.value = false
    }
  }

  const closeAddRecipients = () => {
    showAddRecipients.value = false
  }

  const selectedRemoveRecipients = ref<string[]>([])

  const submitNewRecipients = async () => {
    if (selectedNewRecipients.value.length === 0) return
    recipientsError.value = ''

    try {
      const res = await fetch(`${API_BASE_URL}${API_ROUTES.POSTS}/${postId}/allowed-users`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ user_ids: selectedNewRecipients.value }),
      })

      if (!res.ok) {
        const errData = await res.json().catch(() => null)
        recipientsError.value = errData?.error || 'Failed to add recipients.'
        return
      }

      // Refresh the allowed users list
      allowedUserIDs.value = [...allowedUserIDs.value, ...selectedNewRecipients.value]
      selectedNewRecipients.value = []
    } catch {
      recipientsError.value = 'Failed to add recipients.'
    }
  }

  const submitRemoveRecipients = async () => {
    if (selectedRemoveRecipients.value.length === 0) return
    recipientsError.value = ''

    try {
      const res = await fetch(`${API_BASE_URL}${API_ROUTES.POSTS}/${postId}/allowed-users`, {
        method: 'DELETE',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ user_ids: selectedRemoveRecipients.value }),
      })

      if (!res.ok) {
        const errData = await res.json().catch(() => null)
        recipientsError.value = errData?.error || 'Failed to remove recipients.'
        return
      }

      // Update local state
      allowedUserIDs.value = allowedUserIDs.value.filter(id => !selectedRemoveRecipients.value.includes(id))
      selectedRemoveRecipients.value = []
    } catch {
      recipientsError.value = 'Failed to remove recipients.'
    }
  }

  // ── Likes ──────────────────────────────────────────────────────────────
  const likeCount    = ref(0)
  const dislikeCount = ref(0)
  const myReaction   = ref(0)
  const likeError    = ref('')

  // Sync from post once loaded
  const syncLikesFromPost = () => {
    if (post.value) {
      likeCount.value    = post.value.likeCount    ?? 0
      dislikeCount.value = post.value.dislikeCount ?? 0
      myReaction.value   = post.value.myReaction   ?? 0
    }
  }

  const handleReaction = async (value: 1 | -1) => {
    if (!post.value) return
    likeError.value = ''

    const prevReaction   = myReaction.value
    const prevLikes      = likeCount.value
    const prevDislikes   = dislikeCount.value

    // Optimistic update
    if (prevReaction === value) {
      // Toggle off
      myReaction.value = 0
      if (value === 1) likeCount.value = Math.max(0, likeCount.value - 1)
      else dislikeCount.value = Math.max(0, dislikeCount.value - 1)
    } else {
      // Switch or new
      if (prevReaction === 1) likeCount.value = Math.max(0, likeCount.value - 1)
      if (prevReaction === -1) dislikeCount.value = Math.max(0, dislikeCount.value - 1)
      myReaction.value = value
      if (value === 1) likeCount.value += 1
      else dislikeCount.value += 1
    }

    try {
      if (prevReaction === value) {
        // Remove reaction
        const res = await fetch(`${API_BASE_URL}${API_ROUTES.POSTS}/${postId}/like`, {
          method: 'DELETE',
          credentials: 'include',
        })
        if (res.ok) {
          const data = await res.json()
          likeCount.value    = data.likes
          dislikeCount.value = data.dislikes
          myReaction.value   = data.myReaction
        } else {
          likeCount.value = prevLikes; dislikeCount.value = prevDislikes; myReaction.value = prevReaction
          likeError.value = 'Failed to remove reaction.'
        }
      } else {
        // Upsert reaction
        const res = await fetch(`${API_BASE_URL}${API_ROUTES.POSTS}/${postId}/like`, {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ value }),
        })
        if (res.ok) {
          const data = await res.json()
          likeCount.value    = data.likes
          dislikeCount.value = data.dislikes
          myReaction.value   = data.myReaction
        } else {
          likeCount.value = prevLikes; dislikeCount.value = prevDislikes; myReaction.value = prevReaction
          likeError.value = 'Failed to save reaction.'
        }
      }
    } catch {
      likeCount.value = prevLikes; dislikeCount.value = prevDislikes; myReaction.value = prevReaction
      likeError.value = 'Failed to update reaction.'
    }
  }

  // ── Websocket ─────────────────────────────────────────────────────────
  const socket = ref<WebSocket | null>(null)

  const connectWebSocket = () =>{
    if(socket.value && socket.value.readyState === WebSocket.OPEN) return
    socket.value = new WebSocket(buildWebSocketURL('/ws'))
    socket.value.onopen = () => console.log("websocket connected")
    socket.value.onmessage = async (event) => {
  try {
    const data = JSON.parse(event.data)

        if (data.type === "comment_event") {
          const newCommentPayload = data.payload
          if (newCommentPayload.postId === postId) {
            const exists = comments.value.find((c) => c.id === newCommentPayload.id)
            if (!exists) {
              comments.value.unshift(newCommentPayload)

              nextTick(() => {
                const container = document.querySelector(".comments-scroll")
                if(container){
                  container.scrollTop = 0;
                }
              })
            }
          }
        } else if (data.type === "delete_comment") {
          if (data.payload.postId === postId) {
            comments.value = comments.value.filter((c) => c.id !== data.payload.commentId)
          }
        } else if (data.type === "follower_removed") {
          const { followerId } = data.payload
          // Remove unfollowed user from the followers list
          followers.value = followers.value.filter(f => f.id !== followerId)
          // Also remove from allowedUserIDs (they lose access to private posts)
          allowedUserIDs.value = allowedUserIDs.value.filter(id => id !== followerId)
          // Clear any in-progress selections that reference the removed user
          selectedNewRecipients.value = selectedNewRecipients.value.filter(id => id !== followerId)
          selectedRemoveRecipients.value = selectedRemoveRecipients.value.filter(id => id !== followerId)
        } else if (data.type === "follower_added") {
          const { followingId } = data.payload
          // If the recipients modal is open and we're the post owner, refetch followers
          if (showAddRecipients.value && post.value && post.value.userId === followingId) {
            try {
              const followersRes = await fetch(
                `${API_BASE_URL}/api/users/${followingId}/followers`,
                { credentials: 'include' }
              )
              if (followersRes.ok) {
                const freshData = await followersRes.json()
                followers.value = Array.isArray(freshData) ? freshData : (freshData.followers || [])
              }
            } catch (err) {
              console.log("Failed to refetch followers after follower_added", err)
            }
          }
        }
  } catch (err) {
    console.log("Invalid WebSocket message", err)
  }
}
  }
  // ── Lifecycle ─────────────────────────────────────────────────────────────

  watch(post, () => { syncLikesFromPost() }, { immediate: true })

  onMounted(() => {
    fetchComments()
    connectWebSocket()
    window.addEventListener('keydown', onKeydown)
  })

  // fixing WebSocket never closed on unmount causing connection leak
  onUnmounted(() => {
    window.removeEventListener('keydown', onKeydown)
    socket.value?.close()
    socket.value = null
  })

  return {
    // Post
    post,
    loading,
    error,
    // Comments
    comments,
    commentsLoading,
    newComment,
    imageFile,
    submitting,
    commentError,
    // Lightbox
    lightboxSrc,
    // Methods
    initials,
    openLightbox,
    closeLightbox,
    onImagePicked,
    submitComment,
    deleteComment,
    fetchComments,
    // Add Recipients
    showAddRecipients,
    followers,
    allowedUserIDs,
    selectedNewRecipients,
    recipientsLoading,
    recipientsError,
    selectedRemoveRecipients,
    openAddRecipients,
    closeAddRecipients,
    submitNewRecipients,
    submitRemoveRecipients,
    // Replies
    replyToId,
    replyContent,
    replyError,
    replySubmitting,
    openReply,
    closeReply,
    submitReply,
    // Likes
    likeCount,
    dislikeCount,
    myReaction,
    likeError,
    handleReaction,
    // Re-exported utils so the template only imports from one place
    getImgURL,
    formatPostDate,
  }
}