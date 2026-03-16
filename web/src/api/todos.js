const API_BASE = '/api'

async function handleResponse(res) {
  const text = await res.text()
  if (!res.ok) {
    throw new Error(text || `HTTP ${res.status}`)
  }
  return text ? JSON.parse(text) : null
}

export async function getTodos(parentId = null) {
  const url = parentId != null
    ? `${API_BASE}/todos?parent_id=${parentId}`
    : `${API_BASE}/todos`
  const res = await fetch(url, { cache: 'no-store' })
  return handleResponse(res)
}

export async function createTodo(payload) {
  const body = typeof payload === 'string' ? { title: payload } : payload
  if (!body.title || !String(body.title).trim()) {
    throw new Error('Title is required')
  }
  const res = await fetch(`${API_BASE}/create`, {
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
  const res = await fetch(`${API_BASE}/todos/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })
  return handleResponse(res)
}

export async function deleteTodo(id) {
  const res = await fetch(`${API_BASE}/todos/${id}`, {
    method: 'DELETE',
  })
  if (res.status === 204) return
  const text = await res.text()
  if (!res.ok) throw new Error(text || `HTTP ${res.status}`)
}
