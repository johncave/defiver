<template>
  <div class="max-w-6xl mx-auto px-4 py-10">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-white">Open Tasks</h1>
      <div class="flex items-center gap-3">
        <select v-model="currency" @change="load" class="input w-auto text-sm py-1.5">
          <option value="">All currencies</option>
          <option value="USDT">USDT</option>
          <option value="DNZD">DNZD</option>
        </select>
      </div>
    </div>

    <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div v-for="i in 9" :key="i" class="card animate-pulse h-36 bg-gray-800/50" />
    </div>

    <div v-else-if="tasks.length" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <TaskCard v-for="task in tasks" :key="task.id" :task="task" />
    </div>

    <div v-else class="card text-center py-20 text-gray-500">
      No open tasks matching your filter.
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../api.js'
import TaskCard from '../components/TaskCard.vue'

const tasks = ref([])
const loading = ref(true)
const currency = ref('')

async function load() {
  loading.value = true
  try {
    const params = currency.value ? { currency: currency.value } : {}
    const res = await api.get('/tasks', { params })
    tasks.value = res.data
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
