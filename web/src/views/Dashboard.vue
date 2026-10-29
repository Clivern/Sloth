<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          {{ $t('dashboard.title') }}
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          {{ $t('dashboard.subtitle') }}
        </p>
        <p v-if="user?.email" class="mt-2 text-sm text-theme-text">
          {{ $t('dashboard.welcome_back') }}
          <span class="font-medium">{{ user.name || user.email }}</span>
        </p>
      </header>

      <section class="mb-8 grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
        <div class="section">
          <p class="text-sm text-theme-textLight">{{ $t('dashboard.metrics_calls') }}</p>
          <p class="mt-1 text-2xl font-semibold tabular-nums text-theme-text">
            {{ formatCount(callsTotal) }}
          </p>
          <Sparkline v-if="callsSeries.length > 1" :values="callsSeries" class="mt-3" />
          <p class="mt-2 text-xs text-theme-textLight">
            {{ $t('dashboard.metrics_calls_sub', { window: METRICS_WINDOW }) }}
          </p>
        </div>
        <div class="section">
          <p class="text-sm text-theme-textLight">{{ $t('dashboard.metrics_problems') }}</p>
          <p class="mt-1 text-2xl font-semibold tabular-nums text-theme-text">
            {{ formatCount(problemsCount) }}
          </p>
          <Sparkline
            v-if="problemsSeries.length > 1"
            :values="problemsSeries"
            color="var(--chart-2)"
            class="mt-3"
          />
          <p class="mt-2 text-xs text-theme-textLight">
            {{ $t('dashboard.metrics_problems_sub', { window: METRICS_WINDOW }) }}
          </p>
        </div>
        <div class="section">
          <p class="text-sm text-theme-textLight">{{ $t('dashboard.metrics_latency') }}</p>
          <p class="mt-1 text-2xl font-semibold tabular-nums text-theme-text">
            {{ avgLatency == null ? '—' : `${avgLatency} ms` }}
          </p>
          <Sparkline
            v-if="latencySeries.length > 1"
            :values="latencySeries"
            color="var(--chart-4)"
            class="mt-3"
          />
          <p class="mt-2 text-xs text-theme-textLight">
            {{ $t('dashboard.metrics_latency_sub', { window: METRICS_WINDOW }) }}
          </p>
        </div>
        <div class="section">
          <p class="text-sm text-theme-textLight">{{ $t('dashboard.metrics_agents') }}</p>
          <p class="mt-1 text-2xl font-semibold tabular-nums text-theme-text">
            {{ formatCount(healthyAgents) }}
          </p>
          <p class="mt-2 text-xs text-theme-textLight">
            {{
              agentsTotal == null
                ? '—'
                : $t('dashboard.metrics_agents_sub', {
                    passing: healthyAgents ?? 0,
                    total: agentsTotal,
                  })
            }}
          </p>
        </div>
      </section>

      <section class="mb-8">
        <div class="mb-4 flex flex-wrap items-end justify-between gap-4">
          <div>
            <h2 class="text-lg font-semibold text-theme-text">
              {{ $t('dashboard.agents_traffic_title') }}
            </h2>
            <p class="mt-1 text-sm text-theme-textLight">
              {{ $t('dashboard.agents_traffic_desc') }}
            </p>
          </div>
          <select
            v-model="selectedAgent"
            class="input-field w-auto"
            :aria-label="$t('dashboard.agents_traffic_title')"
            :disabled="!agentNames.length"
          >
            <option v-if="!agentNames.length" value="">
              {{ $t('dashboard.agents_traffic_empty') }}
            </option>
            <option v-for="name in agentNames" :key="name" :value="name">{{ name }}</option>
          </select>
        </div>
        <AgentTraffic
          v-if="selectedAgent"
          :key="selectedAgent"
          :workspace-id="currentWorkspace.id"
          :agent-name="selectedAgent"
        />
      </section>

      <section class="mb-8 grid grid-cols-1 gap-6 lg:grid-cols-2">
        <div class="section overflow-hidden !p-0">
          <div class="border-b border-theme-border px-6 py-4">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div>
                <h2 class="text-lg font-semibold text-theme-text">
                  {{ $t('dashboard.attention_title') }}
                </h2>
                <p class="mt-0.5 text-sm text-theme-textLight">
                  {{ $t('dashboard.attention_subtitle') }}
                </p>
              </div>
              <span
                class="rounded-full border border-amber-200 bg-amber-50 px-2.5 py-0.5 text-xs font-medium text-amber-800"
              >
                {{ $t('dashboard.mock_badge') }}
              </span>
            </div>
          </div>

          <ul class="divide-y divide-theme-border max-h-[28rem] overflow-y-auto">
            <li v-for="item in attentionItems" :key="item.id" class="flex gap-4 px-6 py-4">
              <span
                class="mt-1.5 h-2 w-2 shrink-0 rounded-full"
                :class="attentionDotClass(item.level)"
                aria-hidden="true"
              />
              <div class="min-w-0 flex-1">
                <p class="text-sm font-medium text-theme-text">{{ item.title }}</p>
                <p class="mt-0.5 text-sm text-theme-textLight leading-relaxed">
                  {{ item.description }}
                </p>
                <router-link
                  v-if="item.to"
                  :to="item.to"
                  class="mt-2 inline-flex text-sm font-medium text-primary-800 hover:underline"
                >
                  {{ item.action }}
                </router-link>
              </div>
            </li>
          </ul>
        </div>

        <div v-if="showWhatsNew" class="section overflow-hidden !p-0">
          <div class="border-b border-theme-border px-6 py-4">
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <h2 class="text-lg font-semibold text-theme-text">
                    {{ $t('dashboard.whats_new_title') }}
                  </h2>
                  <span
                    class="rounded-full border border-amber-200 bg-amber-50 px-2.5 py-0.5 text-xs font-medium text-amber-800"
                  >
                    {{ $t('dashboard.mock_badge') }}
                  </span>
                </div>
                <p class="mt-0.5 text-sm text-theme-textLight">
                  {{ $t('dashboard.whats_new_subtitle') }}
                </p>
              </div>
              <button
                type="button"
                class="rounded-md px-2 py-1 text-sm text-theme-textLight transition hover:bg-theme-hover hover:text-theme-text"
                @click="dismissWhatsNew"
              >
                {{ $t('dashboard.whats_new_dismiss') }}
              </button>
            </div>
          </div>
          <div class="px-6 py-5">
            <p class="text-xs font-medium uppercase tracking-wide text-theme-textLight">
              {{ $t('dashboard.whats_new_tip_label') }}
            </p>
            <p class="mt-2 text-base font-medium text-theme-text">
              {{ $t('dashboard.whats_new_tip_title') }}
            </p>
            <p class="mt-1 text-sm leading-relaxed text-theme-textLight">
              {{ $t('dashboard.whats_new_tip_body') }}
            </p>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppNav from '@/comp/AppNav.vue'
