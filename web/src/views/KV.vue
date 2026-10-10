<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="mb-8">
        <h1 class="text-3xl font-semibold text-theme-text">{{ $t('kv_page.title') }}</h1>
        <p class="text-theme-textLight mt-2">{{ $t('kv_page.subtitle') }}</p>
      </div>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
        <div class="flex flex-wrap items-start justify-between gap-4 border-b border-theme-border px-6 py-4">
          <div>
            <h2 class="text-lg font-semibold text-theme-text">{{ $t('kv_page.list_title') }}</h2>
            <p class="mt-1 text-sm text-theme-textLight">{{ $t('kv_page.list_desc') }}</p>
          </div>
          <button
            v-if="canManage"
            type="button"
            class="btn-primary"
            @click="openForm()"
          >
            {{ $t('kv_page.add') }}
          </button>
        </div>

        <div v-if="errorMessage" class="m-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
          {{ errorMessage }}
        </div>

        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-theme-border">
            <thead class="bg-theme-hover">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('kv_page.key') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('kv_page.value') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('kv_page.updated') }}</th>
                <th v-if="canManage" class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('kv_page.actions') }}</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-theme-border">
              <tr v-if="loading">
                <td :colspan="columns" class="px-6 py-4 text-sm text-theme-textLight">{{ $t('common.loading') }}</td>
              </tr>
              <tr v-else-if="!items.length">
                <td :colspan="columns" class="px-6 py-4 text-sm text-theme-textLight">{{ $t('kv_page.empty') }}</td>
              </tr>
              <template v-else>
                <tr v-for="item in items" :key="item.id" class="hover:bg-theme-hover">
                  <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-theme-text font-mono">{{ item.key }}</td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text font-mono">{{ item.value }}</td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">{{ formatRelative(item.updatedAt) }}</td>
                  <td v-if="canManage" class="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <button
                      type="button"
                      class="font-medium text-primary-700 hover:underline"
                      @click="openForm(item)"
                    >
                      {{ $t('kv_page.edit') }}
                    </button>
                    <button
                      type="button"
                      class="ml-4 font-medium text-red-600 hover:underline"
                      @click="askDelete(item)"
                    >
                      {{ $t('kv_page.delete') }}
                    </button>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>
      </section>
    </main>

    <div
      v-if="showForm"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
      role="dialog"
      aria-modal="true"
      aria-labelledby="kv-form-title"
      @click.self="closeForm"
    >
      <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-lg p-6">
        <h2 id="kv-form-title" class="text-lg font-semibold text-theme-text">
          {{ isEditing ? $t('kv_page.edit_title') : $t('kv_page.add_title') }}
        </h2>
        <p class="mt-1 text-sm text-theme-textLight">
          {{ isEditing ? $t('kv_page.edit_desc') : $t('kv_page.add_desc') }}
        </p>

        <div v-if="formError" class="mt-4 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800" role="alert">
          {{ formError }}
        </div>

        <form class="mt-4 space-y-4" @submit.prevent="submit">
          <div>
            <label for="kv-key" class="form-label">{{ $t('kv_page.key') }}</label>
            <input
              id="kv-key"
              v-model="form.key"
              type="text"
              required
              class="input-field font-mono"
              placeholder="agents/invoice-extractor/model"
              autocomplete="off"
              :disabled="saving || isEditing"
            />
            <p class="mt-1 text-xs text-theme-textLight">{{ $t('kv_page.key_hint') }}</p>
          </div>

          <div>
            <label for="kv-value" class="form-label">{{ $t('kv_page.value') }}</label>
            <textarea
              id="kv-value"
              v-model="form.value"
              rows="3"
              required
              class="input-field font-mono text-sm"
              :disabled="saving"
            />
          </div>

          <div>
            <label for="kv-expires" class="form-label">{{ $t('kv_page.expires_at') }}</label>
            <input
              id="kv-expires"
              v-model="form.expiresAt"
              type="datetime-local"
              class="input-field"
              :disabled="saving"
            />
          </div>

          <div class="flex gap-3 justify-end">
            <button type="button" class="btn-secondary" :disabled="saving" @click="closeForm">
              {{ $t('common.cancel') }}
            </button>
            <button type="submit" class="btn-primary" :disabled="saving || !form.key.trim() || !form.value.trim()">
              {{ saving ? $t('kv_page.saving') : $t('common.save') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="deleteTarget"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="kv-delete-title"
          @click.self="closeDelete"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-sm p-6">
            <h2 id="kv-delete-title" class="text-lg font-semibold text-theme-text">
              {{ $t('kv_page.delete_title') }}
            </h2>
            <div v-if="deleteError" class="mt-3 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800" role="alert">
              {{ deleteError }}
            </div>
            <p class="mt-2 text-sm text-theme-textLight">
              {{ $t('kv_page.delete_confirm', { key: deleteTarget?.key }) }}
            </p>
            <div class="mt-6 flex gap-3 justify-end">
              <button type="button" class="btn-secondary" :disabled="deleting" @click="closeDelete">
                {{ $t('common.cancel') }}
              </button>
              <button
                type="button"
                class="btn-primary bg-red-600 hover:bg-red-700 focus:ring-red-500"
                :disabled="deleting"
                @click="confirmDelete"
              >
                {{ deleting ? $t('kv_page.deleting') : $t('kv_page.delete') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppNav from '@/components/AppNav.vue'
import { kv_api } from '@/api'
import { useWorkspaceContext } from '@/lib/permission'
import { formatRelative } from '@/utils/date'

const { t } = useI18n()
const { currentWorkspace, canManage } = useWorkspaceContext()

const items = ref([])
const loading = ref(false)
const errorMessage = ref(null)
const deleteTarget = ref(null)
const deleting = ref(false)
const deleteError = ref(null)

const columns = computed(() => (canManage.value ? 4 : 3))

const showForm = ref(false)
const editingKey = ref(null)
const isEditing = computed(() => editingKey.value !== null)
const saving = ref(false)
const formError = ref(null)
const form = ref(emptyForm())

function emptyForm() {
  return { key: '', value: '', expiresAt: '' }
}

// Converts an ISO timestamp to the local value a datetime-local input expects.
function toLocalInput(iso) {
  if (!iso) return ''

  const date = new Date(iso)
  return new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16)
}

async function loadKeys() {
  loading.value = true
  errorMessage.value = null

  try {
    const res = await kv_api.list(currentWorkspace.id)
    items.value = res.data.items || []
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('kv_page.failed_load')
    items.value = []
  } finally {
    loading.value = false
  }
}

function openForm(item = null) {
  editingKey.value = item ? item.key : null
  form.value = item
    ? { key: item.key, value: item.value, expiresAt: toLocalInput(item.expiresAt) }
    : emptyForm()
  formError.value = null
  showForm.value = true
}

function closeForm() {
  showForm.value = false
}

async function submit() {
  saving.value = true
  formError.value = null

  const key = isEditing.value ? editingKey.value : form.value.key.trim()
  const payload = {
    value: form.value.value,
    expiresAt: form.value.expiresAt ? new Date(form.value.expiresAt).toISOString() : '',
  }

  try {
    await kv_api.put(currentWorkspace.id, key, payload)
    showForm.value = false
    await loadKeys()
  } catch (err) {
    formError.value = err.response?.data?.errorMessage || t('kv_page.failed_save')
  } finally {
    saving.value = false
  }
}

function askDelete(item) {
  deleteTarget.value = item
  deleteError.value = null
}

function closeDelete() {
  deleteTarget.value = null
}

async function confirmDelete() {
  deleting.value = true
  deleteError.value = null

  try {
    await kv_api.delete(currentWorkspace.id, deleteTarget.value.key)
    deleteTarget.value = null
    await loadKeys()
  } catch (err) {
    deleteError.value = err.response?.data?.errorMessage || t('kv_page.failed_delete')
  } finally {
    deleting.value = false
  }
}

onMounted(loadKeys)
</script>
