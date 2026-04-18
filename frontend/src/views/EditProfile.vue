<template>
  <main class="edit-profile-page">
    <section class="edit-profile-shell">
      <header class="edit-profile-hero">
        <div class="edit-profile-hero-copy">
          <span class="edit-profile-kicker">Profile studio</span>
          <h1>Edit your profile so it feels ready to connect.</h1>
          <p>Update your public details, refine your identity, and keep your profile presentation polished.</p>
        </div>

        <div class="edit-profile-hero-meta">
          <article>
            <strong>{{ form.nickname ? '@' + form.nickname : 'Profile' }}</strong>
            <span>Current identity</span>
          </article>
          <article>
            <strong>{{ form.isPublic ? 'Public' : 'Private' }}</strong>
            <span>Privacy mode</span>
          </article>
        </div>
      </header>

      <section class="edit-profile-card">
        <header class="edit-profile-header">
          <h1>Edit profile</h1>
          <p>Update your public details.</p>
        </header>

      <div v-if="isLoading" class="status-message">Loading profile...</div>
      <div v-else-if="loadError" class="status-message error">{{ loadError }}</div>

      <form v-else class="edit-profile-form" @submit.prevent="submitProfile">
        <section class="avatar-editor">
          <div class="avatar-preview">
            <img v-if="displayAvatarSrc" :src="displayAvatarSrc" alt="Current avatar" />
            <div v-else class="avatar-fallback">{{ initials }}</div>
          </div>
          <label class="change-avatar-label" for="avatarFileInput">Change avatar</label>
          <input
            id="avatarFileInput"
            class="hidden-file-input"
            type="file"
            accept="image/png,image/jpeg,image/gif,image/webp"
            @change="onAvatarSelected"
          />
          <small v-if="selectedAvatarName" class="file-name">Selected: {{ selectedAvatarName }}</small>
        </section>

        <div class="grid-two">
          <label class="field">
            <span>First name</span>
            <input v-model.trim="form.firstName" type="text" maxlength="25" required />
          </label>
          <label class="field">
            <span>Last name</span>
            <input v-model.trim="form.lastName" type="text" maxlength="25" required />
          </label>
        </div>

        <div class="grid-two">
          <label class="field">
            <span>Date of birth</span>
            <input v-model="form.dateOfBirth" type="date" :min="dobBounds.min" :max="dobBounds.max" required />
          </label>
          <label class="field">
            <span>Nickname (Optional)</span>
            <input v-model.trim="form.nickname" type="text" maxlength="25" />
          </label>
        </div>

        <label class="field">
          <span>About me (Optional)</span>
          <textarea v-model.trim="form.aboutMe" rows="4" maxlength="255" />
        </label>

        <div class="privacy-row">
          <div class="privacy-copy">
            <p>Account privacy</p>
            <small>{{ form.isPublic ? 'Public profile' : 'Private profile' }}</small>
          </div>
          <label class="privacy-switch" aria-label="Toggle account privacy">
            <input v-model="form.isPublic" type="checkbox" />
            <span class="privacy-slider"></span>
          </label>
        </div>

        <p v-if="errorMessage" class="status-message error">{{ errorMessage }}</p>
        <p v-if="successMessage" class="status-message success">{{ successMessage }}</p>

        <div class="actions">
          <RouterLink class="secondary-btn" to="/profile">Cancel</RouterLink>
          <button type="submit" class="primary-btn" :disabled="isSaving">
            {{ isSaving ? 'Saving...' : 'Save changes' }}
          </button>
        </div>
      </form>
      </section>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { API_ROUTES, apiURL } from '@/api/api'
import {
  getDateOfBirthBounds,
  validateLettersOnlyName,
  validateNickname,
  validateRealisticDateOfBirth,
} from '@/utils/helpers'

type UserProfile = {
  firstName: string
  lastName: string
  dateOfBirth: string
  avatar: string
  nickname: string
  aboutMe: string
  isPublic: boolean
}

type ErrorResponse = {
  error?: string
}

const router = useRouter()

