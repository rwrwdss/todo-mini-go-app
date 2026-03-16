export default function Header({ taskCount, onNewTask }) {
  const label = taskCount === 1 ? '1 task' : `${taskCount} tasks`
  return (
    <header>
      <div className="logo">
        task<span>.</span>grid
      </div>
      <div className="hdr-right">
        <div className="task-count">{label}</div>
        <button type="button" className="add-root-btn" onClick={onNewTask}>
          + New task
        </button>
      </div>
    </header>
  )
}
