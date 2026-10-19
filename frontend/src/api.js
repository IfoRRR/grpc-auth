const API = '/api/grpc-auth'

async function request(path, opts = {}) {
  const headers = { 'Content-Type': 'application/json', ...(opts.headers || {}) }
  const res = await fetch(`${API}${path}`, { ...opts, headers })
  const text = await res.text()
  let data
  try { data = text ? JSON.parse(text) : {} } catch { data = { error: text } }
  if (!res.ok) throw new Error(data.error || res.statusText)
  return data
}

export const api = {
  register: (username, password) =>
    request('/auth/register', { method: 'POST', body: JSON.stringify({ username, password }) }),
  login: (username, password) =>
    request('/auth/login', { method: 'POST', body: JSON.stringify({ username, password }) }),
  listTasks: (token) =>
    request('/tasks', { headers: { Authorization: `Bearer ${token}` } }),
  createTask: (token, title) =>
    request('/tasks', {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify({ title }),
    }),
  probeGRPC: (token) =>
    request('/probe', {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
    }),
}
