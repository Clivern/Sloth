<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-show="open"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
        role="dialog"
        aria-modal="true"
        aria-labelledby="agent-protection-modal-title"
        @click.self="close"
      >
        <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-lg p-6 max-h-[90vh] overflow-y-auto">
          <h2 id="agent-protection-modal-title" class="text-lg font-semibold text-theme-text">
            {{ current ? $t('agents_page.protection_replace_title') : $t('agents_page.protection_add_title') }}
          </h2>
          <p class="mt-1 text-sm text-theme-textLight">
            {{ current ? $t('agents_page.protection_replace_desc') : $t('agents_page.protection_modal_desc') }}
          </p>

          <div v-if="error" class="mt-4 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800" role="alert">
            {{ error }}
          </div>

          <form class="mt-4 space-y-4" @submit.prevent="submit">
            <div>
              <label for="protection-type" class="form-label">{{ $t('agents_page.protection_type') }}</label>
              <select id="protection-type" v-model="form.type" class="input-field" :disabled="saving">
                <option value="api_key">{{ $t('agents_page.protection_type_api_key') }}</option>
                <option value="basic_auth">{{ $t('agents_page.protection_type_basic_auth') }}</option>
              </select>
              <p class="mt-1 text-xs text-theme-textLight">{{ $t(`agents_page.protection_type_${form.type}_hint`) }}</p>
            </div>

            <div v-if="form.type === 'api_key'">
              <label for="protection-header" class="form-label">{{ $t('agents_page.protection_header') }}</label>
              <input
                id="protection-header"
                v-model="form.header"
                type="text"
                maxlength="100"
                class="input-field font-mono"
                placeholder="X-API-Key"
                autocomplete="off"
                :disabled="saving"
              />
            </div>

            <div v-if="form.type === 'basic_auth'">
              <label for="protection-username" class="form-label">{{ $t('agents_page.protection_username') }}</label>
              <input
                id="protection-username"
                v-model="form.username"
                type="text"
                required
                maxlength="100"
                class="input-field font-mono"
                placeholder="cognit"
                autocomplete="off"
                :disabled="saving"
              />
            </div>

            <div>
              <label for="protection-secret" class="form-label">{{ $t('agents_page.protection_secret') }}</label>
              <input
                id="protection-secret"
                v-model="form.secret"
                type="password"
                minlength="8"
                maxlength="512"
                class="input-field font-mono"
                :placeholder="$t('agents_page.protection_secret_placeholder')"
                autocomplete="new-password"
                :disabled="saving"
              />
              <p class="mt-1 text-xs text-theme-textLight">{{ $t('agents_page.protection_secret_hint') }}</p>
            </div>

            <div class="flex gap-3 justify-end">
              <button type="button" class="btn-secondary" :disabled="saving" @click="close">
                {{ $t('common.cancel') }}
              </button>
              <button type="submit" class="btn-primary" :disabled="saving">
                {{ saving ? $t('agents_page.saving') : $t('common.save') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { agent_protection_api } from '@/api'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  workspaceId: {
    type: String,
    required: true,
  },
  agentName: {
    type: String,
    required: true,
  },
  current: {
    type: Object,
    default: null,
  },
})

const emit = defineEmits(['close', 'saved'])

const { t } = useI18n()
const saving = ref(false)
const error = ref(null)
const form = reactive({
  type: 'api_key',
  header: '',
  username: '',
  secret: '',
})

watch(() => props.open, (open) => {
  if (!open) return
  error.value = null
  form.type = props.current?.type || 'api_key'
  form.header = props.current?.header || 'X-API-Key'
  form.username = props.current?.username || 'cognit'
  form.secret = ''
})

function close() {
  if (saving.value) return
  emit('close')
}

async function submit() {
  saving.value = true
  error.value = null

  const body = { type: form.type }
  if (form.type === 'api_key') body.header = form.header.trim()
  if (form.type === 'basic_auth') body.username = form.username.trim()
  if (form.secret) body.secret = form.secret

  try {
    const res = await agent_protection_api.save(props.workspaceId, props.agentName, body)
    emit('saved', res.data)
  } catch (err) {
    error.value = err.response?.data?.errorMessage || t('agents_page.protection_failed_save')
  } finally {
    saving.value = false
  }
}
</script>