const form = reactive<UserProfile>({
  firstName: '',
  lastName: '',
  dateOfBirth: '',
  avatar: '',
  nickname: '',
  aboutMe: '',
  isPublic: true,
})

const selectedAvatarFile = ref<File | null>(null)
const selectedAvatarName = ref('')
const selectedAvatarPreview = ref('')
const isLoading = ref(true)
const isSaving = ref(false)
const loadError = ref('')
const errorMessage = ref('')
const successMessage = ref('')
const currentAvatar = ref('')
const originalIsPublic = ref(true)
const dobBounds = getDateOfBirthBounds()

const initials = computed(() => {
  const first = form.firstName?.[0] || ''
  const last = form.lastName?.[0] || ''
  const value = `${first}${last}`.trim().toUpperCase()
  return value || '?'
})

const displayAvatarSrc = computed(() => {
  if (selectedAvatarPreview.value) return selectedAvatarPreview.value
  if (!currentAvatar.value) return ''
  if (currentAvatar.value.startsWith('http://') || currentAvatar.value.startsWith('https://')) {
    return currentAvatar.value
  }
  return apiURL(currentAvatar.value.startsWith('/') ? currentAvatar.value : `/${currentAvatar.value}`)
})

const loadProfile = async () => {
  isLoading.value = true
  loadError.value = ''

  try {
    const response = await fetch(apiURL(API_ROUTES.USERS_ME), {
      method: 'GET',
      credentials: 'include',
    })

    if (!response.ok) {
      let message = 'Could not load profile.'
      const payload = (await response.json().catch(() => null)) as ErrorResponse | null
      if (payload?.error) message = payload.error
      loadError.value = message
      return
    }

    const data = (await response.json()) as UserProfile
    form.firstName = data.firstName || ''
    form.lastName = data.lastName || ''
    form.dateOfBirth = data.dateOfBirth || ''
    form.avatar = data.avatar || ''
    form.nickname = data.nickname || ''
    form.aboutMe = data.aboutMe || ''
    form.isPublic = Boolean(data.isPublic)
    currentAvatar.value = data.avatar || ''
    originalIsPublic.value = Boolean(data.isPublic)
  } catch {
    loadError.value = 'Network error while loading profile.'
  } finally {
    isLoading.value = false
  }
}

const onAvatarSelected = (event: Event) => {
  const input = event.target as HTMLInputElement | null
  const file = input?.files?.[0] ?? null
  selectedAvatarFile.value = file
  selectedAvatarName.value = file ? file.name : ''
  if (selectedAvatarPreview.value) {
    URL.revokeObjectURL(selectedAvatarPreview.value)
    selectedAvatarPreview.value = ''
  }
  if (file) {
    selectedAvatarPreview.value = URL.createObjectURL(file)
  }
}

const submitProfile = async () => {
  if (isSaving.value) return

  errorMessage.value = ''
  successMessage.value = ''
  isSaving.value = true

  const dobValidationError = validateRealisticDateOfBirth(form.dateOfBirth)
  if (dobValidationError) {
    errorMessage.value = dobValidationError
    isSaving.value = false
    return
  }
  const firstNameError = validateLettersOnlyName('First name', form.firstName)
  if (firstNameError) {
    errorMessage.value = firstNameError
    isSaving.value = false
    return
  }
  const lastNameError = validateLettersOnlyName('Last name', form.lastName)
  if (lastNameError) {
    errorMessage.value = lastNameError
    isSaving.value = false
    return
  }
  const nicknameError = validateNickname(form.nickname)
  if (nicknameError) {
    errorMessage.value = nicknameError
    isSaving.value = false
    return
  }

  try {
    let response: Response
    if (selectedAvatarFile.value) {
      const formData = new FormData()
      formData.append('firstName', form.firstName)
      formData.append('lastName', form.lastName)
      formData.append('dateOfBirth', form.dateOfBirth)
      formData.append('nickname', form.nickname)
      formData.append('aboutMe', form.aboutMe)
      formData.append('avatarFile', selectedAvatarFile.value)

      response = await fetch(apiURL(API_ROUTES.USERS_ME), {
        method: 'PATCH',
        credentials: 'include',
        body: formData,
      })
    } else {
      response = await fetch(apiURL(API_ROUTES.USERS_ME), {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          firstName: form.firstName,
          lastName: form.lastName,
          dateOfBirth: form.dateOfBirth,
          nickname: form.nickname,
          aboutMe: form.aboutMe,
        }),
      })
    }

    if (!response.ok) {
      let message = 'Could not update profile.'
      const payload = (await response.json().catch(() => null)) as ErrorResponse | null
      if (payload?.error) message = payload.error
      errorMessage.value = message
      return
    }

    if (form.isPublic !== originalIsPublic.value) {
      const privacyResponse = await fetch(apiURL(API_ROUTES.USERS_ME_PRIVACY), {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({ isPublic: form.isPublic }),
      })

      if (!privacyResponse.ok) {
        let message = 'Profile updated but privacy could not be changed.'
        const payload = (await privacyResponse.json().catch(() => null)) as ErrorResponse | null
        if (payload?.error) message = payload.error
        errorMessage.value = message
        return
      }
    }

    successMessage.value = 'Profile updated.'
    await router.replace('/profile')
  } catch {
    errorMessage.value = 'Network error while updating profile.'
  } finally {
    isSaving.value = false
  }
}

