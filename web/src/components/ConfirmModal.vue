<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-show="open"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        @click.self="cancel"
      >
        <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-sm p-6">
          <h2 :id="titleId" class="text-lg font-semibold text-theme-text">{{ title }}</h2>
          <div v-if="error" class="mt-3 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800" role="alert">
            {{ error }}
          </div>
          <p class="mt-2 text-sm text-theme-textLight">{{ message }}</p>
          <slot />
          <div class="mt-6 flex gap-3 justify-end">
            <button type="button" class="btn-secondary" :disabled="busy" @click="cancel">
              {{ $t('common.cancel') }}
            </button>
            <button
              type="button"
              class="btn-primary bg-red-600 hover:bg-red-700 focus:ring-red-500"
              :disabled="busy || disabled"
              @click="emit('confirm')"
            >
              {{ busy ? busyLabel : confirmLabel }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { useId, watchEffect } from 'vue'

const props = defineProps({
  open: { type: Boolean, default: false },
  title: { type: String, required: true },
  message: { type: String, default: '' },
  error: { type: String, default: null },
  busy: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
  confirmLabel: { type: String, required: true },
  busyLabel: { type: String, required: true },
})

const emit = defineEmits(['confirm', 'cancel'])
const titleId = useId()

function cancel() {
  if (!props.busy) emit('cancel')
}

// Lock page scroll and close on Escape while the modal is open
watchEffect((onCleanup) => {
  if (!props.open) return
  const onEsc = (e) => { if (e.key === 'Escape') cancel() }
  document.body.style.overflow = 'hidden'
  document.addEventListener('keydown', onEsc)
  onCleanup(() => {
    document.body.style.overflow = ''
    document.removeEventListener('keydown', onEsc)
  })
})
</script>
