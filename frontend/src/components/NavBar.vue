<template>
  <nav class="border-b border-gray-800 bg-gray-950/80 backdrop-blur sticky top-0 z-50">
    <div class="max-w-6xl mx-auto px-4 h-14 flex items-center justify-between">
      <router-link to="/" class="flex items-center gap-2 font-bold text-lg text-white">
        <span class="text-brand-500">⚡</span> Defiver
      </router-link>

      <div class="flex items-center gap-4 text-sm">
        <router-link to="/tasks" class="text-gray-400 hover:text-white transition-colors">Browse Tasks</router-link>
        <router-link to="/agents" class="text-gray-400 hover:text-white transition-colors">Agent Portal</router-link>

        <template v-if="auth.isLoggedIn">
          <router-link to="/dashboard" class="text-gray-400 hover:text-white transition-colors">Dashboard</router-link>
          <span class="text-gray-600">|</span>
          <span class="text-gray-300">{{ auth.displayName }}</span>
          <button @click="auth.logout(); $router.push('/')" class="btn-secondary text-xs px-3 py-1.5">Sign out</button>
        </template>
        <template v-else>
          <router-link to="/login"    class="text-gray-400 hover:text-white transition-colors">Sign in</router-link>
          <router-link to="/register" class="btn-primary text-sm px-3 py-1.5">Sign up</router-link>
        </template>
      </div>
    </div>
  </nav>
</template>

<script setup>
import { useAuthStore } from '../stores/auth.js'
const auth = useAuthStore()
</script>
