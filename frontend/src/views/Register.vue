<template>
  <main class="register-page">
    <section class="register-shell">
      <aside class="register-showcase">
        <img class="auth-logo" src="@/assets/logo-white.svg" alt="Nexus" />
        <p class="showcase-kicker">Build your circle your way</p>
        <h1>Create a profile that feels ready to connect.</h1>
        <p class="showcase-copy">
          Join Nexus to share updates, message friends, join private groups, and shape your own
          social space from day one.
        </p>

        <div
          class="progress-bar-wrap"
          role="progressbar"
          :aria-valuenow="formProgress"
          aria-valuemin="0"
          aria-valuemax="100"
        >
          <div class="progress-bar-fill" :style="{ width: formProgress + '%' }"></div>
        </div>

        <div class="showcase-stats" aria-label="Platform highlights">
          <article>
            <strong>Private by choice</strong>
            <span>Pick a public or private profile before you even post.</span>
          </article>
          <article>
            <strong>Real identity</strong>
            <span>Add your nickname, photo, and short intro to feel familiar fast.</span>
          </article>
          <article>
            <strong>Ready to join</strong>
            <span>Groups, chats, and events are waiting once your account is live.</span>
          </article>
        </div>
      </aside>

      <section class="register-card">
        <header class="register-header">
          <span class="register-badge">Join Nexus</span>
          <h2>Create account</h2>
          <p>Set up your profile and start building your network in a few quick steps.</p>
        </header>

        <form class="register-form" novalidate @submit.prevent="submitRegister">
          <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>
          <p v-if="successMessage" class="success-message">{{ successMessage }}</p>

          <div class="grid-two">
            <label class="field">
              <span>First name</span>
              <input
                id="reg-first-name"
                v-model.trim="form.firstName"
                type="text"
                name="firstName"
                placeholder="First name"
                autocomplete="given-name"
                maxlength="25"
                :class="fieldClass('firstName', firstNameError)"
                @blur="markTouched('firstName')"
                required
              />
              <small v-if="showFieldError('firstName', firstNameError)" class="field-error">{{ firstNameError }}</small>
            </label>

            <label class="field">
              <span>Last name</span>
              <input
                id="reg-last-name"
                v-model.trim="form.lastName"
                type="text"
                name="lastName"
                placeholder="Last name"
                autocomplete="family-name"
                maxlength="25"
                :class="fieldClass('lastName', lastNameError)"
                @blur="markTouched('lastName')"
                required
              />
              <small v-if="showFieldError('lastName', lastNameError)" class="field-error">{{ lastNameError }}</small>
            </label>
          </div>

          <label class="field">
            <span>Nickname (Optional)</span>
            <input
              id="reg-nickname"
              v-model.trim="form.nickname"
              type="text"
              name="nickname"
              placeholder="Your nickname"
              autocomplete="nickname"
              maxlength="25"
              :class="fieldClass('nickname', nicknameError, form.nickname.trim().length > 0)"
              @blur="markTouched('nickname')"
            />
            <small v-if="showFieldError('nickname', nicknameError)" class="field-error">{{ nicknameError }}</small>
          </label>

          <label class="field">
            <span>Email</span>
            <input
              id="reg-email"
              v-model.trim="form.email"
              type="email"
              name="email"
              placeholder="you@example.com"
              autocomplete="email"
              :class="fieldClass('email', emailError)"
              @blur="markTouched('email')"
              required
            />
            <small v-if="showFieldError('email', emailError)" class="field-error">{{ emailError }}</small>
          </label>

          <label class="field">
            <span>Date of birth</span>
            <input
              id="reg-dob"
              v-model="form.dateOfBirth"
              type="date"
              name="dateOfBirth"
              autocomplete="bday"
              :min="dobBounds.min"
              :max="dobBounds.max"
              :class="fieldClass('dateOfBirth', dateOfBirthError)"
              @blur="markTouched('dateOfBirth')"
              required
            />
            <small v-if="showFieldError('dateOfBirth', dateOfBirthError)" class="field-error">{{ dateOfBirthError }}</small>
          </label>

          <div class="grid-two profile-stack">
            <label class="field">
              <span>Avatar (Optional)</span>
              <span class="avatar-dropzone" :class="{ 'has-file': avatarFileName }">
                <img v-if="avatarPreviewUrl" :src="avatarPreviewUrl" alt="Preview" class="avatar-preview" />
                <span v-else class="avatar-dropzone-inner">
                  <span class="avatar-upload-icon">+</span>
                  <span class="avatar-upload-text">Upload avatar</span>
                  <span class="avatar-upload-hint">PNG, JPG, GIF, WebP</span>
                </span>
                <span v-if="avatarFileName" class="avatar-file-name">{{ avatarFileName }}</span>
                <input
                  id="reg-avatar"
                  type="file"
                  name="avatar"
                  accept="image/png,image/jpeg,image/gif,image/webp"
                  class="avatar-file-input"
                  @change="onAvatarSelected"
                />
              </span>
            </label>
          </div>

          <label class="field">
            <span>About me (Optional)</span>
            <textarea
              id="reg-about-me"
              v-model.trim="form.aboutMe"
              name="aboutMe"
              placeholder="Tell people a bit about yourself..."
              rows="3"
              maxlength="255"
              :class="fieldClass('aboutMe', aboutMeError, form.aboutMe.trim().length > 0)"
              @blur="markTouched('aboutMe')"
            />
            <small v-if="showFieldError('aboutMe', aboutMeError)" class="field-error">{{ aboutMeError }}</small>
          </label>

          <div class="grid-two">
            <label class="field">
              <span>Password</span>
              <input
                id="reg-password"
                v-model="form.password"
                type="password"
                name="password"
                placeholder="At least 8 characters"
                autocomplete="new-password"
                :class="fieldClass('password', passwordError)"
                @blur="markTouched('password')"
                required
              />
              <small v-if="showFieldError('password', passwordError)" class="field-error">{{ passwordError }}</small>
            </label>

            <label class="field">
              <span>Confirm password</span>
              <input
                id="reg-confirm-password"
                v-model="confirmPassword"
                type="password"
                name="confirmPassword"
                placeholder="Repeat password"
                autocomplete="new-password"
                :class="fieldClass('confirmPassword', confirmPasswordError)"
                @blur="markTouched('confirmPassword')"
                required
              />
              <small v-if="showFieldError('confirmPassword', confirmPasswordError)" class="field-error">{{ confirmPasswordError }}</small>
            </label>
          </div>

          <div class="privacy-row">
            <div class="privacy-copy">
              <p>Account privacy</p>
              <small>{{ form.isPublic ? 'Public profile' : 'Private profile' }}</small>
            </div>
            <label class="privacy-switch" aria-label="Toggle account privacy">
              <input id="reg-is-public" v-model="form.isPublic" type="checkbox" name="isPublic" />
              <span class="privacy-slider"></span>
            </label>
          </div>

          <button type="submit" :disabled="isSubmitting">
            {{ isSubmitting ? 'Creating account...' : 'Create account' }}
          </button>

          <p class="switch-auth">
            Already have an account?
            <RouterLink to="/login">Sign in</RouterLink>
          </p>
        </form>
      </section>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { API_ROUTES, apiURL } from '@/api/api'
