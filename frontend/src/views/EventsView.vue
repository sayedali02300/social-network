<template>
  <section class="events-shell">
    <header class="topbar-card">
      <div>
        <router-link class="back-link" :to="`/groups/${groupId}`">Back to group</router-link>
        <h1>Group Events</h1>
      </div>
    </header>

    <section class="events-layout">
      <aside class="composer-column">
        <section class="composer-card">
          <div class="composer-head">
            <h2>Create event</h2>
            <span>Visible to members only</span>
          </div>
          <form class="event-form" @submit.prevent="handleCreateEvent">
            <input v-model.trim="form.title" type="text" maxlength="80" placeholder="Sprint planning" required />
            <input v-model="form.eventTime" type="datetime-local" :min="minimumEventTime" required />
            <textarea v-model.trim="form.description" rows="3" placeholder="What is this event for?" />
            <button type="submit" :disabled="submitting">
              {{ submitting ? 'Saving...' : 'Create event' }}
            </button>
          </form>
        </section>
      </aside>

      <main class="feed-column">
        <p v-if="error" class="message error">{{ error }}</p>
        <p v-else-if="feedback" class="message success">{{ feedback }}</p>
        <p v-if="loading" class="message">Loading events...</p>

        <section v-else class="event-feed">
          <p v-if="events.length === 0" class="message">No events yet. Members can create the first one here.</p>
          <article v-for="event in events" :key="event.id" class="event-card">
            <div class="event-top">
              <div>
                <h3>{{ event.title }}</h3>
                <p>{{ event.description || 'No description yet.' }}</p>
              </div>
              <span class="time-pill">{{ formatDate(event.eventTime) }}</span>
            </div>

            <div class="metrics">
              <span>Participants: {{ event.goingCount }}</span>
              <span>Declined: {{ event.notGoingCount }}</span>
              <span>You: {{ event.myResponse ? labelFor(event.myResponse) : 'No response' }}</span>
            </div>

            <div class="actions">
              <div class="action-group">
                <button
                  :class="event.myResponse === 'going' ? 'active-rsvp' : 'secondary'"
                  @click="handleResponse(event.id, 'going')"
                >
                  Participate
                </button>
                <button
                  :class="event.myResponse === 'not_going' ? 'active-rsvp danger-rsvp' : 'danger'"
                  @click="handleResponse(event.id, 'not_going')"
                >
                  Not participate
                </button>
              </div>
              <div class="action-group" v-if="canManageEvent(event)">
                <button
                  v-if="editingEventId !== event.id"
                  class="secondary"
                  @click="startEditEvent(event)"
                >
                  Edit
                </button>
                <button class="danger" @click="handleDeleteEvent(event.id)">
                  Delete
                </button>
              </div>
            </div>

            <div v-if="editingEventId === event.id" class="event-form editor">
              <input v-model.trim="editForm.title" type="text" maxlength="80" required />
              <input v-model="editForm.eventTime" type="datetime-local" :min="minimumEventTime" required />
              <textarea v-model.trim="editForm.description" rows="3" />
              <div class="actions">
                <button class="secondary" @click="cancelEditEvent">Cancel</button>
                <button @click="saveEventChanges(event.id)">Save</button>
              </div>
            </div>
          </article>
        </section>
      </main>
    </section>

    <div v-if="deleteEventConfirmId" class="confirm-overlay" @click="deleteEventConfirmId = ''">
      <div class="confirm-dialog" @click.stop>
        <h3>Delete event</h3>
        <p>Are you sure you want to delete this event?</p>
        <div class="confirm-actions">
          <button class="secondary" @click="deleteEventConfirmId = ''">Cancel</button>
          <button class="danger" @click="confirmDeleteEvent">Delete</button>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'

import {
  createEvent,
  deleteEvent,
  getGroup,
  listEvents,
  respondToEvent,
  updateEvent,
} from '@/api/groups'
import { buildWebSocketURL } from '@/api/websocket'
import { fetchSessionData } from '@/router'
import type { EventItem } from '@/types/groups'
import type { SessionData } from '@/types/User'

