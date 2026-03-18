import { useState, useEffect, useCallback, useRef } from 'react'
import { getTodos, createTodo, updateTodo, deleteTodo } from './api/todos'
import { getSpaces, getSpace, createSpace, inviteMember } from './api/spaces'
import { checkSession, logout } from './api/auth'
import { buildTree, groupRootsByTag, groupRootsByPriority, positionHLines, countDescendants } from './utils/tree'
import Sidebar from './components/Sidebar'
import SpacePicker from './components/SpacePicker'
import TreeNode from './components/TreeNode'
import Modal from './components/Modal'
import ConfirmDeleteModal from './components/ConfirmDeleteModal'
import Empty from './components/Empty'
import StatusBar from './components/StatusBar'
import DashboardPage from './components/DashboardPage'
import CorporateSpaceView from './components/CorporateSpaceView'
import AssignTaskModal from './components/AssignTaskModal'
import InviteMemberModal from './components/InviteMemberModal'
import CreateWorkspaceModal from './components/CreateWorkspaceModal'
import TaskDetailPanel from './components/TaskDetailPanel'
import NotificationBell from './components/NotificationBell'
import './index.css'

export default function App({ onLogout }) {
  const [spaces, setSpaces] = useState([])
  const [currentSpaceId, setCurrentSpaceId] = useState(null)
  const [todos, setTodos] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [modal, setModal] = useState({
    open: false,
    mode: 'create',
    editTask: null,
    parentId: null,
    parentTag: '',
  })
  const [deleteConfirm, setDeleteConfirm] = useState({ open: false, task: null })
  const [view, setView] = useState('tree')
  const treeRootRef = useRef(null)
  const [pan, setPan] = useState({ x: 0, y: 0 })
  const [isPanning, setIsPanning] = useState(false)
  const panStartRef = useRef(null)
  const PAN_MAX = 320

  const currentSpace = spaces.find((s) => s.id === currentSpaceId) || spaces[0]
  const isCorporate = currentSpace?.type === 'corporate'
  const [spaceDetail, setSpaceDetail] = useState(null)
  const [assignModalOpen, setAssignModalOpen] = useState(false)
  const [inviteModalOpen, setInviteModalOpen] = useState(false)
  const [createWorkspaceModalOpen, setCreateWorkspaceModalOpen] = useState(false)
  const [corpTab, setCorpTab] = useState('board')
  const [selectedTaskId, setSelectedTaskId] = useState(null)

  useEffect(() => {
    if (isCorporate && currentSpaceId) {
      getSpace(currentSpaceId)
        .then(setSpaceDetail)
        .catch(() => setSpaceDetail(null))
    } else {
      setSpaceDetail(null)
    }
  }, [isCorporate, currentSpaceId])

  const refreshSpaces = useCallback(() => {
    getSpaces()
      .then((list) => {
        const arr = Array.isArray(list) ? list : []
        setSpaces(arr)
        setCurrentSpaceId((prev) => {
          if (prev != null && arr.some((s) => s.id === prev)) return prev
          const personal = arr.find((s) => s.type === 'personal')
          return personal ? personal.id : arr[0]?.id ?? null
        })
      })
      .catch(() => setSpaces([]))
  }, [])

  useEffect(() => {
    refreshSpaces()
  }, [])

  // Session check every 5 min; on 401 clear token and reload to show login
  useEffect(() => {
    const interval = setInterval(async () => {
      const user = await checkSession()
      if (user == null) {
        logout()
        window.location.reload()
      }
    }, 5 * 60 * 1000)
    return () => clearInterval(interval)
  }, [])

  function openCreateWorkspaceModal() {
    setCreateWorkspaceModalOpen(true)
  }

  async function handleCreateWorkspace(name) {
    try {
      setError(null)
      const space = await createSpace(name)
      setSpaces((prev) => (Array.isArray(prev) ? [...prev, space] : [space]))
      setCurrentSpaceId(space.id)
      setCreateWorkspaceModalOpen(false)
    } catch (e) {
      setError(e.message || 'Failed to create workspace')
    }
  }

  const loadTodos = useCallback(async (background = false) => {
    if (!background) {
      setLoading(true)
      setError(null)
    }
    try {
      const spaceId = currentSpaceId != null ? currentSpaceId : undefined
      const data = await getTodos(spaceId)
      setTodos(Array.isArray(data) ? data : [])
    } catch (e) {
      setError(e.message || 'Failed to load tasks')
    } finally {
      setLoading(false)
    }
  }, [currentSpaceId])

  useEffect(() => {
    loadTodos()
  }, [loadTodos])

  // Corporate space: poll tasks every 4s when tab is visible
  useEffect(() => {
    if (!isCorporate || !currentSpaceId) return
    let tid
    const poll = () => {
      if (document.visibilityState !== 'visible') return
      loadTodos(true)
    }
    const onVis = () => {
      if (document.visibilityState === 'visible') {
        poll()
        tid = setInterval(poll, 4000)
      } else {
        clearInterval(tid)
      }
    }
    onVis()
    document.addEventListener('visibilitychange', onVis)
    return () => {
      clearInterval(tid)
      document.removeEventListener('visibilitychange', onVis)
    }
  }, [isCorporate, currentSpaceId, loadTodos])

  useEffect(() => {
    if (!treeRootRef.current || loading) return
    positionHLines(treeRootRef.current)
    const t = setTimeout(() => positionHLines(treeRootRef.current), 120)
    return () => clearTimeout(t)
  }, [todos, loading])

  useEffect(() => {
    const onResize = () => positionHLines(treeRootRef.current)
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [])

  const handleCanvasMouseDown = useCallback((e) => {
    if (view !== 'tree') return
    const isInteractive = e.target.closest('.node-wrap, .root-card, button, a, .add-child-btn, .tag, .tag-group-label, .tr-check, .tr-title, .tr-tag')
    if (isInteractive) return
    panStartRef.current = { clientX: e.clientX, clientY: e.clientY, panX: pan.x, panY: pan.y }
    setIsPanning(true)
  }, [view, pan.x, pan.y])

  useEffect(() => {
    if (!isPanning) return
    const onMove = (e) => {
      if (!panStartRef.current) return
      const dx = e.clientX - panStartRef.current.clientX
      const dy = e.clientY - panStartRef.current.clientY
      const newX = Math.max(-PAN_MAX, Math.min(PAN_MAX, panStartRef.current.panX + dx))
      const newY = Math.max(-PAN_MAX, Math.min(PAN_MAX, panStartRef.current.panY + dy))
      setPan({ x: newX, y: newY })
    }
    const onUp = () => {
      panStartRef.current = null
      setIsPanning(false)
    }
    window.addEventListener('mousemove', onMove)
    window.addEventListener('mouseup', onUp)
    return () => {
      window.removeEventListener('mousemove', onMove)
      window.removeEventListener('mouseup', onUp)
    }
  }, [isPanning])

  useEffect(() => {
    const onKeyDown = (e) => {
      if (e.key === 'Escape') {
        if (deleteConfirm.open) {
          setDeleteConfirm({ open: false, task: null })
        } else {
          setModal((m) => ({ ...m, open: false }))
        }
      }
      if ((e.ctrlKey || e.metaKey) && e.key === 'Enter' && modal.open) {
        e.preventDefault()
        document.querySelector('.modal .btn-save')?.click()
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [modal.open, deleteConfirm.open])

  const tree = buildTree(todos)
  const tagGroups = groupRootsByTag(tree)
  const existingTags = [...new Set(todos.map((t) => (t.tag || '').trim()).filter(Boolean))].sort()
  const total = todos.length
  const doneCount = todos.filter((t) => t.done).length
  const pendingCount = total - doneCount

  function openNewTask() {
    if (view === 'dashboard') setView('tree')
    setModal({ open: true, mode: 'create', editTask: null, parentId: null })
  }

  function openEdit(task) {
    setModal({ open: true, mode: 'edit', editTask: task, parentId: null })
  }

  function openAddSub(parent) {
    setModal({
      open: true,
      mode: 'create',
      editTask: null,
      parentId: parent.id,
      parentTag: (parent.tag || '').trim(),
    })
  }

  function closeModal() {
    setModal((m) => ({ ...m, open: false }))
  }

  async function handleSave(payload) {
    setError(null)
    try {
      const body = { ...payload }
      if (modal.mode !== 'edit' && currentSpaceId != null) body.space_id = currentSpaceId
      if (modal.mode === 'edit' && modal.editTask) {
        await updateTodo(modal.editTask.id, payload)
      } else {
        await createTodo(body)
      }
      await loadTodos(true)
    } catch (e) {
      setError(e.message || 'Failed to save')
    }
  }

  async function handleToggleDone(task) {
    setError(null)
    try {
      await updateTodo(task.id, { done: !task.done })
      await loadTodos(true)
    } catch (e) {
      setError(e.message || 'Failed to update')
    }
  }

  function isRoot(task) {
    const pid = task.parent_id ?? null
    return pid == null || pid === 0
  }

  async function performDelete(task) {
    setError(null)
    try {
      await deleteTodo(task.id)
      await loadTodos(true)
    } catch (e) {
      setError(e.message || 'Failed to delete')
    }
  }

  function handleDelete(task) {
    if (isRoot(task)) {
      setDeleteConfirm({ open: true, task })
      return
    }
    performDelete(task)
  }

  function handleConfirmDelete(task) {
    if (task) performDelete(task)
    setSelectedTaskId(null)
    setDeleteConfirm({ open: false, task: null })
  }

  const tagCounts = tagGroups.map((g) => ({ tag: g.tag || '', count: g.nodes.length }))

  if (loading && todos.length === 0) {
    return (
      <div className="app-layout">
        <Sidebar view="tree" onNavigate={setView} taskCount={0} tagCounts={[]} onLogout={onLogout} spaces={spaces} currentSpaceId={currentSpaceId} onSpaceSelect={setCurrentSpaceId} onNewWorkspace={openCreateWorkspaceModal} />
        <div className="main">
          <div className="topbar">
            <div className="tb-left">
              <div className="tb-title">My tasks</div>
              <SpacePicker spaces={spaces} currentSpaceId={currentSpaceId} onSelect={setCurrentSpaceId} />
            </div>
            <div className="tb-right">
              <span className="task-count">0 tasks</span>
              <button type="button" className="btn-sm" onClick={openNewTask}>+ New task</button>
            </div>
          </div>
          <div className="canvas">
            <p style={{ color: 'var(--text3)' }}>Loading…</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="app-layout">
      <Sidebar view={view} onNavigate={setView} taskCount={total} tagCounts={tagCounts} onLogout={onLogout} spaces={spaces} currentSpaceId={currentSpaceId} onSpaceSelect={setCurrentSpaceId} onNewWorkspace={openCreateWorkspaceModal} />
      <div className="main">
        <div className="topbar">
          <div className="tb-left">
            <div className="tb-title">{view === 'dashboard' ? 'Dashboard' : 'My tasks'}</div>
            <SpacePicker spaces={spaces} currentSpaceId={currentSpaceId} onSelect={setCurrentSpaceId} />
          </div>
          <div className="tb-right">
            {isCorporate ? (
              <>
                <div className="vtabs">
                  <button type="button" className={`vtab ${corpTab === 'board' ? 'on' : ''}`} onClick={() => setCorpTab('board')}>Board</button>
                  <button type="button" className={`vtab ${corpTab === 'list' ? 'on' : ''}`} onClick={() => setCorpTab('list')}>List</button>
                  <button type="button" className={`vtab ${corpTab === 'members' ? 'on' : ''}`} onClick={() => setCorpTab('members')}>Members</button>
                </div>
                {currentSpace?.role === 'admin' ? (
                  <>
                    <button type="button" className="btn-ghost" onClick={() => setInviteModalOpen(true)}>+ Invite</button>
                    <button type="button" className="btn-sm btn-pk" onClick={() => setAssignModalOpen(true)}>+ Assign task</button>
                  </>
                ) : null}
              </>
            ) : (
              <>
                <span className="task-count">{total === 1 ? '1 task' : `${total} tasks`}</span>
                <button type="button" className="btn-sm" onClick={openNewTask}>+ New task</button>
              </>
            )}
            <NotificationBell onOpenTask={(todoId) => setSelectedTaskId(todoId)} />
          </div>
        </div>

        {view === 'dashboard' ? (
          <DashboardPage
            todos={todos}
            loading={loading}
            onViewAll={() => setView('tree')}
          />
        ) : isCorporate ? (
          <>
            <CorporateSpaceView
              todos={todos}
              spaceDetail={spaceDetail}
              corpTab={corpTab}
              onRefresh={() => loadTodos(true)}
              onSelectTask={(task) => setSelectedTaskId(task?.id ?? null)}
            />
            <AssignTaskModal
              open={assignModalOpen}
              spaceId={currentSpaceId}
              members={spaceDetail?.members || []}
              existingTags={existingTags}
              onSave={async (body) => {
                await createTodo(body)
                setAssignModalOpen(false)
                loadTodos(true)
              }}
              onClose={() => setAssignModalOpen(false)}
            />
            <InviteMemberModal
              open={inviteModalOpen}
              onInvite={async (email, role) => {
                await inviteMember(currentSpaceId, email, role)
                getSpace(currentSpaceId).then(setSpaceDetail)
              }}
              onClose={() => setInviteModalOpen(false)}
            />
          </>
        ) : (
          <>
            <div
              className={`canvas ${isPanning ? 'canvas-panning' : ''}`}
              onMouseDown={handleCanvasMouseDown}
              role="presentation"
            >
              {error ? (
                <p className="todo-error" role="alert" style={{ color: '#e05353', marginBottom: 16 }}>{error}</p>
              ) : null}
              <div
                className="canvas-pan-wrapper"
                style={{ transform: `translate(${pan.x}px, ${pan.y}px)` }}
              >
                <div ref={treeRootRef} className="tree-root">
                {tagGroups.length === 0 ? (
                  <Empty />
                ) : (
                  tagGroups.map((tagGroup) => {
                    const priorityRows = groupRootsByPriority(tagGroup.nodes)
                    return (
                      <div key={tagGroup.tag || '_none'} className="tag-group" data-tag={tagGroup.tag || ''}>
                        <div className="tag-group-label">{tagGroup.tag ? tagGroup.tag : 'No tag'}</div>
                        <div className="tag-group-body">
                          {priorityRows.map((group) => (
                            <div key={group.priority} className="priority-row">
                              {group.nodes.map((node) => (
                                <TreeNode
                                  key={node.id}
                                  node={node}
                                  isRoot
                                  onToggleDone={handleToggleDone}
                                  onEdit={openEdit}
                                  onDelete={handleDelete}
                                  onAddSub={openAddSub}
                                  onSelectTask={(task) => setSelectedTaskId(task?.id ?? null)}
                                />
                              ))}
                            </div>
                          ))}
                        </div>
                      </div>
                    )
                  })
                )}
                </div>
              </div>
            </div>
            <StatusBar total={total} done={doneCount} pending={pendingCount} />
          </>
        )}
      </div>

      <Modal
        open={modal.open}
        mode={modal.mode}
        editTask={modal.editTask}
        parentId={modal.parentId}
        parentTag={modal.parentTag}
        existingTags={existingTags}
        onSave={handleSave}
        onClose={closeModal}
      />
      <ConfirmDeleteModal
        open={deleteConfirm.open}
        task={deleteConfirm.task}
        descendantCount={deleteConfirm.task ? countDescendants(deleteConfirm.task) : 0}
        onConfirm={handleConfirmDelete}
        onClose={() => setDeleteConfirm({ open: false, task: null })}
      />
      <CreateWorkspaceModal
        open={createWorkspaceModalOpen}
        onCreate={handleCreateWorkspace}
        onClose={() => setCreateWorkspaceModalOpen(false)}
      />
      <TaskDetailPanel
        taskId={selectedTaskId}
        onClose={() => setSelectedTaskId(null)}
        onEdit={(task) => { setSelectedTaskId(null); openEdit(task); }}
        onDelete={(task) => { setSelectedTaskId(null); handleDelete(task); }}
        spaceMembers={isCorporate ? (spaceDetail?.members || []) : []}
      />
    </div>
  )
}
