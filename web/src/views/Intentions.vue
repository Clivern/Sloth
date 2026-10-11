<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="mb-8">
        <h1 class="text-3xl font-semibold text-theme-text">{{ $t('intentions_page.title') }}</h1>
        <p class="text-theme-textLight mt-2">{{ $t('intentions_page.subtitle') }}</p>
      </div>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
        <div class="flex flex-wrap items-start justify-between gap-4 border-b border-theme-border px-6 py-4">
          <div>
            <h2 class="text-lg font-semibold text-theme-text">{{ $t('intentions_page.list_title') }}</h2>
            <p class="mt-1 text-sm text-theme-textLight">{{ $t('intentions_page.list_desc') }}</p>
          </div>
          <button
            v-if="canManage"
            type="button"
            class="btn-primary"
            @click="openForm()"
          >
            {{ $t('intentions_page.add') }}
          </button>
        </div>

        <div v-if="errorMessage" class="m-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
          {{ errorMessage }}
        </div>

        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-theme-border">
            <thead class="bg-theme-hover">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('intentions_page.source') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('intentions_page.destination') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('intentions_page.skill') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('intentions_page.action') }}</th>
                <th v-if="canManage" class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('intentions_page.actions') }}</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-theme-border">
              <tr v-if="loading">
                <td :colspan="columns" class="px-6 py-4 text-sm text-theme-textLight">{{ $t('common.loading') }}</td>
              </tr>
              <tr v-else-if="!rules.length">
                <td :colspan="columns" class="px-6 py-4 text-sm text-theme-textLight">{{ $t('intentions_page.empty') }}</td>
              </tr>
              <template v-else>
                <tr v-for="rule in rules" :key="rule.id" class="hover:bg-theme-hover">
                  <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-theme-text font-mono">{{ rule.source }}</td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-theme-text font-mono">{{ rule.destination }}</td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text font-mono">{{ rule.skill || $t('intentions_page.all_skills') }}</td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm">
                    <span
                      class="inline-flex rounded-full px-2 py-0.5 text-xs font-medium"
                      :class="rule.action === 'allow' ? 'bg-emerald-50 text-emerald-800' : 'bg-red-50 text-red-800'"
                    >
                      {{ $t(`intentions_page.action_${rule.action}`) }}
                    </span>
                  </td>
                  <td v-if="canManage" class="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <button
                      type="button"
                      class="font-medium text-primary-700 hover:underline"
                      @click="openForm(rule)"
                    >
                      {{ $t('intentions_page.edit') }}
                    </button>
                    <button
                      type="button"
                      class="ml-4 font-medium text-red-600 hover:underline"
                      @click="askDelete(rule)"
                    >
                      {{ $t('intentions_page.delete') }}
                    </button>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>

        <div
          v-if="total > 0"
          class="bg-white px-6 py-4 border-t border-theme-border flex items-center justify-between"
        >
          <div class="text-sm text-theme-textLight">
            {{ $t('intentions_page.showing', { from: offset + 1, to: Math.min(offset + limit, total), total }) }}
          </div>
          <div class="flex items-center gap-2">
            <button
              type="button"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="offset === 0"
              @click="goToPage(offset - limit)"
            >
              {{ $t('intentions_page.previous') }}
            </button>
            <button
              type="button"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="offset + limit >= total"
              @click="goToPage(offset + limit)"
            >
              {{ $t('intentions_page.next') }}
            </button>
          </div>
        </div>
      </section>
    </main>

    <div
      v-if="showForm"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
      role="dialog"
      aria-modal="true"
      aria-labelledby="intention-form-title"
      @click.self="closeForm"
    >
      <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-lg p-6">
        <h2 id="intention-form-title" class="text-lg font-semibold text-theme-text">
          {{ isEditing ? $t('intentions_page.edit_title') : $t('intentions_page.add_title') }}
        </h2>
        <p class="mt-1 text-sm text-theme-textLight">
          {{ isEditing ? $t('intentions_page.edit_desc') : $t('intentions_page.add_desc') }}
        </p>

        <div v-if="formError" class="mt-4 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800" role="alert">
          {{ formError }}
        </div>

        <form class="mt-4 space-y-4" @submit.prevent="submit">
          <div class="grid gap-4 sm:grid-cols-2">
            <div>
              <label for="intention-source" class="form-label">{{ $t('intentions_page.source') }}</label>
              <select
                id="intention-source"
                v-model="form.source"
                required
                class="input-field font-mono"
                :disabled="saving"
              >
                <option value="" disabled>{{ $t('intentions_page.select_agent') }}</option>
                <option v-for="agent in agents" :key="agent.name" :value="agent.name">{{ agent.name }}</option>
              </select>
            </div>
            <div>
              <label for="intention-destination" class="form-label">{{ $t('intentions_page.destination') }}</label>
              <select
                id="intention-destination"
                v-model="form.destination"
                required
                class="input-field font-mono"
                :disabled="saving"
              >
                <option value="" disabled>{{ $t('intentions_page.select_agent') }}</option>
                <option v-for="agent in agents" :key="agent.name" :value="agent.name">{{ agent.name }}</option>
              </select>
            </div>
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <div>
              <label for="intention-skill" class="form-label">{{ $t('intentions_page.skill') }}</label>
              <input
                id="intention-skill"
                v-model="form.skill"
                type="text"
                class="input-field font-mono"
                placeholder="user.profile.get"
                autocomplete="off"
                :disabled="saving"
              />
              <p class="mt-1 text-xs text-theme-textLight">{{ $t('intentions_page.skill_hint') }}</p>
            </div>
            <div>
              <label for="intention-action" class="form-label">{{ $t('intentions_page.action') }}</label>
              <select
                id="intention-action"
                v-model="form.action"
                class="input-field"
                :disabled="saving"
              >
                <option value="allow">{{ $t('intentions_page.action_allow') }}</option>
                <option value="deny">{{ $t('intentions_page.action_deny') }}</option>
              </select>
            </div>
          </div>

          <div class="flex gap-3 justify-end">
            <button type="button" class="btn-secondary" :disabled="saving" @click="closeForm">
              {{ $t('common.cancel') }}
            </button>
            <button type="submit" class="btn-primary" :disabled="saving">
              {{ saving ? $t('intentions_page.saving') : $t('common.save') }}
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
          aria-labelledby="intention-delete-title"
          @click.self="closeDelete"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-sm p-6">
            <h2 id="intention-delete-title" class="text-lg font-semibold text-theme-text">
              {{ $t('intentions_page.delete_title') }}
            </h2>
            <div v-if="deleteError" class="mt-3 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800" role="alert">
              {{ deleteError }}
            </div>
            <p class="mt-2 text-sm text-theme-textLight">
              {{ $t('intentions_page.delete_confirm', { source: deleteTarget?.source, destination: deleteTarget?.destination }) }}
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
                {{ deleting ? $t('intentions_page.deleting') : $t('intentions_page.delete') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppNav from '@/components/AppNav.vue'
import { agent_api, intention_api } from '@/api'
import { useUrlPagination } from '@/lib/pagination'
import { useWorkspaceContext } from '@/lib/permission'

const { t } = useI18n()
const route = useRoute()
const { currentWorkspace, canManage } = useWorkspaceContext()
const { offset, limit, goToOffset } = useUrlPagination({ limit: 50 })

const rules = ref([])
const agents = ref([])
const total = ref(0)
const loading = ref(false)
const errorMessage = ref(null)

const columns = computed(() => (canManage.value ? 5 : 4))

const showForm = ref(false)
const editingId = ref(null)
const isEditing = computed(() => editingId.value !== null)
const saving = ref(false)
const formError = ref(null)
const form = ref(emptyForm())

const deleteTarget = ref(null)
const deleting = ref(false)
const deleteError = ref(null)

function emptyForm() {
  return { source: '', destination: '', skill: '', action: 'allow' }
}

// Maps a rule from the API to the form fields
function toForm(rule) {
  return { source: rule.source, destination: rule.destination, skill: rule.skill || '', action: rule.action }
}

// The API caps a page at 100 agents, so keep fetching until every agent is loaded
async function loadAgents() {
  const pageSize = 100
  const loaded = []

  while (true) {
    const res = await agent_api.list(currentWorkspace.id, { limit: pageSize, offset: loaded.length })
    const page = res.data.agents || []
    loaded.push(...page)

    if (page.length < pageSize || loaded.length >= (res.data._meta?.total || 0)) break
  }

  agents.value = loaded
}

async function loadRules() {
  loading.value = true
  errorMessage.value = null

  try {
    const res = await intention_api.list(currentWorkspace.id, { limit: limit, offset: offset.value })
    rules.value = res.data.intentions || []
    total.value = res.data._meta?.total || 0
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('intentions_page.failed_load')
    rules.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function goToPage(newOffset) {
  if (newOffset < 0 || newOffset >= total.value) return
  goToOffset(newOffset)
}

function openForm(rule = null) {
  editingId.value = rule ? rule.id : null
  form.value = rule ? toForm(rule) : emptyForm()
  formError.value = null
  showForm.value = true
}

function closeForm() {
  showForm.value = false
}

async function submit() {
  saving.value = true
  formError.value = null

  const payload = {
    source: { agent: form.value.source.trim() },
    destination: { agent: form.value.destination.trim(), skill: form.value.skill.trim() },
    action: form.value.action,
  }

  try {
    if (isEditing.value) {
      await intention_api.update(currentWorkspace.id, editingId.value, payload)
    } else {
      await intention_api.create(currentWorkspace.id, payload)
    }
    showForm.value = false
    await loadRules()
  } catch (err) {
    formError.value = err.response?.data?.errorMessage || t('intentions_page.failed_save')
  } finally {
    saving.value = false
  }
}

function askDelete(rule) {
  deleteTarget.value = rule
  deleteError.value = null
}

function closeDelete() {
  deleteTarget.value = null
}

async function confirmDelete() {
  deleting.value = true
  deleteError.value = null

  try {
    await intention_api.delete(currentWorkspace.id, deleteTarget.value.id)
    deleteTarget.value = null
    await loadRules()
  } catch (err) {
    deleteError.value = err.response?.data?.errorMessage || t('intentions_page.failed_delete')
  } finally {
    deleting.value = false
  }
}

watch(() => route.query.page, loadRules)
onMounted(() => {
  loadRules()
  loadAgents()
})
</script>
