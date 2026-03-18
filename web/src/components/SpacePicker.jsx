import { useState, useRef, useEffect } from 'react'

export default function SpacePicker({ spaces = [], currentSpaceId, onSelect, className = '' }) {
  const [open, setOpen] = useState(false)
  const ref = useRef(null)
  const current = spaces.find((s) => s.id === currentSpaceId) || spaces[0]

  useEffect(() => {
    if (!open) return
    const onOutside = (e) => {
      if (ref.current && !ref.current.contains(e.target)) setOpen(false)
    }
    document.addEventListener('click', onOutside)
    return () => document.removeEventListener('click', onOutside)
  }, [open])

  return (
    <div className={`space-picker ${className}`} ref={ref}>
      <button
        type="button"
        className="tb-ws space-picker-btn"
        onClick={() => setOpen((o) => !o)}
        title={current?.name || 'Select space'}
      >
        <span className="space-picker-label">{current?.name || 'My Space'}</span>
        <svg width="10" height="10" viewBox="0 0 10 10" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ opacity: open ? 1 : 0.7 }}>
          <path d="M2 4l3 3 3-3" />
        </svg>
      </button>
      {open && (
        <div className="space-picker-dropdown">
          {spaces.map((s) => (
            <button
              key={s.id}
              type="button"
              className={`space-picker-item ${s.id === currentSpaceId ? 'active' : ''}`}
              onClick={() => {
                onSelect(s.id)
                setOpen(false)
              }}
            >
              <span>{s.name}</span>
              <span className="space-picker-type">{s.type === 'personal' ? 'Personal' : s.role}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
