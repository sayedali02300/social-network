import { onBeforeUnmount, onMounted, ref, computed, watch } from 'vue'
import type { Post } from '@/types/post'
import { API_BASE_URL } from '@/utils/helpers'
import { API_ROUTES } from '@/api/api'
import { buildWebSocketURL } from '@/api/websocket'

export type PrivacyFilter = 'all' | 'public' | 'almost_private' | 'private'

export function usePosts() {
    const posts = ref<Post[]>([])
    const loading = ref(false)
    const error = ref<string | null>(null)
    const privacyFilter = ref<PrivacyFilter>('all')
    const selectedPrivateAuthorId = ref<string | null>(null)

    // Extract unique authors from private posts for the sub-filter
    const privatePostAuthors = computed(() => {
        const authorMap = new Map<string, { id: string; firstName: string; lastName: string; nickname: string; avatar: string }>()
        for (const post of posts.value) {
            if (post.privacy === 'private' && !authorMap.has(post.userId)) {
                authorMap.set(post.userId, post.author)
            }
        }
        return Array.from(authorMap.values())
    })

    const filteredPosts = computed(() => {
        if (privacyFilter.value === 'all') return posts.value
        let result = posts.value.filter(p => p.privacy === privacyFilter.value)
        if (privacyFilter.value === 'private' && selectedPrivateAuthorId.value) {
            result = result.filter(p => p.userId === selectedPrivateAuthorId.value)
        }
        return result
    })

    // Reset author sub-filter when switching away from private
    watch(privacyFilter, (newVal) => {
        if (newVal !== 'private') {
            selectedPrivateAuthorId.value = null
        }
    })

    const socket = ref<WebSocket | null>(null)
    // fixing WebSocket reconnect loop continuing after component unmount
    let unmounted = false

    const connectWebSocket = () => {
        if (unmounted) return
        if (socket.value && socket.value.readyState === WebSocket.OPEN) return
        socket.value = new WebSocket(buildWebSocketURL('/ws'))
        socket.value.onopen = () => console.log("websocket connected")

        socket.value.onmessage = (event) => {
            try {
            const data = JSON.parse(event.data)

            if (data.type === "post_event") {
                const newPost = data.payload
                const exists = posts.value.find(p => p.id === newPost.id);
                 if (!exists) {
                    posts.value.unshift(newPost);
                }
            } else if (data.type === "delete_post") {
                posts.value = posts.value.filter(p => p.id !== data.payload.postId)
            }
        } catch(err) {
                console.log("Invalid WebSocket message", err)
            }
        }

        socket.value.onclose = () => {
            if (unmounted) return
            console.log('WebSocket disconnected, retrying in 3s...')
            setTimeout(connectWebSocket, 3000)
        }

         socket.value.onerror = (err) => {
            console.error('WebSocket error', err)
            socket.value?.close()
        }
    }

    const fetchPosts = async () => {
        loading.value = true
        error.value = null
        try {
            const response = await fetch(`${API_BASE_URL}${API_ROUTES.FEED}`, {
                credentials: 'include',
            })
            if (!response.ok) throw new Error('Failed to fetch feed')
            posts.value = await response.json()
        } catch (err) {
            error.value = err instanceof Error ? err.message : 'Unknown error'
        } finally {
            loading.value = false
        }
    }

    const deletePost = async (postId: string) => {
        try {
            const response = await fetch(`${API_BASE_URL}${API_ROUTES.POSTS}/${postId}`, {
                method: 'DELETE',
                credentials: 'include',
            })

            if (!response.ok) {
                const errData = await response.json()
                throw new Error(errData.error || 'Failed to delete post')
            }

            // Remove the deleted post from the list immediately
            posts.value = posts.value.filter(p => p.id !== postId)
            
        } catch (err) {
            console.error(err)
            alert(err instanceof Error ? err.message : "Could not delete post")
        }
    }
    onMounted(() => {
        fetchPosts()
        connectWebSocket()
    })
        
    onBeforeUnmount(() => {
        unmounted = true
        socket.value?.close()
        socket.value = null
    })
    
    return { posts, filteredPosts, loading, error, fetchPosts, deletePost, privacyFilter, selectedPrivateAuthorId, privatePostAuthors }
}