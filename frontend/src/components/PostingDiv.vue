<template>
  <div class="form-overlay">
    <form ref="formRef" class="form-container composer-modal" @submit.prevent="onSubmit">
      <header class="composer-head">
        <div>
          <span class="composer-kicker">Create post</span>
          <h2>Share something with your network</h2>
          <p>Write an update, choose who can see it, and add an image if it helps tell the story.</p>
        </div>
        <button type="button" class="close-btn" @click="emit('close')">✕</button>
      </header>

      <div v-if="submitError" class="SubmitPostError">
        <p>{{ submitError }}</p>
      </div>

      <div class="form-group">
        <div class="field-head">
          <label for="Title">Title</label>
          <span class="char-counter" :class="{ 'char-warn': form.title.length >= 50, 'char-limit': form.title.length >= 60 }">
            {{ form.title.length }} / 60
          </span>
        </div>
        <input
          id="Title"
          v-model="form.title"
          name="title"
          type="text"
          maxlength="60"
          placeholder="Give your post a short title"
          :class="{ 'has-error': errors.title }"
          @blur="validateTitle"
          @input="validateTitle"
        />
        <span v-if="errors.title" class="error-text">{{ errors.title }}</span>
      </div>

      <div class="form-group">
        <div class="field-head">
          <label for="Body">Body</label>
          <span class="char-counter" :class="{ 'char-warn': form.body.length >= 4800, 'char-limit': form.body.length >= 5000 }">
            {{ form.body.length }} / 5000
          </span>
        </div>
        <textarea
          id="Body"
          v-model="form.body"
          name="Body"
          rows="6"
          maxlength="5000"
          placeholder="What's on your mind?"
          :class="{ 'has-error': errors.body }"
          @blur="validateBody"
          @input="validateBody"
        ></textarea>
        <span v-if="errors.body" class="error-text">{{ errors.body }}</span>
      </div>

      <div class="form-group">
        <label for="Image">Image</label>
        <input
          id="Image"
          ref="fileInputRef"
          name="Image"
          type="file"
          accept="image/jpeg, image/png, image/gif"
          class="hidden-input"
          @change="handleFileUpload"
        />

        <div v-if="!form.image" class="upload-box" :class="{ 'has-error': errors.image }" @click="triggerFileInput">
          <span class="upload-icon">
            <CameraIcon class="nav-icon" />
          </span>
          <p>Click to upload an image</p>
          <small>JPG, PNG, GIF up to 10MB</small>
        </div>

        <div v-else class="image-preview-card">
          <div class="file-info">
            <span class="file-icon"><PhotoIcon class="nav-icon" /></span>
            <span class="file-name">{{ form.image.name }}</span>
          </div>
          <button type="button" class="remove-btn" title="Remove image" @click="clearImage">
            <TrashIcon class="small-icon" />
          </button>
        </div>
      </div>

      <div class="form-group">
        <label for="Privacy">Privacy</label>
        <select id="Privacy" v-model="form.privacy" name="privacy" class="privacy-select">
          <option value="public">Public</option>
          <option value="almost_private">Followers Only</option>
          <option value="private">Private</option>
        </select>
      </div>

      <div class="composer-grid">
        <div class="form-group">
          <div class="privacy-note">
            <span class="privacy-note-kicker">Visibility</span>
            <p>Choose whether this post is public, followers only, or private to selected people.</p>
          </div>
        </div>
      </div>

      <div v-if="form.privacy === 'private'" class="followers-panel">
        <div class="followers-head">
          <h3>Choose followers</h3>
          <p>Select who can view this private post.</p>
        </div>

        <div class="followers-container">
          <p v-if="isLoadingFollowers" class="loading-text">Loading followers...</p>
          <p v-else-if="followers.length === 0" class="loading-text">No followers yet to add.</p>

          <label v-else v-for="user in followers" :key="user.id" class="follower-option" :for="user.id">
            <div class="follower-info-wrapper">
              <img v-if="user.avatar" :src="`${API_BASE_URL}${user.avatar}`" class="follower-avatar" alt="Avatar" />
              <span v-else class="avatar-fallback sm">
                {{ ((user.firstName?.[0] || '?') + (user.lastName?.[0] || '?')).toUpperCase() }}
              </span>
              <span class="follower-name">{{ user.nickname || `${user.firstName} ${user.lastName}` }}</span>
            </div>
            <input :id="user.id" v-model="form.selectedFollowers" type="checkbox" :value="user.id" class="follower-checkbox" />
          </label>
        </div>

        <span v-if="errors.selectedFollowers" class="error-text">{{ errors.selectedFollowers }}</span>
      </div>

      <div class="composer-actions">
        <button type="button" class="ghost-btn" @click="emit('close')">Cancel</button>
        <button
          type="submit"
          class="submit-btn"
          :disabled="isSubmitting || !!errors.title || !!errors.body || !!errors.image || !!errors.selectedFollowers"
        >
          <span v-if="isSubmitting" class="loader"></span>
          <span v-else>Publish post</span>
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { CameraIcon, PhotoIcon, TrashIcon } from '@heroicons/vue/24/solid'
import { API_BASE_URL } from '@/api/api'
import { usePostForm } from '@/composables/useCreatePost'
import { useClickOutside } from '@/composables/useClickOutside'

