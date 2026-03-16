export default function TaskCard({ task, isRoot, onToggleDone, onEdit, onDelete, onAddSub }) {
  const priority = (task.priority || 'none').toLowerCase()
  const pc = priority !== 'none' ? `p-${priority === 'medium' ? 'med' : priority}` : ''
  const done = task.done ? 'done' : ''
  const rc = isRoot ? 'root-card' : ''

  return (
    <div className={`task-card ${pc} ${done} ${rc}`} data-id={task.id}>
      <div className="task-header">
        <div className="task-left">
          <div
            className={`chk ${task.done ? 'checked' : ''}`}
            onClick={() => onToggleDone(task)}
            role="button"
            tabIndex={0}
            onKeyDown={(e) => e.key === 'Enter' && onToggleDone(task)}
            aria-label={task.done ? 'Mark not done' : 'Mark done'}
          />
          <div className="task-title">{task.title}</div>
        </div>
        <div className="task-actions">
          <button
            type="button"
            className="tbtn"
            title="Edit"
            onClick={() => onEdit(task)}
          >
            <svg width="11" height="11" viewBox="0 0 11 11" fill="none" stroke="currentColor" strokeWidth="1.4">
              <path d="M1 10l2-.5 6.5-6.5-2-2L1.5 8 1 10z" />
            </svg>
          </button>
          <button
            type="button"
            className="tbtn del"
            title="Delete"
            onClick={() => onDelete(task)}
          >
            <svg width="11" height="11" viewBox="0 0 11 11" fill="none" stroke="currentColor" strokeWidth="1.4">
              <path d="M1.5 1.5l8 8M9.5 1.5l-8 8" />
            </svg>
          </button>
        </div>
      </div>
      {task.description ? (
        <div className="task-desc">{task.description}</div>
      ) : null}
      <div className="task-footer">
        <div className="task-meta">
          {task.tag ? <span className="tag pink">{task.tag}</span> : null}
          {task.priority && task.priority !== 'none' ? (
            <span className="tag">{task.priority}</span>
          ) : null}
        </div>
        <button type="button" className="add-child-btn" onClick={() => onAddSub(task)}>
          + sub
        </button>
      </div>
    </div>
  )
}
