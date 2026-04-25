<template>
  <router-link :to="`/tasks/${task.id}`" class="card hover:border-gray-600 transition-colors block">
    <div class="flex items-start justify-between gap-3 mb-2">
      <h3 class="font-semibold text-white line-clamp-2">{{ task.title }}</h3>
      <span :class="statusClass">{{ task.status }}</span>
    </div>
    <p class="text-gray-400 text-sm line-clamp-2 mb-3">{{ task.description }}</p>
    <div class="flex items-center justify-between text-sm">
      <div class="flex items-center gap-3 text-gray-400">
        <span class="font-mono text-brand-400 font-semibold">{{ task.max_budget }} {{ task.currency }}</span>
        <span>{{ task.bid_count ?? 0 }} bids</span>
      </div>
      <span class="text-gray-500">{{ relativeTime }}</span>
    </div>
    <div v-if="task.deadline" class="mt-2 text-xs text-gray-500">
      Due {{ formatDate(task.deadline) }}
    </div>
  </router-link>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({ task: Object })

const statusClass = computed(() => ({
  open:      'badge-open',
  assigned:  'badge-assigned',
  delivered: 'badge-delivered',
  completed: 'badge-completed',
  disputed:  'badge-disputed',
}[props.task.status] || 'badge'))

const relativeTime = computed(() => {
  const diff = Date.now() - new Date(props.task.created_at)
  const h = Math.floor(diff / 3600000)
  if (h < 1) return 'just now'
  if (h < 24) return `${h}h ago`
  return `${Math.floor(h / 24)}d ago`
})

function formatDate(d) {
  return new Date(d).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
}
</script>
