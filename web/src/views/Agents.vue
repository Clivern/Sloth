<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="mb-8">
        <h1 class="text-3xl font-semibold text-theme-text">{{ $t('agents_page.title') }}</h1>
        <p class="text-theme-textLight mt-2">{{ $t('agents_page.subtitle') }}</p>
      </div>

      <div v-if="errorMessage" class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
        {{ errorMessage }}
      </div>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
        <div class="flex flex-wrap items-start justify-between gap-4 border-b border-theme-border px-6 py-4">
          <div>
            <h2 class="text-lg font-semibold text-theme-text">{{ $t('agents_page.list_title') }}</h2>
            <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.list_desc') }}</p>
          </div>
          <button
            v-if="canManage"
            type="button"
            class="btn-primary inline-flex items-center justify-center gap-2"
            @click="showRegisterModal = true"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            {{ $t('agents_page.register') }}
          </button>
        </div>

        <div v-if="loading" class="p-12 text-center">
          <svg class="animate-spin h-8 w-8 mx-auto text-theme-text" fill="none" viewBox="0 0 24 24" aria-hidden="true">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          <p class="text-theme-textLight mt-3">{{ $t('agents_page.loading') }}</p>
        </div>

        <div v-else-if="agents.length === 0" class="p-12 text-center">
          <p class="text-theme-textLight">{{ $t('agents_page.no_agents') }}</p>
          <p class="text-sm text-theme-textLight mt-1">{{ $t('agents_page.no_agents_hint') }}</p>
        </div>

        <div v-else class="overflow-x-auto">
          <table class="min-w-full divide-y divide-theme-border">
            <thead class="bg-theme-hover">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.name') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.version') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.skills') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.instances') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.health') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.updated') }}</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-theme-border">
              <tr v-for="agent in rows" :key="agent.id" class="hover:bg-theme-hover">
                <td class="px-6 py-4">
                  <router-link
                    :to="{ name: 'Agent', params: { name: agent.name } }"
                    class="text-sm font-medium text-theme-text font-mono hover:underline"
                  >
                    {{ agent.name }}
                  </router-link>
                  <p v-if="agent.description" class="mt-0.5 text-sm text-theme-textLight">{{ agent.description }}</p>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text font-mono">{{ agent.version }}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text">
                  {{ $t('agents_page.skills_count', agent.skillCount) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text">
                  {{ $t('agents_page.live_of_total', { live: agent.live, total: agent.instanceCount }) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm">
                  <span
                    class="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium"
                    :class="healthClass(agent.health)"
                  >
                    <span class="h-1.5 w-1.5 rounded-full" :class="healthDotClass(agent.health)" aria-hidden="true" />
                    {{ $t(`agents_page.health_${agent.health}`) }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">{{ formatRelative(agent.updatedAt) }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div
          v-if="total > 0"
          class="bg-white px-6 py-4 border-t border-theme-border flex items-center justify-between"
        >
          <div class="text-sm text-theme-textLight">
            {{ $t('agents_page.showing', { from: offset + 1, to: Math.min(offset + limit, total), total }) }}
          </div>
          <div class="flex items-center gap-2">
            <button
              type="button"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="offset === 0"
              @click="goToPage(offset - limit)"
            >
              {{ $t('agents_page.previous') }}
            </button>
            <button
              type="button"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="offset + limit >= total"
              @click="goToPage(offset + limit)"
            >
              {{ $t('agents_page.next') }}
            </button>
          </div>
        </div>
      </section>
    </main>

    <AgentCardModal
      :open="showRegisterModal"
      :workspace-id="currentWorkspace.id"
      @close="showRegisterModal = false"
      @saved="onRegistered"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppNav from '@/components/AppNav.vue'
import AgentCardModal from '@/components/AgentCardModal.vue'
import { agent_api } from '@/api'
import { showFlash } from '@/lib/flash'
import { useUrlPagination } from '@/lib/pagination'
import { useWorkspaceContext } from '@/lib/permission'
import { countPassing, healthClass, healthDotClass, parseCard } from '@/lib/agent'
import { formatRelative } from '@/utils/date'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { currentWorkspace, canManage } = useWorkspaceContext()
const { offset, limit, goToOffset } = useUrlPagination({ limit: 50 })

const agents = ref([])
const total = ref(0)
const loading = ref(false)
const errorMessage = ref(null)
const showRegisterModal = ref(false)

const rows = computed(() => agents.value.map((agent) => {
  const card = parseCard(agent.card)
  const instances = agent.instances || []
  return {
    id: agent.id,
    name: agent.name,
    version: agent.version,
    description: card.description,
    skillCount: card.skills.length,
    live: countPassing(instances),
    instanceCount: instances.length,
    health: agent.health,
    updatedAt: agent.updatedAt,
  }
}))

async function loadPage() {
  loading.value = true
  errorMessage.value = null

  try {
    const res = await agent_api.list(currentWorkspace.id, { limit, offset: offset.value })
    agents.value = res.data.agents || []
    total.value = res.data._meta?.total || 0
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('agents_page.failed_load')
    agents.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function goToPage(newOffset) {
  if (newOffset < 0 || newOffset >= total.value) return
  goToOffset(newOffset)
}

function onRegistered(agent) {
  showRegisterModal.value = false
  showFlash(t('common.created'))
  router.push({ name: 'Agent', params: { name: agent.name } })
}

watch(() => route.query.page, loadPage)

onMounted(loadPage)
</script>
