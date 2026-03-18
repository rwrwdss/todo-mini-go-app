import { updateTodo } from '../api/todos'

const PRIO_CLASS = { high: 'ph', med: 'pm', medium: 'pm', low: 'pl', none: '' }

export default function CorporateSpaceView({
  todos = [],
  spaceDetail = null,
  corpTab = 'board',
  currentUser = null,
  onRefresh,
  onSelectTask,
}) {
  const members = spaceDetail?.members || []
  const todoList = Array.isArray(todos) ? todos : []
  const todo = todoList.filter((t) => !t.done)
  const done = todoList.filter((t) => t.done)

  async function handleToggle(task) {
    try {
      await updateTodo(task.id, { done: !task.done })
      onRefresh?.()
    } catch (_) {}
  }

  function getAssigneeName(task) {
    const id = task.assignee_id ?? task.assigneeId
    if (!id) return null
    const m = members.find((x) => x.id === id)
    return m ? (m.name || m.email || `#${id}`) : null
  }

  function getAssigneeLabel(task) {
    const name = getAssigneeName(task)
    if (!name) return null
    const id = task.assignee_id ?? task.assigneeId
    const isYou = currentUser?.id != null && id === currentUser.id
    return { text: isYou ? 'You' : `Assigned to: ${name}`, title: name }
  }

  return (
    <div className="corporate-view">
      <div className="content">
        {corpTab === 'board' && (
          <div className="view on">
            <div className="board-wrap">
              <div className="board">
                <div className="col">
                  <div className="col-head">
                    <span className="col-name">To do</span>
                    <span className="col-cnt">{todo.length}</span>
                  </div>
                  <div className="col-body">
                    {todo.map((t) => (
                      <div key={t.id} className={`tc ${PRIO_CLASS[t.priority] || ''} ${t.due_date && t.due_date < new Date().toISOString().slice(0, 10) ? 'tc-overdue' : ''}`} onClick={() => onSelectTask?.(t)} role="button" tabIndex={0} onKeyDown={(e) => e.key === 'Enter' && onSelectTask?.(t)}>
                        <div className="tc-title">{t.title}</div>
                        <div className="tc-meta">
                          {t.tag ? <span className="tc-tag">{t.tag}</span> : null}
                          <span className="tc-pri">{t.priority || 'none'}</span>
                          {t.due_date ? <span className="tc-due">{t.due_date}</span> : null}
                        </div>
                        {getAssigneeLabel(t) ? (
                          <div className="tc-assignees" title={getAssigneeLabel(t).title}>
                            <span className="tc-av">{getAssigneeName(t).charAt(0).toUpperCase()}</span>
                            <span className="tc-assignee-label">{getAssigneeLabel(t).text}</span>
                          </div>
                        ) : null}
                      </div>
                    ))}
                  </div>
                </div>
                <div className="col">
                  <div className="col-head">
                    <span className="col-name">Done</span>
                    <span className="col-cnt">{done.length}</span>
                  </div>
                  <div className="col-body">
                    {done.map((t) => (
                      <div key={t.id} className={`tc tc-done ${PRIO_CLASS[t.priority] || ''}`} onClick={() => onSelectTask?.(t)} role="button" tabIndex={0} onKeyDown={(e) => e.key === 'Enter' && onSelectTask?.(t)}>
                        <div className="tc-title done">{t.title}</div>
                        <div className="tc-meta">
                          {t.tag ? <span className="tc-tag">{t.tag}</span> : null}
                          {t.due_date ? <span className="tc-due">{t.due_date}</span> : null}
                          {getAssigneeLabel(t) ? (
                            <div className="tc-assignees" title={getAssigneeLabel(t).title}>
                              <span className="tc-av">{getAssigneeName(t).charAt(0).toUpperCase()}</span>
                              <span className="tc-assignee-label">{getAssigneeLabel(t).text}</span>
                            </div>
                          ) : null}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {corpTab === 'list' && (
          <div className="view on">
            <div className="list-wrap">
              <div className="list-table">
                <div className="lt-head">
                  <div className="lt-hcell">Task</div>
                  <div className="lt-hcell">Tag</div>
                  <div className="lt-hcell">Priority</div>
                  <div className="lt-hcell">Due</div>
                  <div className="lt-hcell">Assignee</div>
                  <div className="lt-hcell">Done</div>
                </div>
                {todoList.map((t) => (
                  <div key={t.id} className={`lt-row ${t.due_date && t.due_date < new Date().toISOString().slice(0, 10) && !t.done ? 'lt-row-overdue' : ''}`} onClick={() => onSelectTask?.(t)} role="button" tabIndex={0} onKeyDown={(e) => e.key === 'Enter' && onSelectTask?.(t)}>
                    <div className="lt-cell title">
                      <span className="lt-title-text">{t.title}</span>
                    </div>
                    <div className="lt-cell"><span className="lt-tag">{t.tag || '—'}</span></div>
                    <div className="lt-cell">{t.priority || 'none'}</div>
                    <div className="lt-cell">{t.due_date || '—'}</div>
                    <div className="lt-cell" title={getAssigneeName(t) || ''}>{currentUser?.id != null && (t.assignee_id ?? t.assigneeId) === currentUser.id ? 'You' : (getAssigneeName(t) || '—')}</div>
                    <div className="lt-cell">
                      <div className={`lt-chk ${t.done ? 'on' : ''}`} />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {corpTab === 'members' && (
          <div className="view on">
            <div className="members-wrap">
              <div className="members-grid">
                {members.map((m) => (
                  <div key={m.id} className="mg-card">
                    <div className="mg-top">
                      <div className="mg-av">{(m.name || m.email || '?').charAt(0).toUpperCase()}</div>
                      <div>
                        <div className="mg-name">{m.name || m.email || '—'}</div>
                        <div className="mg-role">{m.role}</div>
                      </div>
                    </div>
                    {m.email && m.name ? <div className="mg-email">{m.email}</div> : null}
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
