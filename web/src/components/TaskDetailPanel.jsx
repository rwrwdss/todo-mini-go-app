import { useState, useEffect } from 'react'
import { getTodoById } from '../api/todos'

function formatDate(iso) {
  if (!iso) return '—'
  try {
    const d = new Date(iso)
    return d.toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })
  } catch {
    return iso
  }
}

function formatDateOnly(dateStr) {
  if (!dateStr) return '—'
  try {
    const d = new Date(dateStr + 'T12:00:00')
    return d.toLocaleDateString(undefined, { dateStyle: 'medium' })
  } catch {
    return dateStr
  }
}

function isOverdue(dueDateStr) {
  if (!dueDateStr) return false
  const today = new Date().toISOString().slice(0, 10)
  return dueDateStr < today
}

export default function TaskDetailPanel({ taskId, onClose, onEdit, onDelete, spaceMembers = [], canEdit = true, canDelete = true }) {
  const [task, setTask] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  useEffect(() => {
    if (!taskId) {
      setTask(null)
      setError(null)
      return
    }
    setLoading(true)
    setError(null)
    getTodoById(taskId)
      .then(setTask)
      .catch((e) => {
        setError(e.message || 'Failed to load task')
        setTask(null)
      })
      .finally(() => setLoading(false))
  }, [taskId])

  if (taskId == null) return null

  const assigneeName = task?.assignee_id && spaceMembers.length
    ? (spaceMembers.find((m) => m.id === task.assignee_id)?.name || spaceMembers.find((m) => m.id === task.assignee_id)?.email || `#${task.assignee_id}`)
    : null

  return (
    <div className="task-detail-panel-backdrop" onClick={(e) => e.target === e.currentTarget && onClose?.()}>
      <aside className="task-detail-panel" onClick={(e) => e.stopPropagation()}>
        <div className="task-detail-panel-head">
          <h2 className="task-detail-panel-title">Task details</h2>
          <button type="button" className="task-detail-panel-close" onClick={onClose} aria-label="Close">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" strokeWidth="1.5">
              <path d="M1 1l12 12M13 1L1 13" />
            </svg>
          </button>
        </div>
        {loading && <p className="task-detail-panel-loading">Loading…</p>}
        {error && <p className="task-detail-panel-error">{error}</p>}
        {task && !loading && (
          <div className="task-detail-panel-body">
            <div className="task-detail-field">
              <span className="task-detail-label">Title</span>
              <p className="task-detail-value">{task.title}</p>
            </div>
            <div className="task-detail-field">
              <span className="task-detail-label">Status</span>
              <p className="task-detail-value">{task.done ? 'Done' : 'To do'}</p>
            </div>
            {task.description ? (
              <div className="task-detail-field">
                <span className="task-detail-label">Description</span>
                <p className="task-detail-value task-detail-desc">{task.description}</p>
              </div>
            ) : null}
            {(task.tag || task.priority) && (
              <div className="task-detail-field">
                <span className="task-detail-label">Tag / Priority</span>
                <p className="task-detail-value">
                  {task.tag ? <span className="tag pink">{task.tag}</span> : null}
                  {task.tag && task.priority ? ' · ' : null}
                  {task.priority && task.priority !== 'none' ? task.priority : null}
                </p>
              </div>
            )}
            {task.creator_name != null && task.creator_name !== '' && (
              <div className="task-detail-field">
                <span className="task-detail-label">Created by</span>
                <p className="task-detail-value">{task.creator_name}</p>
              </div>
            )}
            {task.created_at && (
              <div className="task-detail-field">
                <span className="task-detail-label">Created at</span>
                <p className="task-detail-value">{formatDate(task.created_at)}</p>
              </div>
            )}
            <div className="task-detail-field">
              <span className="task-detail-label">Due date</span>
              <p className={`task-detail-value ${task.due_date && isOverdue(task.due_date) ? 'task-detail-overdue' : ''}`}>
                {task.due_date ? formatDateOnly(task.due_date) : 'No due date'}
                {task.due_date && isOverdue(task.due_date) ? ' (overdue)' : ''}
              </p>
            </div>
            {assigneeName && (
              <div className="task-detail-field">
                <span className="task-detail-label">Assignee</span>
                <p className="task-detail-value">{assigneeName}</p>
              </div>
            )}
            <div className="task-detail-panel-actions">
              <button type="button" className="bc" onClick={onClose}>Close</button>
              {canEdit ? <button type="button" className="btn-ghost" onClick={() => onEdit?.(task)}>Edit</button> : null}
              {canDelete ? <button type="button" className="btn-sm" onClick={() => onDelete?.(task)}>Delete</button> : null}
            </div>
          </div>
        )}
      </aside>
    </div>
  )
}