import {
  getDateOfBirthBounds,
  validateLettersOnlyName,
  validateNickname,
  validateRealisticDateOfBirth,
} from '@/utils/helpers'

type RegisterRequest = {
  firstName: string
  lastName: string
  email: string
  dateOfBirth: string
  password: string
  avatar: string
  nickname: string
  aboutMe: string
  isPublic: boolean
}

type ErrorResponse = {
  error?: string
}

type RegisterField =
  | 'firstName'
  | 'lastName'
  | 'email'
  | 'dateOfBirth'
  | 'nickname'
  | 'aboutMe'
  | 'password'
  | 'confirmPassword'

const router = useRouter()
const form = reactive<RegisterRequest>({
  firstName: '',
  lastName: '',
  email: '',
  dateOfBirth: '',
  password: '',
  avatar: '',
  nickname: '',
  aboutMe: '',
  isPublic: true,
})

const confirmPassword = ref('')
const avatarFile = ref<File | null>(null)
const avatarFileName = ref('')
const avatarPreviewUrl = ref('')
const isSubmitting = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const dobBounds = getDateOfBirthBounds()
const touched = reactive<Record<RegisterField, boolean>>({
  firstName: false,
  lastName: false,
  email: false,
  dateOfBirth: false,
  nickname: false,
  aboutMe: false,
  password: false,
  confirmPassword: false,
})
const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

