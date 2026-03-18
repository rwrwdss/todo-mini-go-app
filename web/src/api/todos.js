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

export async function getTodoById(id) {
  const res = await authFetch(`${API_BASE}/todos/${id}`, { cache: 'no-store' })
  return handleResponse(res)
}

export async function getTodos(spaceId = null, parentId = null) {
  const params = new URLSearchParams()
  if (spaceId != null && spaceId > 0) params.set('space_id', String(spaceId))
  if (parentId != null) params.set('parent_id', String(parentId))
  const qs = params.toString()
  const url = qs ? `${API_BASE}/todos?${qs}` : `${API_BASE}/todos`
  const res = await authFetch(url, { cache: 'no-store' })
  return handleResponse(res)
}

export async function getTodosFromOtherSpaces() {
  const res = await authFetch(`${API_BASE}/todos/from-other-spaces`, { cache: 'no-store' })
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
      space_id: body.space_id ?? null,
      assignee_id: body.assignee_id ?? null,
      due_at: body.due_at ?? null,
      due_date: body.due_date ?? null,
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
