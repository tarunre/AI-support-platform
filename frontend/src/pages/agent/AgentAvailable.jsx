import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { ticketService } from '../../services/ticketService'

const PRIORITY_COLOR = {
  LOW: 'bg-gray-100 text-gray-600', MEDIUM: 'bg-blue-100 text-blue-700',
  HIGH: 'bg-orange-100 text-orange-700', URGENT: 'bg-red-100 text-red-700',
}

export default function AgentAvailable() {
  const [tickets, setTickets] = useState([])
  const [loading, setLoading] = useState(true)
  const [claiming, setClaiming] = useState(null)
  const navigate = useNavigate()

  const load = () => {
    ticketService.getUnassigned({ limit: 50 })
      .then((res) => setTickets(res.data.data?.tickets || res.data.data?.items || []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  const handlePickUp = async (ticketId) => {
    setClaiming(ticketId)
    try {
      await ticketService.takeOwnership(ticketId)
      navigate(`/agent/tickets/${ticketId}`)
    } catch { setClaiming(null) }
  }

  return (
    <main className="max-w-4xl mx-auto px-6 py-6">
      <div className="mb-6">
        <h1 className="text-xl font-bold text-gray-900 dark:text-white">Pick Up a Ticket</h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Unassigned tickets — claim one to start working</p>
      </div>

      {loading ? (
        <p className="text-sm text-gray-400 text-center py-16 animate-pulse">Loading…</p>
      ) : tickets.length === 0 ? (
        <div className="text-center py-16">
          <p className="text-4xl mb-3">✨</p>
          <p className="text-sm font-medium text-gray-700 dark:text-gray-200">No unassigned tickets right now.</p>
          <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">Check back soon or refresh the page.</p>
        </div>
      ) : (
        <div className="space-y-3">
          {tickets.map((t) => (
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
                    <span className="text-xs font-medium px-2 py-0.5 rounded-full bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300">
                      {t.category}
                    </span>
                  </div>
                  <p className="text-sm font-medium text-gray-900 dark:text-white line-clamp-1">{t.subject}</p>
                  <p className="text-xs text-gray-400 dark:text-gray-500 mt-0.5 line-clamp-2">{t.description}</p>
                  <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">From: {t.customer_email}</p>
                </div>
                <button
                  onClick={() => handlePickUp(t.id)}
                  disabled={claiming === t.id}
                  className="shrink-0 px-4 py-2 text-sm font-semibold bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition disabled:opacity-50"
                >
                  {claiming === t.id ? 'Claiming…' : '📥 Pick Up'}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </main>
  )
}
