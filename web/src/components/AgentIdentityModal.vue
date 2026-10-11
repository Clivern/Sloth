<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-show="open"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
        role="dialog"
        aria-modal="true"
        aria-labelledby="agent-identity-modal-title"
        @click.self="close"
      >
        <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-lg p-6 max-h-[90vh] overflow-y-auto">
          <h2 id="agent-identity-modal-title" class="text-lg font-semibold text-theme-text">
            {{ $t('agents_page.identity_add_title') }}
          </h2>
          <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.identity_modal_desc') }}</p>

          <div v-if="error" class="mt-4 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800" role="alert">
            {{ error }}
          </div>

          <form class="mt-4 space-y-4" @submit.prevent="submit">
            <div>
              <label for="identity-name" class="form-label">{{ $t('agents_page.name') }}</label>
              <input
                id="identity-name"
                v-model="form.name"
                type="text"
                required
                maxlength="100"
                class="input-field font-mono"
                :placeholder="$t('agents_page.identity_name_placeholder')"
                autocomplete="off"
                :disabled="saving"
              />
              <p class="mt-1 text-xs text-theme-textLight">{{ $t('agents_page.identity_name_hint') }}</p>
            </div>

            <div>
              <label for="identity-type" class="form-label">{{ $t('agents_page.identity_type') }}</label>
              <select id="identity-type" v-model="form.type" class="input-field" :disabled="saving">
                <option value="api_key">{{ $t('agents_page.identity_type_api_key') }}</option>
                <option value="basic_auth">{{ $t('agents_page.identity_type_basic_auth') }}</option>
                <option value="client_credentials">{{ $t('agents_page.identity_type_client_credentials') }}</option>
              </select>
              <p class="mt-1 text-xs text-theme-textLight">{{ $t(`agents_page.identity_type_${form.type}_hint`) }}</p>
            </div>

            <div v-if="form.type === 'basic_auth'">
              <label for="identity-username" class="form-label">{{ $t('agents_page.identity_username') }}</label>
              <input
                id="identity-username"
                v-model="form.username"
                type="text"
                required
                maxlength="100"
                class="input-field font-mono"
                :placeholder="agentName"
                autocomplete="off"
                :disabled="saving"
              />
            </div>

            <div v-if="form.type === 'client_credentials'">
              <label for="identity-token-ttl" class="form-label">{{ $t('agents_page.identity_token_ttl') }}</label>
              <input
                id="identity-token-ttl"
                v-model.number="form.tokenTtl"
                type="number"
                min="60"
                max="86400"
                required
                class="input-field"
                :disabled="saving"
              />
              <p class="mt-1 text-xs text-theme-textLight">{{ $t('agents_page.identity_token_ttl_hint') }}</p>
            </div>

            <div class="flex gap-3 justify-end">
              <button type="button" class="btn-secondary" :disabled="saving" @click="close">
                {{ $t('common.cancel') }}
              </button>
              <button type="submit" class="btn-primary" :disabled="saving || !form.name.trim()">
                {{ saving ? $t('agents_page.saving') : $t('common.create') }}
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
import { agent_identity_api } from '@/api'

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
})

const emit = defineEmits(['close', 'saved'])

const { t } = useI18n()
const saving = ref(false)
const error = ref(null)
const form = reactive({
  name: '',
  type: 'api_key',
  username: '',
  tokenTtl: 300,
})

watch(() => props.open, (open) => {
  if (!open) return
  error.value = null
  form.name = ''
  form.type = 'api_key'
  form.username = props.agentName
  form.tokenTtl = 300
})

function close() {
  if (saving.value) return
  emit('close')
}

async function submit() {
  saving.value = true
  error.value = null

  const body = { name: form.name.trim(), type: form.type }
  if (form.type === 'basic_auth') body.username = form.username.trim()
  if (form.type === 'client_credentials') body.tokenTtl = form.tokenTtl

  try {
    const res = await agent_identity_api.create(props.workspaceId, props.agentName, body)
    emit('saved', res.data)
  } catch (err) {
    error.value = err.response?.data?.errorMessage || t('agents_page.identity_failed_save')
  } finally {
    saving.value = false
  }
}
</script>
