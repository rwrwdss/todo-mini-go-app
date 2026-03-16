export default function TodoList({ todos, loading, error }) {
  if (error) {
    return <p className="todo-error" role="alert">{error}</p>
  }
  if (loading) {
    return <p className="todo-loading">Загрузка…</p>
  }
  if (!todos?.length) {
    return <p className="todo-empty">Нет задач. Добавьте первую.</p>
  }
  return (
    <ul className="todo-list" aria-label="Список задач">
      {todos.map((t) => (
        <li key={t.id} className={`todo-item ${t.done ? 'todo-item--done' : ''}`}>
          <span className="todo-item-status" aria-hidden>
            {t.done ? '✓' : '○'}
          </span>
          <span className="todo-item-title">{t.title}</span>
        </li>
      ))}
    </ul>
  )
}
