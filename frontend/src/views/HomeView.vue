<template>
  <div class="home-root">
    <section class="home">
      <div class="PostsDiv">
        <section class="feed-shell">
          <header class="feed-hero">
            <div class="feed-hero-copy">
              <span class="feed-kicker">Home</span>
              <h1>Welcome back to our community </h1>
              <p>
                Stay updated, share posts, and customize your feed.
              </p>
            </div>

            <div class="feed-hero-meta">
              <article>
                <strong>{{ visiblePosts.length }}</strong>
                <span>Posts in view</span>
              </article>
              <article>
                <strong>{{ privacyFilterLabel }}</strong>
                <span>Current filter</span>
              </article>
            </div>
          </header>

          <div
            class="PostSomething"
            @mousemove="trackMouse($event)"
            @mouseleave="clearMouse($event)"
          >
            <div class="PostSomethingAvatarImg">
              <img
                v-if="currentUser?.avatar"
                :src="getImgURL(currentUser.avatar)"
                alt="Your avatar"
              >
              <div v-else class="composer-avatar-fallback" aria-label="Your avatar">
                {{ currentUserInitials }}
              </div>
            </div>

            <div class="composer-copy">
              <span class="composer-label">Create a post</span>
              <button @click="ClickToPost">Share an update with your network</button>
            </div>

            <button class="composer-action" type="button" @click="ClickToPost">Post</button>
          </div>

          <div v-if="BoolPost">
            <PostingDiv @close="handlePostCreated"></PostingDiv>
          </div>

          <section class="feed-search-bar">
            <label class="feed-search-input-wrap" for="feed-search">
              <svg class="feed-search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9" aria-hidden="true">
                <circle cx="11" cy="11" r="7"></circle>
                <path d="m20 20-3.5-3.5"></path>
              </svg>
              <input
                id="feed-search"
                v-model.trim="searchQuery"
                type="search"
                placeholder="Search posts, authors, or keywords"
              >
            </label>
            <button v-if="searchQuery" class="feed-search-clear" type="button" @click="searchQuery = ''">
              Clear
            </button>
          </section>

          <section class="feed-controls">
            <div class="privacy-filter-bar">
              <button
                v-for="opt in filterOptions"
                :key="opt.value"
                :class="['filter-btn', { active: privacyFilter === opt.value }]"
                @click="privacyFilter = opt.value"
              >
                {{ opt.label }}
              </button>
            </div>

            <div
              v-if="privacyFilter === 'private' && privatePostAuthors.length > 0"
              class="recipient-filter-bar"
            >
              <label class="recipient-label" for="private-author-filter">Private recipients</label>
              <select
                id="private-author-filter"
                v-model="selectedPrivateAuthorId"
                class="recipient-select"
              >
                <option :value="null">All Private Posts</option>
                <option
                  v-for="author in privatePostAuthors"
                  :key="author.id"
                  :value="author.id"
                >
                  {{ author.nickname || (author.firstName + ' ' + author.lastName) }}
                </option>
              </select>
            </div>
          </section>

          <div v-if="loading" class="NoPosts status-card"><p>Loading your feed...</p></div>
          <div v-if="error" class="NoPosts status-card"><p>Error fetching posts.</p></div>

          <div v-if="!loading && !error && visiblePosts.length === 0" class="NoPosts empty-state feed-empty-card">
            <img src="@/assets/empty-states/feed-empty.svg" alt="" class="empty-state-img" />
            <p>
              {{
                searchQuery
                  ? 'No posts match your search right now.'
                  : privacyFilter === 'all'
                  ? 'No posts yet. Be the first to share something.'
                  : 'No posts match this filter right now.'
              }}
            </p>
          </div>

          <div v-if="visiblePosts.length !== 0" class="AllPostCards">
            <article
              v-for="post in visiblePosts"
              :key="post.id"
              class="PostCard"
              @click="goToPost(post.id)"
              @mousemove="trackMouse($event)"
              @mouseleave="clearMouse($event)"
            >
              <div
                v-if="post.imagePath"
                class="PostThumb"
                @click.stop="openLightbox(getImgURL(post.imagePath))"
              >
                <img :src="getImgURL(post.imagePath)" alt="Post Image" />
              </div>

              <div class="HeadOfPost">
                <div class="AuthorInfo">
                  <img
                    v-if="post.author?.avatar"
                    class="AuthorPFP"
                    :src="getImgURL(post.author?.avatar)"
                    alt="Avatar"
                    @click.stop="router.push(`/users/${post.userId}`)"
                  />
                  <div
                    v-else
                    class="avatar-fallback"
                    @click.stop="router.push(`/users/${post.userId}`)"
                  >
                    {{ ((post.author.firstName?.[0] || '?') + (post.author.lastName?.[0] || '?')).toUpperCase() }}
                  </div>

                  <div class="followAndDate">
                    <div class="author-line">
                      <RouterLink class="AuthorName" :to="`/users/${post.userId}`" @click.stop>
                        {{ post.author?.nickname || (post.author?.firstName + ' ' + post.author?.lastName) }}
                      </RouterLink>
                      <span class="privacy-pill">{{ privacyLabel(post.privacy) }}</span>
                    </div>
                    <p class="PostDate">{{ formatPostDate(post.createdAt) }}</p>
                  </div>
                </div>

                <button
                  v-if="post.userId === currentUser?.id"
                  type="button"
                  class="delete-btn"
                  @click.stop="handleDelete(post.id)"
                >
                  Delete
                </button>
              </div>

              <div class="PostInfo">
                <p v-if="post.title" class="titleInfo">{{ post.title }}</p>

                <div
                  class="contentWrapper"
                  :ref="(el) => { if (el) wrapperRefs[post.id] = el as HTMLElement }"
                >
                  <p class="contentInfo">{{ post.content }}</p>
                </div>

                <div
                  v-if="(post.likeCount ?? 0) > 0 || (post.dislikeCount ?? 0) > 0"
                  class="post-reactions"
                >
                  <span v-if="(post.likeCount ?? 0) > 0">👍 {{ post.likeCount }}</span>
                  <span v-if="(post.dislikeCount ?? 0) > 0">👎 {{ post.dislikeCount }}</span>
                </div>
              </div>
            </article>
          </div>
        </section>
      </div>
    </section>

    <Teleport to="body">
      <div v-if="lightboxSrc" class="Lightbox" @click="closeLightbox">
        <button class="LightboxClose" @click="closeLightbox">✕</button>
        <img :src="lightboxSrc" alt="Full image" @click.stop />
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { usePosts, type PrivacyFilter } from '@/composables/usePosts'
import { getImgURL, formatPostDate, trackMouse, clearMouse, debounce } from '@/utils/helpers'
import PostingDiv from '@/components/PostingDiv.vue'
import { fetchSessionData } from '@/router/index.ts'
import type { SessionData, User } from '@/types/User'