import AgentTraffic from '@/comp/AgentTraffic.vue'
import Sparkline from '@/comp/Sparkline.vue'
import { agent_api, traffic_api } from '@/api'
import { user } from '@/lib/auth'
import { useWorkspaceContext } from '@/lib/permission'
import { isSaaS } from '@/lib/edition'

const { t } = useI18n()

const WHATS_NEW_KEY = 'dashboard_whats_new_v1'
const METRICS_WINDOW = 200
const SPARK_DAYS = 14

const { currentWorkspace } = useWorkspaceContext()
const showWhatsNew = ref(localStorage.getItem(WHATS_NEW_KEY) !== '1')

const callsTotal = ref(null)
const problemsCount = ref(null)
const avgLatency = ref(null)
const healthyAgents = ref(null)
const agentsTotal = ref(null)
const callsSeries = ref([])
const problemsSeries = ref([])
const latencySeries = ref([])
const agentNames = ref([])
const selectedAgent = ref('')

function formatCount(value) {
  return value == null ? '—' : value.toLocaleString()
}

function lastDays(count) {
  return [...Array(count)].map((_, i) => {
    const day = new Date()
    day.setDate(day.getDate() - (count - 1 - i))

    return day.toISOString().slice(0, 10)
  })
}

async function loadMetrics() {
  try {
    const res = await traffic_api.list(currentWorkspace.id, {
      limit: METRICS_WINDOW,
      offset: 0,
    })
    const rows = res.data.calls || []
    callsTotal.value = res.data._meta?.total ?? rows.length

    const problems = rows.filter((call) => call.decision === 'deny' || call.result === 'failed')
    problemsCount.value = problems.length

    const latencies = rows.map((call) => call.latencyMs).filter((ms) => ms != null)
    avgLatency.value = latencies.length
      ? Math.round(latencies.reduce((sum, ms) => sum + ms, 0) / latencies.length)
      : null

    const days = lastDays(SPARK_DAYS)
    const perDay = Object.fromEntries(
      days.map((day) => [day, { calls: 0, problems: 0, latency: [] }])
    )

    for (const call of rows) {
      const bucket = perDay[new Date(call.createdAt).toISOString().slice(0, 10)]

      if (bucket) {
        bucket.calls++
        if (call.decision === 'deny' || call.result === 'failed') bucket.problems++
        if (call.latencyMs != null) bucket.latency.push(call.latencyMs)
      }
    }

    callsSeries.value = days.map((day) => perDay[day].calls)
    problemsSeries.value = days.map((day) => perDay[day].problems)
    latencySeries.value = days.map((day) =>
      perDay[day].latency.length
        ? Math.round(
            perDay[day].latency.reduce((sum, ms) => sum + ms, 0) / perDay[day].latency.length
          )
        : 0
    )
  } catch (_) {
    // Metrics stay blank when the traffic cannot be read.
  }
}

