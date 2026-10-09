<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-show="open"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
        role="dialog"
        aria-modal="true"
        aria-labelledby="agent-card-modal-title"
        @click.self="close"
      >
        <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-2xl p-6 max-h-[90vh] overflow-y-auto">
          <h2 id="agent-card-modal-title" class="text-lg font-semibold text-theme-text">
            {{ isEdit ? $t('agents_page.edit_card_title') : $t('agents_page.register_title') }}
          </h2>
          <p class="mt-1 text-sm text-theme-textLight">
            {{ isEdit ? $t('agents_page.edit_card_desc') : $t('agents_page.register_desc') }}
          </p>

          <div v-if="error" class="mt-4 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800" role="alert">
            {{ error }}
          </div>

          <form class="mt-4 space-y-4" @submit.prevent="submit">
            <div class="grid gap-4 sm:grid-cols-[1fr_180px]">
              <div>
                <label for="agent-card-name" class="form-label">{{ $t('agents_page.name') }}</label>
                <input
                  id="agent-card-name"
                  v-model="form.name"
                  type="text"
                  required
                  class="input-field font-mono"
                  placeholder="support-assistant"
                  autocomplete="off"
                  :disabled="saving || isEdit"
                />
                <p class="mt-1 text-xs text-theme-textLight">{{ $t('agents_page.name_hint') }}</p>
              </div>
              <div>
                <label for="agent-card-version" class="form-label">{{ $t('agents_page.version') }}</label>
                <input
                  id="agent-card-version"
                  v-model="form.version"
                  type="text"
                  maxlength="40"
                  class="input-field font-mono"
                  placeholder="1.0.0"
                  autocomplete="off"
                  :disabled="saving"
                />
                <p class="mt-1 text-xs text-theme-textLight">{{ $t('agents_page.version_hint') }}</p>
              </div>
            </div>

            <div>
              <label for="agent-card-json" class="form-label">{{ $t('agents_page.card_json') }}</label>
              <textarea
                id="agent-card-json"
                v-model="form.card"
                rows="14"
                required
                spellcheck="false"
                class="input-field font-mono text-xs leading-5"
                :placeholder="AGENT_CARD_EXAMPLE"
                :disabled="saving"
              />
              <p class="mt-1 text-xs text-theme-textLight">{{ $t('agents_page.card_json_hint') }}</p>
            </div>

            <div class="flex gap-3 justify-end">
              <button type="button" class="btn-secondary" :disabled="saving" @click="close">
                {{ $t('common.cancel') }}
              </button>
              <button
                type="submit"
                class="btn-primary"
                :disabled="saving || !form.name.trim() || !form.card.trim()"
              >
                {{ saving ? $t('agents_page.saving') : (isEdit ? $t('common.save') : $t('agents_page.register')) }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { agent_api } from '@/api'
import { AGENT_CARD_EXAMPLE, AGENT_NAME_PATTERN } from '@/lib/agent'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  workspaceId: {
    type: String,
    required: true,
  },
  // Existing agent to edit; omit to register a new one.
  agent: {
    type: Object,
    default: null,
  },
})

const emit = defineEmits(['close', 'saved'])

const { t } = useI18n()
const isEdit = computed(() => !!props.agent)
const saving = ref(false)
const error = ref(null)
const form = reactive({
  name: '',
  version: '',
  card: '',
})

watch(() => props.open, (open) => {
  if (!open) return
  error.value = null
  form.name = props.agent?.name || ''
  form.version = ''
  form.card = props.agent ? JSON.stringify(props.agent.card ?? {}, null, 2) : ''
})

function close() {
  if (saving.value) return
  emit('close')
}

async function agentExists(name) {
  try {
    await agent_api.get(props.workspaceId, name)
    return true
  } catch (err) {
    if (err.response?.status === 404) return false
    throw err
  }
}

async function submit() {
  const name = form.name.trim()
  if (!AGENT_NAME_PATTERN.test(name)) {
    error.value = t('agents_page.invalid_name')
    return
  }

  let card
  try {
    card = JSON.parse(form.card)
  } catch {
    error.value = t('agents_page.invalid_card_json')
    return
  }
  if (!card || typeof card !== 'object' || Array.isArray(card)) {
    error.value = t('agents_page.invalid_card_json')
    return
  }

  const version = form.version.trim()
  if (!version && !card.version) {
    error.value = t('agents_page.version_required')
    return
  }

  saving.value = true
  error.value = null

  try {
    // PUT is an upsert, so guard against replacing an existing agent from the register form
    if (!isEdit.value && await agentExists(name)) {
      error.value = t('agents_page.already_exists', { name })
      return
    }

    const payload = version ? { card, version } : { card }
    const res = await agent_api.upsert(props.workspaceId, name, payload)
    emit('saved', res.data)
  } catch (err) {
    error.value = err.response?.data?.errorMessage || t('agents_page.failed_save')
  } finally {
    saving.value = false
  }
}
</script>
