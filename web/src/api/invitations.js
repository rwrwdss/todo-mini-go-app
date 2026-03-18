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

export async function getInvitations() {
  const res = await authFetch(`${API_BASE}/invitations`, { cache: 'no-store' })
  return handleResponse(res)
}

function parseInvitationId(invitationId) {
  if (invitationId == null || invitationId === '') return null
  const n = Number(invitationId)
  return Number.isInteger(n) && n >= 1 ? n : null
}

export async function acceptInvitation(invitationId) {
  const id = parseInvitationId(invitationId)
  if (id == null) throw new Error('Invalid invitation id')
  const res = await authFetch(`${API_BASE}/invitations/${id}/accept`, {
    method: 'POST',
  })
  if (res.status === 200) return
  const text = await res.text()
  if (!res.ok) throw new Error(text || `HTTP ${res.status}`)
}

export async function declineInvitation(invitationId) {
  const id = parseInvitationId(invitationId)
  if (id == null) throw new Error('Invalid invitation id')
  const res = await authFetch(`${API_BASE}/invitations/${id}/decline`, {
    method: 'POST',
  })
  if (res.status === 200) return
  const text = await res.text()
  if (!res.ok) throw new Error(text || `HTTP ${res.status}`)
}