type GroupCalendarDeletePayload = {
  groupId: string
  eventId: string
}

const route = useRoute()
const groupId = computed(() => route.params.groupId as string)
const loading = ref(false)
const submitting = ref(false)
const error = ref('')
const feedback = ref('')
const events = ref<EventItem[]>([])
const groupCreatorId = ref('')
const currentUserId = ref('')
const editingEventId = ref('')
const deleteEventConfirmId = ref('')
const minimumEventTime = ref(getCurrentDateTimeLocalMinute())
let socket: WebSocket | null = null
let reconnectTimerId: number | null = null
let socketClosedManually = false
let minTimeIntervalId: number | null = null
const form = reactive({
  title: '',
  description: '',
  eventTime: '',
})
const editForm = reactive({
  title: '',
  description: '',
  eventTime: '',
})

function sortEvents(items: EventItem[]) {
  return [...items].sort((a, b) => new Date(a.eventTime).getTime() - new Date(b.eventTime).getTime())
}

function formatDate(value: string) {
  return new Date(value).toLocaleString()
}

function labelFor(value: EventItem['myResponse']) {
  return value === 'going' ? 'Participate' : 'Not participate'
}

function upsertEvent(event: EventItem) {
  if (event.groupId !== groupId.value) {
    return
  }

  events.value = sortEvents([
    event,
    ...events.value.filter((item) => item.id !== event.id),
  ])
}

function removeEvent(eventId: string) {
  events.value = events.value.filter((event) => event.id !== eventId)
  if (editingEventId.value === eventId) {
    editingEventId.value = ''
  }
}

function isGroupEventPayload(payload: unknown): payload is EventItem {
  if (!payload || typeof payload !== 'object') return false
  const item = payload as Partial<EventItem>
  return typeof item.id === 'string' && typeof item.groupId === 'string' && typeof item.eventTime === 'string'
}

function isGroupCalendarDeletePayload(payload: unknown): payload is GroupCalendarDeletePayload {
  if (!payload || typeof payload !== 'object') return false
  const item = payload as Partial<GroupCalendarDeletePayload>
  return typeof item.groupId === 'string' && typeof item.eventId === 'string'
}

function scheduleSocketReconnect() {
  if (socketClosedManually || reconnectTimerId !== null) {
    return
  }

  reconnectTimerId = window.setTimeout(() => {
    reconnectTimerId = null
    connectRealtimeSocket()
  }, 1500)
}

function connectRealtimeSocket() {
  if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
    return
  }

  socketClosedManually = false
  socket = new WebSocket(buildWebSocketURL('/ws'))

  socket.onmessage = (event) => {
    try {
      const raw = JSON.parse(event.data) as { type?: string; payload?: unknown }
      if (raw.type === 'group_calendar_event' && isGroupEventPayload(raw.payload)) {
        upsertEvent(raw.payload)
        return
      }
      if (raw.type === 'group_calendar_delete' && isGroupCalendarDeletePayload(raw.payload) && raw.payload.groupId === groupId.value) {
        removeEvent(raw.payload.eventId)
      }
    } catch {
      // Ignore invalid websocket payloads.
    }
  }

  socket.onclose = () => {
    socket = null
    scheduleSocketReconnect()
  }

  socket.onerror = () => {
    socket?.close()
  }
}

function disconnectRealtimeSocket() {
  socketClosedManually = true
  if (reconnectTimerId !== null) {
    window.clearTimeout(reconnectTimerId)
    reconnectTimerId = null
  }
  if (socket) {
    socket.close()
    socket = null
  }
}

async function loadEvents() {
  loading.value = true
  error.value = ''

  try {
    const group = await getGroup(groupId.value)
    groupCreatorId.value = group.creatorId
    events.value = await listEvents(groupId.value)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load events.'
  } finally {
    loading.value = false
  }
}

