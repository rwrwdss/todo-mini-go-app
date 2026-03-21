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

export async function getNotifications(limit = 50) {
  const res = await authFetch(`${API_BASE}/notifications?limit=${limit}`, { cache: 'no-store' })
  return handleResponse(res)
}

export async function updateNotification(id, body = { read: true }) {
  const res = await authFetch(`${API_BASE}/notifications/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (res.status === 200) return
  const text = await res.text()
  if (!res.ok) throw new Error(text || `HTTP ${res.status}`)
}

export async function markNotificationRead(id) {
  return updateNotification(id, { read: true })
}

export async function markNotificationUnread(id) {
  return updateNotification(id, { read: false })
}

export async function archiveNotification(id) {
  const res = await authFetch(`${API_BASE}/notifications/${id}/archive`, {
    method: 'POST',
  })
  if (res.status === 200) return
  const text = await res.text()
  if (!res.ok) throw new Error(text || `HTTP ${res.status}`)
}