const firstNameError = computed(() => validateLettersOnlyName('First name', form.firstName) || '')
const lastNameError = computed(() => validateLettersOnlyName('Last name', form.lastName) || '')
const emailError = computed(() => {
  if (!form.email) return 'Email is required.'
  if (!EMAIL_REGEX.test(form.email)) return 'Please enter a valid email address.'
  return ''
})
const dateOfBirthError = computed(() => validateRealisticDateOfBirth(form.dateOfBirth) || '')
const nicknameError = computed(() => validateNickname(form.nickname) || '')
const aboutMeError = computed(() => {
  if (form.aboutMe.trim().length > 255) return 'About me must be 255 characters or fewer.'
  return ''
})
const passwordError = computed(() => {
  if (!form.password) return 'Password is required.'
  if (form.password.length < 8) return 'Password must be at least 8 characters.'
  return ''
})
const confirmPasswordError = computed(() => {
  if (!confirmPassword.value) return 'Confirm password is required.'
  if (form.password !== confirmPassword.value) return 'Passwords do not match.'
  return ''
})
const hasValidationErrors = computed(() =>
  Boolean(
    firstNameError.value ||
      lastNameError.value ||
      emailError.value ||
      dateOfBirthError.value ||
      nicknameError.value ||
      aboutMeError.value ||
      passwordError.value ||
      confirmPasswordError.value,
  ),
)

const formProgress = computed(() => {
  const checks = [
    !firstNameError.value && form.firstName,
    !lastNameError.value && form.lastName,
    !emailError.value && form.email,
    !dateOfBirthError.value && form.dateOfBirth,
    !passwordError.value && form.password,
    !confirmPasswordError.value && confirmPassword.value,
  ]
  const filled = checks.filter(Boolean).length
  return Math.round((filled / checks.length) * 100)
})

const markTouched = (field: RegisterField) => {
  touched[field] = true
}

const touchAllFields = () => {
  const keys = Object.keys(touched) as RegisterField[]
  keys.forEach((key) => {
    touched[key] = true
  })
}

const showFieldError = (field: RegisterField, error: string) => touched[field] && Boolean(error)

const fieldClass = (field: RegisterField, error: string, allowValidState = true) => ({
  'is-invalid': touched[field] && Boolean(error),
  'is-valid': touched[field] && !error && allowValidState,
})

const onAvatarSelected = (event: Event) => {
  const input = event.target as HTMLInputElement | null
  const selected = input?.files?.[0] ?? null
  avatarFile.value = selected
  avatarFileName.value = selected ? selected.name : ''
  if (avatarPreviewUrl.value) URL.revokeObjectURL(avatarPreviewUrl.value)
  avatarPreviewUrl.value = selected ? URL.createObjectURL(selected) : ''
}

