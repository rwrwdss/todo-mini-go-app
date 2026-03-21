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

export async function getSpaceActivity(spaceId, { limit = 100, eventType = '' } = {}) {
  const params = new URLSearchParams()
  params.set('limit', String(limit))
  if (eventType) params.set('event_type', eventType)
  const res = await fetch(`${API_BASE}/spaces/${spaceId}/activity?${params.toString()}`, {
    headers: { ...getAuthHeaders() },
    cache: 'no-store',
  })
  return handleResponse(res)
}
