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

      <section v-if="showGettingStarted" class="mb-8 overflow-hidden rounded-xl border border-theme-border bg-white shadow-sm">
        <div class="flex flex-wrap items-start justify-between gap-3 border-b border-theme-border px-6 py-4">
          <div class="min-w-0">
            <h2 class="text-lg font-semibold text-theme-text">{{ $t('dashboard.getting_started') }}</h2>
            <p class="mt-0.5 text-sm text-theme-textLight">{{ $t('dashboard.getting_started_desc') }}</p>
          </div>
          <button
            type="button"
            class="rounded-md px-2 py-1 text-sm text-theme-textLight transition hover:bg-theme-hover hover:text-theme-text"
            @click="dismissGettingStarted"
          >
            {{ $t('dashboard.getting_started_dismiss') }}
          </button>
        </div>

        <ol class="grid grid-cols-1 gap-px bg-theme-border sm:grid-cols-2 lg:grid-cols-3">
          <li v-for="(step, index) in gettingStartedSteps" :key="step.id" class="flex gap-4 bg-white px-6 py-5">
            <span
              class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-theme-hover text-sm font-semibold text-theme-text"
              aria-hidden="true"
            >
              {{ index + 1 }}
            </span>
            <span class="sr-only">{{ $t('dashboard.step_label', { number: index + 1 }) }}</span>
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-theme-text">{{ step.title }}</p>
              <p class="mt-0.5 text-sm leading-relaxed text-theme-textLight">{{ step.description }}</p>
              <router-link
                :to="step.to"
                class="mt-2 inline-flex text-sm font-medium text-primary-800 hover:underline"
              >
                {{ step.action }}
              </router-link>
            </div>
          </li>
        </ol>
      </section>

      <section class="mb-8 grid grid-cols-1 gap-6 lg:grid-cols-2">
        <div class="section overflow-hidden !p-0">
          <div class="border-b border-theme-border px-6 py-4">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div>
                <h2 class="text-lg font-semibold text-theme-text">{{ $t('dashboard.attention_title') }}</h2>
                <p class="mt-0.5 text-sm text-theme-textLight">{{ $t('dashboard.attention_subtitle') }}</p>
              </div>
              <span class="rounded-full border border-amber-200 bg-amber-50 px-2.5 py-0.5 text-xs font-medium text-amber-800">
                {{ $t('dashboard.mock_badge') }}
              </span>
            </div>
          </div>

          <ul class="divide-y divide-theme-border max-h-[28rem] overflow-y-auto">
            <li
              v-for="item in attentionItems"
              :key="item.id"
              class="flex gap-4 px-6 py-4"
            >
              <span
                class="mt-1.5 h-2 w-2 shrink-0 rounded-full"
                :class="attentionDotClass(item.level)"
                aria-hidden="true"
              />
              <div class="min-w-0 flex-1">
                <p class="text-sm font-medium text-theme-text">{{ item.title }}</p>
                <p class="mt-0.5 text-sm text-theme-textLight leading-relaxed">{{ item.description }}</p>
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

        <div class="section overflow-hidden !p-0">
          <div class="border-b border-theme-border px-6 py-4">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div>
                <h2 class="text-lg font-semibold text-theme-text">{{ $t('dashboard.activity_title') }}</h2>
                <p class="mt-0.5 text-sm text-theme-textLight">{{ $t('dashboard.activity_subtitle') }}</p>
              </div>
              <span class="rounded-full border border-amber-200 bg-amber-50 px-2.5 py-0.5 text-xs font-medium text-amber-800">
                {{ $t('dashboard.mock_badge') }}
              </span>
            </div>
          </div>

          <ul class="divide-y divide-theme-border">
            <li
              v-for="event in recentActivity"
              :key="event.id"
              class="flex gap-3 px-6 py-3.5"
            >
              <span
                class="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-theme-hover text-theme-textLight"
                aria-hidden="true"
              >
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" :d="activityIcons[event.icon]" />
                </svg>
              </span>
              <div class="min-w-0 flex-1">
                <p class="text-sm text-theme-text">
                  <span class="font-medium">{{ event.actor }}</span>
                  {{ event.action }}
                  <span class="font-medium">{{ event.target }}</span>
                </p>
                <p class="mt-0.5 text-xs text-theme-textLight">{{ event.time }}</p>
              </div>
            </li>
          </ul>

          <div class="border-t border-theme-border px-6 py-3">
            <router-link to="/audits" class="text-sm font-medium text-primary-800 hover:underline">
              {{ $t('dashboard.activity_view_all') }}
            </router-link>
          </div>
        </div>
      </section>

      <section class="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div class="space-y-6 lg:col-span-2">
          <div
            v-if="showWhatsNew"
            class="section overflow-hidden !p-0"
          >
            <div class="border-b border-theme-border px-6 py-4">
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div class="min-w-0">
                  <div class="flex flex-wrap items-center gap-2">
                    <h2 class="text-lg font-semibold text-theme-text">{{ $t('dashboard.whats_new_title') }}</h2>
                    <span class="rounded-full border border-amber-200 bg-amber-50 px-2.5 py-0.5 text-xs font-medium text-amber-800">
                      {{ $t('dashboard.mock_badge') }}
                    </span>
                  </div>
                  <p class="mt-0.5 text-sm text-theme-textLight">{{ $t('dashboard.whats_new_subtitle') }}</p>
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
              <p class="mt-2 text-base font-medium text-theme-text">{{ $t('dashboard.whats_new_tip_title') }}</p>
              <p class="mt-1 text-sm leading-relaxed text-theme-textLight">
                {{ $t('dashboard.whats_new_tip_body') }}
              </p>
            </div>
          </div>
        </div>

        <div class="space-y-6">
          <div class="section h-fit">
            <h2 class="section-title">{{ $t('dashboard.workspace_info') }}</h2>
            <dl class="space-y-0">
              <div class="flex justify-between gap-4 py-3 border-b border-theme-border">
                <dt class="text-sm text-theme-textLight">{{ $t('dashboard.workspace_name') }}</dt>
                <dd class="text-sm font-medium text-theme-text text-right">{{ workspace?.name || '—' }}</dd>
              </div>
              <div class="flex justify-between gap-4 py-3 border-b border-theme-border">
                <dt class="text-sm text-theme-textLight">{{ $t('dashboard.workspace_handle') }}</dt>
                <dd class="text-sm font-medium text-theme-text text-right font-mono">{{ workspace?.handle || '—' }}</dd>
              </div>
              <div class="flex justify-between gap-4 py-3 border-b border-theme-border">
                <dt class="text-sm text-theme-textLight">{{ $t('dashboard.workspace_role') }}</dt>
                <dd class="text-sm font-medium text-theme-text capitalize text-right">{{ roleLabel }}</dd>
              </div>
              <div v-if="isSaaS()" class="flex justify-between gap-4 py-3 border-b border-theme-border">
                <dt class="text-sm text-theme-textLight">{{ $t('dashboard.workspace_tokens') }}</dt>
                <dd class="text-sm font-medium text-theme-text text-right">{{ tokenBalanceLabel }}</dd>
              </div>
            </dl>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  billing_api,
  workspace_api,
} from '@/api'
import AppNav from '@/components/AppNav.vue'
import { user } from '@/lib/auth'
import { saveWorkspaceToStorage } from '@/utils/storage'
import { useWorkspaceContext } from '@/lib/permission'
import { isSaaS } from '@/lib/edition'

