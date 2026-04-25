<template>
  <div class="min-h-[70vh] flex items-center justify-center px-4">
    <div class="w-full max-w-sm">
      <h1 class="text-2xl font-bold text-white text-center mb-2">Welcome back</h1>
      <p class="text-gray-400 text-center text-sm mb-8">Sign in to your Defiver account</p>

      <form @submit.prevent="submit" class="space-y-4">
        <div>
          <label class="block text-sm text-gray-400 mb-1">Email</label>
          <input v-model="form.email" type="email" class="input" placeholder="you@example.com" required />
        </div>
        <div>
          <label class="block text-sm text-gray-400 mb-1">Password</label>
          <input v-model="form.password" type="password" class="input" placeholder="••••••••" required />
        </div>

        <div v-if="error" class="text-red-400 text-sm bg-red-900/20 rounded-lg px-3 py-2">{{ error }}</div>

        <button type="submit" :disabled="loading" class="btn-primary w-full py-2.5">
          {{ loading ? 'Signing in…' : 'Sign in' }}
        </button>
      </form>

      <p class="text-center text-gray-500 text-sm mt-6">
        Don't have an account?
        <router-link to="/register" class="text-brand-500 hover:text-brand-400">Sign up</router-link>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import api from '../api.js'
import { useAuthStore } from '../stores/auth.js'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const form = ref({ email: '', password: '' })
const loading = ref(false)
const error = ref('')

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const res = await api.post('/users/login', form.value)
    auth.setAuth(res.data)
    router.push(route.query.redirect || '/dashboard')
  } catch (e) {
    error.value = e.response?.data?.error || 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>
