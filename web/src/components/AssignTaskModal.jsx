import { useState, useEffect, useRef } from 'react'

const PRIORITIES = [
  { value: 'none', label: 'None' },
  { value: 'low', label: 'Low' },
  { value: 'med', label: 'Medium' },
  { value: 'high', label: 'High' },
]

export default function AssignTaskModal({ open, spaceId, members = [], existingTags = [], onSave, onClose }) {
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [priority, setPriority] = useState('none')
  const [tag, setTag] = useState('')
  const [dueDate, setDueDate] = useState('')
  const [assigneeId, setAssigneeId] = useState(null)
  const [titleError, setTitleError] = useState(false)
  const inputRef = useRef(null)

  useEffect(() => {
    if (open) {
      setTitle('')
      setDescription('')
      setPriority('none')
      setTag('')
      setDueDate('')
      setAssigneeId(members[0]?.id ?? null)
      setTitleError(false)
      setTimeout(() => inputRef.current?.focus(), 50)
    }
  }, [open, members])

  function handleSubmit(e) {
    e.preventDefault()
    const t = title.trim()
    if (!t) {
      setTitleError(true)
      return
    }
    if (!assigneeId) {
      return
    }
    onSave({
      title: t,
      description: description.trim(),
      priority: priority === 'medium' ? 'med' : (priority || 'none'),
      tag: tag.trim(),
      due_date: dueDate.trim() || null,
      space_id: spaceId,
      assignee_id: assigneeId,
    })
    onClose()
  }

  if (!open) return null

  return (
    <div className="modal-bg open" onClick={(e) => e.target === e.currentTarget && onClose()}>
      <div className="modal">
        <div className="modal-ttl">Assign task</div>
        <form onSubmit={handleSubmit}>
          <div className="fg">
            <label className="fl" htmlFor="assign-title">Title</label>
            <input
              ref={inputRef}
              id="assign-title"
              type="text"
              className="fi"
              placeholder="Task name..."
              value={title}
              onChange={(e) => { setTitle(e.target.value); setTitleError(false) }}
              autoComplete="off"
              style={titleError ? { borderColor: 'var(--red)' } : {}}
            />
          </div>
          <div className="fg">
            <label className="fl" htmlFor="assign-desc">Description</label>
            <textarea
              id="assign-desc"
              className="fi"
              placeholder="Description..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
            />
          </div>
          <div className="fg">
            <label className="fl" htmlFor="assign-priority">Priority</label>
            <select id="assign-priority" className="fi" value={priority} onChange={(e) => setPriority(e.target.value)}>
              {PRIORITIES.map((p) => (
                <option key={p.value} value={p.value}>{p.label}</option>
              ))}
            </select>
          </div>
          <div className="fg">
            <label className="fl" htmlFor="assign-due">Due date</label>
            <input
              id="assign-due"
              type="date"
              className="fi"
              value={dueDate}
              onChange={(e) => setDueDate(e.target.value)}
            />
          </div>
          <div className="fg">
            <label className="fl" htmlFor="assign-tag">Tag</label>
            <input
              id="assign-tag"
              type="text"
              className="fi"
              placeholder="Tag..."
              value={tag}
              onChange={(e) => setTag(e.target.value)}
              list="assign-tag-list"
            />
            {existingTags.length > 0 ? (
              <datalist id="assign-tag-list">
                {existingTags.map((tagName) => (
                  <option key={tagName} value={tagName} />
                ))}
              </datalist>
            ) : null}
          </div>
          <div className="fg">
            <label className="fl">Assignee</label>
            <div className="assignee-picker">
              {members.map((m) => (
                <button
                  key={m.id}
                  type="button"
                  className={`ap-chip ${assigneeId === m.id ? 'on' : ''}`}
                  onClick={() => setAssigneeId(m.id)}
                >
                  <span className="ap-av">{(m.name || m.email || '?').charAt(0).toUpperCase()}</span>
                  {m.name || m.email}
                </button>
              ))}
            </div>
          </div>
          <div className="modal-footer">
            <button type="button" className="bc" onClick={onClose}>Cancel</button>
            <button type="submit" className="btn-save">Assign</button>
          </div>
        </form>
      </div>
    </div>
  )
}