const { t } = useI18n()

const WHATS_NEW_KEY = 'dashboard_whats_new_v1'
const GETTING_STARTED_KEY = 'dashboard_getting_started_v1'

const { currentWorkspace, canManage } = useWorkspaceContext()
const workspace = ref({ ...currentWorkspace })
const loading = ref(true)
const tokenBalance = ref(null)
const showWhatsNew = ref(localStorage.getItem(WHATS_NEW_KEY) !== '1')
const showGettingStarted = ref(localStorage.getItem(GETTING_STARTED_KEY) !== '1')

function dismissGettingStarted() {
  showGettingStarted.value = false
  localStorage.setItem(GETTING_STARTED_KEY, '1')
}

function dismissWhatsNew() {
  showWhatsNew.value = false
  localStorage.setItem(WHATS_NEW_KEY, '1')
}

const tokenBalanceLabel = computed(() => {
  if (tokenBalance.value == null) return loading.value ? '…' : '—'
  return tokenBalance.value.toLocaleString()
})

const roleLabel = computed(() => workspace.value?.role || '—')

const gettingStartedSteps = computed(() => [
  {
    id: 'access_key',
    title: t('dashboard.step_access_key'),
    description: t('dashboard.step_access_key_desc'),
    action: t('dashboard.step_access_key_link'),
    to: '/settings',
    requiresManage: true,
  },
  {
    id: 'register_agent',
    title: t('dashboard.step_register_agent'),
    description: t('dashboard.step_register_agent_desc'),
    action: t('dashboard.step_register_agent_link'),
    to: '/agents',
  },
  {
    id: 'register_instance',
    title: t('dashboard.step_register_instance'),
    description: t('dashboard.step_register_instance_desc'),
    action: t('dashboard.step_register_instance_link'),
    to: '/agents',
  },
  {
    id: 'health_check',
    title: t('dashboard.step_health_check'),
    description: t('dashboard.step_health_check_desc'),
    action: t('dashboard.step_health_check_link'),
    to: '/agents',
  },
  {
    id: 'intention',
    title: t('dashboard.step_intention'),
    description: t('dashboard.step_intention_desc'),
    action: t('dashboard.step_intention_link'),
    to: '/intentions',
  },
  {
    id: 'traffic',
    title: t('dashboard.step_traffic'),
    description: t('dashboard.step_traffic_desc'),
    action: t('dashboard.step_traffic_link'),
    to: '/traffic',
  },
].filter((step) => !step.requiresManage || canManage.value))

const attentionItems = computed(() => [
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
    to: '/audits',
  },
].filter((item) => !item.requiresSaaS || isSaaS()))

function attentionDotClass(level) {
  return {
    error: 'bg-red-500',
    warn: 'bg-amber-500',
    info: 'bg-sky-500',
  }[level] || 'bg-theme-textLight'
}

const activityIcons = {
  member: 'M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z',
  key: 'M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z',
}

const recentActivity = computed(() => [
  {
    id: 2,
    icon: 'member',
    actor: 'Ahmed',
    action: t('dashboard.activity_invited'),
    target: 'maya@example.com',
    time: t('dashboard.activity_time_3h'),
  },
  {
    id: 3,
    icon: 'key',
    actor: 'Ahmed',
    action: t('dashboard.activity_created_key'),
    target: 'production-ci',
    time: t('dashboard.activity_time_1d'),
  },
])

async function loadDashboard() {
  loading.value = true

  const tasks = [
    workspace_api.get(currentWorkspace.id)
      .then((res) => {
        workspace.value = { ...currentWorkspace, ...res.data, role: res.data.role ?? currentWorkspace.role }
        saveWorkspaceToStorage(workspace.value)
      })
      .catch(() => {}),
  ]

  if (canManage.value && isSaaS()) {
    tasks.push(
      billing_api.status(currentWorkspace.id)
        .then((res) => { tokenBalance.value = res.data.aiTokensBalance || 0 })
        .catch(() => {}),
    )
  }

  await Promise.allSettled(tasks)
  loading.value = false
}

onMounted(loadDashboard)
</script>
