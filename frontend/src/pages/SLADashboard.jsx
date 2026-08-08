import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { slaService } from '../services/slaService'
import { useWebSocket } from '../contexts/WebSocketContext'
import SLABadge from '../components/SLABadge'
import SLACountdown from '../components/SLACountdown'

function StatCard({ label, value, sub, color = 'text-gray-900 dark:text-white' }) {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 shadow-sm p-5 text-center">
      <p className={`text-3xl font-bold ${color}`}>{value ?? '—'}</p>
      <p className="text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500 mt-1 font-medium">{label}</p>
      {sub && <p className="text-xs text-gray-400 dark:text-gray-500 mt-0.5">{sub}</p>}
    </div>
  )
}

function TicketRow({ ticket }) {
  return (
    <tr className="hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
      <td className="px-4 py-3">
        <Link to={`/tickets/${ticket.ticket_id}`} className="text-blue-600 hover:underline text-sm font-medium">
          {ticket.ticket_number}
        </Link>
      </td>
      <td className="px-4 py-3 text-sm text-gray-700 dark:text-gray-200 max-w-xs truncate">{ticket.subject}</td>
      <td className="px-4 py-3">
        <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${
          ticket.priority === 'URGENT' ? 'bg-red-100 text-red-700'
          : ticket.priority === 'HIGH' ? 'bg-orange-100 text-orange-700'
          : ticket.priority === 'MEDIUM' ? 'bg-yellow-100 text-yellow-700'
          : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 dark:text-gray-600'
        }`}>{ticket.priority}</span>
      </td>
      <td className="px-4 py-3"><SLABadge status={ticket.sla_status} /></td>
      <td className="px-4 py-3">
        <SLACountdown
          resolutionDueAt={ticket.resolution_due_at}
          slaStatus={ticket.sla_status}
          compact
        />
      </td>
      <td className="px-4 py-3 text-xs text-gray-400 dark:text-gray-500">
        {ticket.percent_elapsed != null ? `${Math.round(ticket.percent_elapsed)}%` : '—'}
      </td>
    </tr>
  )
}

export default function SLADashboard() {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const { wsService } = useWebSocket()

  const load = () => {
    slaService.getDashboard()
      .then((res) => setData(res.data.data))
      .catch(() => setError('Failed to load SLA data'))
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  // Refresh whenever the backend broadcasts an SLA update
  useEffect(() => {
    const unsub = wsService.on('sla.updated', () => load())
    return () => unsub()
  }, [wsService])

  if (loading) return <div className="p-8 text-gray-400 dark:text-gray-500">Loading SLA dashboard…</div>
  if (error) return <div className="p-8 text-red-600">{error}</div>

  const fmtMin = (min) => {
    if (!min) return '—'
    if (min < 60) return `${Math.round(min)}m`
    return `${Math.round(min / 60)}h ${Math.round(min % 60)}m`
  }

  return (
    <div className="max-w-6xl mx-auto px-6 py-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">SLA Dashboard</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 dark:text-gray-500 mt-0.5">Real-time SLA compliance overview</p>
        </div>
        <Link to="/sla-management"
          className="text-sm text-blue-600 font-medium hover:underline">
          Manage policies →
        </Link>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard
          label="Compliance"
          value={data ? `${Math.round(data.compliance_percent)}%` : '—'}
          color={data?.compliance_percent >= 90 ? 'text-green-600' : data?.compliance_percent >= 70 ? 'text-yellow-600' : 'text-red-600'}
        />
        <StatCard label="Total (with SLA)" value={data?.total_with_sla} />
        <StatCard
          label="Breached"
          value={data?.breached_count}
          color={data?.breached_count > 0 ? 'text-red-600' : 'text-gray-900 dark:text-white'}
        />
        <StatCard label="Completed On Time" value={data?.completed_on_time} color="text-green-600" />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <StatCard
          label="Avg First Response"
          value={fmtMin(data?.avg_first_response_minutes)}
        />
        <StatCard
          label="Avg Resolution Time"
          value={fmtMin(data?.avg_resolution_minutes)}
        />
      </div>

      {/* Breached tickets */}
      <section>
        <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-3 flex items-center gap-2">
          <span className="w-2 h-2 bg-red-500 rounded-full inline-block" />
          Breached Tickets
          {data?.breached?.length > 0 && (
            <span className="ml-1 text-xs bg-red-100 text-red-700 px-2 py-0.5 rounded-full font-medium">
              {data.breached.length}
            </span>
          )}
        </h2>
        <TicketTable tickets={data?.breached || []} emptyMsg="No breached tickets." />
      </section>

      {/* Near breach tickets */}
      <section>
        <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-3 flex items-center gap-2">
          <span className="w-2 h-2 bg-yellow-400 rounded-full inline-block" />
          Near Breach (within 2h)
          {data?.near_breach?.length > 0 && (
            <span className="ml-1 text-xs bg-yellow-100 text-yellow-700 px-2 py-0.5 rounded-full font-medium">
              {data.near_breach.length}
            </span>
          )}
        </h2>
        <TicketTable tickets={data?.near_breach || []} emptyMsg="No tickets near breach." />
      </section>
    </div>
  )
}

function TicketTable({ tickets, emptyMsg }) {
  if (tickets.length === 0) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 shadow-sm p-6 text-center text-gray-400 dark:text-gray-500 text-sm">
        {emptyMsg}
      </div>
    )
  }
  return (
    <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 shadow-sm overflow-hidden">
      <table className="w-full text-sm">
        <thead className="bg-gray-50 dark:bg-gray-900 text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500 uppercase">
          <tr>
            <th className="px-4 py-3 text-left">Ticket</th>
            <th className="px-4 py-3 text-left">Subject</th>
            <th className="px-4 py-3 text-left">Priority</th>
            <th className="px-4 py-3 text-left">Status</th>
            <th className="px-4 py-3 text-left">Time Left</th>
            <th className="px-4 py-3 text-left">% Elapsed</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-50">
          {tickets.map((t) => <TicketRow key={t.ticket_id} ticket={t} />)}
        </tbody>
      </table>
    </div>
  )
}
