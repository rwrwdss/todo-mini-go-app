export default function StatusBar({ total, done, pending }) {
  return (
    <div className="statusbar">
      <div className="stat">Total <span>{total}</span></div>
      <div className="stat">Done <span>{done}</span></div>
      <div className="stat">Pending <span>{pending}</span></div>
    </div>
  )
}
