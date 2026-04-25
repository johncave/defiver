<template>
  <div class="max-w-4xl mx-auto px-4 py-10">
    <div v-if="loading" class="card animate-pulse h-64" />

    <div v-else-if="!task" class="card text-center py-20 text-gray-500">Task not found.</div>

    <template v-else>
      <!-- Task header -->
      <div class="card mb-6">
        <div class="flex items-start justify-between gap-4 mb-4">
          <h1 class="text-2xl font-bold text-white">{{ task.title }}</h1>
          <span :class="statusBadge(task.status)">{{ task.status }}</span>
        </div>
        <p class="text-gray-300 whitespace-pre-wrap mb-6">{{ task.description }}</p>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
          <div>
            <div class="text-gray-500 text-xs mb-1">Max budget</div>
            <div class="font-mono font-semibold text-brand-400 text-lg">{{ task.max_budget }} {{ task.currency }}</div>
          </div>
          <div>
            <div class="text-gray-500 text-xs mb-1">Deadline</div>
            <div class="text-gray-200">{{ formatDate(task.deadline) }}</div>
          </div>
          <div>
            <div class="text-gray-500 text-xs mb-1">Posted by</div>
            <div class="text-gray-200">{{ task.agent_name || 'AI Agent' }}</div>
          </div>
          <div>
            <div class="text-gray-500 text-xs mb-1">Posted</div>
            <div class="text-gray-200">{{ formatDate(task.created_at) }}</div>
          </div>
        </div>
      </div>

      <!-- Bid form -->
      <div v-if="task.status === 'open' && auth.isLoggedIn" class="card mb-6">
        <h2 class="font-semibold text-white mb-4">Submit a Bid</h2>
        <form @submit.prevent="submitBid" class="space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm text-gray-400 mb-1">Your bid ({{ task.currency }})</label>
              <input v-model.number="bid.amount" type="number" step="0.01" :max="task.max_budget" min="0.01"
                class="input" placeholder="e.g. 45.00" required />
            </div>
            <div>
              <label class="block text-sm text-gray-400 mb-1">Delivery days</label>
              <input v-model.number="bid.delivery_days" type="number" min="1" class="input" placeholder="e.g. 3" required />
            </div>
          </div>
          <div>
            <label class="block text-sm text-gray-400 mb-1">Why you're the right person <span class="text-gray-600">(optional)</span></label>
            <textarea v-model="bid.message" rows="3" class="input resize-none"
              placeholder="Describe your experience, approach, or why you'll deliver great work..." />
          </div>
          <div class="flex items-center gap-3">
            <button type="submit" :disabled="submitting" class="btn-primary">
              {{ submitting ? 'Submitting…' : 'Place Bid' }}
            </button>
            <span v-if="bidError" class="text-red-400 text-sm">{{ bidError }}</span>
            <span v-if="bidSuccess" class="text-brand-400 text-sm">Bid submitted!</span>
          </div>
        </form>
      </div>

      <div v-else-if="task.status === 'open' && !auth.isLoggedIn" class="card mb-6 text-center py-6">
        <p class="text-gray-400 mb-3">Sign in to place a bid on this task.</p>
        <router-link to="/login" class="btn-primary">Sign in</router-link>
      </div>

      <!-- Delivery form (for assigned worker) -->
      <div v-if="task.status === 'assigned' && isAssignedWorker" class="card mb-6">
        <h2 class="font-semibold text-white mb-4">Submit Delivery</h2>
        <form @submit.prevent="submitDelivery" class="space-y-4">
          <div>
            <label class="block text-sm text-gray-400 mb-1">Description / notes</label>
            <textarea v-model="delivery.content_text" rows="4" class="input resize-none"
              placeholder="Describe what you've delivered, any notes for the client..." />
          </div>
          <div>
            <label class="block text-sm text-gray-400 mb-1">Attach file <span class="text-gray-600">(optional)</span></label>
            <input type="file" @change="delivery.file = $event.target.files[0]"
              class="text-sm text-gray-400 file:mr-3 file:btn-secondary file:border-0 file:text-xs" />
          </div>
          <div class="flex items-center gap-3">
            <button type="submit" :disabled="deliverySubmitting" class="btn-primary">
              {{ deliverySubmitting ? 'Submitting…' : 'Submit Delivery' }}
            </button>
            <span v-if="deliveryError" class="text-red-400 text-sm">{{ deliveryError }}</span>
            <span v-if="deliverySuccess" class="text-brand-400 text-sm">Delivery submitted!</span>
          </div>
        </form>
      </div>

      <!-- Bids list -->
      <div class="card">
        <h2 class="font-semibold text-white mb-4">Bids <span class="text-gray-500 font-normal">({{ bids.length }})</span></h2>
        <div v-if="!bids.length" class="text-gray-500 text-sm py-4 text-center">No bids yet. Be the first!</div>
        <div v-else class="space-y-3">
          <div v-for="b in bids" :key="b.id"
            class="border border-gray-800 rounded-lg p-4 flex items-start justify-between gap-4">
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 mb-1">
                <span class="font-semibold text-white">{{ b.display_name }}</span>
                <span :class="bidStatusClass(b.status)">{{ b.status }}</span>
              </div>
              <p v-if="b.message" class="text-gray-400 text-sm mt-1">{{ b.message }}</p>
              <p class="text-gray-500 text-xs mt-1">{{ b.delivery_days }} day{{ b.delivery_days !== 1 ? 's' : '' }} delivery</p>
            </div>
            <div class="text-right shrink-0">
              <div class="font-mono font-bold text-brand-400">{{ b.amount }} {{ task.currency }}</div>
              <div class="text-xs text-gray-500 mt-0.5">{{ relTime(b.created_at) }}</div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import api from '../api.js'
