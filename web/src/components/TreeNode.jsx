import TaskCard from './TaskCard'

export default function TreeNode({ node, isRoot, onToggleDone, onEdit, onDelete, onAddSub, onSelectTask, canEdit, canDelete, showAssignedToYou }) {
  const hasChildren = node.children && node.children.length > 0
  const allowEdit = typeof canEdit === 'function' ? canEdit(node) : true
  const allowDelete = typeof canDelete === 'function' ? canDelete(node) : true

  return (
    <div className={`node-wrap ${hasChildren ? 'has-children' : ''}`} data-id={node.id}>
      <TaskCard
        task={node}
        isRoot={isRoot}
        onToggleDone={onToggleDone}
        onEdit={onEdit}
        onDelete={onDelete}
        onAddSub={onAddSub}
        onSelectTask={onSelectTask}
        canEdit={allowEdit}
        canDelete={allowDelete}
        showAssignedToYou={showAssignedToYou}
      />
      {hasChildren ? (
        <div className="children-wrap">
          <div className="h-line" />
          <div className="children-row">
            {node.children.map((child) => (
              <TreeNode
                key={child.id}
                node={child}
                isRoot={false}
                onToggleDone={onToggleDone}
                onEdit={onEdit}
                onDelete={onDelete}
                onAddSub={onAddSub}
                onSelectTask={onSelectTask}
                canEdit={canEdit}
                canDelete={canDelete}
                showAssignedToYou={showAssignedToYou}
              />
            ))}
          </div>
        </div>
      ) : null}
    </div>
  )
}
