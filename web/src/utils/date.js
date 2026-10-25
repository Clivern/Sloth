import { i18n } from '@/i18n'

const units = [
  ['year', 31536000],
  ['month', 2592000],
  ['week', 604800],
  ['day', 86400],
  ['hour', 3600],
  ['minute', 60],
]

// Reading the locale ref inside a template call makes the output follow language changes.
function locale() {
  return i18n.global.locale.value
}

/** Short date, e.g. "Mar 27, 2026". */
export function formatDate(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString(locale(), { dateStyle: 'medium' })
}

/** Date with hours and minutes, e.g. "Mar 27, 2026, 3:04 PM". */
export function formatDateTime(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleString(locale(), { dateStyle: 'medium', timeStyle: 'short' })
}

/** Full timestamp down to the second, for tooltips on relative times. */
export function formatTimestamp(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleString(locale(), { dateStyle: 'medium', timeStyle: 'medium' })
}

/** Relative time in the current locale, e.g. "5 minutes ago" or "il y a 5 minutes". */
export function formatRelative(iso, fallback = '—') {
  if (!iso) return fallback

  const date = new Date(iso)
  if (date.getFullYear() < 1970) return fallback

  const seconds = Math.round((date.getTime() - Date.now()) / 1000)
  const rtf = new Intl.RelativeTimeFormat(locale(), { numeric: 'auto' })

  for (const [unit, size] of units) {
    if (Math.abs(seconds) >= size) return rtf.format(Math.round(seconds / size), unit)
  }

  return rtf.format(0, 'second')
}
