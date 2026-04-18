<template>
  <section class="shell">
    <header class="hero-row">
      <div class="hero-copy">
        <p class="eyebrow">New Group</p>
        <h1>Create a group space that feels ready to gather people.</h1>
        <p>Start a new community, write a clear purpose, and invite the people who should be part of it first.</p>
      </div>

      <div class="hero-meta">
        <article>
          <strong>{{ selectedInviteIds.length }}</strong>
          <span>Invites selected</span>
        </article>
        <article>
          <strong>{{ filteredFollowers.length }}</strong>
          <span>Followers available</span>
        </article>
      </div>
    </header>

    <form class="form-card" @submit.prevent="handleSubmit">
      <header class="form-head">
        <h2>Group details</h2>
        <p>Set up the basics and invite members in one flow.</p>
      </header>

      <label>
        <span>Title</span>
        <input v-model.trim="form.title" type="text" maxlength="80" required />
      </label>

      <label>
        <span>Description</span>
        <textarea v-model.trim="form.description" rows="5" />
      </label>

      <section class="invite-panel">
        <div class="invite-panel-header">
          <div>
            <h2>Invite followers</h2>
          </div>
          <span class="invite-counter">{{ selectedInviteIds.length }} selected</span>
        </div>

        <input
          v-model.trim="inviteSearch"
          type="text"
          placeholder="Search followers"
          :disabled="submitting || inviteLoading"
        />

        <p v-if="inviteError" class="feedback error">{{ inviteError }}</p>
        <p v-else-if="inviteLoading" class="invite-status">Loading followers...</p>
        <p v-else-if="filteredFollowers.length === 0" class="invite-status">No followers available.</p>

        <ul v-else class="invite-list">
          <li v-for="user in filteredFollowers" :key="user.id" class="invite-item">
            <label class="invite-option">
              <input
                type="checkbox"
                :checked="selectedInviteIds.includes(user.id)"
                :disabled="submitting"
                @change="toggleInvite(user.id)"
              />
              <span class="invite-copy">
                <strong>{{ formatUserDisplayName(user) }}</strong>
                <small>{{ user.firstName }} {{ user.lastName }}</small>
              </span>
            </label>
          </li>
        </ul>
      </section>

      <p v-if="error" class="feedback error">{{ error }}</p>
      <p v-if="success" class="feedback success">{{ success }}</p>

      <div class="actions">
        <router-link class="link-button" to="/groups">Cancel</router-link>
        <button type="submit" :disabled="submitting">
          {{ submitting ? 'Creating...' : 'Create group' }}
        </button>
      </div>
    </form>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getUserFollowers, type ConnectionUser } from '@/api/followers'
import { createGroup, createGroupInvite } from '@/api/groups'
import { fetchSessionData } from '@/router'
import type { SessionData } from '@/types/User'

const router = useRouter()
const submitting = ref(false)
const error = ref('')
const success = ref('')
const inviteLoading = ref(false)
const inviteError = ref('')
const inviteSearch = ref('')
const followers = ref<ConnectionUser[]>([])
const selectedInviteIds = ref<string[]>([])
const form = reactive({
  title: '',
  description: '',
})

const filteredFollowers = computed(() => {
  const trimmed = inviteSearch.value.trim().toLowerCase()
  if (!trimmed) {
    return followers.value
  }

  return followers.value.filter((user) => {
    const nickname = (user.nickname || '').toLowerCase()
    const firstName = user.firstName.toLowerCase()
    const lastName = user.lastName.toLowerCase()
    const fullName = `${firstName} ${lastName}`.trim()

    return (
      user.id.toLowerCase().includes(trimmed) ||
      nickname.includes(trimmed) ||
      firstName.includes(trimmed) ||
      lastName.includes(trimmed) ||
      fullName.includes(trimmed)
    )
  })
})

function formatUserDisplayName(user: ConnectionUser) {
  if (user.nickname?.trim()) return `@${user.nickname.trim()}`
  const fullName = `${user.firstName} ${user.lastName}`.trim()
  return fullName || user.id
}

