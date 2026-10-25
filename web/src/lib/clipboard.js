import { ref } from 'vue'

// Copies text and remembers which field was copied for two seconds, so a button can read "Copied".
export function useCopy() {
  const copiedField = ref(null)

  async function copy(text, field) {
    await navigator.clipboard.writeText(text)
    copiedField.value = field
    setTimeout(() => {
      if (copiedField.value === field) copiedField.value = null
    }, 2000)
  }

  return { copiedField, copy }
}