const router = useRouter()

const goToPost = (postId: string) => {
  router.push(`/posts/${postId}`)
}

const {
  posts,
  filteredPosts,
  loading,
  error,
  fetchPosts,
  deletePost,
  privacyFilter,
  selectedPrivateAuthorId,
  privatePostAuthors,
} = usePosts()

const filterOptions: { value: PrivacyFilter; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'public', label: 'Public' },
  { value: 'almost_private', label: 'Followers Only' },
  { value: 'private', label: 'Private' },
]

const BoolPost = ref(false)
const lightboxSrc = ref<string | null>(null)
const currentUser = ref<User | null>(null)
const wrapperRefs = ref<Record<string, HTMLElement>>({})
const searchQuery = ref('')

const currentUserInitials = computed(() => {
  const first = currentUser.value?.firstName?.[0] || ''
  const last = currentUser.value?.lastName?.[0] || ''
  return `${first}${last}`.trim().toUpperCase() || 'U'
})

const privacyFilterLabel = computed(() => {
  return filterOptions.find((option) => option.value === privacyFilter.value)?.label || 'All'
})

const visiblePosts = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  if (!query) return filteredPosts.value

  return filteredPosts.value.filter((post) => {
    const authorName = `${post.author?.firstName || ''} ${post.author?.lastName || ''}`.trim().toLowerCase()
    const nickname = (post.author?.nickname || '').toLowerCase()
    const title = (post.title || '').toLowerCase()
    const content = (post.content || '').toLowerCase()
    const privacy = (post.privacy || '').toLowerCase()

    return (
      title.includes(query) ||
      content.includes(query) ||
      authorName.includes(query) ||
      nickname.includes(query) ||
      privacy.includes(query)
    )
  })
})