function friendlyInviteError(err: unknown, fallback: string) {
  if (!(err instanceof Error)) {
    return fallback
  }

  const message = err.message.toLowerCase()
  if (
    message.includes('pending group invite') ||
    message.includes('invite already pending') ||
    message.includes('unique constraint failed: group_invites.group_id, group_invites.receiver_id')
  ) {
    return 'One of the selected followers already has an active invite for this group.'
  }

  if (message.includes('user is already a group member')) {
    return 'One of the selected users is already a member of this group.'
  }

  return err.message
}

function toggleInvite(userId: string) {
  if (selectedInviteIds.value.includes(userId)) {
    selectedInviteIds.value = selectedInviteIds.value.filter((id) => id !== userId)
    return
  }

  selectedInviteIds.value = [...selectedInviteIds.value, userId]
}

async function loadFollowers() {
  inviteLoading.value = true
  inviteError.value = ''

  try {
    const sessionData: SessionData | null = await fetchSessionData()
    const currentUserId = sessionData?.user.id ?? ''
    if (!currentUserId) {
      followers.value = []
      return
    }

    followers.value = await getUserFollowers(currentUserId)
  } catch (err) {
    followers.value = []
    inviteError.value = err instanceof Error ? err.message : 'Failed to load followers.'
  } finally {
    inviteLoading.value = false
  }
}

async function handleSubmit() {
  submitting.value = true
  error.value = ''
  success.value = ''

  try {
    const group = await createGroup(form)
    if (selectedInviteIds.value.length > 0) {
      const results = await Promise.allSettled(
        selectedInviteIds.value.map((userId) => createGroupInvite(group.id, userId)),
      )
      const failures = results.filter((r) => r.status === 'rejected')
      if (failures.length > 0) {
        const firstErr = (failures[0] as PromiseRejectedResult).reason
        error.value = friendlyInviteError(firstErr, 'Failed to send one or more invitations.')
        if (failures.length === selectedInviteIds.value.length) {
          submitting.value = false
          return
        }
      }
    }
    success.value = 'Group created. Redirecting to details...'
    window.setTimeout(() => {
      void router.push(`/groups/${group.id}`)
    }, 500)
  } catch (err) {
    error.value = friendlyInviteError(err, 'Failed to create group.')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  void loadFollowers()
})
</script>

<style scoped>
.shell {
  width: min(100%, 1320px);
  margin: 0 auto;
  padding: 0.25rem 0 1.75rem;
  display: grid;
  gap: 1rem;
}

.hero-row {
  display: grid;
  grid-template-columns: minmax(0, 1.15fr) minmax(240px, 0.85fr);
  gap: 1rem;
  align-items: end;
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

.hero-row::before,
.hero-row::after {
  content: '';
  position: absolute;
  border-radius: 999px;
  pointer-events: none;
}

.hero-row::before {
  width: 220px;
  height: 220px;
  right: -40px;
  top: -60px;
  background: rgba(255, 255, 255, 0.08);
}

.hero-row::after {
  width: 160px;
  height: 160px;
  left: -40px;
  bottom: -60px;
  background: rgba(125, 211, 252, 0.12);
}

.hero-copy,
.hero-meta {
  position: relative;
  z-index: 1;
}

.eyebrow {
  display: inline-flex;
  align-items: center;
  padding: 0.38rem 0.78rem;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.12);
  color: rgba(255, 255, 255, 0.84);
  font-size: 0.76rem;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  font-weight: 800;
}

h1 {
  margin-top: 1rem;
  color: var(--white);
  max-width: 12ch;
  font-size: clamp(2.2rem, 5vw, 3.5rem);
  line-height: 0.96;
  font-weight: 900;
}

.hero-copy p:last-child {
  margin-top: 0.9rem;
  max-width: 46ch;
  color: rgba(255, 255, 255, 0.82);
  font-size: 1rem;
}

.hero-meta {
  display: grid;
  gap: 0.85rem;
  align-content: end;
}

