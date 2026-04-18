<template>
  <section class="x-shell">

    <!-- Loading -->
    <div v-if="loading" class="x-state">
      <span class="loader"></span>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="x-state">
      <p class="x-error-text">Couldn't load this post.</p>
      <RouterLink to="/" class="x-back-link">← Back to feed</RouterLink>
    </div>

    <!-- Post -->
    <template v-else-if="post">
      <header class="x-hero">
        <div class="x-hero-copy">
          <span class="x-eyebrow">Conversation view</span>
          <h1>Follow the full post and reply flow.</h1>
          <p>Read the post, react, and reply from one focused view without changing how anything works.</p>
        </div>
        <div class="x-hero-meta">
          <article class="x-hero-card">
            <strong>{{ comments.length }}</strong>
            <span>Replies</span>
          </article>
          <article class="x-hero-card">
            <strong>{{ likeCount + dislikeCount }}</strong>
            <span>Reactions</span>
          </article>
        </div>
      </header>

      <div class="x-column">

      <!-- Top bar -->
      <div class="x-topbar">
        <RouterLink to="/" class="x-back-btn" aria-label="Back">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 12H5M12 5l-7 7 7 7"/></svg>
        </RouterLink>
        <span class="x-topbar-title">Post</span>
      </div>

      <!-- Author row -->
      <div class="x-author-row">
        <RouterLink :to="`/users/${post.userId}`">
          <img v-if="post.author?.avatar" :src="getImgURL(post.author.avatar)" class="x-avatar" alt="avatar" />
          <div v-else class="x-avatar x-avatar-fallback">{{ initials(post.author) }}</div>
        </RouterLink>
        <div class="x-author-meta">
          <RouterLink :to="`/users/${post.userId}`" class="x-display-name">
            {{ post.author?.nickname || ((post.author?.firstName || '') + ' ' + (post.author?.lastName || '')).trim() }}
          </RouterLink>
          <span class="x-handle" v-if="post.author?.nickname">@{{ post.author.nickname }}</span>
          <span class="x-handle" v-else>@{{ ((post.author?.firstName || '') + (post.author?.lastName || '')).toLowerCase().replace(/\s+/g,'') }}</span>
        </div>
      </div>

      <!-- Post title -->
      <p v-if="post.title" class="x-post-title">{{ post.title }}</p>

      <!-- Post body -->
      <div class="x-post-body">{{ post.content }}</div>

      <!-- Post image -->
      <div v-if="post.imagePath" class="x-post-image-wrap" @click="openLightbox(getImgURL(post.imagePath))">
        <img :src="getImgURL(post.imagePath)" alt="Post image" class="x-post-image" />
      </div>

      <!-- Timestamp -->
      <div class="x-timestamp">{{ formatPostDate(post.createdAt) }}</div>

      <div class="x-divider"></div>

      <!-- Engagement stats -->
      <div class="x-stats-row">
        <span class="x-stat"><strong>{{ comments.length }}</strong> Comment{{ comments.length !== 1 ? 's' : '' }}</span>
        <span class="x-stat"><strong>{{ likeCount }}</strong> Like{{ likeCount !== 1 ? 's' : '' }}</span>
        <span class="x-stat"><strong>{{ dislikeCount }}</strong> Dislike{{ dislikeCount !== 1 ? 's' : '' }}</span>
      </div>

      <div class="x-divider"></div>

      <!-- Action bar -->
      <div class="x-action-bar">
        <button class="x-action-btn" title="Comment" @click="focusCommentForm">
          <svg class="x-action-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
          <span class="x-action-count">{{ comments.length }}</span>
        </button>
        <button
          class="x-action-btn like-btn"
          :class="{ 'x-active-like': myReaction === 1 }"
          title="Like"
          @click="handleReaction(1)"
        >
          <svg class="x-action-icon" viewBox="0 0 24 24" :fill="myReaction === 1 ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="1.8"><path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"/></svg>
          <span class="x-action-count">{{ likeCount }}</span>
        </button>
        <button
          class="x-action-btn dislike-btn"
          :class="{ 'x-active-dislike': myReaction === -1 }"
          title="Dislike"
          @click="handleReaction(-1)"
        >
          <svg class="x-action-icon" viewBox="0 0 24 24" :fill="myReaction === -1 ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="1.8"><path d="M10 15v4a3 3 0 0 0 3 3l4-9V2H5.72a2 2 0 0 0-2 1.7l-1.38 9a2 2 0 0 0 2 2.3H10z"/><path d="M17 2h2.67A2.31 2.31 0 0 1 22 4v7a2.31 2.31 0 0 1-2.33 2H17"/></svg>
          <span class="x-action-count">{{ dislikeCount }}</span>
        </button>
        <div v-if="likeError" class="x-like-error">{{ likeError }}</div>
      </div>

      <!-- Manage recipients -->
      <div v-if="post.privacy === 'private' && !post.groupId && post.userId === currentUser?.id" class="x-recipients-row">
        <button class="x-recipients-btn" @click="openAddRecipients(post.userId)">Manage Recipients</button>
      </div>

      <div class="x-divider"></div>

      <!-- Reply composer (comment form) -->
      <form class="x-composer" @submit.prevent="submitComment" ref="composerRef">
        <img v-if="currentUser?.avatar" :src="getImgURL(currentUser.avatar)" class="x-avatar x-avatar-sm" alt="Your avatar" />
        <div v-else class="x-avatar x-avatar-sm x-avatar-fallback">{{ currentUser ? ((currentUser.firstName?.[0] || '') + (currentUser.lastName?.[0] || '')).toUpperCase() : '?' }}</div>
        <div class="x-composer-body">
          <textarea
            v-model="newComment"
            class="x-composer-input"
            placeholder="Post your reply…"
            rows="2"
            maxlength="2500"
            :disabled="submitting"
          ></textarea>
          <div class="x-composer-footer">
            <label class="x-attach-btn" title="Attach image">
              <input type="file" accept="image/*" @change="onImagePicked" hidden />
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>
              <span v-if="imageFile" class="x-file-name">{{ imageFile.name }}</span>
            </label>
            <button class="x-reply-post-btn" type="submit" :disabled="(!newComment.trim() && !imageFile) || submitting">
              <span v-if="!submitting">Reply</span>
              <span v-else class="loader btn-loader"></span>
            </button>
          </div>
          <p v-if="commentError" class="x-form-error">{{ commentError }}</p>
        </div>
      </form>

      <div class="x-divider"></div>

      <!-- Comments list -->
      <div v-if="commentsLoading" class="x-state">
        <span class="loader"></span>
      </div>
      <div v-else-if="comments.length === 0" class="x-empty-comments">
        No replies yet. Be the first!
      </div>
      <template v-else>
        <div v-for="c in topLevelComments" :key="c.id" class="x-comment-thread">
          <!-- Top-level comment -->
          <div class="x-comment-row">
            <div class="x-comment-left">
              <RouterLink :to="`/users/${c.author.user_id}`">
                <img v-if="c.author?.avatar" :src="getImgURL(c.author.avatar)" class="x-avatar x-avatar-sm" alt="avatar" />
                <div v-else class="x-avatar x-avatar-sm x-avatar-fallback">{{ ((c.author.firstName?.[0] || '?') + (c.author.lastName?.[0] || '?')).toUpperCase() }}</div>
              </RouterLink>
              <div v-if="repliesFor(c.id).length > 0" class="x-thread-line"></div>
            </div>
            <div class="x-comment-content">
              <div class="x-comment-header">
                <RouterLink :to="`/users/${c.author.user_id}`" class="x-comment-name">
                  {{ c.author?.nickname || ((c.author?.firstName || '') + ' ' + (c.author?.lastName || '')).trim() }}
                </RouterLink>
                <span class="x-comment-handle" v-if="c.author?.nickname">@{{ c.author.nickname }}</span>
                <span class="x-comment-dot">·</span>
                <span class="x-comment-time">{{ formatPostDate(c.createdAt) }}</span>
                <button v-if="c.author.user_id === currentUser?.id" class="x-delete-btn" @click="handleDeleteComment(c.id)">Delete</button>
              </div>
              <p class="x-comment-text">{{ c.content }}</p>
              <div v-if="c.imagePath" class="x-comment-image-wrap">
                <img :src="getImgURL(c.imagePath)" class="x-comment-image" alt="comment image" @click="openLightbox(getImgURL(c.imagePath))" />
              </div>
              <div class="x-comment-actions">
                <button class="x-cmt-action-btn" @click="replyToId === c.id ? closeReply() : openReply(c.id)">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
                  <span>{{ replyToId === c.id ? 'Cancel' : 'Reply' }}</span>
                </button>
              </div>
              <!-- Inline reply form -->
              <div v-if="replyToId === c.id" class="x-inline-reply">
                <p v-if="replyError" class="x-form-error">{{ replyError }}</p>
                <textarea v-model="replyContent" class="x-reply-input" placeholder="Post your reply…" rows="2" maxlength="2500" :disabled="replySubmitting"></textarea>
                <div class="x-inline-reply-footer">
                  <button class="x-reply-post-btn" :disabled="!replyContent.trim() || replySubmitting" @click="submitReply">
                    <span v-if="!replySubmitting">Reply</span>
                    <span v-else>…</span>
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Replies -->
          <div v-for="reply in repliesFor(c.id)" :key="reply.id" class="x-comment-row x-reply-row">
            <div class="x-comment-left">
              <RouterLink :to="`/users/${reply.author.user_id}`">
                <img v-if="reply.author?.avatar" :src="getImgURL(reply.author.avatar)" class="x-avatar x-avatar-xs" alt="avatar" />
                <div v-else class="x-avatar x-avatar-xs x-avatar-fallback">{{ ((reply.author.firstName?.[0] || '?') + (reply.author.lastName?.[0] || '?')).toUpperCase() }}</div>
              </RouterLink>
            </div>
            <div class="x-comment-content">
              <div class="x-comment-header">
                <RouterLink :to="`/users/${reply.author.user_id}`" class="x-comment-name">
                  {{ reply.author?.nickname || ((reply.author?.firstName || '') + ' ' + (reply.author?.lastName || '')).trim() }}
                </RouterLink>
                <span class="x-comment-handle" v-if="reply.author?.nickname">@{{ reply.author.nickname }}</span>
                <span class="x-comment-dot">·</span>
                <span class="x-comment-time">{{ formatPostDate(reply.createdAt) }}</span>
                <button v-if="reply.author.user_id === currentUser?.id" class="x-delete-btn" @click="handleDeleteComment(reply.id)">Delete</button>
              </div>
              <p class="x-comment-text">{{ reply.content }}</p>
              <div v-if="reply.imagePath" class="x-comment-image-wrap">
                <img :src="getImgURL(reply.imagePath)" class="x-comment-image" alt="reply image" @click="openLightbox(getImgURL(reply.imagePath))" />
              </div>
            </div>
          </div>

          <div class="x-divider"></div>
        </div>
      </template>

      </div><!-- end x-column -->
    </template>

    <!-- Add Recipients Modal -->
    <Teleport to="body">
      <div v-if="showAddRecipients" class="form-overlay" @click.self="closeAddRecipients">
        <div class="form-container recipients-modal">
          <h3>Manage Recipients</h3>
          <div v-if="recipientsLoading" class="loading-text">Loading followers...</div>
          <div v-else-if="recipientsError" class="x-form-error">{{ recipientsError }}</div>
          <div v-else-if="followers.length === 0" class="loading-text">You don't have any followers</div>
          <div v-else>
            <div v-if="currentRecipients.length > 0" class="recipients-section">
              <h4 class="section-label">Current Recipients</h4>
              <div class="followers-container">
                <label v-for="f in currentRecipients" :key="f.id" class="follower-option">
                  <div class="follower-info-wrapper">
                    <img v-if="f.avatar" :src="getImgURL(f.avatar)" class="follower-avatar" alt="" />
                    <div v-else class="avatar-fallback sm">{{ ((f.firstName?.[0] || '?') + (f.lastName?.[0] || '?')).toUpperCase() }}</div>
                    <span class="follower-name">{{ f.nickname || (f.firstName + ' ' + f.lastName) }}</span>
                  </div>
                  <input type="checkbox" :value="f.id" v-model="selectedRemoveRecipients" class="follower-checkbox" />
                </label>
              </div>
              <button class="remove-recipients-btn" :disabled="selectedRemoveRecipients.length === 0" @click="submitRemoveRecipients">Remove Selected</button>
            </div>
            <div class="recipients-section">
              <h4 class="section-label">Add Followers</h4>
              <div class="followers-container">
                <div v-if="availableFollowers.length === 0" class="loading-text">All your followers already have access.</div>
                <label v-for="f in availableFollowers" :key="f.id" class="follower-option">
                  <div class="follower-info-wrapper">
                    <img v-if="f.avatar" :src="getImgURL(f.avatar)" class="follower-avatar" alt="" />
                    <div v-else class="avatar-fallback sm">{{ ((f.firstName?.[0] || '?') + (f.lastName?.[0] || '?')).toUpperCase() }}</div>
                    <span class="follower-name">{{ f.nickname || (f.firstName + ' ' + f.lastName) }}</span>
                  </div>
                  <input type="checkbox" :value="f.id" v-model="selectedNewRecipients" class="follower-checkbox" />
                </label>
              </div>
              <button class="submit-btn" style="margin-top:0.5em" :disabled="selectedNewRecipients.length === 0" @click="submitNewRecipients">Add Selected</button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Lightbox -->
    <Teleport to="body">
      <div v-if="lightboxSrc" class="Lightbox" @click="closeLightbox">
        <button class="LightboxClose" @click="closeLightbox">✕</button>
        <img :src="lightboxSrc" alt="Full image" @click.stop />
      </div>
    </Teleport>

  </section>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import { usePostPageLogic } from '@/composables/usePostPageLogic'
