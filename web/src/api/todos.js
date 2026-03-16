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

export async function getTodos(parentId = null) {
  const url = parentId != null
    ? `${API_BASE}/todos?parent_id=${parentId}`
    : `${API_BASE}/todos`
  const res = await authFetch(url, { cache: 'no-store' })
  return handleResponse(res)
}

export async function createTodo(payload) {
  const body = typeof payload === 'string' ? { title: payload } : payload
  if (!body.title || !String(body.title).trim()) {
    throw new Error('Title is required')
  }
  const res = await authFetch(`${API_BASE}/create`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      title: body.title.trim(),
      done: body.done ?? false,
      description: body.description ?? '',
      priority: body.priority ?? 'none',
      tag: body.tag ?? '',
      parent_id: body.parent_id ?? null,
    }),
  })
  return handleResponse(res)
}

export async function updateTodo(id, payload) {
  const res = await authFetch(`${API_BASE}/todos/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })
  return handleResponse(res)
}

export async function deleteTodo(id) {
  const res = await authFetch(`${API_BASE}/todos/${id}`, {
    method: 'DELETE',
  })
  if (res.status === 204) return
  const text = await res.text()
  if (!res.ok) throw new Error(text || `HTTP ${res.status}`)
}
