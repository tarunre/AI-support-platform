import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { ticketService } from '../../services/ticketService'
import { commentService } from '../../services/commentService'

const STATUSES = ['OPEN', 'IN_PROGRESS', 'RESOLVED', 'CLOSED']
const STATUS_COLOR = {
  OPEN: 'bg-yellow-100 text-yellow-700', IN_PROGRESS: 'bg-blue-100 text-blue-700',
  RESOLVED: 'bg-emerald-100 text-emerald-700', CLOSED: 'bg-gray-100 text-gray-600',
}
const PRIORITY_COLOR = {
  LOW: 'bg-gray-100 text-gray-600', MEDIUM: 'bg-blue-100 text-blue-700',
  HIGH: 'bg-orange-100 text-orange-700', URGENT: 'bg-red-100 text-red-700',
}

export default function AgentTicketDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [ticket, setTicket] = useState(null)
  const [comments, setComments] = useState([])
  const [loading, setLoading] = useState(true)
  const [statusUpdating, setStatusUpdating] = useState(false)
  const [claiming, setClaiming] = useState(false)
  const [comment, setComment] = useState('')
  const [posting, setPosting] = useState(false)

  const load = async () => {
    try {
      const [tRes, cRes] = await Promise.all([
        ticketService.getTicket(id),
        commentService.getComments(id),
      ])
      setTicket(tRes.data.data)
      setComments(cRes.data.data || [])
    } catch { navigate('/agent/queue') }
    finally { setLoading(false) }
  }

  useEffect(() => { load() }, [id]) // eslint-disable-line

  const handleStatusChange = async (e) => {
    const newStatus = e.target.value
    setStatusUpdating(true)
    try {
      await ticketService.updateStatus(id, newStatus)
      setTicket((t) => ({ ...t, status: newStatus }))
    } catch { /* ignore */ }
    finally { setStatusUpdating(false) }
  }

  const handlePickUp = async () => {
    setClaiming(true)
    try {
      const res = await ticketService.takeOwnership(id)
      setTicket(res.data.data)
    } catch { /* ignore */ }
    finally { setClaiming(false) }
  }

  const handleComment = async (e) => {
    e.preventDefault()
    if (!comment.trim()) return
    setPosting(true)
    try {
      const res = await commentService.createComment(id, { message: comment.trim(), comment_type: 'PUBLIC' })
      setComments((prev) => [...prev, res.data.data])
      setComment('')
    } catch { /* ignore */ }
    finally { setPosting(false) }
  }

  if (loading) return (
    <div className="flex items-center justify-center py-24">
      <p className="text-sm text-gray-400 animate-pulse">Loading…</p>
    </div>
  )
  if (!ticket) return null

  return (
    <main className="max-w-3xl mx-auto px-6 py-6">
      {/* Back */}
      <Link to="/agent/queue" className="inline-flex items-center gap-1 text-xs text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 mb-5 transition">
        ← Back to My Queue
      </Link>

      {/* Header */}
      <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 p-5 mb-4 shadow-sm">
        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0">
            <div className="flex items-center gap-2 mb-2">
              <span className="font-mono text-xs text-emerald-600 dark:text-emerald-400">{ticket.ticket_number}</span>
              <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${PRIORITY_COLOR[ticket.priority]}`}>{ticket.priority}</span>
              {!ticket.assigned_to && (
                <span className="text-xs font-medium px-2 py-0.5 rounded-full bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400">Unassigned</span>
              )}
            </div>
            <h1 className="text-base font-semibold text-gray-900 dark:text-white">{ticket.subject}</h1>
          </div>
          {/* Actions */}
          <div className="shrink-0 flex items-center gap-2">
            {!ticket.assigned_to && (
              <button
                onClick={handlePickUp}
                disabled={claiming}
                className="px-4 py-1.5 bg-emerald-600 text-white text-xs font-semibold rounded-lg hover:bg-emerald-700 disabled:opacity-50 transition"
              >
                {claiming ? 'Claiming…' : '📥 Pick Up'}
              </button>
            )}
            {/* Status selector */}
            <select
              value={ticket.status}
              onChange={handleStatusChange}
              disabled={statusUpdating}
              className={`text-xs font-medium px-3 py-1.5 rounded-lg border-0 cursor-pointer focus:outline-none focus:ring-2 focus:ring-emerald-500 ${STATUS_COLOR[ticket.status]} disabled:opacity-50`}
            >
              {STATUSES.map((s) => <option key={s} value={s}>{s.replace('_', ' ')}</option>)}
            </select>
          </div>
        </div>
      </div>

      {/* Description */}
      <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 p-5 mb-4 shadow-sm">
        <h3 className="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-3">Description</h3>
        <p className="text-sm text-gray-800 dark:text-gray-100 whitespace-pre-wrap break-words">{ticket.description}</p>
      </div>

      {/* Customer */}
      <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 p-5 mb-4 shadow-sm">
        <h3 className="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-3">Customer</h3>
        <div className="space-y-1.5">
          <div className="flex gap-3">
            <dt className="w-16 text-xs text-gray-400 dark:text-gray-500 shrink-0">Name</dt>
            <dd className="text-sm text-gray-800 dark:text-gray-100 font-medium">{ticket.customer_name || '—'}</dd>
          </div>
          <div className="flex gap-3">
            <dt className="w-16 text-xs text-gray-400 dark:text-gray-500 shrink-0">Email</dt>
            <dd className="text-sm text-gray-800 dark:text-gray-100">{ticket.customer_email}</dd>
          </div>
        </div>
      </div>

      {/* Comments */}
      <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 p-5 shadow-sm">
        <h3 className="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-4">
          Comments ({comments.length})
        </h3>

        {comments.length === 0 ? (
          <p className="text-sm text-gray-400 dark:text-gray-500 mb-4">No comments yet.</p>
        ) : (
          <ul className="space-y-3 mb-4">
            {comments.map((c) => (
              <li key={c.id} className="flex gap-3">
                <div className="w-7 h-7 rounded-full bg-emerald-100 dark:bg-emerald-900/40 flex items-center justify-center text-xs font-semibold text-emerald-700 dark:text-emerald-400 shrink-0">
                  {c.user?.name?.[0]?.toUpperCase() || 'A'}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-xs text-gray-400 dark:text-gray-500 mb-1">{c.user?.name || 'Agent'}</p>
                  <p className="text-sm text-gray-800 dark:text-gray-100 whitespace-pre-wrap break-words">{c.message}</p>
                </div>
              </li>
            ))}
          </ul>
        )}

        {/* Add comment */}
        <form onSubmit={handleComment} className="flex gap-2">
          <textarea
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            rows={2}
            placeholder="Add a comment…"
            className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm resize-none focus:outline-none focus:ring-2 focus:ring-emerald-500 dark:bg-gray-700 dark:text-gray-100"
          />
          <button
            type="submit"
            disabled={posting || !comment.trim()}
            className="px-4 py-2 bg-emerald-600 text-white text-sm font-medium rounded-lg hover:bg-emerald-700 disabled:opacity-50 transition self-end"
          >
            {posting ? '…' : 'Post'}
          </button>
        </form>
      </div>
    </main>
  )
}