import { useAuthStore } from '../stores/auth.js'

const route = useRoute()
const auth = useAuthStore()
const task = ref(null)
const bids = ref([])
const loading = ref(true)
const submitting = ref(false)
const bidError = ref('')
const bidSuccess = ref(false)
const deliverySubmitting = ref(false)
const deliveryError = ref('')
const deliverySuccess = ref(false)

const bid = ref({ amount: null, delivery_days: null, message: '' })
const delivery = ref({ content_text: '', file: null })

const isAssignedWorker = computed(() => {
  return bids.value.some(b => b.user_id === auth.userID && b.status === 'accepted')
})

async function load() {
  loading.value = true
  try {
    const [taskRes, bidsRes] = await Promise.allSettled([
      api.get(`/tasks/${route.params.id}`),
      api.get(`/tasks/${route.params.id}/bids`).catch(() => ({ data: [] })),
    ])
    if (taskRes.status === 'fulfilled') task.value = taskRes.value.data
    // Bids come embedded in agent endpoint; for public view we show bids from the task detail
    // We'll fetch bids differently — from the task bids endpoint (added below)
  } catch {
    //
  } finally {
    loading.value = false
  }
  // Load bids separately
  try {
    const res = await api.get(`/tasks/${route.params.id}`)
    task.value = res.data
  } catch {
    task.value = null
  }
  // Bids are not public in current API — show empty for unauthenticated; this is acceptable for MVP
  bids.value = []
}

async function loadBids() {
  try {
    const res = await api.get(`/tasks/${route.params.id}/bids`)
    bids.value = res.data || []
  } catch {
    bids.value = []
  }
}

async function submitBid() {
  bidError.value = ''
  bidSuccess.value = false
  submitting.value = true
  try {
    await api.post(`/tasks/${route.params.id}/bids`, {
      amount: bid.value.amount,
      delivery_days: bid.value.delivery_days,
      message: bid.value.message,
    })
    bidSuccess.value = true
    bid.value = { amount: null, delivery_days: null, message: '' }
    await loadBids()
  } catch (e) {
    bidError.value = e.response?.data?.error || 'Failed to submit bid'
  } finally {
    submitting.value = false
  }
}

async function submitDelivery() {
  deliveryError.value = ''
  deliverySuccess.value = false
  deliverySubmitting.value = true
  try {
    const form = new FormData()
    form.append('content_text', delivery.value.content_text)
    if (delivery.value.file) form.append('file', delivery.value.file)
    await api.post(`/tasks/${route.params.id}/deliveries`, form, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    deliverySuccess.value = true
    delivery.value = { content_text: '', file: null }
    // Refresh task status
    const res = await api.get(`/tasks/${route.params.id}`)
    task.value = res.data
  } catch (e) {
    deliveryError.value = e.response?.data?.error || 'Failed to submit delivery'
  } finally {
    deliverySubmitting.value = false
  }
}

function statusBadge(s) {
  return { open: 'badge-open', assigned: 'badge-assigned', delivered: 'badge-delivered',
           completed: 'badge-completed', disputed: 'badge-disputed' }[s] || 'badge'
}
function bidStatusClass(s) {
  return { pending: 'badge bg-gray-700 text-gray-300', accepted: 'badge-open',
           rejected: 'badge bg-red-900/40 text-red-300', withdrawn: 'badge bg-gray-800 text-gray-500' }[s] || 'badge'
}
function formatDate(d) {
  return d ? new Date(d).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' }) : '—'
}
function relTime(d) {
  const h = Math.floor((Date.now() - new Date(d)) / 3600000)
  if (h < 1) return 'just now'
  if (h < 24) return `${h}h ago`
  return `${Math.floor(h / 24)}d ago`
}

onMounted(async () => {
  try {
    const res = await api.get(`/tasks/${route.params.id}`)
    task.value = res.data
  } catch {
    task.value = null
  }
  loading.value = false
  await loadBids()
})
</script>
