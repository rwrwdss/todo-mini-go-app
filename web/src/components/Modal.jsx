import { useState, useEffect, useRef } from 'react'

const PRIORITIES = [
  { value: 'none', label: 'None' },
  { value: 'low', label: 'Low' },
  { value: 'med', label: 'Medium' },
  { value: 'high', label: 'High' },
]

export default function Modal({ open, mode, editTask, parentId, parentTag = '', existingTags = [], onSave, onClose }) {
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [priority, setPriority] = useState('none')
  const [tag, setTag] = useState('')
  const [dueDateTime, setDueDateTime] = useState('')
  const [titleError, setTitleError] = useState(false)
  const inputRef = useRef(null)

  const isEdit = mode === 'edit' && editTask

  function toDateTimeLocal(isoOrDate) {
    if (!isoOrDate) return ''
    const d = new Date(isoOrDate.indexOf('T') >= 0 ? isoOrDate : isoOrDate + 'T12:00:00')
    if (Number.isNaN(d.getTime())) return ''
    const y = d.getFullYear()
    const m = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    const h = String(d.getHours()).padStart(2, '0')
    const min = String(d.getMinutes()).padStart(2, '0')
    return `${y}-${m}-${day}T${h}:${min}`
  }

  useEffect(() => {
    if (open) {
      if (isEdit) {
        setTitle(editTask.title || '')
        setDescription(editTask.description || '')
        setPriority((editTask.priority || 'none').toLowerCase())
        setTag(editTask.tag || '')
        setDueDateTime(toDateTimeLocal(editTask.due_at || editTask.due_date || ''))
      } else {
        setTitle('')
        setDescription('')
        setPriority('none')
        setTag(parentId ? (parentTag || '').trim() : '')
        setDueDateTime('')
      }
      setTitleError(false)
      setTimeout(() => inputRef.current?.focus(), 50)
    }
  }, [open, isEdit, editTask, parentId, parentTag])

  function handleSubmit(e) {
    e.preventDefault()
    const t = title.trim()
    if (!t) {
      setTitleError(true)
      return
    }
    const prio = priority === 'medium' ? 'med' : (priority || 'none')
    const dueAt = dueDateTime.trim()
      ? new Date(dueDateTime.trim()).toISOString()
      : null
    onSave({
      title: t,
      description: description.trim(),
      priority: prio,
      tag: tag.trim(),
      due_at: dueAt,
      due_date: dueAt ? dueAt.slice(0, 10) : null,
      ...(isEdit ? {} : { parent_id: parentId ?? null }),
    })
    onClose()
  }

  if (!open) return null

  const modalTitle = isEdit ? 'Edit task' : parentId ? 'New subtask' : 'New task'

  return (
    <div className="modal-bg open" onClick={(e) => e.target === e.currentTarget && onClose()}>
      <div className="modal">
        <div className="modal-ttl">{modalTitle}</div>
        <form onSubmit={handleSubmit}>
          <div className="fg">
            <label className="fl" htmlFor="modal-title">Title</label>
            <input
              ref={inputRef}
              id="modal-title"
              type="text"
              className="fi"
              placeholder="Task name..."
              value={title}
              onChange={(e) => { setTitle(e.target.value); setTitleError(false) }}
              autoComplete="off"
              style={titleError ? { borderColor: '#e05353' } : {}}
            />
          </div>
          <div className="fg">
            <label className="fl" htmlFor="modal-desc">Description</label>
            <textarea
              id="modal-desc"
              className="fta"
              placeholder="Optional details..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
          <div className="fg">
            <label className="fl" htmlFor="modal-prio">Priority</label>
            <select
              id="modal-prio"
              className="fsel"
              value={priority}
              onChange={(e) => setPriority(e.target.value)}
            >
              {PRIORITIES.map((p) => (
                <option key={p.value} value={p.value}>{p.label}</option>
              ))}
            </select>
          </div>
          <div className="fg">
            <label className="fl" htmlFor="modal-due">Due date & time</label>
            <input
              id="modal-due"
              type="datetime-local"
              className="fi fi-datetime"
              value={dueDateTime}
              onChange={(e) => setDueDateTime(e.target.value)}
            />
          </div>
          <div className="fg tag-combo">
            <label className="fl" htmlFor="modal-tag">Tag</label>
            {parentId && tag ? (
              <div className="modal-tag-filled">
                <span className="tag pink">{tag}</span>
              </div>
            ) : (
              <input
                id="modal-tag"
                type="text"
                className="fi fi--tag"
                placeholder="design, dev, research..."
                value={tag}
                onChange={(e) => setTag(e.target.value)}
                list="tags-datalist"
                autoComplete="off"
              />
            )}
            <datalist id="tags-datalist">
              {existingTags.map((t) => (
                <option key={t} value={t} />
              ))}
            </datalist>
          </div>
          <div className="modal-acts">
            <button type="button" className="btn-cancel" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn-save">Save</button>
          </div>
        </form>
      </div>
    </div>
  )
}
