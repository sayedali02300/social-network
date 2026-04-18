import { API_BASE_URL } from '@/api/api';
export { API_BASE_URL };

export const getImgURL = (path: string) => {
  if (!path) return '';
  const CleanPath = path.replace("./", "/");
  return API_BASE_URL+CleanPath;
}

export const formatPostDate = (dateStr: string) => {
  if (!dateStr) return '';

  const postDate = new Date(dateStr);
  const now = new Date();

  const diffMs = now.getTime() - postDate.getTime();
  const diffMinutes = Math.floor(diffMs / (1000 * 60));
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60));

  if (diffMinutes < 1) return 'Just now';
  if (diffHours < 1) return `${diffMinutes}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;

  const options: Intl.DateTimeFormatOptions = {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  };

  if (postDate.getFullYear() !== now.getFullYear()) {
    options.year = 'numeric';
  }

  return postDate.toLocaleString(undefined, options);
};

export function debounce<Args extends unknown[]>(
  fn: (...args: Args) => void, 
  delay: number) {
  let timeoutId: ReturnType<typeof setTimeout> | null = null;
  
  return function (...args: Args) {
    if (timeoutId !== null) {
      clearTimeout(timeoutId);
    }
    timeoutId = setTimeout(() => fn(...args), delay);
  };
}

export const trackMouse = (e: MouseEvent) => {
  const el = e.currentTarget as HTMLElement
  const rect = el.getBoundingClientRect()
  el.style.setProperty('--mx', `${e.clientX - rect.left}px`)
  el.style.setProperty('--my', `${e.clientY - rect.top}px`)
}

export const clearMouse = (e: MouseEvent) => {
  const el = e.currentTarget as HTMLElement
  el.style.setProperty('--mx', '-999px')
  el.style.setProperty('--my', '-999px')
}

const DOB_MAX_AGE_YEARS = 120
const NAME_MAX_LENGTH = 25
const LETTERS_AND_SPACES_REGEX = /^[A-Za-z]+(?:\s+[A-Za-z]+)*$/

export const formatDateForInput = (date: Date) => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export const getDateOfBirthBounds = () => {
  const today = new Date()
  const max = formatDateForInput(today)

  const minDate = new Date()
  minDate.setFullYear(today.getFullYear() - DOB_MAX_AGE_YEARS)
  const min = formatDateForInput(minDate)

  return { min, max }
}

export const validateRealisticDateOfBirth = (value: string): string | null => {
  if (!value) return 'Date of birth is required.'

  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value)
  if (!match) return 'Date of birth must use YYYY-MM-DD format.'

  const year = Number(match[1])
  const month = Number(match[2])
  const day = Number(match[3])

  const candidate = new Date(year, month - 1, day)
  const isSameDate = (
    candidate.getFullYear() === year &&
    candidate.getMonth() === month - 1 &&
    candidate.getDate() === day)

  if (!isSameDate) return 'Date of birth is invalid.'

  const { min, max } = getDateOfBirthBounds()
  const minDate = new Date(`${min}T00:00:00`)
  const maxDate = new Date(`${max}T23:59:59`)

  const today = new Date()

  if (candidate > maxDate) return 'Date of birth cannot be in the future.'

  const age = today.getFullYear() - candidate.getFullYear();
  const m = today.getMonth() - candidate.getMonth();
  const isUnderage = m < 0 || (m === 0 && today.getDate() < candidate.getDate());
  
  if (age < 13 || (age === 13 && isUnderage)) {
    return 'Age must be at least 13 years old.'
  }

  if (candidate < minDate) return `Date of birth must be on or after ${min}.`

  return null
}

export const validateLettersOnlyName = (label: 'First name' | 'Last name', value: string): string | null => {
  const trimmed = value.trim()
  if (!trimmed) return `${label} is required.`
  if (trimmed.length > NAME_MAX_LENGTH) return `${label} must be 25 characters or fewer.`
  if (!LETTERS_AND_SPACES_REGEX.test(trimmed)) return `${label} must contain letters and spaces only.`
  return null
}

export const validateNickname = (value: string): string | null => {
  if (value.trim().length > NAME_MAX_LENGTH) return 'Nickname must be 25 characters or fewer.'
  return null
}