const emit = defineEmits(['close'])
const formRef = ref<HTMLFormElement | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)

const {
  form,
  errors,
  isSubmitting,
  submitError,
  followers,
  isLoadingFollowers,
  validateTitle,
  validateBody,
  handleFileUpload,
  submitPost,
  removeImage,
} = usePostForm()

const onSubmit = () => {
  submitPost(() => {
    setTimeout(() => emit('close'), 500)
  })
}

useClickOutside(formRef, () => {
  emit('close')
})

const triggerFileInput = () => {
  fileInputRef.value?.click()
}

const clearImage = () => {
  removeImage()
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
  }
}
</script>

<style scoped>
.form-overlay {
  position: fixed;
  inset: 0;
  z-index: 500;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.5rem;
  background: rgba(8, 18, 37, 0.48);
  backdrop-filter: blur(8px);
}

.composer-modal {
  width: min(calc(100vw - 1rem), 1080px);
  max-width: none;
  max-height: min(88vh, 920px);
  overflow-y: auto;
  padding: 1.4rem;
  border-radius: 32px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background:
    radial-gradient(circle at top right, rgba(219, 234, 254, 0.8), transparent 24%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.92));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 32px 80px rgba(15, 23, 42, 0.22);
}

.composer-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1.2rem;
}

.composer-kicker {
  display: inline-flex;
  align-items: center;
  padding: 0.35rem 0.72rem;
  border-radius: 999px;
  background: var(--brand-50);
  color: var(--brand-700);
  font-size: 0.76rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.composer-head h2 {
  margin-top: 0.85rem;
  color: var(--gray-900);
  font-size: clamp(1.9rem, 4vw, 2.5rem);
  line-height: 0.98;
  font-weight: 900;
}

.composer-head p {
  margin-top: 0.7rem;
  max-width: 48ch;
  color: var(--gray-500);
  font-size: 0.94rem;
}

.close-btn {
  width: 42px;
  height: 42px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.84);
  color: var(--gray-600);
  font-size: 1rem;
  cursor: pointer;
}

.SubmitPostError p {
  margin-bottom: 1rem;
  padding: 0.85rem 1rem;
  border-radius: 16px;
  background: var(--danger-bg);
  color: var(--danger);
  font-weight: 700;
}

.form-group {
  display: grid;
  gap: 0.45rem;
  margin-bottom: 1rem;
}

.composer-grid {
  display: none;
}

.field-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

label,
.followers-head h3 {
  color: var(--gray-900);
  font-size: 0.78rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

input,
textarea,
select {
  width: 100%;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 20px;
  padding: 1rem 1.05rem;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  color: var(--gray-900);
  font-size: 0.97rem;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 8px 22px rgba(15, 23, 42, 0.04);
  transition:
    border-color var(--dur-fast) var(--ease-standard),
    box-shadow var(--dur-fast) var(--ease-standard),
    transform var(--dur-fast) var(--ease-standard);
}

textarea {
  min-height: 180px;
  resize: vertical;
  font-family: inherit;
  line-height: 1.55;
}

input::placeholder,
textarea::placeholder {
  color: #94a3b8;
}

input:focus,
textarea:focus,
select:focus {
  outline: none;
  border-color: rgba(37, 99, 235, 0.55);
  box-shadow:
    0 0 0 5px rgba(37, 99, 235, 0.1),
    0 16px 34px rgba(37, 99, 235, 0.12);
  transform: translateY(-1px);
}

.has-error {
  border-color: rgba(220, 38, 38, 0.7) !important;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 1), rgba(254, 242, 242, 0.9));
}

