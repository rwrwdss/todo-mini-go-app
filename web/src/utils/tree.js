/**
 * Builds a tree from flat list of todos. Mutates nodes adding `children` array.
 * Roots: parent_id is null, 0, or undefined.
 */
export function buildTree(todos) {
  if (!Array.isArray(todos) || !todos.length) return []
  const list = todos.map((t) => ({ ...t, children: [] }))
  const byId = new Map(list.map((t) => [t.id, t]))
  const roots = []
  for (const t of list) {
    const pid = t.parent_id ?? null
    if (pid == null || pid === 0) {
      roots.push(t)
    } else {
      const parent = byId.get(pid)
      if (parent && parent.children) parent.children.push(t)
      else roots.push(t)
    }
  }
  return roots
}

/**
 * Position horizontal connector lines between sibling nodes.
 * Call after tree is rendered (useEffect + ref to tree root).
 */
export function positionHLines(container) {
  if (!container) return
  container.querySelectorAll('.children-wrap').forEach((cw) => {
    const row = cw.querySelector('.children-row')
    const hl = cw.querySelector('.h-line')
    if (!row || !hl) return
    const nodes = Array.from(row.children)
    if (nodes.length < 1) return

    const rowRect = row.getBoundingClientRect()

    if (nodes.length === 1) {
      hl.style.display = 'none'
      return
    }

    hl.style.display = ''
    const firstCard = nodes[0].querySelector(':scope > .task-card')
    const lastCard = nodes[nodes.length - 1].querySelector(':scope > .task-card')
    if (!firstCard || !lastCard) return
    const fr = firstCard.getBoundingClientRect()
    const lr = lastCard.getBoundingClientRect()

    const left = (fr.left + fr.right) / 2 - rowRect.left
    const right = (lr.left + lr.right) / 2 - rowRect.left

    hl.style.left = left + 'px'
    hl.style.width = right - left + 'px'
  })
}
