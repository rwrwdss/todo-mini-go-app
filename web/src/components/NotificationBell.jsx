import { useState, useEffect, useRef } from 'react'
import { getNotifications, markNotificationRead } from '../api/notifications'

function formatNotificationTime(iso) {
  if (!iso) return ''
  try {
    const d = new Date(iso)
    const now = new Date()
    const diffMs = now - d
    const diffM = Math.floor(diffMs / 60000)
    if (diffM < 1) return 'Just now'
    if (diffM < 60) return `${diffM}m ago`
    const diffH = Math.floor(diffM / 60)
    if (diffH < 24) return `${diffH}h ago`
    return d.toLocaleDateString()
  } catch {
    return iso
  }
}

export default function NotificationBell({ onOpenTask }) {
  const [open, setOpen] = useState(false)
  const [list, setList] = useState([])
  const [loading, setLoading] = useState(false)
  const ref = useRef(null)

  function fetchList() {
    setLoading(true)
    getNotifications()
      .then((data) => setList(Array.isArray(data) ? data : []))
      .catch(() => setList([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    if (open) fetchList()
  }, [open])

  useEffect(() => {
    if (!open) return
    const onOutside = (e) => {
      if (ref.current && !ref.current.contains(e.target)) setOpen(false)
    }
    document.addEventListener('click', onOutside)
    return () => document.removeEventListener('click', onOutside)
  }, [open])

  const unreadCount = list.filter((n) => !n.read_at).length

  function handleMarkRead(n) {
    if (n.read_at) return
    markNotificationRead(n.id).then(() => fetchList())
  }

  function handleClickItem(n) {
    handleMarkRead(n)
    onOpenTask?.(n.todo_id)
    setOpen(false)
  }

  return (
    <div className="notification-bell" ref={ref}>
      <button
        type="button"
        className="notification-bell-btn"
        onClick={() => setOpen((o) => !o)}
        title="Notifications"
        aria-label={unreadCount ? `${unreadCount} unread notifications` : 'Notifications'}
      >
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
          <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
          <path d="M13.73 21a2 2 0 0 1-3.46 0" />
        </svg>
        {unreadCount > 0 ? (
          <span className="notification-bell-badge">{unreadCount > 99 ? '99+' : unreadCount}</span>
        ) : null}
      </button>
      {open && (
        <div className="notification-bell-dropdown">
          <div className="notification-bell-header">Notifications</div>
          {loading ? (
            <p className="notification-bell-loading">Loading…</p>
          ) : list.length === 0 ? (
            <p className="notification-bell-empty">No notifications</p>
          ) : (
            <ul className="notification-bell-list">
              {list.map((n) => (
                <li
                  key={n.id}
                  className={`notification-bell-item ${n.read_at ? '' : 'unread'}`}
                  role="button"
                  tabIndex={0}
                  onClick={() => handleClickItem(n)}
                  onKeyDown={(e) => e.key === 'Enter' && handleClickItem(n)}
                >
                  <span className="notification-bell-item-type">{n.type === 'overdue' ? 'Overdue' : 'Due soon'}</span>
                  <span className="notification-bell-item-title">"{n.title || 'Task'}"</span>
                  <span className="notification-bell-item-time">{formatNotificationTime(n.created_at)}</span>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  )
}