onMounted(loadProfile)

onBeforeUnmount(() => {
  if (selectedAvatarPreview.value) {
    URL.revokeObjectURL(selectedAvatarPreview.value)
  }
})
</script>

<style scoped>
.edit-profile-page {
  min-height: calc(100dvh - var(--navbar-height, 60px));
  padding: 0.25rem 0 1.75rem;
}

.edit-profile-shell {
  width: min(100%, 1320px);
  margin: 0 auto;
  display: grid;
  gap: 1rem;
}

.edit-profile-hero {
  display: grid;
  grid-template-columns: minmax(0, 1.15fr) minmax(240px, 0.85fr);
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

.edit-profile-hero::before,
.edit-profile-hero::after {
  content: '';
  position: absolute;
  border-radius: 999px;
  pointer-events: none;
}

.edit-profile-hero::before {
  width: 220px;
  height: 220px;
  right: -40px;
  top: -60px;
  background: rgba(255, 255, 255, 0.08);
}

.edit-profile-hero::after {
  width: 160px;
  height: 160px;
  left: -40px;
  bottom: -60px;
  background: rgba(125, 211, 252, 0.12);
}

.edit-profile-hero-copy,
.edit-profile-hero-meta {
  position: relative;
  z-index: 1;
}

.edit-profile-kicker {
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

.edit-profile-hero h1 {
  margin-top: 1rem;
  max-width: 12ch;
  font-size: clamp(2.2rem, 5vw, 3.5rem);
  line-height: 0.96;
  font-weight: 900;
  color: var(--white);
}

.edit-profile-hero p {
  margin-top: 0.9rem;
  max-width: 46ch;
  color: rgba(255, 255, 255, 0.82);
  font-size: 1rem;
}

.edit-profile-hero-meta {
  display: grid;
  gap: 0.85rem;
  align-content: end;
}

.edit-profile-hero-meta article {
  padding: 1rem 1.05rem;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.12);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.edit-profile-hero-meta strong {
  display: block;
  color: var(--white);
  font-size: 1.55rem;
  font-weight: 900;
  overflow-wrap: anywhere;
}

.edit-profile-hero-meta span {
  color: rgba(255, 255, 255, 0.76);
  font-size: 0.86rem;
  font-weight: 700;
}

.edit-profile-card {
  width: 100%;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(248, 250, 252, 0.86));
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 32px;
  padding: 1.25rem;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.08);
}

.edit-profile-header {
  margin-bottom: 1.2rem;
}

.edit-profile-header h1 {
  color: var(--gray-900);
  font-size: 1.7rem;
  font-weight: 900;
}

.edit-profile-header p {
  color: var(--gray-500);
  font-weight: 600;
}

.edit-profile-form {
  display: grid;
  gap: 1rem;
}