.hero-meta article {
  padding: 1rem 1.05rem;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.12);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.hero-meta strong {
  display: block;
  color: var(--white);
  font-size: 1.55rem;
  font-weight: 900;
}

.hero-meta span {
  color: rgba(255, 255, 255, 0.76);
  font-size: 0.86rem;
  font-weight: 700;
}

.form-card {
  width: 100%;
  display: grid;
  gap: 1.2rem;
  padding: 1.25rem;
  border-radius: 32px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(248, 250, 252, 0.86));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.08);
}

.form-head h2 {
  color: var(--gray-900);
  font-size: 1.7rem;
  font-weight: 900;
}

.form-head p {
  margin-top: 0.3rem;
  color: var(--gray-500);
  font-weight: 600;
}

.invite-panel {
  display: grid;
  gap: 0.9rem;
  padding: 1rem;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 24px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.88));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.08);
}

.invite-panel-header {
  display: flex;
  justify-content: space-between;
  gap: 0.8rem;
  align-items: start;
}

.invite-panel-header h2 {
  color: var(--gray-900);
  font-size: 1.2rem;
  font-weight: 900;
}

.invite-panel-header p {
  color: var(--text-secondary);
  font-size: 0.88rem;
}

.invite-counter {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0.3rem 0.8rem;
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.14), rgba(14, 165, 233, 0.08));
  color: var(--brand-700);
  font-size: 0.8rem;
  font-weight: 800;
  white-space: nowrap;
}

.invite-status {
  color: var(--text-secondary);
  font-weight: 600;
}

.invite-list {
  list-style: none;
  display: grid;
  gap: 0.8rem;
  padding: 0.45rem;
  margin: 0;
  max-height: 12rem;
  overflow-y: auto;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 20px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
}

.invite-item {
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 18px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.92));
}

.invite-option {
  display: flex;
  gap: 0.75rem;
  align-items: center;
  padding: 0.75rem 0.85rem;
  cursor: pointer;
}

.invite-option input {
  flex: 0 0 auto;
  width: 1rem;
  height: 1rem;
  min-width: 1rem;
  margin: 0;
}

.invite-copy {
  display: grid;
  gap: 0.15rem;
  min-width: 0;
}

.invite-copy strong {
  color: var(--text-primary);
}

.invite-copy small {
  color: var(--text-secondary);
}

label {
  display: grid;
  gap: 0.45rem;
}

label span {
  color: var(--gray-900);
  font-weight: 800;
  font-size: 0.95rem;
}

input,
textarea {
  width: 100%;
  border-radius: 20px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  padding: 0.95rem 1rem;
  font: inherit;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 8px 22px rgba(15, 23, 42, 0.04);
}

input:focus,
textarea:focus {
  outline: none;
  border-color: rgba(37, 99, 235, 0.55);
  box-shadow:
    0 0 0 5px rgba(37, 99, 235, 0.1),
    0 16px 34px rgba(37, 99, 235, 0.12);
}

.feedback {
  border-radius: 20px;
  padding: 0.75rem 0.9rem;
  font-weight: 700;
}

.feedback.error {
  color: var(--status-error);
  background: rgba(220, 38, 38, 0.08);
}

.feedback.success {
  color: var(--status-success);
  background: rgba(22, 163, 74, 0.08);
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

@media (max-width: 900px) {
  .hero-row {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .actions {
    justify-content: stretch;
    flex-direction: column-reverse;
  }

  button,
  .link-button {
    width: 100%;
    text-align: center;
  }

  .hero-row,
  .form-card,
  .invite-panel {
    border-radius: 24px;
  }
}

button,
.link-button {
  border: none;
  border-radius: 18px;
  padding: 0.82rem 1rem;
  text-decoration: none;
  font-weight: 800;
  font: inherit;
}

button {
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  color: var(--text-white);
  cursor: pointer;
  box-shadow: 0 16px 30px rgba(37, 99, 235, 0.24);
}

button:disabled {
  opacity: 0.6;
  cursor: wait;
}

.link-button {
  color: var(--gray-700);
  border: 1px solid rgba(148, 163, 184, 0.18);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
}
</style>