const submitRegister = async () => {
  if (isSubmitting.value) return

  errorMessage.value = ''
  successMessage.value = ''
  touchAllFields()

  if (hasValidationErrors.value) {
    errorMessage.value = 'Please fix the highlighted fields.'
    return
  }

  isSubmitting.value = true
  try {
    const requestInit: RequestInit = {
      method: 'POST',
      credentials: 'include',
    }

    if (avatarFile.value) {
      const formData = new FormData()
      formData.append('firstName', form.firstName)
      formData.append('lastName', form.lastName)
      formData.append('email', form.email)
      formData.append('dateOfBirth', form.dateOfBirth)
      formData.append('password', form.password)
      formData.append('nickname', form.nickname)
      formData.append('aboutMe', form.aboutMe)
      formData.append('isPublic', String(form.isPublic))
      formData.append('avatar', avatarFile.value)
      requestInit.body = formData
    } else {
      requestInit.headers = { 'Content-Type': 'application/json' }
      requestInit.body = JSON.stringify(form)
    }

    const response = await fetch(apiURL(API_ROUTES.AUTH_REGISTER), requestInit)

    if (!response.ok) {
      let message = 'Could not create account'
      const payload = (await response.json().catch(() => null)) as ErrorResponse | null
      if (payload?.error) message = payload.error
      errorMessage.value = message
      return
    }

    successMessage.value = 'Account created. Redirecting to login...'
    await router.replace('/login')
  } catch {
    errorMessage.value = 'Network error. Please try again.'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.register-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 1.5rem;
  background:
    radial-gradient(circle at top left, rgba(59, 130, 246, 0.34), transparent 28%),
    radial-gradient(circle at bottom right, rgba(14, 165, 233, 0.22), transparent 32%),
    linear-gradient(135deg, #081225 0%, #0f1f3a 45%, #eff6ff 45%, #f8fbff 100%);
}

.register-shell {
  width: min(1380px, 100%);
  min-height: 860px;
  display: grid;
  grid-template-columns: 1.12fr 0.88fr;
  border-radius: 32px;
  overflow: hidden;
  box-shadow: 0 28px 90px rgba(8, 18, 37, 0.28);
  background: rgba(255, 255, 255, 0.82);
  backdrop-filter: blur(16px);
}

.register-showcase {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 1.2rem;
  padding: 3.25rem;
  min-height: 100%;
  color: var(--white);
  background:
    linear-gradient(180deg, rgba(37, 99, 235, 0.1), rgba(8, 18, 37, 0.08)),
    linear-gradient(160deg, #0f62fe 0%, #1245a8 45%, #081225 100%);
}

.register-showcase::before,
.register-showcase::after {
  content: '';
  position: absolute;
  border-radius: 50%;
  pointer-events: none;
}

.register-showcase::before {
  width: 320px;
  height: 320px;
  top: -90px;
  right: -60px;
  background: rgba(255, 255, 255, 0.1);
}

.register-showcase::after {
  width: 240px;
  height: 240px;
  bottom: -80px;
  left: -60px;
  background: rgba(125, 211, 252, 0.12);
}

.auth-logo {
  width: 132px;
  height: auto;
  position: relative;
  z-index: 1;
}

.showcase-kicker {
  position: relative;
  z-index: 1;
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.72);
}

.register-showcase h1 {
  position: relative;
  z-index: 1;
  max-width: 12ch;
  font-size: clamp(2.3rem, 4.8vw, 4rem);
  line-height: 1.02;
  font-weight: 900;
  color: var(--white);
}

.showcase-copy {
  position: relative;
  z-index: 1;
  max-width: 44ch;
  font-size: 1rem;
  color: rgba(255, 255, 255, 0.82);
}

.progress-bar-wrap {
  position: relative;
  z-index: 1;
  width: min(100%, 420px);
  height: 8px;
  background: rgba(255, 255, 255, 0.14);
  border-radius: 999px;
  overflow: hidden;
  margin-top: 0.3rem;
}

.progress-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #7dd3fc, #ffffff);
  border-radius: inherit;
  transition: width 300ms ease;
}

.showcase-stats {
  position: relative;
  z-index: 1;
  display: grid;
  gap: 0.9rem;
  margin-top: 1.2rem;
}

.showcase-stats article {
  padding: 1rem 1.05rem;
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.08);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.showcase-stats strong {
  display: block;
  margin-bottom: 0.2rem;
  font-size: 0.98rem;
  font-weight: 800;
}

.showcase-stats span {
  color: rgba(255, 255, 255, 0.74);
  font-size: 0.92rem;
}

.register-card {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100%;
  padding: 3rem 2.6rem;
  background:
    radial-gradient(circle at top, rgba(219, 234, 254, 0.65), transparent 28%),
    rgba(255, 255, 255, 0.92);
}

.register-header,
.register-form {
  width: min(100%, 560px);
}

.register-header {
  margin-bottom: 1.5rem;
}

.register-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.35rem 0.75rem;
  border-radius: var(--radius-full);
  background: var(--brand-50);
  color: var(--brand-700);
  font-size: 0.78rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.register-header h2 {
  margin-top: 1rem;
  color: var(--gray-900);
  font-size: clamp(2rem, 3.4vw, 2.7rem);
  line-height: 1;
  font-weight: 900;
}

.register-header p {
  margin-top: 0.75rem;
  color: var(--gray-500);
  font-size: 0.96rem;
}

.register-form {
  display: grid;
  gap: 1rem;
}

.grid-two {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.85rem;
}

.profile-stack {
  grid-template-columns: 1fr;
}

.field {
  display: grid;
  gap: 0.42rem;
}

.field span {
  color: var(--text-primary);
  font-size: 0.78rem;
  font-weight: 800;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.field input,
.field textarea {
  width: 100%;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 20px;
  padding: 1rem 1.05rem;
  font-size: 0.98rem;
  font-weight: 500;
  color: var(--gray-900);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.95),
    0 8px 22px rgba(15, 23, 42, 0.04);
  transition:
    border-color 150ms ease,
    box-shadow 150ms ease,
    transform 150ms ease,
    background 150ms ease;
}

.field textarea {
  resize: vertical;
  font-family: inherit;
  min-height: 108px;
}

.field input::placeholder,
.field textarea::placeholder {
  color: #94a3b8;
}

.field input:focus,
.field textarea:focus {
  outline: none;
  border-color: rgba(37, 99, 235, 0.55);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 1), rgba(239, 246, 255, 0.92));
  box-shadow:
    0 0 0 5px rgba(37, 99, 235, 0.1),
    0 16px 34px rgba(37, 99, 235, 0.12);
  transform: translateY(-2px);
}

