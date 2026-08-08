import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { ticketService } from '../../services/ticketService'

const STATUS_NEXT = {
  OPEN:        { label: 'Start Working', next: 'IN_PROGRESS', color: 'bg-blue-600 hover:bg-blue-700' },
  IN_PROGRESS: { label: 'Mark Resolved', next: 'RESOLVED',    color: 'bg-emerald-600 hover:bg-emerald-700' },
  RESOLVED:    { label: 'Reopen',        next: 'OPEN',        color: 'bg-yellow-500 hover:bg-yellow-600' },
}

const PRIORITY_COLOR = {
  LOW: 'bg-gray-100 text-gray-600', MEDIUM: 'bg-blue-100 text-blue-700',
  HIGH: 'bg-orange-100 text-orange-700', URGENT: 'bg-red-100 text-red-700',
}

export default function AgentQueue() {
  const [tickets, setTickets] = useState([])
  const [loading, setLoading] = useState(true)
  const [updating, setUpdating] = useState(null)

  const load = () => {
    ticketService.getTeamTickets({ limit: 50 })
      .then((res) => setTickets(res.data.data?.items || []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
    const interval = setInterval(load, 15000)
    return () => clearInterval(interval)
  }, [])

  const handleStatus = async (ticketId, nextStatus) => {
    setUpdating(ticketId)
    try {
      await ticketService.updateStatus(ticketId, nextStatus)
      setTickets((prev) => prev.map((t) => t.id === ticketId ? { ...t, status: nextStatus } : t))
    } catch { /* ignore */ }
    finally { setUpdating(null) }
  }

  return (
    <main className="max-w-4xl mx-auto px-6 py-6">
      <div className="mb-6">
        <h1 className="text-xl font-bold text-gray-900 dark:text-white">Team Queue</h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Tickets assigned to your team or AI-routed to you</p>
      </div>

      {loading ? (
        <p className="text-sm text-gray-400 text-center py-16 animate-pulse">Loading…</p>
      ) : tickets.length === 0 ? (
        <div className="text-center py-16">
          <p className="text-4xl mb-3">🎉</p>
          <p className="text-sm font-medium text-gray-700 dark:text-gray-200">All clear! No tickets assigned to you.</p>
          <Link to="/agent/available" className="mt-3 inline-block text-sm text-emerald-600 hover:underline">Pick up a ticket →</Link>
        </div>
      ) : (
        <div className="space-y-3">
          {tickets.map((t) => {
            const action = STATUS_NEXT[t.status]
            return (
              <div
                key={t.id}
                className="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 p-4 shadow-sm"
              >
                <div className="flex items-start justify-between gap-4">
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="font-mono text-xs text-emerald-600 dark:text-emerald-400">{t.ticket_number}</span>
                      <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${PRIORITY_COLOR[t.priority] || 'bg-gray-100 text-gray-600'}`}>
                        {t.priority}
                      </span>
                      <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${
                        t.status === 'OPEN' ? 'bg-yellow-100 text-yellow-700' :
                        t.status === 'IN_PROGRESS' ? 'bg-blue-100 text-blue-700' :
                        'bg-emerald-100 text-emerald-700'
                      }`}>{t.status}</span>
                    </div>
                    <Link to={`/agent/tickets/${t.id}`} className="text-sm font-medium text-gray-900 dark:text-white hover:text-emerald-600 dark:hover:text-emerald-400 line-clamp-1">
                      {t.subject}
                    </Link>
                    <p className="text-xs text-gray-400 dark:text-gray-500 mt-0.5">{t.customer_email}</p>
                  </div>
                  <div className="flex items-center gap-2 shrink-0">
                    <Link
                      to={`/agent/tickets/${t.id}`}
                      className="px-3 py-1.5 text-xs border border-gray-200 dark:border-gray-600 rounded-lg text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition"
                    >
                      View
                    </Link>
                    {action && (
                      <button
                        onClick={() => handleStatus(t.id, action.next)}
                        disabled={updating === t.id}
                        className={`px-3 py-1.5 text-xs text-white rounded-lg transition disabled:opacity-50 ${action.color}`}
                      >
                        {updating === t.id ? '…' : action.label}
                      </button>
                    )}
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </main>
  )
}
