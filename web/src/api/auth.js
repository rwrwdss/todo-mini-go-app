const API_BASE = '/api'

function getStoredToken() {
  return typeof window !== 'undefined' ? localStorage.getItem('token') : null
}

export function getAuthHeaders() {
  const token = getStoredToken()
  if (!token) return {}
  return { Authorization: `Bearer ${token}` }
}

export async function login(email, password) {
  const res = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: email.trim().toLowerCase(), password }),
  })
  const text = await res.text()
  if (!res.ok) throw new Error(text || 'Login failed')
  return JSON.parse(text)
}

export async function register(email, password, name) {
  const res = await fetch(`${API_BASE}/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      email: email.trim().toLowerCase(),
      password,
      name: (name || '').trim(),
    }),
  })
  const text = await res.text()
  if (!res.ok) throw new Error(text || 'Registration failed')
  return JSON.parse(text)
}

export function logout() {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
}

export function saveAuth(token, user) {
  localStorage.setItem('token', token)
  if (user) localStorage.setItem('user', JSON.stringify(user))
}