import type { User, SessionData} from '@/types/User'
import type { Comment } from '@/types/comment'
import { fetchSessionData } from '@/router'
import { onMounted, ref, computed } from 'vue'

const route  = useRoute()
const postId = route.params.id as string

const currentUser = ref<User | null>(null)

onMounted(async () => {
  const sessionData: SessionData | null = await fetchSessionData()
  if (sessionData) {
    currentUser.value = sessionData.user
    console.log('Logged-in user:', currentUser.value)
  }
})

const handleDeleteComment = async (commentId: string) => {
  if (!confirm('Are you sure you want to delete this comment?')) return

  const success = await deleteComment(commentId)
  if (success) {
    await fetchComments();
    console.log('Comment deleted')
  } else {
    alert('Failed to delete comment')
  }
}

const {
  post, loading, error,
  comments, commentsLoading,
  newComment, imageFile, submitting, commentError,
  lightboxSrc,
  initials, openLightbox, closeLightbox,
  onImagePicked, submitComment,
  getImgURL, formatPostDate, deleteComment, fetchComments,
  showAddRecipients, followers, allowedUserIDs, selectedNewRecipients,
  selectedRemoveRecipients,
  recipientsLoading, recipientsError,
  openAddRecipients, closeAddRecipients, submitNewRecipients, submitRemoveRecipients,
  likeCount, dislikeCount, myReaction, likeError, handleReaction,
  replyToId, replyContent, replyError, replySubmitting, openReply, closeReply, submitReply,
} = usePostPageLogic(postId)

