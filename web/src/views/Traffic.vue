<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="mb-8">
        <h1 class="text-3xl font-semibold text-theme-text">{{ $t('traffic_page.title') }}</h1>
        <p class="text-theme-textLight mt-2">{{ $t('traffic_page.subtitle') }}</p>
      </div>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
        <div class="border-b border-theme-border px-6 py-4">
          <h2 class="text-lg font-semibold text-theme-text">{{ $t('traffic_page.list_title') }}</h2>
          <p class="mt-1 text-sm text-theme-textLight">{{ $t('traffic_page.list_desc') }}</p>
        </div>

        <div v-if="errorMessage" class="m-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
          {{ errorMessage }}
        </div>

        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-theme-border">
            <thead class="bg-theme-hover">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('traffic_page.when') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('traffic_page.source') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('traffic_page.destination') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('traffic_page.skill') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('traffic_page.verb') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('traffic_page.decision') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('traffic_page.result') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('traffic_page.latency') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('traffic_page.task') }}</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-theme-border">
              <tr v-if="loading">
                <td colspan="9" class="px-6 py-4 text-center text-sm text-theme-textLight">{{ $t('common.loading') }}</td>
              </tr>
              <tr v-else-if="!calls.length">
                <td colspan="9" class="px-6 py-4 text-center text-sm text-theme-textLight">{{ $t('traffic_page.empty') }}</td>
              </tr>
              <template v-else>
                <tr v-for="call in calls" :key="call.id" class="hover:bg-theme-hover">
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                    <time :datetime="call.createdAt" :title="formatDate(call.createdAt)">{{ formatRelative(call.createdAt) }}</time>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm font-medium font-mono">
                    <router-link
                      :to="{ name: 'Agent', params: { name: call.sourceInstance } }"
                      class="text-primary-800 hover:underline"
                    >
                      {{ call.sourceInstance }}
                    </router-link>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm font-medium font-mono">
                    <router-link
                      v-if="call.destinationAgent"
                      :to="{ name: 'Agent', params: { name: call.destinationAgent } }"
                      class="text-primary-800 hover:underline"
                    >
                      {{ call.destinationAgent }}
                    </router-link>
                    <span v-else class="text-theme-textLight">—</span>
                    <div v-if="call.destinationInstance" class="text-xs font-normal text-theme-textLight">{{ call.destinationInstance }}</div>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text font-mono">{{ call.skill || '—' }}</td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text font-mono">{{ call.verb }}</td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm">
                    <span
                      class="inline-flex rounded-full px-2 py-0.5 text-xs font-medium"
                      :class="call.decision === 'allow' ? 'bg-emerald-50 text-emerald-800' : 'bg-red-50 text-red-800'"
                    >
                      {{ $t(`traffic_page.decision_${call.decision}`) }}
                    </span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm">
                    <span
                      class="inline-flex rounded-full px-2 py-0.5 text-xs font-medium"
                      :class="resultClass(call.result)"
                    >
                      {{ $t(`traffic_page.result_${call.result}`) }}
                    </span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text">{{ call.latencyMs ? `${call.latencyMs} ms` : '—' }}</td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text font-mono">{{ call.taskId || '—' }}</td>
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
            {{ $t('traffic_page.showing', { from: offset + 1, to: Math.min(offset + limit, total), total }) }}
          </div>
          <div class="flex items-center gap-2">
            <button
              type="button"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="offset === 0"
              @click="goToPage(offset - limit)"
            >
              {{ $t('traffic_page.previous') }}
            </button>
            <button
              type="button"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="offset + limit >= total"
              @click="goToPage(offset + limit)"
            >
              {{ $t('traffic_page.next') }}
            </button>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppNav from '@/comp/AppNav.vue'
import { traffic_api } from '@/api'
import { useUrlPagination } from '@/lib/pagination'
import { useWorkspaceContext } from '@/lib/permission'
import { errorText } from '@/lib/error'
import { formatRelative } from '@/utils/date'

const { t, locale } = useI18n()
const route = useRoute()
const { currentWorkspace } = useWorkspaceContext()
const { offset, limit, goToOffset } = useUrlPagination({ limit: 50 })

const calls = ref([])
const total = ref(0)
const loading = ref(false)
const errorMessage = ref(null)

function formatDate(iso) {
  return new Date(iso).toLocaleString(locale.value, { month: 'short', day: 'numeric', year: 'numeric', hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function resultClass(result) {
  return {
    completed: 'bg-emerald-50 text-emerald-800',
    failed: 'bg-red-50 text-red-800',
    denied: 'bg-red-50 text-red-800',
    timeout: 'bg-amber-50 text-amber-800',
    no_instance: 'bg-amber-50 text-amber-800',
  }[result] || 'bg-theme-hover text-theme-textLight'
}

async function loadTraffic() {
  loading.value = true
  errorMessage.value = null

  try {
    const res = await traffic_api.list(currentWorkspace.id, { limit: limit, offset: offset.value })
    calls.value = res.data.calls || []
    total.value = res.data._meta?.total || 0
  } catch (err) {
    errorMessage.value = errorText(err, t('traffic_page.failed_load'))
    calls.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function goToPage(newOffset) {
  if (newOffset < 0 || newOffset >= total.value) return
  goToOffset(newOffset)
}

watch(() => route.query.page, loadTraffic)
onMounted(loadTraffic)
</script>