async function loadAgentHealth() {
  try {
    const loaded = []

    for (;;) {
      const res = await agent_api.list(currentWorkspace.id, {
        limit: 100,
        offset: loaded.length,
      })
      const page = res.data.agents || []
      loaded.push(...page)

      if (page.length < 100 || loaded.length >= (res.data._meta?.total || 0)) break
    }

    agentsTotal.value = loaded.length
    healthyAgents.value = loaded.filter((agent) => agent.health === 'passing').length
    agentNames.value = loaded.map((agent) => agent.name).sort()

    if (!selectedAgent.value && agentNames.value.length) {
      selectedAgent.value = agentNames.value[0]
    }
  } catch (_) {
    // Metrics stay blank when the agents cannot be read.
  }
}

onMounted(() => {
  loadMetrics()
  loadAgentHealth()
})

function dismissWhatsNew() {
  showWhatsNew.value = false
  localStorage.setItem(WHATS_NEW_KEY, '1')
}

const attentionItems = computed(() =>
  [
    {
      id: 'key',
      level: 'error',
      title: t('dashboard.attention_key_title'),
      description: t('dashboard.attention_key_desc'),
      action: t('dashboard.attention_key_action'),
      to: '/settings',
    },
    {
      id: 'invite',
      level: 'info',
      title: t('dashboard.attention_invite_title'),
      description: t('dashboard.attention_invite_desc'),
      action: t('dashboard.attention_invite_action'),
      to: '/members',
    },
    {
      id: 'integration',
      level: 'info',
      title: t('dashboard.attention_integration_title'),
      description: t('dashboard.attention_integration_desc'),
      action: t('dashboard.attention_integration_action'),
      to: '/integrations',
    },
    {
      id: 'trial',
      level: 'warn',
      title: t('dashboard.attention_trial_title'),
      description: t('dashboard.attention_trial_desc'),
      action: t('dashboard.attention_trial_action'),
      to: '/billing',
      requiresSaaS: true,
    },
    {
      id: 'stale_key',
      level: 'info',
      title: t('dashboard.attention_stale_key_title'),
      description: t('dashboard.attention_stale_key_desc'),
      action: t('dashboard.attention_stale_key_action'),
      to: '/settings',
    },
    {
      id: 'latency',
      level: 'warn',
      title: t('dashboard.attention_latency_title'),
      description: t('dashboard.attention_latency_desc'),
      action: t('dashboard.attention_latency_action'),
      to: '/traffic',
    },
  ].filter((item) => !item.requiresSaaS || isSaaS())
)

function attentionDotClass(level) {
  return (
    {
      error: 'bg-red-500',
      warn: 'bg-amber-500',
      info: 'bg-sky-500',
    }[level] || 'bg-theme-textLight'
  )
}
</script>
