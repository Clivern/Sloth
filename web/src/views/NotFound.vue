<template>
  <div class="min-h-screen flex items-center justify-center bg-theme-bg px-4">
    <div class="max-w-md w-full text-center">
      <div class="mb-8">
        <router-link to="/">
          <AppLogo :alt="$t('common.logo_alt')" class="h-24 w-auto mx-auto text-theme-text" />
        </router-link>
      </div>
      <p class="text-6xl font-semibold text-theme-textLight mb-2">404</p>
      <h1 class="text-xl font-semibold text-theme-text mb-2">{{ $t('not_found.title') }}</h1>
      <p class="text-sm text-theme-textLight mb-6">
        {{ $t('not_found.description') }}
      </p>
      <div class="flex flex-col sm:flex-row gap-3 justify-center">
        <router-link to="/dashboard" class="btn-primary">
          {{ $t('not_found.go_to_dashboard') }}
        </router-link>
        <router-link to="/login" class="btn-secondary">
          {{ $t('not_found.sign_in') }}
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import AppLogo from '@/comp/AppLogo.vue'
</script>
