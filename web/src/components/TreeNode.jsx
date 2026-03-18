import TaskCard from './TaskCard'

export default function TreeNode({ node, isRoot, onToggleDone, onEdit, onDelete, onAddSub, onSelectTask }) {
  const hasChildren = node.children && node.children.length > 0

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
              />
            ))}
          </div>
        </div>
      ) : null}
    </div>
  )
}