function toDateTimeLocal(value: string) {
  const date = new Date(value)
  const offsetMs = date.getTimezoneOffset() * 60 * 1000
  return new Date(date.getTime() - offsetMs).toISOString().slice(0, 16)
}

function getCurrentDateTimeLocalMinute() {
  const now = new Date()
  now.setSeconds(0, 0)
  return toDateTimeLocal(now.toISOString())
}

function refreshMinimumEventTime() {
  minimumEventTime.value = getCurrentDateTimeLocalMinute()
}

function isPastEventTime(value: string) {
  if (!value) {
    return false
  }

  const selected = new Date(value)
  if (Number.isNaN(selected.getTime())) {
    return false
  }

  const now = new Date()
  now.setSeconds(0, 0)
  return selected.getTime() < now.getTime()
}

function canManageEvent(event: EventItem) {
  return event.creatorId === currentUserId.value || groupCreatorId.value === currentUserId.value
}

function startEditEvent(event: EventItem) {
  editingEventId.value = event.id
  editForm.title = event.title
  editForm.description = event.description
  editForm.eventTime = toDateTimeLocal(event.eventTime)
}

function cancelEditEvent() {
  editingEventId.value = ''
}

async function saveEventChanges(eventId: string) {
  error.value = ''
  feedback.value = ''
  refreshMinimumEventTime()

  if (isPastEventTime(editForm.eventTime)) {
    error.value = 'Event time cannot be in the past.'
    return
  }

  try {
    await updateEvent(eventId, {
      title: editForm.title,
      description: editForm.description,
      eventTime: editForm.eventTime,
    })
    editingEventId.value = ''
    feedback.value = 'Event updated.'
    await loadEvents()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to update event.'
  }
}

function handleDeleteEvent(eventId: string) {
  deleteEventConfirmId.value = eventId
}

async function confirmDeleteEvent() {
  const eventId = deleteEventConfirmId.value
  deleteEventConfirmId.value = ''
  if (!eventId) return

  error.value = ''
  feedback.value = ''

  try {
    await deleteEvent(eventId)
    if (editingEventId.value === eventId) {
      editingEventId.value = ''
    }
    feedback.value = 'Event deleted.'
    await loadEvents()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to delete event.'
  }
}

async function handleCreateEvent() {
  submitting.value = true
  error.value = ''
  feedback.value = ''
  refreshMinimumEventTime()

  if (isPastEventTime(form.eventTime)) {
    error.value = 'Event time cannot be in the past.'
    submitting.value = false
    return
  }

  try {
    await createEvent(groupId.value, form)
    form.title = ''
    form.description = ''
    form.eventTime = ''
    feedback.value = 'Event created.'
    await loadEvents()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to create event.'
  } finally {
    submitting.value = false
  }
}

async function handleResponse(eventId: string, response: 'going' | 'not_going') {
  error.value = ''
  feedback.value = ''

  try {
    await respondToEvent(eventId, response)
    feedback.value = response === 'going' ? 'RSVP saved as participating.' : 'RSVP saved as not participating.'
    await loadEvents()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to submit RSVP.'
  }
}

onMounted(async () => {
  connectRealtimeSocket()
  minTimeIntervalId = window.setInterval(refreshMinimumEventTime, 60_000)
  const sessionData: SessionData | null = await fetchSessionData()
  currentUserId.value = sessionData?.user.id ?? ''
  await loadEvents()
})

onBeforeUnmount(() => {
  disconnectRealtimeSocket()
  if (minTimeIntervalId !== null) {
    window.clearInterval(minTimeIntervalId)
    minTimeIntervalId = null
  }
})
</script>

<style scoped>
.events-shell {
  --surface: #ffffff;
  --surface-soft: #f7f9fc;
  --surface-line: #dde3ec;
  --ink: #0f172a;
  --ink-muted: #5b6474;
  --brand: #1877f2;
  --brand-soft: #e7f1ff;
  width: min(100%, 1320px);
  margin: 0 auto;
  padding: 0.25rem 0 2rem;
  display: grid;
  gap: 1.25rem;
}

