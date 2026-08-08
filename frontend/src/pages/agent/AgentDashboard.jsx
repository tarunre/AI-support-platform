import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'
import { ticketService } from '../../services/ticketService'

function StatCard({ label, value, color, icon }) {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 p-5 shadow-sm">
      <div className="flex items-center justify-between mb-2">
        <p className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">{label}</p>
        <span className="text-xl">{icon}</span>
      </div>
      <p className={`text-3xl font-bold ${color}`}>{value ?? 0}</p>
    </div>
  )
}

export default function AgentDashboard() {
  const { user } = useAuth()
  const [myTickets, setMyTickets] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    ticketService.getTeamTickets({ limit: 50 })
      .then((res) => setMyTickets(res.data.data?.items || []))
      .catch(() => {})
      .finally(() => setLoading(false))
    const interval = setInterval(() => {
      ticketService.getTeamTickets({ limit: 50 })
        .then((res) => setMyTickets(res.data.data?.items || []))
        .catch(() => {})
    }, 15000)
    return () => clearInterval(interval)
  }, [])

  const open       = myTickets.filter((t) => t.status === 'OPEN').length
  const inProgress = myTickets.filter((t) => t.status === 'IN_PROGRESS').length
  const resolved   = myTickets.filter((t) => t.status === 'RESOLVED').length
  const closed     = myTickets.filter((t) => t.status === 'CLOSED').length
  const recent     = [...myTickets].sort((a, b) => new Date(b.created_at) - new Date(a.created_at)).slice(0, 5)

  return (
    <main className="max-w-4xl mx-auto px-6 py-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-xl font-bold text-gray-900 dark:text-white">
          Welcome back, {user?.name?.split(' ')[0]} 👋
        </h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
          {user?.team || 'Support'} Team · {new Date().toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric' })}
        </p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4 mb-8">
        <StatCard label="Open"        value={open}        color="text-yellow-500"  icon="🟡" />
        <StatCard label="In Progress" value={inProgress}  color="text-blue-500"    icon="🔵" />
        <StatCard label="Resolved"    value={resolved}    color="text-emerald-500" icon="✅" />
        <StatCard label="Closed"      value={closed}      color="text-gray-400"    icon="🔒" />
      </div>

      {/* Recent assigned tickets */}
      <div>
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-sm font-semibold text-gray-700 dark:text-gray-200">My Recent Tickets</h2>
          <Link to="/agent/queue" className="text-xs text-emerald-600 hover:underline">View all →</Link>
        </div>
        <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 overflow-hidden">
          {loading ? (
            <p className="text-sm text-gray-400 text-center py-8 animate-pulse">Loading…</p>
          ) : recent.length === 0 ? (
            <p className="text-sm text-gray-400 text-center py-8">No tickets assigned yet.</p>
          ) : (
            <ul className="divide-y divide-gray-50 dark:divide-gray-700">
              {recent.map((t) => (
                <li key={t.id}>
                  <Link
                    to={`/agent/tickets/${t.id}`}
                    className="flex items-center justify-between px-5 py-3.5 hover:bg-gray-50 dark:hover:bg-gray-700 transition"
                  >
                    <div className="flex items-center gap-3 min-w-0">
                      <span className="font-mono text-xs text-emerald-600 dark:text-emerald-400 shrink-0">{t.ticket_number}</span>
                      <span className="text-sm text-gray-800 dark:text-gray-100 truncate">{t.subject}</span>
                    </div>
                    <span className={`text-xs font-medium px-2 py-0.5 rounded-full shrink-0 ml-3 ${
                      t.status === 'OPEN'        ? 'bg-yellow-100 text-yellow-700' :
                      t.status === 'IN_PROGRESS' ? 'bg-blue-100 text-blue-700' :
                      t.status === 'RESOLVED'    ? 'bg-emerald-100 text-emerald-700' :
                      'bg-gray-100 text-gray-600'
                    }`}>{t.status}</span>
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      {/* Quick actions */}
      <div className="mt-6 grid grid-cols-2 gap-4">
        <Link
          to="/agent/available"
          className="flex items-center gap-3 p-4 bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800 rounded-xl hover:bg-emerald-100 dark:hover:bg-emerald-900/30 transition"
        >
          <span className="text-2xl">📥</span>
          <div>
            <p className="text-sm font-semibold text-emerald-700 dark:text-emerald-400">Pick Up a Ticket</p>
            <p className="text-xs text-emerald-600/70 dark:text-emerald-500">View unassigned queue</p>
          </div>
        </Link>
        <Link
          to="/agent/queue"
          className="flex items-center gap-3 p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-xl hover:bg-blue-100 dark:hover:bg-blue-900/30 transition"
        >
          <span className="text-2xl">📋</span>
          <div>
            <p className="text-sm font-semibold text-blue-700 dark:text-blue-400">My Queue</p>
            <p className="text-xs text-blue-600/70 dark:text-blue-500">{myTickets.length} tickets assigned</p>
          </div>
        </Link>
      </div>
    </main>
  )
}
