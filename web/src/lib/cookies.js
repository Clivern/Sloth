import { ref } from 'vue'

export const CONSENT_STORAGE_KEY = 'sloth_cookie_consent'
export const CONSENT_VERSION = 1

export const showCookieBanner = ref(false)
export const showCookiePreferences = ref(false)

const defaultCategories = () => ({
  essential: true,
  preferences: false,
  analytics: false,
})

export function readConsent() {
  try {
    const raw = localStorage.getItem(CONSENT_STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw)
    if (!parsed || parsed.version !== CONSENT_VERSION) return null
    return parsed
  } catch (_) {
    return null
  }
}

export function hasConsent() {
  return readConsent() !== null
}

export function isCategoryAllowed(category) {
  if (category === 'essential') return true
  const consent = readConsent()
  if (!consent) return false
  return consent.categories?.[category] === true
}

export function saveConsent(categories) {
  const payload = {
    version: CONSENT_VERSION,
    updatedAt: new Date().toISOString(),
    categories: {
      essential: true,
      preferences: !!categories.preferences,
      analytics: !!categories.analytics,
    },
  }
  try {
    localStorage.setItem(CONSENT_STORAGE_KEY, JSON.stringify(payload))
  } catch (_) {}
  showCookieBanner.value = false
  showCookiePreferences.value = false
  return payload
}

export function acceptAllCookies() {
  return saveConsent({ preferences: true, analytics: true })
}

export function rejectOptionalCookies() {
  return saveConsent({ preferences: false, analytics: false })
}

export function openCookiePreferences() {
  showCookieBanner.value = false
  showCookiePreferences.value = true
}

export function initCookieConsent() {
  if (!readConsent()) {
    showCookieBanner.value = true
  }
}

export function consentCategoriesForEdit() {
  const consent = readConsent()
  if (!consent) return defaultCategories()
  return {
    essential: true,
    preferences: !!consent.categories?.preferences,
    analytics: !!consent.categories?.analytics,
  }
}