.avatar-editor {
  display: grid;
  justify-items: center;
  gap: 0.6rem;
  padding-bottom: 0.75rem;
}

.avatar-preview img,
.avatar-fallback {
  width: 108px;
  height: 108px;
  border-radius: 999px;
}

.avatar-preview img {
  object-fit: cover;
  border: 4px solid rgba(255, 255, 255, 0.96);
  box-shadow: 0 18px 34px rgba(15, 23, 42, 0.14);
}

.avatar-fallback {
  display: grid;
  place-items: center;
  background: linear-gradient(135deg, rgba(191, 219, 254, 0.95), rgba(219, 234, 254, 0.9));
  color: var(--brand-700);
  font-weight: 900;
  font-size: 2rem;
  box-shadow: 0 18px 34px rgba(15, 23, 42, 0.14);
}

.change-avatar-label {
  color: var(--brand-600);
  font-weight: 800;
  cursor: pointer;
  text-decoration: none;
}

.change-avatar-label:hover {
  text-decoration: underline;
}

.hidden-file-input {
  display: none;
}

.grid-two {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

.field {
  display: grid;
  gap: 0.35rem;
}

.field span {
  color: var(--gray-900);
  font-weight: 800;
  font-size: 0.95rem;
}

.field input,
.field textarea {
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 20px;
  padding: 0.95rem 1rem;
  font-size: 0.95rem;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
  font-family: inherit;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 8px 22px rgba(15, 23, 42, 0.04);
}

.field textarea {
  width: 100%;
  max-width: 710px;
}

.file-name {
  color: var(--gray-500);
  font-size: 0.82rem;
  font-weight: 600;
}

.privacy-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.9rem 1rem;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 24px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.88));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 44px rgba(15, 23, 42, 0.08);
}

.privacy-copy p {
  color: var(--gray-900);
  font-weight: 800;
}

.privacy-copy small {
  color: var(--gray-500);
  font-weight: 600;
}

.privacy-switch {
  position: relative;
  display: inline-flex;
  width: 48px;
  height: 28px;
}

.privacy-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.privacy-slider {
  position: absolute;
  inset: 0;
  border-radius: 999px;
  background: #cbd5e1;
  transition: 0.2s ease;
  cursor: pointer;
}

.privacy-slider::before {
  content: '';
  position: absolute;
  width: 22px;
  height: 22px;
  left: 3px;
  top: 3px;
  border-radius: 999px;
  background: #ffffff;
  transition: 0.2s ease;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.25);
}

.privacy-switch input:checked + .privacy-slider {
  background: var(--brand-500);
}

.privacy-switch input:checked + .privacy-slider::before {
  transform: translateX(20px);
}

.field input:focus,
.field textarea:focus {
  outline: none;
  border-color: rgba(37, 99, 235, 0.55);
  box-shadow:
    0 0 0 5px rgba(37, 99, 235, 0.1),
    0 16px 34px rgba(37, 99, 235, 0.12);
}

.status-message {
  color: var(--gray-600);
  font-weight: 700;
}

.status-message.error {
  color: var(--status-error);
}

.status-message.success {
  color: var(--status-success);
}

.actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
}

.primary-btn,
.secondary-btn {
  border: none;
  border-radius: 18px;
  padding: 0.82rem 1rem;
  font-weight: 800;
  font-size: 0.9rem;
  cursor: pointer;
  text-decoration: none;
}

.primary-btn {
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  color: var(--text-white);
  box-shadow: 0 16px 30px rgba(37, 99, 235, 0.24);
}

.primary-btn:hover {
  transform: translateY(-1px);
}

.primary-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.secondary-btn {
  background:
    linear-gradient(180deg, rgba(238, 242, 255, 0.96), rgba(224, 231, 255, 0.88));
  color: var(--brand-700);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
}

@media (max-width: 900px) {
  .edit-profile-hero {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 650px) {
  .grid-two {
    grid-template-columns: 1fr;
  }

  .actions {
    flex-direction: column-reverse;
  }

  .edit-profile-card,
  .edit-profile-hero,
  .privacy-row {
    border-radius: 24px;
  }
}
</style>
