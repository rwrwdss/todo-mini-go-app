import { useState, useEffect, useRef } from 'react'

export default function InviteMemberModal({ open, onInvite, onClose }) {
  const [email, setEmail] = useState('')
  const [role, setRole] = useState('member')
  const [error, setError] = useState('')
  const inputRef = useRef(null)

  useEffect(() => {
    if (open) {
      setEmail('')
      setRole('member')
      setError('')
      setTimeout(() => inputRef.current?.focus(), 50)
    }
  }, [open])

  async function handleSubmit(e) {
    e.preventDefault()
    const em = email.trim().toLowerCase()
    if (!em) {
      setError('Email is required')
      return
    }
    setError('')
    try {
      await onInvite(em, role)
      onClose()
    } catch (err) {
      setError(err.message || 'Failed to invite')
    }
  }

  if (!open) return null

  return (
    <div className="modal-bg open" onClick={(e) => e.target === e.currentTarget && onClose()}>
      <div className="modal">
        <div className="modal-ttl">Invite member</div>
        <form onSubmit={handleSubmit}>
          <div className="fg">
            <label className="fl" htmlFor="invite-email">Email</label>
            <input
              ref={inputRef}
              id="invite-email"
              type="email"
              className="fi"
              placeholder="colleague@example.com"
              value={email}
              onChange={(e) => { setEmail(e.target.value); setError('') }}
              autoComplete="email"
              style={error ? { borderColor: 'var(--red)' } : {}}
            />
          </div>
          <div className="fg">
            <label className="fl" htmlFor="invite-role">Role</label>
            <select id="invite-role" className="fi" value={role} onChange={(e) => setRole(e.target.value)}>
              <option value="member">Member</option>
              <option value="admin">Admin</option>
            </select>
          </div>
          {error ? <p className="modal-error" style={{ color: 'var(--red)', fontSize: 12, marginBottom: 12 }}>{error}</p> : null}
          <div className="modal-footer">
            <button type="button" className="bc" onClick={onClose}>Cancel</button>
            <button type="submit" className="btn-save">Invite</button>
          </div>
        </form>
      </div>
    </div>
  )
}
