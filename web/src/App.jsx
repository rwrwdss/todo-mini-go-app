import { useState, useEffect, useCallback, useRef } from 'react'
import { getTodos, createTodo, updateTodo, deleteTodo } from './api/todos'
import { buildTree, groupRootsByTag, groupRootsByPriority, positionHLines, countDescendants } from './utils/tree'
import Sidebar from './components/Sidebar'
import TreeNode from './components/TreeNode'
import Modal from './components/Modal'
import ConfirmDeleteModal from './components/ConfirmDeleteModal'
import Empty from './components/Empty'
import StatusBar from './components/StatusBar'
import DashboardPage from './components/DashboardPage'
import './index.css'

export default function App({ onLogout }) {
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

  const loadTodos = useCallback(async (background = false) => {
    if (!background) {
      setLoading(true)
      setError(null)
    }
    try {
      const data = await getTodos()
      setTodos(Array.isArray(data) ? data : [])
    } catch (e) {
      setError(e.message || 'Failed to load tasks')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadTodos()
  }, [loadTodos])

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
      if (modal.mode === 'edit' && modal.editTask) {
        await updateTodo(modal.editTask.id, payload)
      } else {
        await createTodo(payload)
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
    setDeleteConfirm({ open: false, task: null })
  }

  const tagCounts = tagGroups.map((g) => ({ tag: g.tag || '', count: g.nodes.length }))

  if (loading && todos.length === 0) {
    return (
      <div className="app-layout">
        <Sidebar view="tree" onNavigate={setView} taskCount={0} tagCounts={[]} onLogout={onLogout} />
        <div className="main">
          <div className="topbar">
            <div className="tb-title">My tasks</div>
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
      <Sidebar view={view} onNavigate={setView} taskCount={total} tagCounts={tagCounts} onLogout={onLogout} />
      <div className="main">
        <div className="topbar">
          <div className="tb-title">{view === 'dashboard' ? 'Dashboard' : 'My tasks'}</div>
          <div className="tb-right">
            <span className="task-count">{total === 1 ? '1 task' : `${total} tasks`}</span>
            <button type="button" className="btn-sm" onClick={openNewTask}>+ New task</button>
          </div>
        </div>

        {view === 'dashboard' ? (
          <DashboardPage
            todos={todos}
            loading={loading}
            onViewAll={() => setView('tree')}
          />
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
    </div>
  )
}