const privacyLabel = (value: string) => {
  switch (value) {
    case 'public':
      return 'Public'
    case 'almost_private':
      return 'Followers only'
    case 'private':
      return 'Private'
    default:
      return value
  }
}

onMounted(async () => {
  const sessionData: SessionData | null = await fetchSessionData()
  if (sessionData) {
    currentUser.value = sessionData.user
  }
})

const updateLineClamps = () => {
  const wrappers = Object.values(wrapperRefs.value)
  if (!wrappers.length) return

  wrappers.forEach((wrapper) => {
    const content = wrapper.querySelector('.contentInfo') as HTMLElement | null
    if (!content) return

    const availableHeight = wrapper.clientHeight
    const computedStyle = window.getComputedStyle(content)
    const lineHeight = parseFloat(computedStyle.lineHeight)

    const linesThatFit = Math.floor(availableHeight / lineHeight)
    content.style.webkitLineClamp = linesThatFit > 0 ? linesThatFit.toString() : '1'
  })
}

const debouncedUpdateLineClamps = debounce(updateLineClamps, 200)

watch(
  () => posts.value,
  async () => {
    await nextTick()
    updateLineClamps()
  },
  { deep: true },
)

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

const ClickToPost = () => {
  BoolPost.value = true
}

const handlePostCreated = () => {
  BoolPost.value = false
  fetchPosts()
}

const handleDelete = async (postId: string) => {
  if (confirm('Are you sure you want to delete this post?')) {
    await deletePost(postId)
  }
}

onMounted(() => {
  fetchPosts()
  window.addEventListener('keydown', onKeydown)
  window.addEventListener('resize', debouncedUpdateLineClamps)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown)
  window.removeEventListener('resize', debouncedUpdateLineClamps)
})
</script>

<style scoped>
.home-root {
  display: block;
}

.home {
  min-height: calc(100dvh - var(--navbar-height, 60px));
}

.PostsDiv {
  width: 100%;
  min-height: calc(100dvh - 150px);
  overflow-y: auto;
  padding: 0.25rem 0 1.75rem;
}

.feed-shell {
  width: min(100%, 1320px);
  margin: 0 auto;
  display: grid;
  gap: 1rem;
}

.feed-hero {
  display: grid;
  grid-template-columns: minmax(0, 1.15fr) minmax(260px, 0.85fr);
  gap: 1rem;
  padding: 1.5rem;
  border-radius: 32px;
  background:
    radial-gradient(circle at top left, rgba(59, 130, 246, 0.2), transparent 28%),
    linear-gradient(135deg, rgba(8, 18, 37, 0.96), rgba(15, 70, 180, 0.96));
  box-shadow: 0 28px 70px rgba(15, 23, 42, 0.18);
  color: var(--white);
  overflow: hidden;
  position: relative;
}

.feed-hero::before,
.feed-hero::after {
  content: '';
  position: absolute;
  border-radius: 999px;
  pointer-events: none;
}

.feed-hero::before {
  width: 220px;
  height: 220px;
  right: -40px;
  top: -60px;
  background: rgba(255, 255, 255, 0.08);
}

.feed-hero::after {
  width: 160px;
  height: 160px;
  left: -40px;
  bottom: -60px;
  background: rgba(125, 211, 252, 0.12);
}

.feed-hero-copy,
.feed-hero-meta {
  position: relative;
  z-index: 1;
}

.feed-kicker {
  display: inline-flex;
  align-items: center;
  padding: 0.38rem 0.78rem;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.12);
  color: rgba(255, 255, 255, 0.84);
  font-size: 0.76rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.feed-hero h1 {
  margin-top: 1rem;
  max-width: 11ch;
  font-size: clamp(2.2rem, 5vw, 3.5rem);
  line-height: 0.96;
  font-weight: 900;
  color: var(--white);
}