const availableFollowers = computed(() =>
  followers.value.filter(f => !allowedUserIDs.value.includes(f.id))
)

const currentRecipients = computed(() =>
  followers.value.filter(f => allowedUserIDs.value.includes(f.id))
)

// Group comments: top-level (no parentId) and replies (has parentId)
const topLevelComments = computed(() =>
  comments.value.filter((c: Comment) => !c.parentId)
)
const repliesFor = (commentId: string) =>
  comments.value.filter((c: Comment) => c.parentId === commentId)

const composerRef = ref<HTMLElement | null>(null)
const focusCommentForm = () => {
  composerRef.value?.querySelector('textarea')?.focus()
}
</script>

<style scoped>
/* ── Shell ── */
.x-shell {
  min-height: 100vh;
  width: min(100%, 1320px);
  margin: 0 auto;
  padding: 0.25rem 0 2rem;
  display: grid;
  gap: 1.25rem;
  background: transparent;
}

.x-hero {
  position: relative;
  overflow: hidden;
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(260px, 0.8fr);
  gap: 1rem;
  padding: 1.8rem 1.95rem;
  border-radius: 2rem;
  border: 1px solid rgba(96, 165, 250, 0.36);
  background:
    radial-gradient(circle at top right, rgba(96, 165, 250, 0.28), transparent 32%),
    radial-gradient(circle at bottom left, rgba(147, 197, 253, 0.22), transparent 24%),
    linear-gradient(135deg, #17306b 0%, #1f4eb6 100%);
  box-shadow: 0 34px 70px rgba(37, 99, 235, 0.26);
}

.x-hero::after {
  content: "";
  position: absolute;
  top: -3rem;
  right: -2rem;
  width: 14rem;
  height: 14rem;
  border-radius: 999px;
  background: rgba(147, 197, 253, 0.14);
}

.x-hero-copy,
.x-hero-meta {
  position: relative;
  z-index: 1;
}

.x-hero-copy {
  display: grid;
  gap: 0.75rem;
}

.x-eyebrow {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  border-radius: 999px;
  padding: 0.65rem 1rem;
  background: rgba(219, 234, 254, 0.18);
  border: 1px solid rgba(191, 219, 254, 0.16);
  color: #eff6ff;
  font-size: 0.82rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  backdrop-filter: blur(12px);
}

.x-hero h1 {
  margin: 0;
  font-size: clamp(2.3rem, 4vw, 4.2rem);
  line-height: 0.98;
  letter-spacing: -0.04em;
  color: #fff;
  text-wrap: balance;
}

.x-hero p {
  max-width: 42rem;
  margin: 0;
  font-size: 1.05rem;
  line-height: 1.75;
  color: rgba(239, 246, 255, 0.86);
}

.x-hero-meta {
  display: grid;
  align-content: end;
  gap: 0.9rem;
}

.x-hero-card {
  display: grid;
  gap: 0.25rem;
  padding: 1.25rem 1.15rem;
  border-radius: 1.5rem;
  border: 1px solid rgba(147, 197, 253, 0.26);
  background: rgba(255, 255, 255, 0.12);
  color: #eff6ff;
  backdrop-filter: blur(14px);
}

.x-hero-card strong {
  font-size: 2rem;
  font-weight: 800;
}

.x-hero-card span {
  font-size: 0.92rem;
  font-weight: 700;
  color: rgba(219, 234, 254, 0.9);
}

.x-column {
  width: 100%;
  min-height: 100vh;
  border-radius: 2rem;
  border: 1px solid rgba(191, 219, 254, 0.72);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.985), rgba(241, 247, 255, 0.94));
  box-shadow: 0 26px 60px rgba(148, 163, 184, 0.2);
  overflow: hidden;
}

