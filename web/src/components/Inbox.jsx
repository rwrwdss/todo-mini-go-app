import { useEffect, useRef, useState } from 'react'
import { acceptInvitation, declineInvitation } from '../api/invitations'
import { archiveNotification, getNotifications, markNotificationRead, markNotificationUnread } from '../api/notifications'

const POLL_INTERVAL_MS = 15000

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

function notificationTypeLabel(type) {
  switch (type) {
    case 'space_invitation':
      return 'Invitation'
    case 'task_overdue':
      return 'Overdue'
    case 'task_due_soon':
      return 'Due soon'
    case 'task_assigned':
      return 'Assigned'
    case 'task_created':
      return 'New task'
    default:
      return 'Notification'
  }
}

export default function Inbox({ onOpenTask, onOpenSpace, onAcceptInvitation }) {
  const [open, setOpen] = useState(false)
  const [list, setList] = useState([])
  const [loading, setLoading] = useState(false)
  const [actingId, setActingId] = useState(null)
  const ref = useRef(null)

  function fetchList(showLoading = true) {
    if (showLoading) setLoading(true)
    getNotifications()
      .then((data) => setList(Array.isArray(data) ? data : []))
      .catch(() => setList([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    if (open) fetchList()
  }, [open])

  useEffect(() => {
    const t = setInterval(() => fetchList(false), POLL_INTERVAL_MS)
    return () => clearInterval(t)
  }, [])

  useEffect(() => {
    if (!open) return
    const onOutside = (e) => {
      if (ref.current && !ref.current.contains(e.target)) setOpen(false)
    }
    document.addEventListener('click', onOutside)
    return () => document.removeEventListener('click', onOutside)
  }, [open])

  const unreadCount = list.filter((n) => !n.read_at).length

  function handleClickItem(n) {
    if (n.type === 'space_invitation') return
    if (!n.read_at) {
      markNotificationRead(n.id).catch(() => {})
    }
    if (n.todo_id) onOpenTask?.(n.todo_id)
    if (!n.todo_id && n.space_id) onOpenSpace?.(n.space_id)
    setOpen(false)
  }

  async function handleAccept(n) {
    if (!n.invitation_id || actingId) return
    setActingId(n.id)
    try {
      await acceptInvitation(n.invitation_id)
      await markNotificationRead(n.id).catch(() => {})
      onAcceptInvitation?.()
      fetchList(false)
    } catch {
      fetchList(false)
    } finally {
      setActingId(null)
    }
  }

  async function handleDecline(n) {
    if (!n.invitation_id || actingId) return
    setActingId(n.id)
    try {
      await declineInvitation(n.invitation_id)
      await markNotificationRead(n.id).catch(() => {})
      fetchList(false)
    } catch {
      fetchList(false)
    } finally {
      setActingId(null)
    }
  }

  async function handleArchive(n) {
    if (actingId) return
    setActingId(n.id)
    try {
      await archiveNotification(n.id)
      fetchList(false)
    } finally {
      setActingId(null)
    }
  }

  async function handleToggleRead(n) {
    if (actingId) return
    setActingId(n.id)
    try {
      if (n.read_at) await markNotificationUnread(n.id)
      else await markNotificationRead(n.id)
      fetchList(false)
    } finally {
      setActingId(null)
    }
  }

  return (
    <div className="notification-bell" ref={ref}>
      <button
        type="button"
        className="inbox-btn"
        onClick={() => setOpen((o) => !o)}
        title="Inbox"
        aria-label={unreadCount ? `${unreadCount} unread notifications` : 'Inbox'}
      >
        Inbox
        {unreadCount > 0 ? (
          <span className="notification-bell-badge">{unreadCount > 99 ? '99+' : unreadCount}</span>
        ) : null}
      </button>
      {open && (
        <div className="notification-bell-dropdown">
          <div className="notification-bell-header">Inbox</div>
          {loading ? (
            <p className="notification-bell-loading">Loading...</p>
          ) : list.length === 0 ? (
            <p className="notification-bell-empty">No notifications</p>
          ) : (
            <ul className="notification-bell-list">
              {list.map((n) => (
                <li
                  key={n.id}
                  className={`notification-bell-item ${n.read_at ? '' : 'unread'} ${n.type === 'space_invitation' ? 'notification-bell-item-invitation' : ''}`}
                  role={n.type === 'space_invitation' ? undefined : 'button'}
                  tabIndex={n.type === 'space_invitation' ? undefined : 0}
                  onClick={n.type === 'space_invitation' ? undefined : () => handleClickItem(n)}
                  onKeyDown={n.type === 'space_invitation' ? undefined : (e) => e.key === 'Enter' && handleClickItem(n)}
                >
                  {n.type === 'space_invitation' ? (
                    <>
                      <span className="notification-bell-item-type">{notificationTypeLabel(n.type)}</span>
                      <span className="notification-bell-item-title">Invitation to space "{n.space_name || 'Workspace'}"</span>
                      <span className="notification-bell-item-time">{formatNotificationTime(n.created_at)}</span>
                      <div className="notification-bell-item-actions" onClick={(e) => e.stopPropagation()}>
                        <button type="button" className="notification-bell-btn-accept" onClick={() => handleAccept(n)} disabled={actingId !== null}>
                          Accept
                        </button>
                        <button type="button" className="notification-bell-btn-decline" onClick={() => handleDecline(n)} disabled={actingId !== null}>
                          Decline
                        </button>
                      </div>
                    </>
                  ) : (
                    <>
                      <span className="notification-bell-item-type">{notificationTypeLabel(n.type)}</span>
                      <span className="notification-bell-item-title">"{n.title || 'Task'}"</span>
                      <span className="notification-bell-item-time">{formatNotificationTime(n.created_at)}</span>
                      <div className="notification-bell-item-actions" onClick={(e) => e.stopPropagation()}>
                        <button type="button" className="notification-bell-btn-decline" onClick={() => handleToggleRead(n)} disabled={actingId !== null}>
                          {n.read_at ? 'Unread' : 'Read'}
                        </button>
                        <button type="button" className="notification-bell-btn-decline" onClick={() => handleArchive(n)} disabled={actingId !== null}>
                          Archive
                        </button>
                      </div>
                    </>
                  )}
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  )
}
