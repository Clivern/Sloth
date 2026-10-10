/** Short date for tables (e.g. "Mar 27, 2026"). Pass `locale` from `useI18n()` for correct formatting. */
export function formatCreatedAt(iso, locale = 'en') {
  if (!iso) return '—'
  try {
    const d = new Date(iso)
    if (Number.isNaN(d.getTime())) return '—'
    return d.toLocaleDateString(locale || 'en', { month: 'short', day: 'numeric', year: 'numeric' })
  } catch {
    return '—'
  }
}

export function formatRelative(dateString, fallback = '—') {
  if (dateString == null || dateString === '') return fallback
  try {
    const date = new Date(dateString)
    if (isNaN(date.getTime()) || date.getFullYear() < 1970) return fallback

    const now = new Date()
    const diffMs = now - date
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)
    const diffDays = Math.floor(diffMs / 86400000)

    if (diffMins < 1) return 'Just now'
    if (diffMins < 60) return `${diffMins} ${diffMins === 1 ? 'minute' : 'minutes'} ago`
    if (diffHours < 24) return `${diffHours} ${diffHours === 1 ? 'hour' : 'hours'} ago`
    if (diffDays < 7) return `${diffDays} ${diffDays === 1 ? 'day' : 'days'} ago`

    const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
    const month = months[date.getMonth()]
    const day = date.getDate()
    const year = date.getFullYear()
    const isCurrentYear = year === now.getFullYear()

    return isCurrentYear ? `${month} ${day}` : `${month} ${day}, ${year}`
  } catch {
    return fallback
  }
}
