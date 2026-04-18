<template>
  <main class="login-page">
    <section class="login-shell">
      <aside class="login-showcase">
        <img class="auth-logo" src="@/assets/logo-white.svg" alt="Nexus" />
        <p class="showcase-kicker">Social network for your real circles</p>
        <h1>Pick up the conversation where your people left it.</h1>
        <p class="showcase-copy">
          Jump back into your feed, private groups, chats, and events with one secure sign in.
        </p>

        <div class="showcase-stats" aria-label="Platform highlights">
          <article>
            <strong>Live chats</strong>
            <span>Stay close to friends and groups in real time.</span>
          </article>
          <article>
            <strong>Private spaces</strong>
            <span>Share updates with the people you actually choose.</span>
          </article>
          <article>
            <strong>Smarter identity</strong>
            <span>Use your email or your nickname to sign in faster.</span>
          </article>
        </div>
      </aside>

      <section class="login-card">
        <header class="login-header">
          <span class="login-badge">Welcome back</span>
          <h2>Sign in to Nexus</h2>
          <p>Use your email or nickname to get back to your feed, messages, and groups.</p>
        </header>

        <form class="login-form" novalidate @submit.prevent="submitLogin">
          <label class="field">
            <span>Email or nickname</span>
            <input
              id="login-identifier"
              v-model.trim="form.identifier"
              type="text"
              name="identifier"
              placeholder="you@example.com or @nickname"
              autocomplete="username"
              :class="fieldClass('identifier', identifierError)"
              @blur="markTouched('identifier')"
              required
            />
            <small v-if="showFieldError('identifier', identifierError)" class="field-error">{{ identifierError }}</small>
          </label>

          <label class="field">
            <span>Password</span>
            <input
              id="login-password"
              v-model="form.password"
              type="password"
              name="password"
              placeholder="Enter your password"
              autocomplete="current-password"
              :class="fieldClass('password', passwordError)"
              @blur="markTouched('password')"
              required
            />
            <small v-if="showFieldError('password', passwordError)" class="field-error">{{ passwordError }}</small>
          </label>

          <p class="field-hint">You can type your nickname with or without `@`.</p>

          <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>

          <button type="submit" :disabled="isSubmitting">
            {{ isSubmitting ? 'Signing in...' : 'Sign in' }}
          </button>

          <p class="switch-auth">
            Don't have an account?
            <RouterLink to="/register">Create one</RouterLink>
          </p>
        </form>
      </section>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { API_ROUTES, apiURL } from '@/api/api'

type LoginRequest = {
  identifier: string
  password: string
}

type ErrorResponse = {
  error?: string
}

type LoginField = 'identifier' | 'password'

const router = useRouter()
const route = useRoute()
const form = reactive<LoginRequest>({
  identifier: '',
  password: '',
})

const touched = reactive<Record<LoginField, boolean>>({
  identifier: false,
  password: false,
})

const isSubmitting = ref(false)
const errorMessage = ref('')

const normalizedIdentifier = computed(() => form.identifier.trim().replace(/^@+/, ''))

const identifierError = computed(() => {
  if (!normalizedIdentifier.value) return 'Email or nickname is required.'
  if (normalizedIdentifier.value.length > 50) return 'Email or nickname is too long.'
  return ''
})

const passwordError = computed(() => {
  if (!form.password) return 'Password is required.'
  if (form.password.length < 8) return 'Password must be at least 8 characters.'
  return ''
})

const hasValidationErrors = computed(() => Boolean(identifierError.value || passwordError.value))

const markTouched = (field: LoginField) => {
  touched[field] = true
}

const showFieldError = (field: LoginField, error: string) => touched[field] && Boolean(error)

const fieldClass = (field: LoginField, error: string) => ({
  'is-invalid': touched[field] && Boolean(error),
  'is-valid': touched[field] && !error,
})

