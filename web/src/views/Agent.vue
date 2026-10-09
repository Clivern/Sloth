<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <router-link to="/agents" class="text-sm font-medium text-primary-800 hover:underline">
        {{ $t('agents_page.back_to_agents') }}
      </router-link>

      <div v-if="loading" class="mt-8 rounded-xl border border-theme-border bg-white p-12 text-center shadow-sm">
        <svg class="animate-spin h-8 w-8 mx-auto text-theme-text" fill="none" viewBox="0 0 24 24" aria-hidden="true">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
        </svg>
        <p class="text-theme-textLight mt-3">{{ $t('agents_page.loading_agent') }}</p>
      </div>

      <div v-else-if="notFound" class="mt-8 rounded-xl border border-theme-border bg-white p-8 shadow-sm">
        <h1 class="text-2xl font-semibold text-theme-text">{{ $t('agents_page.not_found') }}</h1>
      </div>

      <div v-else-if="!agent" class="mt-8 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
        {{ errorMessage }}
      </div>

      <template v-else>
        <div class="mt-4 mb-8 flex flex-wrap items-start justify-between gap-4">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <h1 class="text-3xl font-semibold text-theme-text font-mono break-all">{{ agent.name }}</h1>
              <span
                class="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium"
                :class="healthClass(agent.health)"
              >
                <span class="h-1.5 w-1.5 rounded-full" :class="healthDotClass(agent.health)" aria-hidden="true" />
                {{ $t(`agents_page.health_${agent.health}`) }}
              </span>
            </div>
            <p v-if="card.description" class="text-theme-textLight mt-2">{{ card.description }}</p>
          </div>
          <div v-if="canManage" class="flex items-center gap-2">
            <button type="button" class="btn-secondary" @click="showEditModal = true">
              {{ $t('agents_page.edit_card') }}
            </button>
            <button
              type="button"
              class="btn-primary bg-red-600 hover:bg-red-700 focus:ring-red-500"
              @click="openDeleteModal"
            >
              {{ $t('agents_page.delete') }}
            </button>
          </div>
        </div>

        <div v-if="errorMessage" class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
          {{ errorMessage }}
        </div>

        <div class="space-y-6">
          <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
            <div class="border-b border-theme-border px-6 py-4">
              <h2 class="text-lg font-semibold text-theme-text">{{ $t('agents_page.card') }}</h2>
              <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.card_desc') }}</p>
            </div>
            <dl class="divide-y divide-theme-border">
              <div class="grid gap-2 px-6 py-4 sm:grid-cols-[180px_1fr]">
                <dt class="text-sm text-theme-textLight">{{ $t('agents_page.version') }}</dt>
                <dd class="text-sm font-mono text-theme-text">{{ agent.version }}</dd>
              </div>
              <div v-if="card.url" class="grid gap-2 px-6 py-4 sm:grid-cols-[180px_1fr]">
                <dt class="text-sm text-theme-textLight">{{ $t('agents_page.url') }}</dt>
                <dd class="text-sm font-mono text-theme-text break-all">{{ card.url }}</dd>
              </div>
              <div v-if="card.provider" class="grid gap-2 px-6 py-4 sm:grid-cols-[180px_1fr]">
                <dt class="text-sm text-theme-textLight">{{ $t('agents_page.provider') }}</dt>
                <dd class="text-sm text-theme-text">{{ card.provider }}</dd>
              </div>
              <div class="grid gap-2 px-6 py-4 sm:grid-cols-[180px_1fr]">
                <dt class="text-sm text-theme-textLight">{{ $t('agents_page.skills') }}</dt>
                <dd>
                  <ul v-if="card.skills.length" class="space-y-3">
                    <li v-for="skill in card.skills" :key="skill.id">
                      <div class="flex flex-wrap items-center gap-2">
                        <span class="rounded-full bg-theme-hover px-2 py-0.5 text-xs font-medium text-theme-text font-mono">{{ skill.id }}</span>
                        <span v-if="skill.name" class="text-sm text-theme-text">{{ skill.name }}</span>
                        <span
                          v-for="tag in skill.tags"
                          :key="tag"
                          class="rounded-full border border-theme-border px-2 py-0.5 text-xs text-theme-textLight"
                        >
                          {{ tag }}
                        </span>
                      </div>
                      <p v-if="skill.description" class="mt-1 text-sm text-theme-textLight">{{ skill.description }}</p>
                    </li>
                  </ul>
                  <span v-else class="text-sm text-theme-textLight">—</span>
                </dd>
              </div>
              <div class="grid gap-2 px-6 py-4 sm:grid-cols-[180px_1fr]">
                <dt class="text-sm text-theme-textLight">{{ $t('agents_page.input_modes') }}</dt>
                <dd class="text-sm text-theme-text">{{ card.inputModes.length ? card.inputModes.join(', ') : '—' }}</dd>
              </div>
              <div class="grid gap-2 px-6 py-4 sm:grid-cols-[180px_1fr]">
                <dt class="text-sm text-theme-textLight">{{ $t('agents_page.output_modes') }}</dt>
                <dd class="text-sm text-theme-text">{{ card.outputModes.length ? card.outputModes.join(', ') : '—' }}</dd>
              </div>
              <div class="grid gap-2 px-6 py-4 sm:grid-cols-[180px_1fr]">
                <dt class="text-sm text-theme-textLight">{{ $t('agents_page.capabilities') }}</dt>
                <dd class="text-sm text-theme-text">{{ card.capabilities.length ? card.capabilities.join(', ') : '—' }}</dd>
              </div>
              <div class="grid gap-2 px-6 py-4 sm:grid-cols-[180px_1fr]">
                <dt class="text-sm text-theme-textLight">{{ $t('agents_page.checksum') }}</dt>
                <dd class="text-xs font-mono text-theme-textLight break-all">{{ agent.cardChecksum }}</dd>
              </div>
              <div class="grid gap-2 px-6 py-4 sm:grid-cols-[180px_1fr]">
                <dt class="text-sm text-theme-textLight">{{ $t('agents_page.updated') }}</dt>
                <dd class="text-sm text-theme-text">{{ formatRelative(agent.updatedAt) }}</dd>
              </div>
            </dl>
            <details class="border-t border-theme-border">
              <summary class="cursor-pointer px-6 py-3 text-sm font-medium text-theme-text hover:bg-theme-hover">
                {{ $t('agents_page.raw_card') }}
              </summary>
              <pre class="overflow-x-auto bg-theme-hover px-6 py-4 text-xs leading-5 font-mono text-theme-text">{{ rawCard }}</pre>
            </details>
          </section>

          <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
            <div class="flex flex-wrap items-start justify-between gap-4 border-b border-theme-border px-6 py-4">
              <div>
                <h2 class="text-lg font-semibold text-theme-text">{{ $t('agents_page.checks_title') }}</h2>
                <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.checks_desc') }}</p>
              </div>
              <button v-if="canManage" type="button" class="btn-secondary text-sm" @click="openCheckModal(null)">
                {{ $t('agents_page.check_add') }}
              </button>
            </div>

            <div v-if="loadingChecks && checks.length === 0" class="p-8 text-center text-sm text-theme-textLight">
              {{ $t('agents_page.loading_checks') }}
            </div>
            <div v-else-if="checks.length === 0" class="p-8 text-center">
              <p class="text-sm text-theme-textLight">{{ $t('agents_page.no_checks') }}</p>
              <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.no_checks_hint') }}</p>
            </div>
            <div v-else class="overflow-x-auto">
              <table class="min-w-full divide-y divide-theme-border">
                <thead class="bg-theme-hover">
                  <tr>
                    <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.check') }}</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.check_type') }}</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.check_settings') }}</th>
                    <th v-if="canManage" class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.actions') }}</th>
                  </tr>
                </thead>
                <tbody class="bg-white divide-y divide-theme-border">
                  <tr v-for="check in checks" :key="check.checkId" class="hover:bg-theme-hover">
                    <td class="px-6 py-4 whitespace-nowrap">
                      <p class="text-sm font-medium text-theme-text">{{ check.name }}</p>
                      <p class="text-xs font-mono text-theme-textLight">{{ check.checkId }}</p>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text">{{ $t(`agents_page.check_type_${check.type}`) }}</td>
                    <td class="px-6 py-4 text-sm text-theme-text font-mono">{{ describeCheck(check) }}</td>
                    <td v-if="canManage" class="px-6 py-4 whitespace-nowrap text-right text-sm">
                      <span class="inline-flex items-center gap-3">
                        <button type="button" class="text-primary-600 hover:text-primary-700 hover:underline" @click="openCheckModal(check)">
                          {{ $t('common.edit') }}
                        </button>
                        <button
                          type="button"
                          class="text-red-600 hover:text-red-700 hover:underline disabled:opacity-50 disabled:cursor-not-allowed"
                          :disabled="deletingCheckId === check.checkId"
                          @click="deleteCheck(check)"
                        >
                          {{ deletingCheckId === check.checkId ? $t('agents_page.deleting') : $t('agents_page.delete') }}
                        </button>
                      </span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>

          <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
            <div class="flex flex-wrap items-start justify-between gap-4 border-b border-theme-border px-6 py-4">
              <div>
                <h2 class="text-lg font-semibold text-theme-text">{{ $t('agents_page.instances') }}</h2>
                <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.instances_desc') }}</p>
              </div>
              <button
                type="button"
                class="btn-secondary text-sm disabled:opacity-50"
                :disabled="loadingInstances"
                @click="loadInstances"
              >
                {{ $t('agents_page.refresh') }}
              </button>
            </div>

            <div v-if="loadingInstances && instances.length === 0" class="p-12 text-center text-sm text-theme-textLight">
              {{ $t('agents_page.loading_instances') }}
            </div>
            <div v-else-if="instances.length === 0" class="p-12 text-center">
              <p class="text-sm text-theme-textLight">{{ $t('agents_page.no_instances') }}</p>
              <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.no_instances_hint') }}</p>
            </div>
            <div v-else class="overflow-x-auto">
              <table class="min-w-full divide-y divide-theme-border">
                <thead class="bg-theme-hover">
                  <tr>
                    <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.instance') }}</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.address') }}</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.datacenter') }}</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.health') }}</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.lease_expires') }}</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.checks_title') }}</th>
                    <th v-if="canManage" class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.actions') }}</th>
                  </tr>
                </thead>
                <tbody class="bg-white divide-y divide-theme-border">
                  <template v-for="instance in instances" :key="instance.id">
                  <tr class="hover:bg-theme-hover">
                    <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-theme-text font-mono">{{ instance.instanceId }}</td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text font-mono">{{ instance.address }}:{{ instance.port }}</td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text">{{ instance.datacenter || '—' }}</td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm">
                      <span
                        class="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium"
                        :class="healthClass(instance.status)"
                      >
                        <span class="h-1.5 w-1.5 rounded-full" :class="healthDotClass(instance.status)" aria-hidden="true" />
                        {{ $t(`agents_page.health_${instance.status}`) }}
                      </span>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">{{ formatExpires(instance.leaseExpiresAt) }}</td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm">
                      <button
                        type="button"
                        class="inline-flex items-center gap-1 text-primary-600 hover:text-primary-700 hover:underline"
                        :aria-expanded="!!expanded[instance.id]"
                        @click="toggleInstance(instance.id)"
                      >
                        {{ $t('agents_page.checks_failing', { failing: failingCount(instance), total: (instance.checks || []).length }) }}
                        <svg class="h-3.5 w-3.5 transition-transform" :class="expanded[instance.id] ? 'rotate-180' : ''" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                        </svg>
                      </button>
                    </td>
                    <td v-if="canManage" class="px-6 py-4 whitespace-nowrap text-right text-sm">
                      <button
                        type="button"
                        class="text-red-600 hover:text-red-700 hover:underline disabled:opacity-50 disabled:cursor-not-allowed"
                        :disabled="deregisteringId === instance.instanceId"
                        @click="deregister(instance)"
                      >
                        {{ deregisteringId === instance.instanceId ? $t('agents_page.deregistering') : $t('agents_page.deregister') }}
                      </button>
                    </td>
                  </tr>
                  <tr v-if="expanded[instance.id]">
                    <td :colspan="canManage ? 7 : 6" class="bg-theme-hover px-6 py-4">
                      <ul class="space-y-3">
                        <li v-for="check in instance.checks" :key="check.checkId" class="grid gap-1 sm:grid-cols-[220px_110px_1fr]">
                          <div>
                            <p class="text-sm font-medium text-theme-text">{{ check.source === 'lease' ? $t('agents_page.check_lease') : check.name }}</p>
                            <p class="text-xs font-mono text-theme-textLight">{{ check.checkId }} · {{ $t(`agents_page.check_type_${check.type}`) }}</p>
                          </div>
                          <div>
                            <span
                              class="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium"
                              :class="healthClass(check.status)"
                            >
                              <span class="h-1.5 w-1.5 rounded-full" :class="healthDotClass(check.status)" aria-hidden="true" />
                              {{ $t(`agents_page.health_${check.status}`) }}
                            </span>
                          </div>
                          <div class="min-w-0">
                            <p class="text-sm text-theme-text break-words font-mono">{{ check.output || '—' }}</p>
                            <p class="text-xs text-theme-textLight">
                              {{ $t('agents_page.check_last_run', { when: formatRelative(check.lastRunAt, $t('agents_page.check_never')) }) }}
                            </p>
                          </div>
                        </li>
                      </ul>
                    </td>
                  </tr>
                  </template>
                </tbody>
              </table>
            </div>
          </section>
        </div>
      </template>
    </main>

    <AgentCheckModal
      v-if="agent"
      :open="checkModalOpen"
      :workspace-id="currentWorkspace.id"
      :agent-name="agent.name"
      :check="editingCheck"
      :existing-ids="checks.map((check) => check.checkId)"
      @close="checkModalOpen = false"
      @saved="onCheckSaved"
    />

    <AgentCardModal
      v-if="agent"
      :open="showEditModal"
      :workspace-id="currentWorkspace.id"
      :agent="agent"
      @close="showEditModal = false"
      @saved="onCardSaved"
    />

    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="showDeleteModal"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="delete-agent-title"
          @click.self="closeDeleteModal"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-sm p-6">
            <h2 id="delete-agent-title" class="text-lg font-semibold text-theme-text">
              {{ $t('agents_page.delete_title') }}
            </h2>
            <div v-if="deleteError" class="mt-3 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800" role="alert">
              {{ deleteError }}
            </div>
            <p class="mt-2 text-sm text-theme-textLight">
              {{ $t('agents_page.delete_agent_confirm', { name: agent?.name }) }}
            </p>
            <div class="mt-4">
              <label for="delete-agent-confirm" class="block text-sm font-medium text-theme-text mb-1.5">
                {{ $t('agents_page.type_to_confirm', { name: agent?.name }) }}
              </label>
              <input
                id="delete-agent-confirm"
                v-model="deleteConfirmName"
                type="text"
                class="input-field font-mono"
                :placeholder="agent?.name"
                autocomplete="off"
                :disabled="deleting"
              />
            </div>
            <div class="mt-6 flex gap-3 justify-end">
              <button type="button" class="btn-secondary" :disabled="deleting" @click="closeDeleteModal">
                {{ $t('common.cancel') }}
              </button>
              <button
                type="button"
                class="btn-primary bg-red-600 hover:bg-red-700 focus:ring-red-500"
                :disabled="deleting || deleteConfirmName.trim() !== agent?.name"
                @click="confirmDelete"
              >
                {{ deleting ? $t('agents_page.deleting') : $t('agents_page.delete') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppNav from '@/components/AppNav.vue'
import AgentCardModal from '@/components/AgentCardModal.vue'
import AgentCheckModal from '@/components/AgentCheckModal.vue'
import { agent_api, agent_check_api, agent_instance_api } from '@/api'
import { showFlash } from '@/lib/flash'
import { useWorkspaceContext } from '@/lib/permission'
import { healthClass, healthDotClass, parseCard } from '@/lib/agent'
import { formatExpires, formatRelative } from '@/utils/date'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { currentWorkspace, canManage } = useWorkspaceContext()

const agent = ref(null)
const instances = ref([])
const loading = ref(false)
const loadingInstances = ref(false)
const notFound = ref(false)
const errorMessage = ref(null)
const deregisteringId = ref(null)

const checks = ref([])
const loadingChecks = ref(false)
const checkModalOpen = ref(false)
const editingCheck = ref(null)
const deletingCheckId = ref(null)
const expanded = reactive({})

const showEditModal = ref(false)
const showDeleteModal = ref(false)
const deleteConfirmName = ref('')
const deleting = ref(false)
const deleteError = ref(null)

const agentName = computed(() => String(route.params.name || ''))
const card = computed(() => parseCard(agent.value?.card))
const rawCard = computed(() => JSON.stringify(agent.value?.card ?? {}, null, 2))

async function loadAgent() {
  loading.value = true
  notFound.value = false
  errorMessage.value = null
  agent.value = null
  instances.value = []
  checks.value = []

  try {
    const res = await agent_api.get(currentWorkspace.id, agentName.value)
    agent.value = res.data
    await Promise.all([loadChecks(), loadInstances()])
  } catch (err) {
    if (err.response?.status === 404) {
      notFound.value = true
    } else {
      errorMessage.value = err.response?.data?.errorMessage || t('agents_page.failed_load_agent')
    }
  } finally {
    loading.value = false
  }
}

async function loadInstances() {
  loadingInstances.value = true

  try {
    const res = await agent_instance_api.list(currentWorkspace.id, agentName.value)
    instances.value = res.data.instances || []
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('agents_page.failed_load_instances')
  } finally {
    loadingInstances.value = false
  }
}

async function loadChecks() {
  loadingChecks.value = true

  try {
    const res = await agent_check_api.list(currentWorkspace.id, agentName.value)
    checks.value = res.data.checks || []
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('agents_page.failed_load_checks')
  } finally {
    loadingChecks.value = false
  }
}

// Template changes alter instance statuses, so refresh the agent and instances too
async function refreshHealth() {
  const res = await agent_api.get(currentWorkspace.id, agentName.value)
  agent.value = res.data
  await Promise.all([loadChecks(), loadInstances()])
}

function openCheckModal(check) {
  editingCheck.value = check
  checkModalOpen.value = true
}

async function onCheckSaved() {
  checkModalOpen.value = false
  showFlash(t('common.saved'))
  try {
    await refreshHealth()
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('agents_page.failed_load_agent')
  }
}

async function deleteCheck(check) {
  if (!window.confirm(t('agents_page.check_delete_confirm', { name: check.name }))) return
  deletingCheckId.value = check.checkId
  errorMessage.value = null

  try {
    await agent_check_api.delete(currentWorkspace.id, agentName.value, check.checkId)
    showFlash(t('common.deleted'))
    await refreshHealth()
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('agents_page.check_failed_delete')
  } finally {
    deletingCheckId.value = null
  }
}

function describeCheck(check) {
  if (check.type === 'ttl') {
    return t('agents_page.check_describe_ttl', { ttl: check.ttl })
  }
  const port = check.port ? `:${check.port}` : ''
  const target = check.type === 'http' ? `${check.scheme}://…${port}${check.path}` : `tcp …${port}`
  return t('agents_page.check_describe_probe', { target, interval: check.interval, timeout: check.timeout })
}

function toggleInstance(id) {
  expanded[id] = !expanded[id]
}

function failingCount(instance) {
  return (instance.checks || []).filter((check) => check.status !== 'passing').length
}

function onCardSaved(updated) {
  showEditModal.value = false
  agent.value = updated
  showFlash(t('common.saved'))
}

async function deregister(instance) {
  if (!window.confirm(t('agents_page.deregister_confirm', { id: instance.instanceId }))) return
  deregisteringId.value = instance.instanceId
  errorMessage.value = null

  try {
    await agent_instance_api.delete(currentWorkspace.id, agentName.value, instance.instanceId)
    // Refresh the agent too, since its health is derived from the remaining instances
    const res = await agent_api.get(currentWorkspace.id, agentName.value)
    agent.value = res.data
    await loadInstances()
    showFlash(t('common.deleted'))
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('agents_page.failed_deregister')
  } finally {
    deregisteringId.value = null
  }
}

function openDeleteModal() {
  deleteConfirmName.value = ''
  deleteError.value = null
  showDeleteModal.value = true
}

function closeDeleteModal() {
  if (deleting.value) return
  showDeleteModal.value = false
  deleteConfirmName.value = ''
  deleteError.value = null
}

async function confirmDelete() {
  if (!agent.value) return
  deleting.value = true
  deleteError.value = null

  try {
    await agent_api.delete(currentWorkspace.id, agent.value.name)
    showDeleteModal.value = false
    showFlash(t('common.deleted'))
    router.push('/agents')
  } catch (err) {
    deleteError.value = err.response?.data?.errorMessage || t('agents_page.failed_delete')
  } finally {
    deleting.value = false
  }
}

watch(agentName, loadAgent, { immediate: true })
</script>
