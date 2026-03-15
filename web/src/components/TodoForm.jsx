import { useState } from 'react'

export default function TodoForm({ onAdded, disabled }) {
  const [title, setTitle] = useState('')
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e) {
    e.preventDefault()
    const trimmed = title.trim()
    if (!trimmed || disabled || submitting) return
    setSubmitting(true)
    try {
      await onAdded(trimmed)
      setTitle('')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form className="todo-form" onSubmit={handleSubmit}>
      <input
        type="text"
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        placeholder="Новая задача"
        disabled={disabled}
        aria-label="Текст задачи"
      />
      <button type="submit" disabled={disabled || submitting || !title.trim()}>
        {submitting ? '…' : 'Добавить'}
      </button>
    </form>
  )
}
