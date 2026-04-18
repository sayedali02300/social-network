<template>
  <section class="page">
    <header class="header-row">
      <div>
        <router-link class="back-link" :to="`/groups/${groupId}`">Back to group</router-link>
        <h1>Group members</h1>
      </div>
    </header>

    <p v-if="loading" class="message">Loading members...</p>
    <p v-else-if="error" class="message error">{{ error }}</p>

    <ul v-else class="member-list">
      <li v-for="member in members" :key="member.userId">
        <div>
          <strong>{{ member.nickname }}</strong>
          <p>{{ formatMemberRole(member.role) }}</p>
        </div>
        <button
          v-if="member.role !== 'creator' && (isCreator || member.userId === currentUserId)"
          class="danger"
          @click="handleRemove(member.userId)"
        >
          {{ member.userId === currentUserId ? 'Leave group' : 'Remove' }}
        </button>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'

import { getGroup, listGroupMembers, removeGroupMember } from '@/api/groups'
import { fetchSessionData } from '@/router'
import type { GroupMember } from '@/types/groups'
import type { SessionData } from '@/types/User'

const route = useRoute()
const members = ref<GroupMember[]>([])
const loading = ref(false)
const error = ref('')
const currentUserId = ref('')
const creatorId = ref('')

const groupId = computed(() => route.params.groupId as string)
const isCreator = computed(() => creatorId.value === currentUserId.value)

function formatMemberRole(role?: string) {
  switch ((role || '').toLowerCase()) {
    case 'creator':
      return 'Creator'
    case 'admin':
      return 'Admin'
    case 'member':
      return 'Member'
    default:
      return 'Member'
  }
}

async function loadMembers() {
  loading.value = true
  error.value = ''

  try {
    const [groupDetails, groupMembers] = await Promise.all([
      getGroup(groupId.value),
      listGroupMembers(groupId.value),
    ])
    creatorId.value = groupDetails.creatorId
    members.value = groupMembers
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load members.'
  } finally {
    loading.value = false
  }
}

async function handleRemove(userId: string) {
  try {
    await removeGroupMember(groupId.value, userId)
    await loadMembers()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to remove member.'
  }
}

onMounted(async () => {
  const sessionData: SessionData | null = await fetchSessionData()
  currentUserId.value = sessionData?.user.id ?? ''
  await loadMembers()
})
</script>

<style scoped>
.page {
  padding: 2rem clamp(1rem, 2vw, 2.5rem) 3rem;
  display: grid;
  gap: 1.5rem;
}

.header-row {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: center;
  flex-wrap: wrap;
}

.back-link {
  color: var(--primary-blue);
  font-weight: 700;
  text-decoration: none;
}

h1 {
  color: var(--text-primary);
}

.message,
.member-list li {
  border-radius: 1rem;
  border: 1px solid var(--border-color);
  background: var(--bg-card);
}

.message {
  padding: 1rem;
}

.message.error {
  color: var(--status-error);
}

.member-list {
  list-style: none;
  display: grid;
  gap: 0.75rem;
  padding: 0;
}

.member-list li {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.1rem;
}

strong {
  color: var(--text-primary);
}

p {
  color: var(--text-secondary);
}

button {
  border: none;
  border-radius: 999px;
  padding: 0.7rem 0.95rem;
  font: inherit;
  font-weight: 700;
  cursor: pointer;
}

.danger {
  color: var(--status-error);
  background: rgba(220, 38, 38, 0.1);
}
</style>
