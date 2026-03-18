import { getAuthHeaders } from './auth'

const API_BASE = '/api'

async function handleResponse(res) {
  const text = await res.text()
  if (res.status === 401) {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    window.location.reload()
  }
  if (!res.ok) {
    throw new Error(text || `HTTP ${res.status}`)
  }
  return text ? JSON.parse(text) : null
}

function authFetch(url, opts = {}) {
  return fetch(url, {
    ...opts,
    headers: { ...getAuthHeaders(), ...opts.headers },
  })
}

export async function getSpaces() {
  const res = await authFetch(`${API_BASE}/spaces`, { cache: 'no-store' })
  return handleResponse(res)
}

export async function getSpace(id) {
  const res = await authFetch(`${API_BASE}/spaces/${id}`, { cache: 'no-store' })
  return handleResponse(res)
}

export async function createSpace(name) {
  const res = await authFetch(`${API_BASE}/spaces`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name: (name || '').trim() }),
  })
  return handleResponse(res)
}

export async function inviteMember(spaceId, email, role = 'member') {
  const res = await authFetch(`${API_BASE}/spaces/${spaceId}/members`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: (email || '').trim().toLowerCase(), role: role || 'member' }),
  })
  if (res.status === 204 || res.status === 200) return
  const text = await res.text()
  if (!res.ok) throw new Error(text || `HTTP ${res.status}`)
}