const submitLogin = async () => {
  if (isSubmitting.value) return

  errorMessage.value = ''
  touched.identifier = true
  touched.password = true

  if (hasValidationErrors.value) {
    errorMessage.value = 'Please fix the highlighted fields.'
    return
  }

  isSubmitting.value = true

  try {
    const response = await fetch(apiURL(API_ROUTES.AUTH_LOGIN), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({
        identifier: normalizedIdentifier.value,
        password: form.password,
      }),
    })

    if (!response.ok) {
      let message = 'Could not sign in'
      const payload = (await response.json().catch(() => null)) as ErrorResponse | null
      if (payload?.error) message = payload.error
      errorMessage.value = message
      return
    }

    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    await router.replace(redirect || '/')
  } catch {
    errorMessage.value = 'Network error. Please try again.'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 1.5rem;
  background:
    radial-gradient(circle at top left, rgba(59, 130, 246, 0.34), transparent 28%),
    radial-gradient(circle at bottom right, rgba(14, 165, 233, 0.22), transparent 32%),
    linear-gradient(135deg, #081225 0%, #0f1f3a 45%, #eff6ff 45%, #f8fbff 100%);
}

.login-shell {
  width: min(1240px, 100%);
  min-height: 680px;
  display: grid;
  grid-template-columns: 1.18fr 0.82fr;
  border-radius: 32px;
  overflow: hidden;
  box-shadow: 0 28px 90px rgba(8, 18, 37, 0.28);
  background: rgba(255, 255, 255, 0.76);
  backdrop-filter: blur(16px);
}

.login-showcase {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 1.25rem;
  padding: 3.25rem;
  color: var(--white);
  background:
    linear-gradient(180deg, rgba(37, 99, 235, 0.1), rgba(8, 18, 37, 0.08)),
    linear-gradient(160deg, #0f62fe 0%, #1245a8 45%, #081225 100%);
}

.login-showcase::before,
.login-showcase::after {
  content: '';
  position: absolute;
  border-radius: 50%;
  pointer-events: none;
}

.login-showcase::before {
  width: 320px;
  height: 320px;
  top: -90px;
  right: -60px;
  background: rgba(255, 255, 255, 0.1);
}

.login-showcase::after {
  width: 240px;
  height: 240px;
  bottom: -80px;
  left: -60px;
  background: rgba(125, 211, 252, 0.12);
}

.auth-logo {
  width: 132px;
  height: auto;
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

.login-showcase h1 {
  position: relative;
  z-index: 1;
  max-width: 13ch;
  font-size: clamp(2.35rem, 4.6vw, 3.95rem);
  line-height: 1.02;
  font-weight: 900;
}

.showcase-copy {
  position: relative;
  z-index: 1;
  max-width: 44ch;
  font-size: 1rem;
  color: rgba(255, 255, 255, 0.82);
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

.login-card {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 2.4rem;
  background:
    radial-gradient(circle at top, rgba(219, 234, 254, 0.65), transparent 28%),
    rgba(255, 255, 255, 0.92);
}

.login-card::before {
  content: '';
  width: min(100%, 420px);
  height: 100%;
  position: absolute;
  inset: 0;
  margin: auto;
  pointer-events: none;
}

.login-header,
.login-form {
  width: min(100%, 360px);
}

.login-header {
  margin-bottom: 1.75rem;
}

.login-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.35rem 0.75rem;
  border-radius: var(--radius-full);
  background: var(--brand-50);
  color: var(--brand-700);
  font-size: 0.78rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.login-header h2 {
  margin-top: 1rem;
  color: var(--gray-900);
  font-size: clamp(1.9rem, 3.2vw, 2.6rem);
  line-height: 0.98;
  font-weight: 900;
}

.login-header p {
  margin-top: 0.75rem;
  color: var(--gray-500);
  font-size: 0.95rem;
}

.login-form {
  display: grid;
  gap: 1rem;
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

.field input {
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

.field input::placeholder {
  color: #94a3b8;
}

.field input:focus {
  outline: none;
  border-color: rgba(37, 99, 235, 0.55);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 1), rgba(239, 246, 255, 0.92));
  box-shadow:
    0 0 0 5px rgba(37, 99, 235, 0.1),
    0 16px 34px rgba(37, 99, 235, 0.12);
  transform: translateY(-2px);
}

.field input.is-invalid {
  border-color: rgba(220, 38, 38, 0.7);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 1), rgba(254, 242, 242, 0.9));
}

.field input.is-valid {
  border-color: rgba(22, 163, 74, 0.58);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 1), rgba(240, 253, 244, 0.92));
}

.field-error {
  color: var(--danger);
  font-size: var(--text-sm);
  font-weight: 700;
}

.field-hint {
  color: var(--gray-500);
  font-size: 0.84rem;
  line-height: 1.45;
  margin-top: -0.15rem;
}

.error-message {
  color: var(--danger);
  font-size: 0.9rem;
  font-weight: 700;
  background: var(--danger-bg);
  padding: 0.8rem 0.95rem;
  border-radius: 16px;
  border: 1px solid rgba(220, 38, 38, 0.12);
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

@media (max-width: 900px) {
  .login-page {
    padding: 1rem;
    background:
      radial-gradient(circle at top left, rgba(59, 130, 246, 0.28), transparent 30%),
      linear-gradient(180deg, #081225 0%, #113164 42%, #eff6ff 42%, #f8fbff 100%);
  }

  .login-shell {
    min-height: auto;
    grid-template-columns: 1fr;
  }

  .login-showcase,
  .login-card {
    padding: 2rem 1.35rem;
  }

  .login-showcase h1 {
    max-width: 15ch;
    font-size: clamp(2.1rem, 9vw, 3rem);
  }
}

@media (max-width: 540px) {
  .login-shell {
    border-radius: 24px;
  }

  .showcase-stats {
    grid-template-columns: 1fr;
  }

  .field input,
  button[type='submit'] {
    border-radius: 16px;
  }
}
</style>
