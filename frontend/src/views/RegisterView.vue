<template>
  <div class="min-h-[70vh] flex items-center justify-center px-4">
    <div class="w-full max-w-sm">
      <h1 class="text-2xl font-bold text-white text-center mb-2">Create account</h1>
      <p class="text-gray-400 text-center text-sm mb-8">Join Defiver and start earning</p>

      <form @submit.prevent="submit" class="space-y-4">
        <div>
          <label class="block text-sm text-gray-400 mb-1">Display name</label>
          <input v-model="form.display_name" type="text" class="input" placeholder="Your name or handle" required />
        </div>
        <div>
          <label class="block text-sm text-gray-400 mb-1">Email</label>
          <input v-model="form.email" type="email" class="input" placeholder="you@example.com" required />
        </div>
        <div>
          <label class="block text-sm text-gray-400 mb-1">Password</label>
          <input v-model="form.password" type="password" class="input" placeholder="At least 8 characters" minlength="8" required />
        </div>

        <div v-if="error" class="text-red-400 text-sm bg-red-900/20 rounded-lg px-3 py-2">{{ error }}</div>

        <button type="submit" :disabled="loading" class="btn-primary w-full py-2.5">
          {{ loading ? 'Creating account…' : 'Create account' }}
        </button>
      </form>

      <p class="text-center text-gray-500 text-sm mt-6">
        Already have an account?
        <router-link to="/login" class="text-brand-500 hover:text-brand-400">Sign in</router-link>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '../api.js'
import { useAuthStore } from '../stores/auth.js'

const router = useRouter()
const auth = useAuthStore()
const form = ref({ display_name: '', email: '', password: '' })
const loading = ref(false)
const error = ref('')

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const res = await api.post('/users/register', form.value)
    auth.setAuth(res.data)
    router.push('/dashboard')
  } catch (e) {
    error.value = e.response?.data?.error || 'Registration failed'
  } finally {
    loading.value = false
  }
}
</script>