.topbar-card {
  position: relative;
  overflow: hidden;
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(260px, 0.8fr);
  gap: 1rem;
  align-items: start;
  border-radius: 2rem;
  border: 1px solid rgba(96, 165, 250, 0.36);
  background:
    radial-gradient(circle at top right, rgba(96, 165, 250, 0.28), transparent 32%),
    radial-gradient(circle at bottom left, rgba(147, 197, 253, 0.22), transparent 24%),
    linear-gradient(135deg, #17306b 0%, #1f4eb6 100%);
  box-shadow: 0 34px 70px rgba(37, 99, 235, 0.26);
  padding: 1.9rem 1.95rem 1.7rem;
}

.topbar-card::after {
  content: "";
  position: absolute;
  top: -3rem;
  right: -2rem;
  width: 14rem;
  height: 14rem;
  border-radius: 999px;
  background: rgba(147, 197, 253, 0.14);
}

.back-link {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  border-radius: 999px;
  padding: 0.65rem 1rem;
  background: rgba(219, 234, 254, 0.18);
  border: 1px solid rgba(191, 219, 254, 0.16);
  color: #eff6ff;
  font-size: 0.96rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  text-decoration: none;
  backdrop-filter: blur(12px);
  margin-bottom: 0.95rem;
}

.topbar-card h1 {
  color: #fff;
  font-size: clamp(2.35rem, 4vw, 4.1rem);
  font-weight: 800;
  line-height: 0.98;
  letter-spacing: -0.04em;
  margin: 0;
  text-wrap: balance;
}

.events-layout {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.15rem;
  align-items: start;
}

.composer-column {
  width: 100%;
  max-width: none;
  margin: 0 auto;
}

.composer-card {
  border-radius: 1.8rem;
  border: 1px solid rgba(191, 219, 254, 0.72);
  background: linear-gradient(180deg, rgba(255,255,255,0.985), rgba(239,246,255,0.94));
  padding: 1.25rem;
  box-shadow: 0 26px 60px rgba(148, 163, 184, 0.2);
}

.composer-head {
  display: flex;
  justify-content: space-between;
  gap: 0.65rem;
  align-items: center;
  flex-wrap: wrap;
  margin-bottom: 0.9rem;
}

.composer-head h2 {
  color: var(--ink);
  font-size: 1.3rem;
  font-weight: 800;
  letter-spacing: -0.03em;
  margin: 0;
}

.composer-head span {
  color: #64748b;
  font-size: 0.82rem;
  font-weight: 700;
}

.feed-column {
  display: grid;
  gap: 1rem;
  width: 100%;
  max-width: none;
  margin: 0 auto;
}

.event-feed {
  display: grid;
  gap: 1rem;
}

.event-form {
  display: grid;
  gap: 0.85rem;
}

input,
textarea {
  width: 100%;
  border-radius: 1.15rem;
  border: 1px solid rgba(203, 213, 225, 0.92);
  padding: 0.95rem 1rem;
  font: inherit;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: inset 0 1px 2px rgba(148, 163, 184, 0.08);
}

button {
  border: 1px solid transparent;
  border-radius: 999px;
  padding: 0.82rem 1.08rem;
  font: inherit;
  font-weight: 700;
  cursor: pointer;
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  color: #fff;
  box-shadow: 0 16px 32px rgba(37, 99, 235, 0.2);
}

button:disabled {
  opacity: 0.6;
  cursor: wait;
}

button.secondary {
  background: linear-gradient(180deg, rgba(255,255,255,0.96), rgba(239,246,255,0.94));
  color: #263449;
  border: 1px solid rgba(191, 219, 254, 0.82);
  box-shadow: 0 14px 28px rgba(148, 163, 184, 0.14);
}

button.danger {
  background: rgba(239, 68, 68, 0.12);
  color: var(--status-error);
  border-color: rgba(248, 113, 113, 0.28);
  box-shadow: 0 16px 32px rgba(248, 113, 113, 0.14);
}

.event-card {
  border-radius: 1.65rem;
  border: 1px solid rgba(191, 219, 254, 0.62);
  background: linear-gradient(180deg, rgba(255,255,255,0.985), rgba(241,247,255,0.94));
  padding: 1.15rem 1.2rem;
  display: grid;
  gap: 0.9rem;
  box-shadow: 0 20px 45px rgba(148, 163, 184, 0.16);
   min-width: 0;
  overflow: hidden;
}

.event-top,
.metrics,
.actions {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  flex-wrap: wrap;
  align-items: center;
}

.actions {
  justify-content: space-between;
}

.metrics {
  margin: 0.35rem 0;
}

.action-group {
  display: flex;
  gap: 0.55rem;
  align-items: center;
  flex-wrap: wrap;
}
.event-top > div {
  min-width: 0;
  overflow: hidden;
  flex: 1;
}

.event-top {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  flex-wrap: wrap;
  align-items: center;
  min-width: 0;
}
.event-top h3 {
  color: var(--ink);
  font-size: 1.18rem;
  font-weight: 800;
  margin: 0;
  letter-spacing: -0.02em;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.event-top p,
.metrics span {
  color: var(--ink-muted);
  font-size: 0.94rem;
}
.event-top p {
  color: var(--ink-muted);
  font-size: 0.94rem;
  display: -webkit-box;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  overflow-wrap: break-word;
  word-break: break-word;
  margin: 0;
}
.editor {
  border-top: 1px solid rgba(191, 219, 254, 0.72);
  padding-top: 0.9rem;
}

.time-pill {
  padding: 0.45rem 0.82rem;
  border-radius: 999px;
  background: rgba(219, 234, 254, 0.92);
  color: #0b4bab;
  font-weight: 700;
  font-size: 0.8rem;
   flex-shrink: 0;
  white-space: nowrap;
}

.message {
  padding: 1rem 1.05rem;
  border-radius: 1.3rem;
  border: 1px solid rgba(191, 219, 254, 0.62);
  background: linear-gradient(180deg, rgba(255,255,255,0.98), rgba(248,250,252,0.96));
  color: #64748b;
  font-weight: 700;
  min-height: 4.5rem;
  display: flex;
  align-items: center;
}

.message.error {
  color: #9f1239;
  border-color: #fecdd3;
}

.message.success {
  color: #166534;
  border-color: #bbf7d0;
}

.active-rsvp {
  background: rgba(219, 234, 254, 0.95);
  color: #1d4ed8;
  border: 1px solid rgba(96, 165, 250, 0.44);
}

.active-rsvp.danger-rsvp {
  background: #fee2e2;
  color: #b91c1c;
  border-color: #fca5a5;
}

@media (max-width: 900px) {
  .composer-column {
    position: static;
  }

  .topbar-card {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .topbar-card {
    align-items: flex-start;
    border-radius: 1.7rem;
    padding: 1.4rem 1.15rem;
  }

  .events-shell {
    width: 100%;
    padding: 0 0 1.5rem;
  }
}

.confirm-overlay {
  position: fixed;
  inset: 0;
  background: rgba(17, 24, 39, 0.45);
  display: grid;
  place-items: center;
  z-index: 50;
  padding: 1rem;
}

.confirm-dialog {
  width: min(420px, 100%);
  background: linear-gradient(180deg, rgba(255,255,255,0.985), rgba(239,246,255,0.96));
  border: 1px solid rgba(191, 219, 254, 0.72);
  border-radius: 1.5rem;
  padding: 1.2rem;
  display: grid;
  gap: 0.9rem;
  box-shadow: 0 30px 70px rgba(15, 23, 42, 0.24);
}

.confirm-dialog h3 { color: var(--text-primary); }
.confirm-dialog p  { color: var(--text-secondary); }

.confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.65rem;
}
</style>
