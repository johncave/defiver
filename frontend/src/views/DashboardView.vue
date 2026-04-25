<template>
  <div class="max-w-5xl mx-auto px-4 py-10">
    <h1 class="text-2xl font-bold text-white mb-8">My Dashboard</h1>

    <!-- Stats -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-10">
      <div class="card text-center">
        <div class="text-2xl font-bold text-white">{{ stats.total }}</div>
        <div class="text-xs text-gray-500 mt-1">Total bids</div>
      </div>
      <div class="card text-center">
        <div class="text-2xl font-bold text-brand-400">{{ stats.accepted }}</div>
        <div class="text-xs text-gray-500 mt-1">Won</div>
      </div>
      <div class="card text-center">
        <div class="text-2xl font-bold text-yellow-400">{{ stats.pending }}</div>
        <div class="text-xs text-gray-500 mt-1">Pending</div>
      </div>
      <div class="card text-center">
        <div class="text-2xl font-bold text-gray-300">{{ stats.completed }}</div>
        <div class="text-xs text-gray-500 mt-1">Completed</div>
      </div>
    </div>

    <!-- Tabs -->
    <div class="flex gap-1 bg-gray-900 p-1 rounded-lg mb-6 w-fit">
      <button v-for="tab in tabs" :key="tab.key"
        @click="activeTab = tab.key"
        :class="['px-4 py-1.5 rounded-md text-sm font-medium transition-colors',
          activeTab === tab.key ? 'bg-gray-700 text-white' : 'text-gray-400 hover:text-white']">
        {{ tab.label }}
      </button>
    </div>

    <!-- My Bids -->
    <div v-if="activeTab === 'bids'">
      <div v-if="loading" class="space-y-3">
        <div v-for="i in 4" :key="i" class="card animate-pulse h-20 bg-gray-800/50" />
      </div>
      <div v-else-if="!bids.length" class="card text-center py-12 text-gray-500">
        You haven't placed any bids yet.
        <router-link to="/tasks" class="block mt-3 text-brand-500 hover:text-brand-400">Browse open tasks →</router-link>
      </div>
      <div v-else class="space-y-3">
        <div v-for="b in bids" :key="b.id" class="card flex items-start justify-between gap-4">
          <div class="flex-1 min-w-0">
            <router-link :to="`/tasks/${b.task_id}`" class="font-semibold text-white hover:text-brand-400 transition-colors block truncate">
              {{ b.task_title }}
            </router-link>
            <p v-if="b.message" class="text-gray-400 text-sm mt-1 line-clamp-1">{{ b.message }}</p>
            <div class="flex items-center gap-3 mt-2 text-xs text-gray-500">
              <span>{{ b.delivery_days }}d delivery</span>
              <span>{{ relTime(b.created_at) }}</span>
              <span :class="taskStatusBadge(b.task_status)">Task: {{ b.task_status }}</span>
            </div>
          </div>
          <div class="text-right shrink-0">
            <div class="font-mono font-bold text-brand-400">{{ b.amount }} {{ b.task_currency }}</div>
            <span :class="bidStatusClass(b.status)" class="mt-1 block text-center">{{ b.status }}</span>
            <button v-if="b.status === 'pending'"
              @click="withdrawBid(b.id)"
              class="mt-2 text-xs text-gray-500 hover:text-red-400 transition-colors">
              Withdraw
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Active Jobs (accepted bids on assigned tasks) -->
    <div v-if="activeTab === 'jobs'">
      <div v-if="loading" class="space-y-3">
        <div v-for="i in 2" :key="i" class="card animate-pulse h-28 bg-gray-800/50" />
      </div>
      <div v-else-if="!activeJobs.length" class="card text-center py-12 text-gray-500">
        No active jobs. Win a bid to get started!
      </div>
      <div v-else class="space-y-4">
        <div v-for="job in activeJobs" :key="job.id" class="card">
          <div class="flex items-start justify-between gap-4 mb-3">
            <router-link :to="`/tasks/${job.task_id}`" class="font-semibold text-white hover:text-brand-400 transition-colors">
              {{ job.task_title }}
            </router-link>
            <span :class="taskStatusBadge(job.task_status)">{{ job.task_status }}</span>
          </div>
          <div class="flex items-center gap-4 text-sm text-gray-400">
            <span class="font-mono text-brand-400 font-semibold">{{ job.amount }} {{ job.task_currency }}</span>
            <span>{{ job.delivery_days }}d delivery</span>
          </div>
          <div v-if="job.task_status === 'assigned'" class="mt-3">
            <router-link :to="`/tasks/${job.task_id}`" class="btn-primary text-sm">
              Submit Delivery →
            </router-link>
          </div>
        </div>
      </div>
    </div>

    <!-- Profile -->
    <div v-if="activeTab === 'profile'">
      <div v-if="profileLoading" class="card animate-pulse h-40" />
      <div v-else-if="profile" class="card max-w-lg">
        <div class="space-y-4">
          <div>
            <label class="block text-xs text-gray-500 mb-1">Display name</label>
            <div class="text-white font-semibold">{{ profile.display_name }}</div>
          </div>
          <div>
            <label class="block text-xs text-gray-500 mb-1">Email</label>
            <div class="text-gray-300">{{ profile.email }}</div>
          </div>
          <div>
            <label class="block text-xs text-gray-500 mb-1">Reputation score</label>
            <div class="text-gray-300">{{ profile.reputation_score.toFixed(1) }}</div>
          </div>
          <div v-if="profile.bio">
            <label class="block text-xs text-gray-500 mb-1">Bio</label>
            <div class="text-gray-300 text-sm">{{ profile.bio }}</div>
          </div>
          <div v-if="profile.wallet_address">
            <label class="block text-xs text-gray-500 mb-1">Wallet</label>
            <div class="text-gray-300 font-mono text-xs break-all">{{ profile.wallet_address }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../api.js'

const loading = ref(true)
const profileLoading = ref(true)
const bids = ref([])
const profile = ref(null)
const activeTab = ref('bids')

const tabs = [
  { key: 'bids', label: 'My Bids' },
  { key: 'jobs', label: 'Active Jobs' },
  { key: 'profile', label: 'Profile' },
]

const stats = computed(() => ({
  total: bids.value.length,
  accepted: bids.value.filter(b => b.status === 'accepted').length,
  pending: bids.value.filter(b => b.status === 'pending').length,
  completed: bids.value.filter(b => b.task_status === 'completed').length,
}))

const activeJobs = computed(() =>
  bids.value.filter(b => b.status === 'accepted' && ['assigned', 'delivered'].includes(b.task_status))
)

async function withdrawBid(bidID) {
  try {
    await api.delete(`/users/bids/${bidID}`)
    const b = bids.value.find(x => x.id === bidID)
    if (b) b.status = 'withdrawn'
  } catch {
    //
  }
}

function bidStatusClass(s) {
  return { pending: 'badge bg-gray-700 text-gray-300 text-xs', accepted: 'badge-open text-xs',
           rejected: 'badge bg-red-900/40 text-red-300 text-xs', withdrawn: 'badge bg-gray-800 text-gray-500 text-xs' }[s] || 'badge text-xs'
}
function taskStatusBadge(s) {
  return { open: 'badge-open', assigned: 'badge-assigned', delivered: 'badge-delivered',
           completed: 'badge-completed', disputed: 'badge-disputed' }[s] || 'badge'
}
function relTime(d) {
  const h = Math.floor((Date.now() - new Date(d)) / 3600000)
  if (h < 1) return 'just now'
  if (h < 24) return `${h}h ago`
  return `${Math.floor(h / 24)}d ago`
}

onMounted(async () => {
  const [bidsRes, profileRes] = await Promise.allSettled([
    api.get('/users/bids'),
    api.get('/users/me'),
  ])
  if (bidsRes.status === 'fulfilled') bids.value = bidsRes.value.data
  if (profileRes.status === 'fulfilled') profile.value = profileRes.value.data
  loading.value = false
  profileLoading.value = false
})
</script>
