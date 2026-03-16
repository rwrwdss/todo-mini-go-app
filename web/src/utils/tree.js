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

const PRIORITY_ORDER = ['high', 'med', 'low', 'none']

function normalizePriority(p) {
  const s = (p || 'none').toLowerCase()
  return s === 'medium' ? 'med' : s
}

/**
 * Groups root nodes by tag. Returns array of { tag: string, nodes: Node[] }.
 * Tagged groups first (alphabetically), then group with empty tag last (for "No tag").
 * Skips empty groups.
 */
export function groupRootsByTag(roots) {
  if (!Array.isArray(roots) || !roots.length) return []
  const byTag = new Map()
  for (const node of roots) {
    const t = (node.tag || '').trim()
    const key = t === '' ? '_none' : t
    if (!byTag.has(key)) byTag.set(key, [])
    byTag.get(key).push(node)
  }
  const result = []
  const tagged = [...byTag.entries()].filter(([k]) => k !== '_none').sort((a, b) => a[0].localeCompare(b[0]))
  for (const [key, nodes] of tagged) {
    result.push({ tag: key, nodes })
  }
  if (byTag.has('_none')) {
    result.push({ tag: '', nodes: byTag.get('_none') })
  }
  return result
}

/**
 * Groups root nodes by priority. Returns array of { priority, nodes } in order: high, med, low, none.
 * Skips empty groups.
 */
export function groupRootsByPriority(roots) {
  if (!Array.isArray(roots) || !roots.length) return []
  const groups = { high: [], med: [], low: [], none: [] }
  for (const node of roots) {
    const key = normalizePriority(node.priority)
    if (groups[key]) groups[key].push(node)
    else groups.none.push(node)
  }
  return PRIORITY_ORDER.filter((p) => groups[p].length > 0).map((priority) => ({
    priority,
    nodes: groups[priority],
  }))
}

/**
 * Returns total count of descendants of a tree node (children + their descendants).
 */
export function countDescendants(node) {
  if (!node || !Array.isArray(node.children) || node.children.length === 0) return 0
  return node.children.length + node.children.reduce((sum, c) => sum + countDescendants(c), 0)
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
