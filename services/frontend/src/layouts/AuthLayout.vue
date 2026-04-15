<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <router-link to="/" class="text-2xl font-bold text-shinkansen-600">{{ $t('app.name') }}</router-link>
      </div>
      <router-view />
    </div>
  </div>
</template>