.feed-hero p {
  margin-top: 0.9rem;
  max-width: 46ch;
  color: rgba(255, 255, 255, 0.82);
  font-size: 1rem;
}

.feed-hero-meta {
  display: grid;
  gap: 0.85rem;
  align-content: end;
}

.feed-hero-meta article {
  padding: 1rem 1.05rem;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.12);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.feed-hero-meta strong {
  display: block;
  color: var(--white);
  font-size: 1.55rem;
  font-weight: 900;
}

.feed-hero-meta span {
  color: rgba(255, 255, 255, 0.76);
  font-size: 0.86rem;
  font-weight: 700;
}

.feed-search-bar,
.privacy-filter-bar,
.recipient-filter-bar {
  width: 100%;
}

.PostSomething {
  position: relative;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 1rem;
  width: 100%;
  padding: 1rem 1.1rem;
  border-radius: 28px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(248, 250, 252, 0.86));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.08);
  overflow: hidden;
}

.PostSomething::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle 200px at var(--mx, -999px) var(--my, -999px), rgba(59, 130, 246, 0.12), transparent 58%);
  opacity: 0.85;
  pointer-events: none;
}

.PostSomethingAvatarImg,
.composer-copy,
.composer-action {
  position: relative;
  z-index: 1;
}

.PostSomethingAvatarImg {
  width: 3.15rem;
  height: 3.15rem;
  border-radius: 50%;
  overflow: hidden;
  box-shadow: 0 12px 22px rgba(37, 99, 235, 0.12);
}

.PostSomethingAvatarImg img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.composer-avatar-fallback {
  width: 100%;
  height: 100%;
  display: grid;
  place-items: center;
  border-radius: 50%;
  background: linear-gradient(135deg, rgba(191, 219, 254, 0.95), rgba(219, 234, 254, 0.9));
  color: var(--brand-700);
  font-weight: 900;
  font-size: 1.3rem;
  border: 3px solid rgba(255, 255, 255, 0.96);
}

.composer-copy {
  display: grid;
  gap: 0.24rem;
}

.composer-label {
  color: var(--gray-900);
  font-size: 0.76rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.composer-copy button {
  padding: 0;
  border: none;
  background: transparent;
  color: var(--gray-500);
  font-size: 1rem;
  text-align: left;
  cursor: text;
}

.composer-action {
  border: none;
  padding: 0.88rem 1.25rem;
  border-radius: 18px;
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  color: var(--white);
  font-size: 0.92rem;
  font-weight: 800;
  cursor: pointer;
  box-shadow: 0 16px 32px rgba(37, 99, 235, 0.24);
  transition:
    transform var(--dur-fast) var(--ease-standard),
    box-shadow var(--dur-fast) var(--ease-standard);
}

.composer-action:hover {
  transform: translateY(-1px);
  box-shadow: 0 20px 34px rgba(37, 99, 235, 0.3);
}

.feed-controls {
  display: grid;
  gap: 0.8rem;
}

.feed-search-bar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 0.75rem;
  align-items: center;
  padding: 0.75rem;
  border-radius: 24px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(239, 246, 255, 0.76));
  border: 1px solid rgba(148, 163, 184, 0.16);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 12px 28px rgba(15, 23, 42, 0.05);
}

.feed-search-input-wrap {
  display: flex;
  align-items: center;
  gap: 0.72rem;
  min-width: 0;
  padding: 0.95rem 1rem;
  border-radius: 18px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 8px 22px rgba(15, 23, 42, 0.04);
}

.feed-search-icon {
  width: 1.1rem;
  height: 1.1rem;
  flex-shrink: 0;
  color: var(--gray-500);
}

.feed-search-input-wrap input {
  width: 100%;
  min-width: 0;
  border: none;
  outline: none;
  background: transparent;
  color: var(--gray-800);
  font: inherit;
  font-size: 0.95rem;
  font-weight: 600;
}

.feed-search-input-wrap:focus-within {
  border-color: rgba(37, 99, 235, 0.55);
  box-shadow:
    0 0 0 5px rgba(37, 99, 235, 0.1),
    0 16px 34px rgba(37, 99, 235, 0.12);
}