.privacy-select {
  cursor: pointer;
}

.privacy-note {
  display: none;
}

.upload-box {
  display: grid;
  place-items: center;
  gap: 0.35rem;
  min-height: 150px;
  border: 1px dashed rgba(96, 165, 250, 0.5);
  border-radius: 22px;
  background:
    radial-gradient(circle at top, rgba(219, 234, 254, 0.95), rgba(239, 246, 255, 0.75));
  text-align: center;
  cursor: pointer;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 12px 24px rgba(37, 99, 235, 0.06);
  transition:
    transform var(--dur-fast) var(--ease-standard),
    box-shadow var(--dur-fast) var(--ease-standard),
    border-color var(--dur-fast) var(--ease-standard);
}

.upload-box:hover {
  transform: translateY(-1px);
  border-color: var(--brand-500);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 18px 30px rgba(37, 99, 235, 0.12);
}

.upload-icon {
  display: inline-grid;
  place-items: center;
  width: 40px;
  height: 40px;
  border-radius: 999px;
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  color: var(--white);
  box-shadow: 0 12px 24px rgba(37, 99, 235, 0.22);
}

.upload-box p {
  color: var(--gray-900);
  font-weight: 800;
}

.upload-box small {
  color: var(--gray-500);
  font-weight: 600;
}

.image-preview-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 1rem;
  border: 1px solid rgba(148, 163, 184, 0.18);
  border-radius: 20px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(239, 246, 255, 0.72));
}

.file-info {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  min-width: 0;
}

.file-icon {
  color: var(--brand-600);
}

.file-name {
  color: var(--gray-700);
  font-size: 0.92rem;
  font-weight: 700;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.remove-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  border: 1px solid rgba(239, 68, 68, 0.24);
  border-radius: 12px;
  background: rgba(254, 242, 242, 0.88);
  color: var(--danger);
  cursor: pointer;
}

.followers-panel {
  margin-top: 0.3rem;
  padding: 1rem;
  border-radius: 24px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(248, 250, 252, 0.88));
}

.followers-head {
  margin-bottom: 0.8rem;
}

.followers-head p {
  margin-top: 0.35rem;
  color: var(--gray-500);
  font-size: 0.9rem;
}

.followers-container {
  display: grid;
  gap: 0.5rem;
  max-height: 220px;
  overflow-y: auto;
  padding-right: 0.25rem;
}

.follower-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.8rem 0.85rem;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.88);
  border: 1px solid rgba(148, 163, 184, 0.12);
}

.follower-info-wrapper {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  min-width: 0;
}

.follower-avatar {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  object-fit: cover;
}

.follower-name {
  color: var(--gray-800);
  font-size: 0.92rem;
  font-weight: 700;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.follower-checkbox {
  width: 18px;
  height: 18px;
  accent-color: var(--brand-500);
  flex-shrink: 0;
}

.loading-text {
  color: var(--gray-500);
  font-size: 0.9rem;
  font-weight: 600;
  padding: 0.6rem 0.2rem;
}

.composer-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 1rem;
}

.ghost-btn,
.submit-btn {
  min-height: 48px;
  border-radius: 18px;
  padding: 0.85rem 1.2rem;
  font-size: 0.94rem;
  font-weight: 800;
  cursor: pointer;
}

.ghost-btn {
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: rgba(255, 255, 255, 0.82);
  color: var(--gray-700);
}

.submit-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 152px;
  border: none;
  background: linear-gradient(135deg, var(--brand-500), #1e40af);
  color: var(--white);
  box-shadow: 0 16px 30px rgba(37, 99, 235, 0.24);
}

.submit-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
  box-shadow: none;
}

.char-counter {
  font-size: 0.8rem;
  font-weight: 700;
  color: #94a3b8;
}

.char-counter.char-warn {
  color: #d97706;
}

.char-counter.char-limit,
.error-text {
  color: #dc2626;
}

.error-text {
  font-size: 0.84rem;
  font-weight: 700;
}

.hidden-input {
  display: none;
}

.small-icon {
  width: 1rem;
  height: 1rem;
}

@media (max-width: 720px) {
  .composer-modal {
    width: min(calc(100vw - 0.5rem), 1080px);
    padding: 1rem;
    border-radius: 24px;
  }

  .composer-head {
    align-items: flex-start;
  }

  .composer-actions {
    flex-direction: column-reverse;
  }

  .ghost-btn,
  .submit-btn {
    width: 100%;
  }
}
</style>