.x-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  padding: 4rem 1rem;
  width: 100%;
}

.x-error-text { color: var(--status-error, #dc2626); font-weight: 600; }
.x-back-link { color: var(--brand-500); text-decoration: none; font-size: 0.9rem; }

/* ── Top bar ── */
.x-topbar {
  display: flex;
  align-items: center;
  gap: 1.25rem;
  padding: 1.05rem 1.1rem;
  z-index: 10;
  background: rgba(255,255,255,0.72);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid rgba(191, 219, 254, 0.7);
}

.x-back-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.8rem;
  height: 2.8rem;
  border-radius: 999px;
  color: var(--text-primary);
  background: rgba(239, 246, 255, 0.78);
  border: 1px solid rgba(191, 219, 254, 0.8);
  transition: background 0.15s, transform 0.15s;
  text-decoration: none;
}
.x-back-btn:hover { background: rgba(219, 234, 254, 0.92); transform: translateY(-1px); }

.x-topbar-title {
  font-size: 1.1rem;
  font-weight: 800;
  color: var(--text-primary);
}

/* ── Author row ── */
.x-author-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1.35rem 1.2rem 0.5rem;
}

.x-author-meta {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
}

.x-display-name {
  font-size: 0.98rem;
  font-weight: 800;
  color: var(--text-primary);
  text-decoration: none;
}
.x-display-name:hover { text-decoration: underline; }

