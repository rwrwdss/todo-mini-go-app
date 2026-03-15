import { useState, useEffect, useCallback } from 'react'
import { getTodos, createTodo } from './api/todos'
import TodoForm from './components/TodoForm'
import TodoList from './components/TodoList'
import './App.css'

export default function App() {
  const [todos, setTodos] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  const loadTodos = useCallback(async (background = false) => {
    if (!background) {
      setLoading(true)
      setError(null)
    }
    try {
      const data = await getTodos()
      setTodos(Array.isArray(data) ? data : [])
    } catch (e) {
      setError(e.message || 'Не удалось загрузить задачи')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadTodos()
  }, [loadTodos])

  async function handleAdded(title) {
    setError(null)
    try {
      await createTodo(title)
    } catch (e) {
      setError(e.message || 'Не удалось создать задачу')
    }
    await loadTodos(true)
  }

  return (
    <main className="app">
      <h1>Todo App</h1>
      <TodoForm onAdded={handleAdded} disabled={loading} />
      <TodoList todos={todos} loading={loading} error={error} />
    </main>
  )
}