.feed-search-clear {
  border: 1px solid rgba(148, 163, 184, 0.16);
  padding: 0.88rem 1.15rem;
  border-radius: 18px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(239, 246, 255, 0.92));
  color: var(--gray-700);
  font-size: 0.9rem;
  font-weight: 800;
  cursor: pointer;
}

.privacy-filter-bar {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.5rem;
  padding: 0.45rem;
  border-radius: 24px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.92), rgba(239, 246, 255, 0.72));
  border: 1px solid rgba(148, 163, 184, 0.16);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 12px 28px rgba(15, 23, 42, 0.05);
}

.filter-btn {
  padding: 0.82rem 0.8rem;
  border-radius: 18px;
  border: 1px solid transparent;
  background: transparent;
  color: var(--text-secondary);
  font-size: 0.9rem;
  font-weight: 800;
  cursor: pointer;
  transition:
    background var(--dur-base) var(--ease-standard),
    color var(--dur-base) var(--ease-standard),
    box-shadow var(--dur-base) var(--ease-standard),
    transform var(--dur-base) var(--ease-standard);
}

.filter-btn:hover {
  transform: translateY(-1px);
  color: var(--gray-900);
}

.filter-btn.active {
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.14), rgba(14, 165, 233, 0.08));
  color: var(--brand-700);
  box-shadow: 0 12px 24px rgba(37, 99, 235, 0.08);
}

.recipient-filter-bar {
  display: grid;
  gap: 0.42rem;
}

.recipient-label {
  color: var(--gray-700);
  font-size: 0.76rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.recipient-select {
  width: 100%;
  padding: 0.95rem 1rem;
  border-radius: 18px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  color: var(--gray-700);
  font-size: 0.92rem;
  font-weight: 600;
  outline: none;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 8px 22px rgba(15, 23, 42, 0.04);
}

.recipient-select:focus {
  border-color: rgba(37, 99, 235, 0.55);
  box-shadow:
    0 0 0 5px rgba(37, 99, 235, 0.1),
    0 16px 34px rgba(37, 99, 235, 0.12);
}

.AllPostCards {
  display: grid;
  gap: 1rem;
}

.PostCard {
  width: 100%;
  border-radius: 28px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.88));
  border: 1px solid rgba(148, 163, 184, 0.16);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.08);
  overflow: hidden;
  position: relative;
  isolation: isolate;
  cursor: pointer;
  transition:
    transform var(--dur-base) var(--ease-standard),
    box-shadow var(--dur-base) var(--ease-standard),
    border-color var(--dur-base) var(--ease-standard);
}

.PostCard::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle 220px at var(--mx, -999px) var(--my, -999px), rgba(59, 130, 246, 0.1), transparent 58%);
  opacity: 0;
  transition: opacity var(--dur-base) var(--ease-standard);
  pointer-events: none;
}

.PostCard:hover {
  transform: translateY(-2px);
  border-color: rgba(96, 165, 250, 0.22);
  box-shadow: 0 26px 58px rgba(15, 23, 42, 0.12);
}

.PostCard:hover::before {
  opacity: 1;
}

.PostThumb {
  width: 100%;
  height: 270px;
  overflow: hidden;
  background: #dbeafe;
}

.PostThumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 300ms ease;
}

.PostThumb:hover img {
  transform: scale(1.03);
}

.HeadOfPost {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  padding: 1.2rem 1.25rem 0;
}

.AuthorInfo {
  display: flex;
  align-items: center;
  gap: 0.9rem;
}

.AuthorPFP,
.avatar-fallback {
  width: 3rem;
  height: 3rem;
  border-radius: 50%;
  flex-shrink: 0;
}

.AuthorPFP {
  object-fit: cover;
}

.avatar-fallback {
  display: grid;
  place-items: center;
  background: var(--brand-100);
  color: var(--brand-600);
  font-weight: 800;
  font-size: 0.95rem;
}

.followAndDate {
  display: grid;
  gap: 0.3rem;
}

.author-line {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  flex-wrap: wrap;
}

