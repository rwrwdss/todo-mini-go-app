export default function ConfirmDeleteModal({ open, task, descendantCount, onConfirm, onClose }) {
  if (!open) return null

  const n = descendantCount || 0
  const message =
    n > 0
      ? `Deleting this root task will also delete all subtasks (${n} total). Continue?`
      : 'Delete this task?'

  return (
    <div
      className="modal-bg open"
      onClick={(e) => e.target === e.currentTarget && onClose()}
      role="dialog"
      aria-modal="true"
      aria-labelledby="confirm-delete-title"
    >
      <div className="modal confirm-delete-modal">
        <div className="modal-ttl" id="confirm-delete-title">
          Delete task
        </div>
        <p className="confirm-delete-msg">{message}</p>
        <div className="modal-acts">
          <button type="button" className="btn-cancel" onClick={onClose}>
            Cancel
          </button>
          <button type="button" className="btn-danger" onClick={() => onConfirm(task)}>
            Delete
          </button>
        </div>
      </div>
    </div>
  )
}
