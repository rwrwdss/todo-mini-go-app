import { useEffect, useState } from 'react'
import { getSpaceActivity } from '../api/activity'

function formatEvent(item) {
  const actor = item.actor_name || (item.actor_id ? `#${item.actor_id}` : 'System')
  const subject = item.subject_name || (item.subject_user_id ? `#${item.subject_user_id}` : 'user')
  switch (item.event_type) {
    case 'todo_created':
      return `${actor} created task`
    case 'todo_assigned':
      return `${actor} assigned task to ${subject}`
    case 'todo_done_toggled':
      return `${actor} changed task status`
    case 'todo_updated':
      return `${actor} updated task`
    case 'member_joined_space':
      return `${subject} joined space`
    default:
      return `${actor} -> ${item.event_type}`
  }
}

export default function SpaceActivityView({ spaceId, active }) {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!active || !spaceId) return undefined
    let cancelled = false
    const load = async () => {
      setLoading(true)
      try {
        const data = await getSpaceActivity(spaceId, { limit: 200 })
        if (!cancelled) {
          setItems(Array.isArray(data) ? data : [])
          setError('')
        }
      } catch (e) {
        if (!cancelled) setError(e?.message || 'Failed to load activity')
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    load()
    const t = setInterval(load, 15000)
    return () => {
      cancelled = true
      clearInterval(t)
    }
  }, [spaceId, active])

  return (
    <div className="members-wrap">
      {loading ? <p className="tree-other-empty">Loading...</p> : null}
      {error ? <p className="tree-other-empty tree-other-error">{error}</p> : null}
      {!loading && !error && items.length === 0 ? <p className="tree-other-empty">No activity yet</p> : null}
      <div className="activity-list">
        {items.map((item) => (
          <div key={item.id} className="activity-item">
            <div className="activity-text">{formatEvent(item)}</div>
            <div className="activity-meta">
              {item.todo_id ? `Task #${item.todo_id}` : 'Space event'} · {item.created_at ? new Date(item.created_at).toLocaleString() : ''}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