.AuthorName {
  color: var(--gray-900);
  text-decoration: none;
  font-weight: 800;
}

.AuthorName:hover {
  color: var(--brand-600);
}

.privacy-pill {
  display: inline-flex;
  align-items: center;
  padding: 0.22rem 0.55rem;
  border-radius: 999px;
  background: rgba(37, 99, 235, 0.1);
  color: var(--brand-700);
  font-size: 0.72rem;
  font-weight: 800;
}

.PostDate {
  color: var(--gray-500);
  font-size: 0.84rem;
  font-weight: 600;
}

.delete-btn {
  border: 1px solid rgba(239, 68, 68, 0.38);
  background: transparent;
  color: #dc2626;
  border-radius: 999px;
  padding: 0.45rem 0.85rem;
  font-size: 0.8rem;
  font-weight: 700;
  cursor: pointer;
  transition:
    transform var(--dur-fast) var(--ease-standard),
    background-color var(--dur-fast) var(--ease-standard),
    color var(--dur-fast) var(--ease-standard),
    box-shadow var(--dur-fast) var(--ease-standard);
}

.delete-btn:hover {
  transform: translateY(-1px);
  background: #dc2626;
  color: var(--white);
  box-shadow: 0 10px 22px rgba(220, 38, 38, 0.2);
}

.PostInfo {
  padding: 0.95rem 1.25rem 1.2rem;
  display: flex;
  flex-direction: column;
}

.titleInfo {
  margin: 0 0 0.8rem;
  padding: 0.82rem 0.95rem;
  border-radius: 18px;
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.1), rgba(14, 165, 233, 0.05));
  color: var(--brand-700);
  font-weight: 800;
  font-size: 1rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.contentWrapper {
  overflow: hidden;
}

.contentInfo {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.65;
  word-break: break-word;
  margin: 0;
  color: var(--gray-700);
}

.post-reactions {
  display: inline-flex;
  align-items: center;
  gap: 0.9rem;
  margin-top: 1rem;
  padding-top: 0.95rem;
  border-top: 1px solid rgba(148, 163, 184, 0.14);
  color: var(--gray-600);
  font-size: 0.88rem;
  font-weight: 700;
}

.status-card,
.feed-empty-card {
  width: 100%;
  min-height: 260px;
  border-radius: 32px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.9), rgba(248, 250, 252, 0.82));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.06);
}

.NoPosts {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  text-align: center;
  padding: 2rem 1rem;
}

.NoPosts p {
  background: transparent;
  color: var(--gray-600);
  padding: 0;
  font-weight: 700;
  max-width: 28ch;
}

.feed-empty-card :deep(.empty-state-img),
.empty-state-img {
  width: 120px;
  height: 120px;
}

.Lightbox {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.88);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  animation: fadeIn 0.2s ease;
  cursor: zoom-out;
}

.Lightbox img {
  max-width: 90vw;
  max-height: 88vh;
  object-fit: contain;
  border-radius: 10px;
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.6);
}

.LightboxClose {
  position: absolute;
  top: 1.2rem;
  right: 1.4rem;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.25);
  color: #fff;
  font-size: 1.1rem;
  width: 2.2rem;
  height: 2.2rem;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@media (max-width: 899px) {
  .feed-hero {
    grid-template-columns: 1fr;
  }

  .feed-hero h1 {
    max-width: 13ch;
  }
}

@media (max-width: 599px) {
  .PostsDiv {
    padding-bottom: 1rem;
  }

  .feed-shell {
    gap: 0.8rem;
  }

  .feed-hero,
  .PostSomething,
  .PostCard,
  .status-card,
  .feed-empty-card {
    border-radius: 24px;
  }

  .feed-hero,
  .PostSomething,
  .HeadOfPost,
  .PostInfo {
    padding-left: 1rem;
    padding-right: 1rem;
  }

  .PostSomething {
    grid-template-columns: auto 1fr;
  }

  .composer-action {
    grid-column: 1 / -1;
    width: 100%;
  }

  .privacy-filter-bar {
    grid-template-columns: 1fr 1fr;
  }

  .PostThumb {
    height: 200px;
  }
}
</style>
