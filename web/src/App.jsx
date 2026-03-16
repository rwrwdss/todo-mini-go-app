import { useState, useEffect, useCallback, useRef } from 'react'
import { getTodos, createTodo, updateTodo, deleteTodo } from './api/todos'
import { buildTree, positionHLines } from './utils/tree'
import Header from './components/Header'
import TreeNode from './components/TreeNode'
import Modal from './components/Modal'
import Empty from './components/Empty'
import StatusBar from './components/StatusBar'
import './index.css'

export default function App() {
  const [todos, setTodos] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [modal, setModal] = useState({
    open: false,
    mode: 'create',
    editTask: null,
    parentId: null,
  })
  const treeRootRef = useRef(null)

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

  useEffect(() => {
    const onKeyDown = (e) => {
      if (e.key === 'Escape') {
        setModal((m) => ({ ...m, open: false }))
      }
      if ((e.ctrlKey || e.metaKey) && e.key === 'Enter' && modal.open) {
        e.preventDefault()
        document.querySelector('.modal .btn-save')?.click()
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [modal.open])

  const tree = buildTree(todos)
  const total = todos.length
  const doneCount = todos.filter((t) => t.done).length
  const pendingCount = total - doneCount

  function openNewTask() {
    setModal({ open: true, mode: 'create', editTask: null, parentId: null })
  }

  function openEdit(task) {
    setModal({ open: true, mode: 'edit', editTask: task, parentId: null })
  }

  function openAddSub(parent) {
    setModal({ open: true, mode: 'create', editTask: null, parentId: parent.id })
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

  async function handleDelete(task) {
    setError(null)
    try {
      await deleteTodo(task.id)
      await loadTodos(true)
    } catch (e) {
      setError(e.message || 'Failed to delete')
    }
  }

  if (loading && todos.length === 0) {
    return (
      <>
        <Header taskCount={0} onNewTask={openNewTask} />
        <div className="canvas">
          <p style={{ color: 'var(--text3)' }}>Loading…</p>
        </div>
        <StatusBar total={0} done={0} pending={0} />
      </>
    )
  }

  return (
    <>
      <Header taskCount={total} onNewTask={openNewTask} />
      <div className="canvas">
        {error ? (
          <p className="todo-error" role="alert" style={{ color: '#e05353', marginBottom: 16 }}>{error}</p>
        ) : null}
        <div ref={treeRootRef} className="tree-root">
          {tree.length === 0 ? (
            <Empty />
          ) : (
            tree.map((node) => (
              <TreeNode
                key={node.id}
                node={node}
                isRoot
                onToggleDone={handleToggleDone}
                onEdit={openEdit}
                onDelete={handleDelete}
                onAddSub={openAddSub}
              />
            ))
          )}
        </div>
      </div>
      <StatusBar total={total} done={doneCount} pending={pendingCount} />
      <Modal
        open={modal.open}
        mode={modal.mode}
        editTask={modal.editTask}
        parentId={modal.parentId}
        onSave={handleSave}
        onClose={closeModal}
      />
    </>
  )
}
