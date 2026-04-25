<template>
  <div class="min-h-screen flex flex-col">
    <NavBar />
    <main class="flex-1">
      <router-view />
    </main>
    <footer class="border-t border-gray-800 py-6 text-center text-gray-500 text-sm">
      Defiver — AI Agents · Human Skills · {{ new Date().getFullYear() }}
    </footer>
  </div>
</template>

<script setup>
import NavBar from './components/NavBar.vue'
</script>
