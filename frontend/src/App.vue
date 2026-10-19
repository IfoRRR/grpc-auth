<script setup>
import { ref, computed, onMounted } from 'vue'
import { api } from './api'

const mode = ref('login')
const username = ref('demo')
const password = ref('demo1234')
const token = ref(localStorage.getItem('grpc_auth_token') || '')
const user = ref(localStorage.getItem('grpc_auth_user') || '')
const tasks = ref([])
const newTitle = ref('')
const probe = ref(null)
const toast = ref('')
const loading = ref(false)

const loggedIn = computed(() => !!token.value)

function showToast(msg) {
  toast.value = msg
  setTimeout(() => { toast.value = '' }, 2500)
}

async function auth() {
  loading.value = true
  try {
    const fn = mode.value === 'login' ? api.login : api.register
    const data = await fn(username.value, password.value)
    token.value = data.token
    user.value = data.username
    localStorage.setItem('grpc_auth_token', data.token)
    localStorage.setItem('grpc_auth_user', data.username)
    showToast(mode.value === 'login' ? 'Вход OK' : 'Регистрация OK')
    await loadTasks()
  } catch (e) {
    showToast(e.message)
  } finally {
    loading.value = false
  }
}

function logout() {
  token.value = ''
  user.value = ''
  tasks.value = []
  probe.value = null
  localStorage.removeItem('grpc_auth_token')
  localStorage.removeItem('grpc_auth_user')
}

async function loadTasks() {
  if (!token.value) return
  const data = await api.listTasks(token.value)
  tasks.value = data.tasks || []
}

async function addTask() {
  if (!newTitle.value.trim()) return
  loading.value = true
  try {
    await api.createTask(token.value, newTitle.value.trim())
    newTitle.value = ''
    await loadTasks()
    showToast('Задача создана (REST)')
  } catch (e) {
    showToast(e.message)
  } finally {
    loading.value = false
  }
}

async function runProbe() {
  loading.value = true
  try {
    probe.value = await api.probeGRPC(token.value)
    showToast('gRPC ListTasks OK')
  } catch (e) {
    showToast(e.message)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  if (token.value) {
    try { await loadTasks() } catch { logout() }
  }
})
</script>

<template>
  <nav class="demo-nav">
    <a href="/" class="demo-nav-link">← Портфолио</a>
    <span class="demo-nav-title">backend lab · 04 / 04</span>
  </nav>

  <div class="page">
    <header class="hero glass">
      <span class="label mono">grpc auth lab</span>
      <h1>gRPC + JWT</h1>
      <p>REST для браузера · тот же Bearer в gRPC metadata · unary interceptor</p>
    </header>

    <section v-if="!loggedIn" class="panel glass">
      <div class="tabs">
        <button :class="{ ghost: mode !== 'login' }" @click="mode = 'login'">Вход</button>
        <button :class="{ ghost: mode !== 'register' }" @click="mode = 'register'">Регистрация</button>
      </div>
      <label>Username<input v-model="username" autocomplete="username" /></label>
      <label>Password<input v-model="password" type="password" autocomplete="current-password" /></label>
      <button :disabled="loading" @click="auth">
        {{ loading ? 'Проверяем JWT…' : (mode === 'login' ? 'Войти' : 'Создать аккаунт') }}
      </button>
      <p class="hint">Demo: <code>demo / demo1234</code> — зарегистрируйте один раз</p>
    </section>

    <section v-else class="panel glass">
      <div class="row">
        <div>
          <div class="mono muted">logged in</div>
          <strong>{{ user }}</strong>
        </div>
        <button class="ghost" @click="logout">Выйти</button>
      </div>

      <div class="row">
        <input v-model="newTitle" placeholder="Новая задача…" @keyup.enter="addTask" />
        <button :disabled="loading" @click="addTask">REST Create</button>
      </div>

      <ul class="tasks">
        <li v-for="t in tasks" :key="t.id" class="mono">{{ t.title }} <span class="muted">#{{ t.id.slice(0, 8) }}</span></li>
        <li v-if="!tasks.length" class="muted">Пока нет задач</li>
      </ul>

      <button :disabled="loading" @click="runProbe">
        {{ loading ? 'Вызываем gRPC…' : 'Probe gRPC ListTasks' }}
      </button>

      <pre v-if="probe" class="probe mono">{{ JSON.stringify(probe, null, 2) }}</pre>
    </section>

    <footer class="hint mono">
      JWT → Authorization header (REST) · metadata authorization (gRPC) · TaskService за interceptor
    </footer>

    <div v-if="toast" class="toast">{{ toast }}</div>
  </div>
</template>

<style scoped>
.page {
  max-width: 720px;
  margin: 0 auto;
  padding: 1rem 1rem 3rem;
}

.hero {
  padding: 1.25rem 1.5rem;
  margin-bottom: 1rem;
}

.label {
  color: var(--accent);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-size: 0.72rem;
}

h1 { margin: 0.35rem 0; font-size: 1.6rem; }
.hero p { margin: 0; color: var(--muted); }

.panel {
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.tabs { display: flex; gap: 0.5rem; }

label {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  font-size: 0.85rem;
  color: var(--muted);
}

.row {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  flex-wrap: wrap;
}

.tasks {
  list-style: none;
  margin: 0;
  padding: 0;
}

.tasks li {
  padding: 0.45rem 0;
  border-bottom: 1px solid var(--border);
}

.probe {
  background: rgba(0, 0, 0, 0.35);
  padding: 0.75rem;
  border-radius: 10px;
  overflow: auto;
  max-height: 220px;
  font-size: 0.75rem;
}

.hint {
  color: var(--muted);
  font-size: 0.78rem;
  margin: 0;
}

.muted { color: var(--muted); }

.toast {
  position: fixed;
  bottom: 1rem;
  left: 50%;
  transform: translateX(-50%);
  background: #111;
  border: 1px solid var(--border);
  padding: 0.6rem 1rem;
  border-radius: 999px;
  font-size: 0.85rem;
}

.demo-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  max-width: 720px;
  margin: 0 auto;
}

.demo-nav-link {
  color: var(--accent);
  text-decoration: none;
  font-size: 0.9rem;
}

.demo-nav-title {
  color: var(--muted);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}
</style>
