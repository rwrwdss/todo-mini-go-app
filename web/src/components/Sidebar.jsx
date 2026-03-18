const DashboardIcon = () => (
  <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" strokeWidth="1.4">
    <rect x="1" y="1" width="5" height="5" />
    <rect x="8" y="1" width="5" height="5" />
    <rect x="1" y="8" width="5" height="5" />
    <rect x="8" y="8" width="5" height="5" />
  </svg>
)

const TasksIcon = () => (
  <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" strokeWidth="1.4">
    <path d="M1 7h4M1 3h4M1 11h4" />
    <rect x="7" y="1" width="6" height="6" rx="0" />
    <path d="M7 11l2 2 4-4" />
  </svg>
)

const TagIcon = () => (
  <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" strokeWidth="1.4">
    <path d="M7 1l2 4h4l-3 3 1 4-4-2-4 2 1-4-3-3h4z" />
  </svg>
)

function useStoredUser() {
  try {
    const u = localStorage.getItem('user')
    if (!u) return null
    const o = JSON.parse(u)
    return {
      name: o.name || '',
      email: o.email || '',
      initial: (o.name || o.email || 'U').trim().charAt(0).toUpperCase(),
    }
  } catch (_) {
    return null
  }
}

export default function Sidebar({ view, onNavigate, taskCount, tagCounts, onLogout, spaces = [], currentSpaceId, onSpaceSelect, onNewWorkspace }) {
  const tagList = Array.isArray(tagCounts) ? tagCounts : []
  const user = useStoredUser()
  const personalSpaces = Array.isArray(spaces) ? spaces.filter((s) => s.type === 'personal') : []
  const workspaceSpaces = Array.isArray(spaces) ? spaces.filter((s) => s.type === 'corporate') : []

  return (
    <aside className="sidebar">
      <div className="sb-logo">
        <button type="button" className="logo" onClick={() => onNavigate('tree')} style={{ cursor: 'pointer', border: 'none', background: 'none' }}>
          Task<em>.</em>grid
        </button>
      </div>

      <div className="sb-section">
        <div className="sb-label">Overview</div>
        <button type="button" className={`sb-item ${view === 'dashboard' ? 'active' : ''}`} onClick={() => onNavigate('dashboard')}>
          <DashboardIcon />
          Dashboard
        </button>
        <button type="button" className={`sb-item ${view === 'tree' ? 'active' : ''}`} onClick={() => onNavigate('tree')}>
          <TasksIcon />
          My tasks
          {taskCount != null && <span className="sb-badge">{taskCount}</span>}
        </button>
      </div>

      {personalSpaces.length > 0 ? (
        <div className="sb-section">
          <div className="sb-label">My Space</div>
          {personalSpaces.map((s) => (
            <button
              key={s.id}
              type="button"
              className={`sb-item ${currentSpaceId === s.id ? 'active' : ''}`}
              onClick={() => { onSpaceSelect?.(s.id); onNavigate('tree') }}
            >
              {s.name}
            </button>
          ))}
        </div>
      ) : null}

      <div className="sb-section">
        <div className="sb-label">Workspaces</div>
        {workspaceSpaces.map((s) => (
          <button
            key={s.id}
            type="button"
            className={`sb-item ${currentSpaceId === s.id ? 'active' : ''}`}
            onClick={() => { onSpaceSelect?.(s.id); onNavigate('tree') }}
          >
            {s.name}
            <span className="sb-badge">{s.role}</span>
          </button>
        ))}
        {onNewWorkspace ? (
          <button type="button" className="sb-item sb-item-new" onClick={onNewWorkspace}>
            + New workspace
          </button>
        ) : null}
      </div>

      {tagList.length > 0 ? (
        <div className="sb-section">
          <div className="sb-label">By tag</div>
          {tagList.map(({ tag, count }) => (
            <button key={tag || '_none'} type="button" className="sb-item" onClick={() => onNavigate('tree')}>
              <TagIcon />
              {tag || 'No tag'}
              <span className="sb-badge">{count}</span>
            </button>
          ))}
        </div>
      ) : null}

      <div className="sb-footer">
        <div className="sb-avatar" title={user?.email || 'User'}>
          {user ? user.initial : '?'}
        </div>
        <div className="sb-footer-user">
          <div className="sb-user-name">{user?.name || user?.email || 'User'}</div>
          {user?.email && user?.name ? (
            <div className="sb-user-email">{user.email}</div>
          ) : null}
          {onLogout ? (
            <button type="button" className="sb-sign-out" onClick={onLogout}>
              Sign out
            </button>
          ) : (
            <div className="sb-user-plan">Local</div>
          )}
        </div>
      </div>
    </aside>
  )
}
