const API_BASE = '/api'

async function handleResponse(res) {
  const text = await res.text()
  if (!res.ok) {
    throw new Error(text || `HTTP ${res.status}`)
  }
  return text ? JSON.parse(text) : null
}

export async function getTodos() {
  const res = await fetch(`${API_BASE}/todos`, { cache: 'no-store' })
  return handleResponse(res)
}

export async function createTodo(title) {
  const res = await fetch(`${API_BASE}/create`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title }),
  })
  return handleResponse(res)
}
