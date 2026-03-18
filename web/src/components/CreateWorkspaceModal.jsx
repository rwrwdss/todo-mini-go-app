import { useState, useEffect, useRef } from 'react'

const NAME_MAX = 40

export default function CreateWorkspaceModal({ open, onCreate, onClose }) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [error, setError] = useState('')
  const inputRef = useRef(null)

  useEffect(() => {
    if (open) {
      setName('')
      setDescription('')
      setError('')
      setTimeout(() => inputRef.current?.focus(), 50)
    }
  }, [open])

  function handleSubmit(e) {
    e.preventDefault()
    const trimmed = name.trim()
    if (!trimmed) {
      setError('Workspace name is required')
      return
    }
    if (trimmed.length > NAME_MAX) {
      setError(`Name must be at most ${NAME_MAX} characters`)
      return
    }
    setError('')
    onCreate(trimmed)
    onClose()
  }

  if (!open) return null

  return (
    <div className="modal-bg open" onClick={(e) => e.target === e.currentTarget && onClose()}>
      <div className="modal modal-create-workspace">
        <div className="create-ws-header">
          <span className="create-ws-label">New workspace</span>
          <h2 className="create-ws-title">Create workspace</h2>
        </div>
        <form onSubmit={handleSubmit}>
          <div className="fg">
            <label className="fl" htmlFor="create-ws-name">Workspace name</label>
            <div className="create-ws-input-wrap">
              <input
                ref={inputRef}
                id="create-ws-name"
                type="text"
                className="fi"
                placeholder="e.g. Product team, Q2 launch..."
                value={name}
                onChange={(e) => { setName(e.target.value.slice(0, NAME_MAX)); setError('') }}
                autoComplete="off"
                maxLength={NAME_MAX}
                style={error ? { borderColor: 'var(--red)' } : {}}
              />
              <span className="create-ws-count">{name.length}/{NAME_MAX}</span>
            </div>
          </div>
          <div className="fg">
            <label className="fl" htmlFor="create-ws-desc">Description (optional)</label>
            <textarea
              id="create-ws-desc"
              className="fi fta"
              placeholder="What is this workspace for?"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
            />
          </div>
          {error ? <p className="modal-error" style={{ color: 'var(--red)', fontSize: 12, marginBottom: 12 }}>{error}</p> : null}
          <div className="modal-footer">
            <button type="button" className="bc" onClick={onClose}>Cancel</button>
            <button type="submit" className="btn-save">Create</button>
          </div>
        </form>
      </div>
    </div>
  )
}
