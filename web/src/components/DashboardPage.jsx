import { useMemo } from 'react'

const PRIORITY_ORDER = ['high', 'med', 'low', 'none']

function norm(p) {
  const s = (p || 'none').toLowerCase()
  return s === 'medium' ? 'med' : s
}

export default function DashboardPage({ todos, loading, onViewAll, onNewTask }) {
  const stats = useMemo(() => {
    const total = todos.length
    const doneCount = todos.filter((t) => t.done).length
    const pendingCount = total - doneCount
    const byPriority = { high: 0, med: 0, low: 0, none: 0 }
    todos.forEach((t) => {
      const key = norm(t.priority)
      if (byPriority[key] !== undefined) byPriority[key]++
      else byPriority.none++
    })
    const recent = [...todos].sort((a, b) => (b.id || 0) - (a.id || 0)).slice(0, 6)
    const byTag = new Map()
    todos.forEach((t) => {
      const tag = (t.tag || '').trim() || null
      const key = tag || '_none'
      if (!byTag.has(key)) byTag.set(key, { tag: tag || '', total: 0, done: 0 })
      const g = byTag.get(key)
      g.total++
      if (t.done) g.done++
    })
    const workspaces = [...byTag.entries()].map(([key, g]) => ({
      tag: g.tag || 'No tag',
      count: g.total,
      pct: g.total ? Math.round((g.done / g.total) * 100) : 0,
    }))
    return {
      total,
      doneCount,
      pendingCount,
      byPriority,
      recent,
      workspaces,
    }
  }, [todos])

  const donutGradient = useMemo(() => {
    const { byPriority } = stats
    const total = stats.total || 1
    let acc = 0
    const parts = PRIORITY_ORDER.map((p) => {
      const n = byPriority[p] || 0
      const pct = (n / total) * 100
      const start = acc
      acc += pct
      return { p, start, end: acc }
    })
    return parts
      .map(({ p, start, end }) => {
        const color = p === 'high' ? 'var(--pink)' : p === 'med' ? 'var(--yellow)' : p === 'low' ? 'var(--green)' : 'var(--surface2)'
        return `${color} ${start}% ${end}%`
      })
      .join(', ')
  }, [stats])

  if (loading && todos.length === 0) {
    return (
      <div className="dash-content">
        <p style={{ color: 'var(--text3)' }}>Loading…</p>
      </div>
    )
  }

  const maxBar = Math.max(1, ...PRIORITY_ORDER.map((p) => stats.byPriority[p] || 0))

  return (
    <div className="dash-content">
      <div className="kpi-row">
        <div className="kpi kpi-pk">
          <div className="kpi-label">Total tasks</div>
          <div className="kpi-val">{stats.total}</div>
          <div className="kpi-delta neu">—</div>
        </div>
        <div className="kpi kpi-gr">
          <div className="kpi-label">Completed</div>
          <div className="kpi-val">{stats.doneCount}</div>
          <div className="kpi-delta neu">—</div>
        </div>
        <div className="kpi kpi-yw">
          <div className="kpi-label">In progress</div>
          <div className="kpi-val">{stats.pendingCount}</div>
          <div className="kpi-delta neu">—</div>
        </div>
        <div className="kpi">
          <div className="kpi-label">Overdue</div>
          <div className="kpi-val">0</div>
          <div className="kpi-delta neu">No due dates</div>
        </div>
      </div>

      <div className="charts-row">
        <div className="chart-card">
          <div className="card-title">Tasks by priority <span>current</span></div>
          <div className="bar-chart">
            {PRIORITY_ORDER.map((p) => {
              const n = stats.byPriority[p] || 0
              const h = maxBar ? (n / maxBar) * 80 + 4 : 4
              const isMax = n === maxBar && n > 0
              return (
                <div key={p} className="bar-group">
                  <div className={`bar ${isMax ? 'pk' : ''}`} style={{ height: `${h}px` }} title={`${p}: ${n}`} />
                  <div className="bar-lbl">{p}</div>
                </div>
              )
            })}
          </div>
        </div>
        <div className="chart-card">
          <div className="card-title">Tasks by priority <span>share</span></div>
          <div className="donut-wrap">
            <div className="donut" style={{ background: `conic-gradient(${donutGradient})` }} />
            <div className="donut-legend">
              {PRIORITY_ORDER.map((p) => {
                const n = stats.byPriority[p] || 0
                const pct = stats.total ? Math.round((n / stats.total) * 100) : 0
                const color = p === 'high' ? 'var(--pink)' : p === 'med' ? 'var(--yellow)' : p === 'low' ? 'var(--green)' : 'var(--surface2)'
                return (
                  <div key={p} className="dl-item">
                    <div className="dl-dot" style={{ background: color }} />
                    <span className="dl-text">{p}</span>
                    <span className="dl-val">{pct}%</span>
                  </div>
                )
              })}
            </div>
          </div>
        </div>
      </div>

      <div className="table-row">
        <div className="table-card">
          <div className="tc-head">
            <div className="tc-title">Recent tasks</div>
            <button type="button" className="tc-link" onClick={onViewAll}>View all</button>
          </div>
          {stats.recent.length === 0 ? (
            <div style={{ padding: '18px 22px', color: 'var(--text3)', fontSize: 12 }}>No tasks yet</div>
          ) : (
            stats.recent.map((t) => {
              const pri = norm(t.priority)
              return (
                <div key={t.id} className="dash-task-row" onClick={onViewAll} role="button" tabIndex={0} onKeyDown={(e) => e.key === 'Enter' && onViewAll()}>
                  <div className={`tr-check ${t.done ? 'done' : ''}`} />
                  <div className={`tr-title ${t.done ? 'done' : ''}`}>{t.title || 'Untitled'}</div>
                  <div className="tr-tags">
                    {t.tag ? <span className="tr-tag pk">{t.tag}</span> : null}
                  </div>
                  <div className={`tr-pri ${pri === 'high' ? 'h' : pri === 'med' ? 'm' : pri === 'low' ? 'l' : ''}`} title={pri} />
                </div>
              )
            })
          )}
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', background: 'var(--border)', gap: 1 }}>
          <div className="table-card" style={{ flex: 1 }}>
            <div className="tc-head">
              <div className="tc-title">Activity feed</div>
            </div>
            <div className="act-row">
              <div className="act-dot gr" />
              <div className="act-body">
                <div className="act-text"><b>You</b> can complete tasks in My tasks</div>
                <div className="act-time">—</div>
              </div>
            </div>
            <div className="act-row">
              <div className="act-dot pk" />
              <div className="act-body">
                <div className="act-text"><b>Dashboard</b> shows live stats from your tasks</div>
                <div className="act-time">—</div>
              </div>
            </div>
            <div className="act-row">
              <div className="act-dot yw" />
              <div className="act-body">
                <div className="act-text">Use tags to group tasks</div>
                <div className="act-time">—</div>
              </div>
            </div>
          </div>

          <div className="table-card">
            <div className="tc-head">
              <div className="tc-title">By tag</div>
            </div>
            {stats.workspaces.length === 0 ? (
              <div style={{ padding: '12px 22px', color: 'var(--text3)', fontSize: 12 }}>No tags yet</div>
            ) : (
              stats.workspaces.map((w) => (
                <div key={w.tag} className="ws-row">
                  <div className="ws-icon">{(w.tag || 'N').charAt(0).toUpperCase()}</div>
                  <div style={{ flex: 1 }}>
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                      <div className="ws-name">{w.tag}</div>
                      <div className="ws-count">{w.count} tasks</div>
                    </div>
                    <div style={{ marginTop: 5, display: 'flex', alignItems: 'center', gap: 8 }}>
                      <div className="prog-bar" style={{ flex: 1 }}>
                        <div className="prog-fill" style={{ width: `${w.pct}%` }} />
                      </div>
                      <span style={{ fontFamily: "'DM Mono', monospace", fontSize: 8, color: 'var(--text3)' }}>{w.pct}%</span>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
