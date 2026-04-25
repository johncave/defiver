import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userID = ref(Number(localStorage.getItem('userID')) || null)
  const displayName = ref(localStorage.getItem('displayName') || '')

  const isLoggedIn = computed(() => !!token.value)

  function setAuth(data) {
    token.value = data.token
    userID.value = data.user_id
    displayName.value = data.display_name
    localStorage.setItem('token', data.token)
    localStorage.setItem('userID', data.user_id)
    localStorage.setItem('displayName', data.display_name)
  }

  function logout() {
    token.value = ''
    userID.value = null
    displayName.value = ''
    localStorage.removeItem('token')
    localStorage.removeItem('userID')
    localStorage.removeItem('displayName')
  }

  return { token, userID, displayName, isLoggedIn, setAuth, logout }
})