.field input.is-invalid,
.field textarea.is-invalid {
  border-color: rgba(220, 38, 38, 0.7);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 1), rgba(254, 242, 242, 0.9));
}

.field input.is-valid,
.field textarea.is-valid {
  border-color: rgba(22, 163, 74, 0.58);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 1), rgba(240, 253, 244, 0.92));
}

.field-error {
  color: var(--danger);
  font-size: var(--text-sm);
  font-weight: 700;
}

.avatar-dropzone {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  min-height: 132px;
  border: 1px dashed rgba(96, 165, 250, 0.5);
  border-radius: 20px;
  background:
    radial-gradient(circle at top, rgba(219, 234, 254, 0.95), rgba(239, 246, 255, 0.75));
  cursor: pointer;
  overflow: hidden;
  transition:
    border-color var(--dur-fast),
    background var(--dur-fast),
    transform var(--dur-fast),
    box-shadow var(--dur-fast);
  padding: 1rem;
  text-align: center;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.95),
    0 10px 24px rgba(37, 99, 235, 0.06);
}

.avatar-dropzone:hover,
.avatar-dropzone:focus-within {
  border-color: var(--brand-500);
  background:
    radial-gradient(circle at top, rgba(219, 234, 254, 1), rgba(224, 242, 254, 0.9));
  transform: translateY(-2px);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.95),
    0 16px 30px rgba(37, 99, 235, 0.12);
}

.avatar-dropzone.has-file {
  border-color: var(--brand-500);
}

.avatar-file-input {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
  width: 100%;
  height: 100%;
}

.avatar-dropzone-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.25rem;
  pointer-events: none;
}

.avatar-upload-icon {
  display: inline-grid;
  place-items: center;
  width: 34px;
  height: 34px;
  border-radius: 999px;
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  color: var(--white);
  font-size: 1.1rem;
  font-weight: 800;
  box-shadow: 0 10px 20px rgba(37, 99, 235, 0.24);
}

.avatar-upload-text {
  font-size: 0.9rem;
  font-weight: 800;
  color: var(--text-primary);
}

