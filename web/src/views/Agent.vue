<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
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
        <div class="mb-6 flex flex-wrap items-start justify-between gap-4">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-3">
              <h1 class="text-3xl font-semibold text-theme-text font-mono break-all">{{ agent.name }}</h1>
              <span
                v-if="agent.health"
                class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
                :class="healthClass(agent.health)"
              >
                <span class="h-1.5 w-1.5 rounded-full" :class="healthDotClass(agent.health)" aria-hidden="true" />
                {{ $t(`agents_page.health_${agent.health}`) }}
              </span>
              <span class="rounded-full border border-theme-border bg-white px-2.5 py-1 text-xs font-mono text-theme-textLight">
                v{{ agent.version }}
              </span>
            </div>
            <p v-if="card.description" class="mt-2 max-w-3xl text-theme-textLight">{{ card.description }}</p>
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

        <div class="mb-6 grid grid-cols-2 gap-4 lg:grid-cols-4">
          <a href="#instances" class="rounded-xl border border-theme-border bg-white p-4 shadow-sm transition-colors hover:border-primary-300">
            <p class="text-xs font-medium uppercase tracking-wider text-theme-textLight">{{ $t('agents_page.instances') }}</p>
            <p class="mt-1 text-2xl font-semibold text-theme-text">{{ instancesTotal }}</p>
            <div class="mt-1 min-h-5">
              <StatusBreakdown v-if="instancesTotal" :counts="agent.statusCounts || {}" hide-zero />
            </div>
          </a>
          <a href="#checks" class="rounded-xl border border-theme-border bg-white p-4 shadow-sm transition-colors hover:border-primary-300">
            <p class="text-xs font-medium uppercase tracking-wider text-theme-textLight">{{ $t('agents_page.checks_title') }}</p>
            <p class="mt-1 text-2xl font-semibold text-theme-text">{{ checks.length }}</p>
            <p class="mt-1 text-sm text-theme-textLight">{{ checkTypesSummary }}</p>
          </a>
          <a href="#protection" class="rounded-xl border border-theme-border bg-white p-4 shadow-sm transition-colors hover:border-primary-300">
            <p class="text-xs font-medium uppercase tracking-wider text-theme-textLight">{{ $t('agents_page.protection_title') }}</p>
            <p class="mt-1 text-2xl font-semibold" :class="protection?.isActive ? 'text-emerald-700' : 'text-theme-text'">
              {{ protection ? (protection.isActive ? $t('agents_page.protection_active') : $t('agents_page.protection_inactive')) : $t('agents_page.protection_not_set') }}
            </p>
            <p class="mt-1 text-sm text-theme-textLight">
              {{ protection ? $t(`agents_page.protection_type_${protection.type}`) : $t('agents_page.protection_not_set_hint') }}
            </p>
          </a>
          <a href="#identities" class="rounded-xl border border-theme-border bg-white p-4 shadow-sm transition-colors hover:border-primary-300">
            <p class="text-xs font-medium uppercase tracking-wider text-theme-textLight">{{ $t('agents_page.identities_title') }}</p>
            <p class="mt-1 text-2xl font-semibold text-theme-text">{{ identities.length }}</p>
            <p class="mt-1 text-sm text-theme-textLight">
              {{ $t('agents_page.identities_active_count', { count: identities.filter((identity) => identity.isActive).length }) }}
            </p>
          </a>
        </div>

        <div class="grid gap-6 lg:grid-cols-3">
          <div class="min-w-0 space-y-6 lg:col-span-2">
            <section id="instances" class="scroll-mt-6 rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
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
                      <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.health') }}</th>
                      <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.checks_title') }}</th>
                      <th v-if="canManage" class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('agents_page.actions') }}</th>
                    </tr>
                  </thead>
                  <tbody class="bg-white divide-y divide-theme-border">
                    <template v-for="instance in instances" :key="instance.id">
                    <tr class="hover:bg-theme-hover">
                      <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-theme-text font-mono">{{ instance.instanceId }}</td>
                      <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text font-mono">{{ instance.address }}:{{ instance.port }}</td>
                      <td class="px-6 py-4 whitespace-nowrap text-sm">
                        <span
                          class="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium"
                          :class="healthClass(instance.status)"
                        >
                          <span class="h-1.5 w-1.5 rounded-full" :class="healthDotClass(instance.status)" aria-hidden="true" />
                          {{ $t(`agents_page.health_${instance.status}`) }}
                        </span>
                      </td>
                      <td class="px-6 py-4 whitespace-nowrap text-sm">
                        <span v-if="!(instance.checks || []).length" class="text-theme-textLight">{{ $t('agents_page.instance_no_checks') }}</span>
                        <button
                          v-else
                          type="button"
                          class="inline-flex items-center gap-1 text-primary-600 hover:text-primary-700 hover:underline"
                          :aria-expanded="!!expanded[instance.id]"
                          @click="toggleInstance(instance.id)"
                        >
                          <StatusBreakdown :counts="countByStatus(instance.checks)" hide-zero />
                          <svg class="h-3.5 w-3.5 transition-transform" :class="expanded[instance.id] ? 'rotate-180' : ''" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                          </svg>
                        </button>
                      </td>
                      <td v-if="canManage" class="px-6 py-4 whitespace-nowrap text-right text-sm">
                        <button
                          type="button"
                          class="text-red-600 hover:text-red-700 hover:underline"
                          @click="askDeregister(instance)"
                        >
                          {{ $t('agents_page.deregister') }}
                        </button>
                      </td>
                    </tr>
                    <tr v-if="expanded[instance.id]">
                      <td :colspan="canManage ? 5 : 4" class="bg-theme-hover px-6 py-4">
                        <ul class="space-y-3">
                          <li v-for="check in instance.checks" :key="check.checkId" class="grid gap-1 sm:grid-cols-[220px_110px_1fr]">
                            <div>
                              <p class="text-sm font-medium text-theme-text">{{ check.name }}</p>
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

              <div
                v-if="instancesTotal > 0"
                class="bg-white px-6 py-4 border-t border-theme-border flex items-center justify-between"
              >
                <div class="text-sm text-theme-textLight">
                  {{ $t('agents_page.showing', { from: instancesOffset + 1, to: Math.min(instancesOffset + instancesLimit, instancesTotal), total: instancesTotal }) }}
                </div>
                <div class="flex items-center gap-2">
                  <button
                    type="button"
                    class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
                    :disabled="instancesOffset === 0"
                    @click="goToInstancesPage(instancesOffset - instancesLimit)"
                  >
                    {{ $t('agents_page.previous') }}
                  </button>
                  <button
                    type="button"
                    class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
                    :disabled="instancesOffset + instancesLimit >= instancesTotal"
                    @click="goToInstancesPage(instancesOffset + instancesLimit)"
                  >
                    {{ $t('agents_page.next') }}
                  </button>
                </div>
              </div>
            </section>

            <section id="checks" class="scroll-mt-6 rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
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
                            class="text-red-600 hover:text-red-700 hover:underline"
                            @click="askDeleteCheck(check)"
                          >
                            {{ $t('agents_page.delete') }}
                          </button>
                        </span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </section>

            <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
              <div class="border-b border-theme-border px-6 py-4">
                <h2 class="text-lg font-semibold text-theme-text">{{ $t('agents_page.skills') }}</h2>
                <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.skills_desc') }}</p>
              </div>
              <ul v-if="card.skills.length" class="divide-y divide-theme-border">
                <li v-for="skill in card.skills" :key="skill.id" class="px-6 py-4">
                  <div class="flex flex-wrap items-center gap-2">
                    <span v-if="skill.name" class="text-sm font-medium text-theme-text">{{ skill.name }}</span>
                    <span class="rounded bg-theme-hover px-1.5 py-0.5 text-xs font-mono text-theme-textLight">{{ skill.id }}</span>
                  </div>
                  <p v-if="skill.description" class="mt-1 text-sm text-theme-textLight">{{ skill.description }}</p>
                  <div v-if="skill.tags.length" class="mt-2 flex flex-wrap gap-1.5">
                    <span
                      v-for="tag in skill.tags"
                      :key="tag"
                      class="rounded-full border border-theme-border px-2 py-0.5 text-xs text-theme-textLight"
                    >
                      {{ tag }}
                    </span>
                  </div>
                </li>
              </ul>
              <p v-else class="p-8 text-center text-sm text-theme-textLight">{{ $t('agents_page.no_skills') }}</p>
            </section>
          </div>

          <div class="min-w-0 space-y-6">
            <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
              <div class="border-b border-theme-border px-5 py-4">
                <h2 class="text-lg font-semibold text-theme-text">{{ $t('agents_page.card') }}</h2>
                <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.card_desc') }}</p>
              </div>
              <div v-if="!agent.card" class="p-5">
                <div class="flex items-start gap-3 rounded-lg border border-dashed border-theme-border p-4">
                  <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-theme-hover text-theme-textLight">
                    <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6l4 2m6-2a10 10 0 11-20 0 10 10 0 0120 0z" />
                    </svg>
                  </span>
                  <div class="min-w-0">
                    <p class="text-sm font-medium text-theme-text">{{ $t('agents_page.card_pending') }}</p>
                    <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.card_pending_desc') }}</p>
                  </div>
                </div>
                <dl class="mt-4 text-sm">
                  <dt class="text-xs font-medium uppercase tracking-wider text-theme-textLight">{{ $t('agents_page.card_path') }}</dt>
                  <dd class="mt-1.5"><code class="block break-all rounded-md border border-theme-border bg-theme-hover px-3 py-2 text-xs font-mono text-theme-text">{{ agent.cardPath }}</code></dd>
                </dl>
              </div>
              <dl v-if="agent.card" class="space-y-4 px-5 py-4 text-sm">
                <div v-if="card.url">
                  <dt class="text-theme-textLight">{{ $t('agents_page.gateway_url') }}</dt>
                  <dd class="mt-1 flex items-start gap-2">
                    <code class="min-w-0 flex-1 break-all rounded bg-theme-hover px-2 py-1 text-xs font-mono text-theme-text">{{ card.url }}</code>
                    <button type="button" class="shrink-0 text-xs font-medium text-primary-600 hover:text-primary-700" @click="copyUrl">
                      {{ urlCopied ? $t('agents_page.copied') : $t('agents_page.copy') }}
                    </button>
                  </dd>
                </div>
                <div v-if="card.provider">
                  <dt class="text-theme-textLight">{{ $t('agents_page.provider') }}</dt>
                  <dd class="mt-1 text-theme-text">{{ card.provider }}</dd>
                </div>
                <div>
                  <dt class="text-theme-textLight">{{ $t('agents_page.input_modes') }}</dt>
                  <dd class="mt-1 flex flex-wrap gap-1.5">
                    <span v-for="mode in card.inputModes" :key="mode" class="rounded bg-theme-hover px-1.5 py-0.5 text-xs font-mono text-theme-text">{{ mode }}</span>
                    <span v-if="!card.inputModes.length" class="text-theme-textLight">—</span>
                  </dd>
                </div>
                <div>
                  <dt class="text-theme-textLight">{{ $t('agents_page.output_modes') }}</dt>
                  <dd class="mt-1 flex flex-wrap gap-1.5">
                    <span v-for="mode in card.outputModes" :key="mode" class="rounded bg-theme-hover px-1.5 py-0.5 text-xs font-mono text-theme-text">{{ mode }}</span>
                    <span v-if="!card.outputModes.length" class="text-theme-textLight">—</span>
                  </dd>
                </div>
                <div>
                  <dt class="text-theme-textLight">{{ $t('agents_page.capabilities') }}</dt>
                  <dd class="mt-1 flex flex-wrap gap-1.5">
                    <span
                      v-for="capability in card.capabilities"
                      :key="capability"
                      class="rounded-full bg-primary-50 px-2 py-0.5 text-xs font-medium text-primary-700"
                    >
                      {{ capability }}
                    </span>
                    <span v-if="!card.capabilities.length" class="text-theme-textLight">—</span>
                  </dd>
                </div>
                <div>
                  <dt class="text-theme-textLight">{{ $t('agents_page.card_path') }}</dt>
                  <dd class="mt-1"><code class="break-all rounded bg-theme-hover px-2 py-1 text-xs font-mono text-theme-text">{{ agent.cardPath }}</code></dd>
                </div>
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <dt class="text-theme-textLight">{{ $t('agents_page.updated') }}</dt>
                    <dd class="mt-1 text-theme-text">{{ formatRelative(agent.updatedAt) }}</dd>
                  </div>
                  <div class="min-w-0">
                    <dt class="text-theme-textLight">{{ $t('agents_page.checksum') }}</dt>
                    <dd class="mt-1 truncate text-xs font-mono text-theme-textLight" :title="agent.cardChecksum">{{ agent.cardChecksum }}</dd>
                  </div>
                </div>
              </dl>
              <details v-if="agent.card" class="border-t border-theme-border">
                <summary class="cursor-pointer px-5 py-3 text-sm font-medium text-theme-text hover:bg-theme-hover">
                  {{ $t('agents_page.raw_card') }}
                </summary>
                <pre class="max-h-96 overflow-auto bg-theme-hover px-5 py-4 text-xs leading-5 font-mono text-theme-text">{{ rawCard }}</pre>
              </details>
            </section>

            <section id="protection" class="scroll-mt-6 rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
              <div class="flex flex-wrap items-start justify-between gap-4 border-b border-theme-border px-5 py-4">
                <div>
                  <h2 class="text-lg font-semibold text-theme-text">{{ $t('agents_page.protection_title') }}</h2>
                  <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.protection_desc') }}</p>
                </div>
                <button v-if="canManage && protection" type="button" class="btn-secondary text-sm" @click="protectionModalOpen = true">
                  {{ protection ? $t('agents_page.protection_replace') : $t('agents_page.protection_add') }}
                </button>
              </div>

              <div v-if="newProtectionSecret" class="border-b border-amber-200 bg-amber-50 px-5 py-4 space-y-2">
                <p class="text-sm font-medium text-amber-900">{{ $t('agents_page.protection_secret_title') }}</p>
                <div class="flex flex-wrap items-center gap-2">
                  <code class="flex-1 min-w-0 px-3 py-2 bg-white border border-amber-200 rounded text-sm text-theme-text break-all font-mono">{{ newProtectionSecret }}</code>
                  <button type="button" class="btn-secondary shrink-0" @click="copyProtectionSecret">
                    {{ protectionSecretCopied ? $t('agents_page.copied') : $t('agents_page.copy') }}
                  </button>
                </div>
                <button
                  type="button"
                  class="text-sm text-amber-800 hover:text-amber-900 font-medium"
                  @click="newProtectionSecret = null; protectionSecretCopied = false"
                >
                  {{ $t('common.close') }}
                </button>
              </div>

              <div v-if="loadingProtection && !protection" class="p-8 text-center text-sm text-theme-textLight">
                {{ $t('agents_page.loading_protection') }}
              </div>
              <div v-else-if="!protection" class="p-5">
                <div class="flex items-start gap-3 rounded-lg border border-dashed border-theme-border p-4">
                  <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-theme-hover text-theme-textLight">
                    <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M12 3l7 3v6c0 4.5-3 8-7 9-4-1-7-4.5-7-9V6l7-3z" />
                    </svg>
                  </span>
                  <div class="min-w-0">
                    <p class="text-sm font-medium text-theme-text">{{ $t('agents_page.no_protection') }}</p>
                    <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.no_protection_hint') }}</p>
                    <button v-if="canManage" type="button" class="btn-secondary mt-3 text-sm" @click="protectionModalOpen = true">
                      {{ $t('agents_page.protection_add') }}
                    </button>
                  </div>
                </div>
              </div>
              <div v-else class="space-y-4 px-5 py-4">
                <dl class="grid grid-cols-2 gap-x-6 gap-y-3 text-sm">
                  <div>
                    <dt class="text-theme-textLight">{{ $t('agents_page.protection_type') }}</dt>
                    <dd class="text-theme-text">{{ $t(`agents_page.protection_type_${protection.type}`) }}</dd>
                  </div>
                  <div>
                    <dt class="text-theme-textLight">{{ $t('agents_page.protection_status') }}</dt>
                    <dd>
                      <span
                        class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
                        :class="protection.isActive ? 'bg-green-50 text-green-700' : 'bg-theme-hover text-theme-textLight'"
                      >
                        {{ protection.isActive ? $t('agents_page.protection_active') : $t('agents_page.protection_inactive') }}
                      </span>
                    </dd>
                  </div>
                  <div v-if="protection.header">
                    <dt class="text-theme-textLight">{{ $t('agents_page.protection_header') }}</dt>
                    <dd class="text-theme-text font-mono">{{ protection.header }}</dd>
                  </div>
                  <div v-if="protection.username">
                    <dt class="text-theme-textLight">{{ $t('agents_page.protection_username') }}</dt>
                    <dd class="text-theme-text font-mono">{{ protection.username }}</dd>
                  </div>
                  <div>
                    <dt class="text-theme-textLight">{{ $t('agents_page.protection_updated') }}</dt>
                    <dd class="text-theme-text">{{ formatRelative(protection.updatedAt) }}</dd>
                  </div>
                </dl>
                <div v-if="canManage" class="flex items-center gap-3 border-t border-theme-border pt-3 text-sm">
                  <button
                    type="button"
                    class="text-primary-600 hover:text-primary-700 hover:underline disabled:opacity-50"
                    :disabled="togglingProtection"
                    @click="toggleProtection"
                  >
                    {{ protection.isActive ? $t('agents_page.protection_disable') : $t('agents_page.protection_enable') }}
                  </button>
                  <button
                    type="button"
                    class="text-red-600 hover:text-red-700 hover:underline"
                    @click="protectionDeleteOpen = true; protectionDeleteError = null"
                  >
                    {{ $t('agents_page.protection_remove') }}
                  </button>
                </div>
              </div>
            </section>

            <section id="identities" class="scroll-mt-6 rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
              <div class="flex flex-wrap items-start justify-between gap-4 border-b border-theme-border px-5 py-4">
                <div>
                  <h2 class="text-lg font-semibold text-theme-text">{{ $t('agents_page.identities_title') }}</h2>
                  <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.identities_desc') }}</p>
                </div>
                <button v-if="canManage && identities.length" type="button" class="btn-secondary text-sm" @click="identityModalOpen = true">
                  {{ $t('agents_page.identity_add') }}
                </button>
              </div>

              <div v-if="newIdentity" class="border-b border-amber-200 bg-amber-50 px-5 py-4 space-y-2">
                <p class="text-sm font-medium text-amber-900">{{ $t('agents_page.identity_secret_title', { name: newIdentity.name }) }}</p>
                <p v-if="newIdentity.username" class="text-sm text-amber-900">
                  {{ $t('agents_page.identity_username') }}: <code class="font-mono">{{ newIdentity.username }}</code>
                </p>
                <p v-if="newIdentity.type === 'client_credentials'" class="text-sm text-amber-900">
                  {{ $t('agents_page.identity_client_id') }}: <code class="font-mono break-all">{{ newIdentity.id }}</code>
                </p>
                <div class="flex flex-wrap items-center gap-2">
                  <code class="flex-1 min-w-0 px-3 py-2 bg-white border border-amber-200 rounded text-sm text-theme-text break-all font-mono">{{ newIdentity.secret }}</code>
                  <button type="button" class="btn-secondary shrink-0" @click="copySecret">
                    {{ secretCopied ? $t('agents_page.copied') : $t('agents_page.copy') }}
                  </button>
                </div>
                <button
                  type="button"
                  class="text-sm text-amber-800 hover:text-amber-900 font-medium"
                  @click="newIdentity = null; secretCopied = false"
                >
                  {{ $t('common.close') }}
                </button>
              </div>

              <div v-if="loadingIdentities && identities.length === 0" class="p-8 text-center text-sm text-theme-textLight">
                {{ $t('agents_page.loading_identities') }}
              </div>
              <div v-else-if="identities.length === 0" class="p-5">
                <div class="flex items-start gap-3 rounded-lg border border-dashed border-theme-border p-4">
                  <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-theme-hover text-theme-textLight">
                    <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                    </svg>
                  </span>
                  <div class="min-w-0">
                    <p class="text-sm font-medium text-theme-text">{{ $t('agents_page.no_identities') }}</p>
                    <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.no_identities_hint') }}</p>
                    <button v-if="canManage" type="button" class="btn-secondary mt-3 text-sm" @click="identityModalOpen = true">
                      {{ $t('agents_page.identity_add') }}
                    </button>
                  </div>
                </div>
              </div>
              <ul v-else class="divide-y divide-theme-border">
                <li v-for="identity in identities" :key="identity.id" class="px-5 py-3">
                  <div class="flex items-start justify-between gap-3">
                    <div class="min-w-0">
                      <p class="truncate text-sm font-medium text-theme-text font-mono">{{ identity.name }}</p>
                      <p class="mt-0.5 text-xs text-theme-textLight">
                        {{ $t(`agents_page.identity_type_${identity.type}`) }}
                        <span v-if="identity.username" class="font-mono">· {{ identity.username }}</span>
                        <span v-if="identity.tokenTtl">· {{ $t('agents_page.identity_token_ttl_value', { seconds: identity.tokenTtl }) }}</span>
                        · {{ formatRelative(identity.createdAt) }}
                      </p>
                    </div>
                    <span
                      class="inline-flex shrink-0 items-center rounded-full px-2 py-0.5 text-xs font-medium"
                      :class="identity.isActive ? 'bg-emerald-50 text-emerald-800' : 'bg-theme-hover text-theme-textLight'"
                    >
                      {{ identity.isActive ? $t('agents_page.identity_active') : $t('agents_page.identity_inactive') }}
                    </span>
                  </div>
                  <div v-if="canManage" class="mt-2 flex items-center gap-3 text-xs">
                    <button
                      type="button"
                      class="text-primary-600 hover:text-primary-700 hover:underline disabled:opacity-50"
                      :disabled="togglingIdentity === identity.id"
                      @click="toggleIdentity(identity)"
                    >
                      {{ identity.isActive ? $t('agents_page.identity_disable') : $t('agents_page.identity_enable') }}
                    </button>
                    <button
                      type="button"
                      class="text-red-600 hover:text-red-700 hover:underline"
                      @click="askDeleteIdentity(identity)"
                    >
                      {{ $t('agents_page.identity_revoke') }}
                    </button>
                  </div>
                </li>
              </ul>
            </section>
          </div>
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

    <AgentProtectionModal
      v-if="agent"
      :open="protectionModalOpen"
      :workspace-id="currentWorkspace.id"
      :agent-name="agent.name"
      :current="protection"
      @close="protectionModalOpen = false"
      @saved="onProtectionSaved"
    />

    <AgentIdentityModal
      v-if="agent"
      :open="identityModalOpen"
      :workspace-id="currentWorkspace.id"
      :agent-name="agent.name"
      @close="identityModalOpen = false"
      @saved="onIdentitySaved"
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
          v-show="deregisterTarget"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="deregister-instance-title"
          @click.self="closeDeregister"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-sm p-6">
            <h2 id="deregister-instance-title" class="text-lg font-semibold text-theme-text">
              {{ $t('agents_page.deregister_title') }}
            </h2>
            <div v-if="deregisterError" class="mt-3 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800" role="alert">
              {{ deregisterError }}
            </div>
            <p class="mt-2 text-sm text-theme-textLight">
              {{ $t('agents_page.deregister_confirm', { id: deregisterTarget?.instanceId }) }}
            </p>
            <div class="mt-6 flex gap-3 justify-end">
              <button type="button" class="btn-secondary" :disabled="deregistering" @click="closeDeregister">
                {{ $t('common.cancel') }}
              </button>
              <button
                type="button"
                class="btn-primary bg-red-600 hover:bg-red-700 focus:ring-red-500"
                :disabled="deregistering"
                @click="confirmDeregister"
              >
                {{ deregistering ? $t('agents_page.deregistering') : $t('agents_page.deregister') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="checkDeleteTarget"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="delete-check-title"
          @click.self="closeDeleteCheck"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-sm p-6">
            <h2 id="delete-check-title" class="text-lg font-semibold text-theme-text">
              {{ $t('agents_page.check_delete_title') }}
            </h2>
            <div v-if="checkDeleteError" class="mt-3 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800" role="alert">
              {{ checkDeleteError }}
            </div>
            <p class="mt-2 text-sm text-theme-textLight">
              {{ $t('agents_page.check_delete_confirm', { name: checkDeleteTarget?.name }) }}
            </p>
            <div class="mt-6 flex gap-3 justify-end">
              <button type="button" class="btn-secondary" :disabled="checkDeleting" @click="closeDeleteCheck">
                {{ $t('common.cancel') }}
              </button>
              <button
                type="button"
                class="btn-primary bg-red-600 hover:bg-red-700 focus:ring-red-500"
                :disabled="checkDeleting"
                @click="confirmDeleteCheck"
              >
                {{ checkDeleting ? $t('agents_page.deleting') : $t('agents_page.delete') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="protectionDeleteOpen"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="delete-protection-title"
          @click.self="closeDeleteProtection"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-sm p-6">
            <h2 id="delete-protection-title" class="text-lg font-semibold text-theme-text">
              {{ $t('agents_page.protection_delete_title') }}
            </h2>
            <div v-if="protectionDeleteError" class="mt-3 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800" role="alert">
              {{ protectionDeleteError }}
            </div>
            <p class="mt-2 text-sm text-theme-textLight">
              {{ $t('agents_page.protection_delete_confirm') }}
            </p>
            <div class="mt-6 flex gap-3 justify-end">
              <button type="button" class="btn-secondary" :disabled="protectionDeleting" @click="closeDeleteProtection">
                {{ $t('common.cancel') }}
              </button>
              <button
                type="button"
                class="btn-primary bg-red-600 hover:bg-red-700 focus:ring-red-500"
                :disabled="protectionDeleting"
                @click="confirmDeleteProtection"
              >
                {{ protectionDeleting ? $t('agents_page.deleting') : $t('agents_page.protection_remove') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="identityDeleteTarget"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="delete-identity-title"
          @click.self="closeDeleteIdentity"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-sm p-6">
            <h2 id="delete-identity-title" class="text-lg font-semibold text-theme-text">
              {{ $t('agents_page.identity_delete_title') }}
            </h2>
            <div v-if="identityDeleteError" class="mt-3 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800" role="alert">
              {{ identityDeleteError }}
            </div>
            <p class="mt-2 text-sm text-theme-textLight">
              {{ $t('agents_page.identity_delete_confirm', { name: identityDeleteTarget?.name }) }}
            </p>
            <div class="mt-6 flex gap-3 justify-end">
              <button type="button" class="btn-secondary" :disabled="identityDeleting" @click="closeDeleteIdentity">
                {{ $t('common.cancel') }}
              </button>
              <button
                type="button"
                class="btn-primary bg-red-600 hover:bg-red-700 focus:ring-red-500"
                :disabled="identityDeleting"
                @click="confirmDeleteIdentity"
              >
                {{ identityDeleting ? $t('agents_page.deleting') : $t('agents_page.identity_revoke') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

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
import StatusBreakdown from '@/components/StatusBreakdown.vue'
import { computed, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppNav from '@/components/AppNav.vue'
import AgentCardModal from '@/components/AgentCardModal.vue'
import AgentCheckModal from '@/components/AgentCheckModal.vue'
import AgentIdentityModal from '@/components/AgentIdentityModal.vue'
import AgentProtectionModal from '@/components/AgentProtectionModal.vue'
import { agent_api, agent_check_api, agent_identity_api, agent_instance_api, agent_protection_api } from '@/api'
import { showFlash } from '@/lib/flash'
import { useWorkspaceContext } from '@/lib/permission'
import { countByStatus, healthClass, healthDotClass, parseCard } from '@/lib/agent'
import { formatRelative } from '@/utils/date'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { currentWorkspace, canManage } = useWorkspaceContext()

const agent = ref(null)
const instances = ref([])
const instancesLimit = 20
const instancesOffset = ref(0)
const instancesTotal = ref(0)
const loading = ref(false)
const loadingInstances = ref(false)
const notFound = ref(false)
const errorMessage = ref(null)
const deregisterTarget = ref(null)
const deregistering = ref(false)
const deregisterError = ref(null)

const checks = ref([])
const loadingChecks = ref(false)
const checkModalOpen = ref(false)
const editingCheck = ref(null)
const checkDeleteTarget = ref(null)
const checkDeleting = ref(false)
const checkDeleteError = ref(null)
const expanded = reactive({})
const urlCopied = ref(false)

const protection = ref(null)
const loadingProtection = ref(false)
const protectionModalOpen = ref(false)
const newProtectionSecret = ref(null)
const protectionSecretCopied = ref(false)
const togglingProtection = ref(false)
const protectionDeleteOpen = ref(false)
const protectionDeleting = ref(false)
const protectionDeleteError = ref(null)

const identities = ref([])
const loadingIdentities = ref(false)
const identityModalOpen = ref(false)
const newIdentity = ref(null)
const secretCopied = ref(false)
const togglingIdentity = ref(null)
const identityDeleteTarget = ref(null)
const identityDeleting = ref(false)
const identityDeleteError = ref(null)

const showEditModal = ref(false)
const showDeleteModal = ref(false)
const deleteConfirmName = ref('')
const deleting = ref(false)
const deleteError = ref(null)

const agentName = computed(() => String(route.params.name || ''))
const card = computed(() => parseCard(agent.value?.card))
const rawCard = computed(() => JSON.stringify(agent.value?.card ?? {}, null, 2))
const checkTypesSummary = computed(() => {
  const types = [...new Set(checks.value.map((check) => t(`agents_page.check_type_${check.type}`)))]
  return types.length ? types.join(', ') : t('agents_page.no_checks_short')
})

async function loadAgent() {
  loading.value = true
  notFound.value = false
  errorMessage.value = null
  agent.value = null
  instances.value = []
  instancesOffset.value = 0
  checks.value = []
  protection.value = null
  newProtectionSecret.value = null
  identities.value = []
  newIdentity.value = null

  try {
    const res = await agent_api.get(currentWorkspace.id, agentName.value)
    agent.value = res.data
    await Promise.all([loadChecks(), loadProtection(), loadIdentities(), loadInstances()])
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
    const res = await agent_instance_api.list(currentWorkspace.id, agentName.value, {
      limit: instancesLimit,
      offset: instancesOffset.value,
    })
    instances.value = res.data.instances || []
    instancesTotal.value = res.data._meta?.total || 0

    // The page emptied after a deregister, so step back to the last page with rows
    if (!instances.value.length && instancesOffset.value > 0) {
      instancesOffset.value = Math.max(0, instancesOffset.value - instancesLimit)
      await loadInstances()
    }
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

async function loadProtection() {
  loadingProtection.value = true

  try {
    const res = await agent_protection_api.get(currentWorkspace.id, agentName.value)
    protection.value = res.data
  } catch (err) {
    if (err.response?.status === 404) {
      protection.value = null
      return
    }
    errorMessage.value = err.response?.data?.errorMessage || t('agents_page.failed_load_protection')
  } finally {
    loadingProtection.value = false
  }
}

function onProtectionSaved(saved) {
  protectionModalOpen.value = false
  newProtectionSecret.value = saved.secret || null
  protectionSecretCopied.value = false
  protection.value = { ...saved, secret: undefined }
  showFlash(t('common.saved'))
}

function copyProtectionSecret() {
  navigator.clipboard.writeText(newProtectionSecret.value).then(() => {
    protectionSecretCopied.value = true
    setTimeout(() => { protectionSecretCopied.value = false }, 2000)
  })
}

async function toggleProtection() {
  togglingProtection.value = true

  try {
    const res = await agent_protection_api.updateStatus(currentWorkspace.id, agentName.value, {
      isActive: !protection.value.isActive,
    })
    protection.value = res.data
    showFlash(t('common.saved'))
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('agents_page.protection_failed_save')
  } finally {
    togglingProtection.value = false
  }
}

function closeDeleteProtection() {
  protectionDeleteOpen.value = false
}

async function confirmDeleteProtection() {
  protectionDeleting.value = true
  protectionDeleteError.value = null

  try {
    await agent_protection_api.delete(currentWorkspace.id, agentName.value)
    protection.value = null
    newProtectionSecret.value = null
    protectionDeleteOpen.value = false
    showFlash(t('common.deleted'))
  } catch (err) {
    protectionDeleteError.value = err.response?.data?.errorMessage || t('agents_page.protection_failed_delete')
  } finally {
    protectionDeleting.value = false
  }
}

async function loadIdentities() {
  loadingIdentities.value = true

  try {
    const res = await agent_identity_api.list(currentWorkspace.id, agentName.value)
    identities.value = res.data.identities || []
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('agents_page.failed_load_identities')
  } finally {
    loadingIdentities.value = false
  }
}

async function onIdentitySaved(identity) {
  identityModalOpen.value = false
  newIdentity.value = identity
  secretCopied.value = false
  await loadIdentities()
}

function copySecret() {
  navigator.clipboard.writeText(newIdentity.value.secret).then(() => {
    secretCopied.value = true
    setTimeout(() => { secretCopied.value = false }, 2000)
  })
}

async function toggleIdentity(identity) {
  togglingIdentity.value = identity.id

  try {
    const res = await agent_identity_api.update(currentWorkspace.id, agentName.value, identity.id, {
      name: identity.name,
      isActive: !identity.isActive,
    })
    Object.assign(identity, res.data)
    showFlash(t('common.saved'))
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('agents_page.identity_failed_save')
  } finally {
    togglingIdentity.value = null
  }
}

function askDeleteIdentity(identity) {
  identityDeleteTarget.value = identity
  identityDeleteError.value = null
}

function closeDeleteIdentity() {
  identityDeleteTarget.value = null
}

async function confirmDeleteIdentity() {
  identityDeleting.value = true
  identityDeleteError.value = null

  try {
    await agent_identity_api.delete(currentWorkspace.id, agentName.value, identityDeleteTarget.value.id)
    if (newIdentity.value?.id === identityDeleteTarget.value.id) newIdentity.value = null
    identityDeleteTarget.value = null
    showFlash(t('common.deleted'))
    await loadIdentities()
  } catch (err) {
    identityDeleteError.value = err.response?.data?.errorMessage || t('agents_page.identity_failed_delete')
  } finally {
    identityDeleting.value = false
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

function askDeleteCheck(check) {
  checkDeleteTarget.value = check
  checkDeleteError.value = null
}

function closeDeleteCheck() {
  checkDeleteTarget.value = null
}

async function confirmDeleteCheck() {
  checkDeleting.value = true
  checkDeleteError.value = null

  try {
    await agent_check_api.delete(currentWorkspace.id, agentName.value, checkDeleteTarget.value.checkId)
    checkDeleteTarget.value = null
    showFlash(t('common.deleted'))
    await refreshHealth()
  } catch (err) {
    checkDeleteError.value = err.response?.data?.errorMessage || t('agents_page.check_failed_delete')
  } finally {
    checkDeleting.value = false
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

function copyUrl() {
  navigator.clipboard.writeText(card.value.url).then(() => {
    urlCopied.value = true
    setTimeout(() => { urlCopied.value = false }, 2000)
  })
}

function goToInstancesPage(newOffset) {
  if (newOffset < 0 || newOffset >= instancesTotal.value) return
  instancesOffset.value = newOffset
  loadInstances()
}

function toggleInstance(id) {
  expanded[id] = !expanded[id]
}

function onCardSaved(updated) {
  showEditModal.value = false
  agent.value = updated
  showFlash(t('common.saved'))
}

function askDeregister(instance) {
  deregisterTarget.value = instance
  deregisterError.value = null
}

function closeDeregister() {
  deregisterTarget.value = null
}

async function confirmDeregister() {
  deregistering.value = true
  deregisterError.value = null

  try {
    await agent_instance_api.delete(currentWorkspace.id, agentName.value, deregisterTarget.value.instanceId)
    deregisterTarget.value = null
    // Refresh the agent too, since its health is derived from the remaining instances
    const res = await agent_api.get(currentWorkspace.id, agentName.value)
    agent.value = res.data
    await loadInstances()
    showFlash(t('common.deleted'))
  } catch (err) {
    deregisterError.value = err.response?.data?.errorMessage || t('agents_page.failed_deregister')
  } finally {
    deregistering.value = false
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
