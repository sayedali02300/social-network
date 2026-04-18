import { API_BASE_URL,API_ROUTES } from '@/api/api'
import { ref, onMounted } from 'vue'
import type { Post } from '@/types/post.ts'
export function usePost(id: string) {
  const post = ref<Post | null>(null)
  const loading = ref(true)
  const error = ref(false)

  onMounted(async () => {
    try {
      const res = await fetch(`${API_BASE_URL}${API_ROUTES.POSTS}/${id}`, {
        method: 'GET',
        credentials: 'include',
    })

      if (!res.ok) {
        console.error(`Failed to load post: ${res.status} ${await res.text()}`)
        error.value = true
        return
      }

      post.value = await res.json()
    } catch (e) {
      console.error("usePost fetch error:", e)
      error.value = true
    } finally {
      loading.value = false
    }
  })

  return { post, loading, error }
}