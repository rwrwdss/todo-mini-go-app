export default function Empty() {
  const dots = Array(15).fill(0).map((_, i) => <div key={i} className="empty-dot" />)
  return (
    <div className="empty">
      <div className="empty-dots">{dots}</div>
      <p>No tasks yet</p>
    </div>
  )
}