.x-handle {
  font-size: 0.85rem;
  color: var(--text-secondary, #6b7280);
}

/* ── Avatar ── */
.x-avatar {
  border-radius: 50%;
  flex-shrink: 0;
  object-fit: cover;
}
.x-avatar-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--brand-100, #dbeafe);
  color: var(--brand-600, #2563eb);
  font-weight: 700;
  font-size: 0.8rem;
  letter-spacing: 0.03em;
}
/* sizes */
.x-author-row .x-avatar { width: 48px; height: 48px; }
.x-avatar-sm { width: 40px; height: 40px; font-size: 0.75rem; }
.x-avatar-xs { width: 32px; height: 32px; font-size: 0.65rem; }

/* ── Post content ── */
.x-post-title {
  margin: 0.55rem 1.2rem 0.25rem;
  font-size: 1.38rem;
  font-weight: 800;
  color: var(--text-primary);
  overflow-wrap: break-word;
  letter-spacing: -0.03em;
}

.x-post-body {
  margin: 0.25rem 1.2rem 0.9rem;
  font-size: 1.05rem;
  line-height: 1.75;
  color: var(--text-primary);
  white-space: pre-wrap;
  overflow-wrap: break-word;
}

.x-post-image-wrap {
  margin: 0 1.2rem 0.9rem;
  border-radius: 18px;
  overflow: hidden;
  cursor: zoom-in;
  border: 1px solid rgba(191, 219, 254, 0.72);
}
.x-post-image {
  width: 100%;
  display: block;
  max-height: 480px;
  object-fit: cover;
  transition: opacity 0.2s;
}
.x-post-image:hover { opacity: 0.92; }

/* ── Timestamp ── */
.x-timestamp {
  padding: 0.55rem 1.2rem 0.9rem;
  font-size: 0.82rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: #64748b;
}

/* ── Divider ── */
.x-divider {
  height: 1px;
  background: rgba(191, 219, 254, 0.72);
  margin: 0;
}

/* ── Stats ── */
.x-stats-row {
  display: flex;
  gap: 1.5rem;
  padding: 0.9rem 1.2rem;
  flex-wrap: wrap;
}
.x-stat {
  font-size: 0.9rem;
  color: var(--text-secondary, #6b7280);
}
.x-stat strong {
  color: var(--text-primary);
  font-weight: 800;
}

/* ── Action bar ── */
.x-action-bar {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  padding: 0.85rem 1.2rem;
  justify-content: flex-start;
  flex-wrap: wrap;
}

.x-action-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  background: rgba(239, 246, 255, 0.72);
  border: 1px solid rgba(191, 219, 254, 0.64);
  cursor: pointer;
  padding: 0.7rem 1rem;
  border-radius: 9999px;
  color: #475569;
  font-size: 0.9rem;
  font-weight: 700;
  transition: color 0.15s, background 0.15s, border-color 0.15s, transform 0.15s;
}
.x-action-btn:hover { background: rgba(219, 234, 254, 0.92); border-color: rgba(96, 165, 250, 0.46); transform: translateY(-1px); }

.x-action-icon {
  width: 20px;
  height: 20px;
}
.x-action-count { font-size: 0.85rem; }

/* Like active — red */
.like-btn:hover { color: #f91880; background: rgba(249,24,128,0.07); }
.x-active-like { color: #f91880 !important; }

/* Dislike active — blue */
.dislike-btn:hover { color: var(--brand-500, #2563eb); background: rgba(37,99,235,0.07); }
.x-active-dislike { color: var(--brand-500, #2563eb) !important; }

.x-like-error {
  font-size: 0.78rem;
  color: var(--status-error, #dc2626);
  margin-left: 0.5rem;
}

/* ── Recipients row ── */
.x-recipients-row { padding: 0.5rem 1rem; }
.x-recipients-btn {
  padding: 0.7rem 1rem;
  border-radius: 9999px;
  border: 1px solid rgba(191, 219, 254, 0.84);
  background: linear-gradient(180deg, rgba(255,255,255,0.96), rgba(239,246,255,0.94));
  color: var(--brand-500, #2563eb);
  font-size: 0.85rem;
  font-weight: 700;
  cursor: pointer;
  box-shadow: 0 14px 28px rgba(148, 163, 184, 0.14);
}
.x-recipients-btn:hover { background: rgba(37,99,235,0.07); }

/* ── Composer (reply form) ── */
.x-composer {
  display: flex;
  gap: 0.75rem;
  margin: 1rem 1.2rem;
  padding: 1rem 1.05rem;
  align-items: flex-start;
  border: 1px solid rgba(191, 219, 254, 0.7);
  border-radius: 1.5rem;
  background: linear-gradient(180deg, rgba(255,255,255,0.985), rgba(239,246,255,0.94));
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.88);
}

.x-composer-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.x-composer-input {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid rgba(203, 213, 225, 0.92);
  border-radius: 1.1rem;
  background: rgba(255, 255, 255, 0.94);
  font-family: inherit;
  font-size: 1rem;
  line-height: 1.6;
  color: var(--text-primary);
  resize: none;
  outline: none;
  padding: 0.9rem 1rem;
  placeholder-color: var(--text-secondary);
}
.x-composer-input:focus {
  border-color: rgba(59, 130, 246, 0.82);
  box-shadow: 0 0 0 4px rgba(191, 219, 254, 0.62);
}
.x-composer-input:disabled { opacity: 0.5; }

.x-composer-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-top: 1px solid rgba(191, 219, 254, 0.72);
  padding-top: 0.75rem;
}

.x-attach-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  cursor: pointer;
  color: var(--brand-500, #2563eb);
  font-size: 0.82rem;
  border-radius: 9999px;
  padding: 0.3rem;
  transition: background 0.15s;
}
.x-attach-btn:hover { background: rgba(37,99,235,0.08); }
.x-file-name {
  font-size: 0.78rem;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.x-reply-post-btn {
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  color: #fff;
  border: 1px solid transparent;
  border-radius: 9999px;
  padding: 0.72rem 1.2rem;
  font-size: 0.9rem;
  font-weight: 700;
  cursor: pointer;
  transition: background 0.15s, transform 0.15s, box-shadow 0.15s;
  box-shadow: 0 16px 32px rgba(37, 99, 235, 0.22);
}
.x-reply-post-btn:hover:not(:disabled) { background: #1d4ed8; transform: translateY(-1px); }
.x-reply-post-btn:disabled { opacity: 0.5; cursor: default; }

.x-form-error {
  font-size: 0.82rem;
  color: var(--status-error, #dc2626);
  margin: 0;
}

/* ── Empty comments ── */
.x-empty-comments {
  margin: 0 1.2rem 1.2rem;
  padding: 1.4rem 1rem;
  color: #64748b;
  font-size: 0.95rem;
  text-align: center;
  border: 1px dashed rgba(148, 163, 184, 0.48);
  border-radius: 1.35rem;
  background: linear-gradient(180deg, rgba(255,255,255,0.96), rgba(248,250,252,0.96));
}

/* ── Comment thread ── */
.x-comment-thread { }

.x-comment-row {
  display: flex;
  gap: 0.65rem;
  margin: 0 1.2rem;
  padding: 1rem 1rem 0.75rem;
  border-radius: 1.35rem;
  background: rgba(255, 255, 255, 0.7);
}
.x-reply-row {
  margin-left: 3rem;
}

.x-comment-left {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex-shrink: 0;
}

.x-thread-line {
  width: 2px;
  flex: 1;
  min-height: 16px;
  background: rgba(191, 219, 254, 0.9);
  margin-top: 4px;
  border-radius: 1px;
}

.x-comment-content {
  flex: 1;
  min-width: 0;
  padding-bottom: 0.5rem;
}

.x-comment-header {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.3rem;
  margin-bottom: 0.25rem;
}

.x-comment-name {
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--text-primary);
  text-decoration: none;
}
.x-comment-name:hover { text-decoration: underline; }

.x-comment-handle,
.x-comment-time {
  font-size: 0.82rem;
  color: var(--text-secondary, #6b7280);
}
.x-comment-dot {
  font-size: 0.82rem;
  color: var(--text-secondary, #6b7280);
}

.x-delete-btn {
  margin-left: auto;
  border: none;
  background: rgba(254, 242, 242, 0.85);
  color: var(--status-error, #dc2626);
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
  padding: 0.3rem 0.55rem;
  border-radius: 999px;
}
.x-delete-btn:hover { background: rgba(220,38,38,0.08); }

.x-comment-text {
  margin: 0 0 0.4rem;
  font-size: 0.95rem;
  line-height: 1.6;
  color: var(--text-primary);
  white-space: pre-wrap;
  overflow-wrap: break-word;
}

.x-comment-image-wrap {
  margin-bottom: 0.5rem;
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid rgba(191, 219, 254, 0.72);
}
.x-comment-image {
  width: 100%;
  max-height: 220px;
  object-fit: cover;
  display: block;
  cursor: zoom-in;
}

.x-comment-actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.2rem;
}

.x-cmt-action-btn {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  border: 1px solid rgba(191, 219, 254, 0.64);
  background: rgba(239, 246, 255, 0.74);
  color: var(--text-secondary, #6b7280);
  font-size: 0.82rem;
  font-weight: 600;
  cursor: pointer;
  padding: 0.45rem 0.7rem;
  border-radius: 9999px;
  transition: color 0.15s, background 0.15s;
}
.x-cmt-action-btn:hover {
  color: var(--brand-500, #2563eb);
  background: rgba(219, 234, 254, 0.92);
}

/* ── Inline reply ── */
.x-inline-reply {
  margin-top: 0.6rem;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  padding: 0.9rem;
  border: 1px solid rgba(191, 219, 254, 0.64);
  border-radius: 1.1rem;
  background: rgba(239, 246, 255, 0.62);
}
.x-reply-input {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid rgba(203, 213, 225, 0.92);
  border-radius: 12px;
  padding: 0.7rem 0.85rem;
  font-family: inherit;
  font-size: 0.9rem;
  line-height: 1.5;
  color: var(--text-primary);
  background: rgba(255, 255, 255, 0.95);
  resize: none;
}
.x-reply-input:focus { outline: none; border-color: var(--brand-500, #2563eb); box-shadow: 0 0 0 4px rgba(191, 219, 254, 0.62); }
.x-inline-reply-footer { display: flex; justify-content: flex-end; }

/* ── Loader ── */
.btn-loader { width: 14px; height: 14px; }

@media (max-width: 900px) {
  .x-shell {
    width: 100%;
    padding: 0 0 1.5rem;
  }

  .x-hero {
    grid-template-columns: 1fr;
    border-radius: 1.7rem;
    padding: 1.45rem 1.2rem;
  }

  .x-column {
    border-radius: 1.5rem;
  }

  .x-comment-row,
  .x-empty-comments,
  .x-composer {
    margin-left: 0.9rem;
    margin-right: 0.9rem;
  }

  .x-reply-row {
    margin-left: 1.7rem;
  }
}
</style>