.avatar-upload-hint {
  font-size: 0.77rem;
  color: var(--gray-500);
  font-weight: 600;
}

.avatar-preview {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid var(--brand-500);
}

.avatar-file-name {
  font-size: var(--text-xs);
  color: var(--brand-600);
  font-weight: 700;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.privacy-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 1rem 1.05rem;
  border: 1px solid rgba(148, 163, 184, 0.18);
  border-radius: 20px;
  background:
    linear-gradient(180deg, rgba(239, 246, 255, 0.95), rgba(219, 234, 254, 0.72));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.95),
    0 12px 26px rgba(37, 99, 235, 0.06);
}

.privacy-copy p {
  color: var(--text-primary);
  font-weight: 800;
  font-size: 0.98rem;
}

.privacy-copy small {
  color: var(--gray-500);
  font-weight: 600;
  font-size: 0.82rem;
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
  border-radius: var(--radius-full);
  background: var(--gray-300);
  transition: background 200ms ease;
  cursor: pointer;
}

.privacy-slider::before {
  content: '';
  position: absolute;
  width: 22px;
  height: 22px;
  left: 3px;
  top: 3px;
  border-radius: var(--radius-full);
  background: var(--white);
  transition: transform 200ms ease;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.25);
}

.privacy-switch input:checked + .privacy-slider {
  background: var(--brand-500);
}

.privacy-switch input:checked + .privacy-slider::before {
  transform: translateX(20px);
}

.error-message,
.success-message {
  font-size: 0.9rem;
  font-weight: 700;
  padding: 0.8rem 0.95rem;
  border-radius: 16px;
}

.error-message {
  color: var(--danger);
  background: var(--danger-bg);
  border: 1px solid rgba(220, 38, 38, 0.12);
}

.success-message {
  color: var(--success);
  background: var(--success-bg);
  border: 1px solid rgba(22, 163, 74, 0.12);
}

button[type='submit'] {
  border: none;
  border-radius: 18px;
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  color: var(--white);
  font-weight: 800;
  font-size: 1rem;
  padding: 0.95rem 1rem;
  width: 100%;
  cursor: pointer;
  box-shadow: 0 16px 30px rgba(37, 99, 235, 0.24);
  transition:
    transform var(--dur-fast) var(--ease-standard),
    box-shadow var(--dur-fast) var(--ease-standard),
    filter var(--dur-fast) var(--ease-standard);
}

button[type='submit']:not(:disabled):hover {
  transform: translateY(-1px);
  box-shadow: 0 20px 34px rgba(37, 99, 235, 0.3);
  filter: saturate(1.06);
}

button[type='submit']:active:not(:disabled) {
  transform: translateY(0);
}

button[type='submit']:disabled {
  opacity: 0.68;
  cursor: not-allowed;
  box-shadow: none;
}

.switch-auth {
  color: var(--gray-500);
  font-size: 0.94rem;
  text-align: center;
}

.switch-auth a {
  color: var(--brand-600);
  font-weight: 800;
  text-decoration: none;
}

.switch-auth a:hover {
  text-decoration: underline;
}

@media (max-width: 980px) {
  .register-page {
    padding: 1rem;
    background:
      radial-gradient(circle at top left, rgba(59, 130, 246, 0.28), transparent 30%),
      linear-gradient(180deg, #081225 0%, #113164 38%, #eff6ff 38%, #f8fbff 100%);
  }

  .register-shell {
    min-height: auto;
    grid-template-columns: 1fr;
  }

  .register-showcase,
  .register-card {
    padding: 2rem 1.35rem;
  }

  .register-showcase h1 {
    max-width: 15ch;
    font-size: clamp(2.1rem, 9vw, 3rem);
  }
}

@media (max-width: 640px) {
  .register-shell {
    border-radius: 24px;
  }

  .register-header,
  .register-form {
    width: 100%;
  }

  .grid-two {
    grid-template-columns: 1fr;
  }

  .field input,
  .field textarea,
  .avatar-dropzone,
  .privacy-row,
  button[type='submit'] {
    border-radius: 16px;
  }
}
</style>
