import { useState, useEffect, useCallback } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { ticketService } from '../../services/ticketService'
import StatusBadge from '../../components/tickets/StatusBadge'
import PriorityBadge from '../../components/tickets/PriorityBadge'
import Toast, { useToast } from '../../components/Toast'
import { formatDate } from '../../utils/format'

const STATUSES = ['OPEN', 'IN_PROGRESS', 'RESOLVED', 'CLOSED']
const PRIORITIES = ['LOW', 'MEDIUM', 'HIGH', 'URGENT']

function TicketList() {
  const navigate = useNavigate()
  const { toast, showToast } = useToast()

  const [tickets, setTickets] = useState([])
  const [meta, setMeta] = useState({ total_count: 0, current_page: 1, total_pages: 1 })
  const [loading, setLoading] = useState(true)

  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [priorityFilter, setPriorityFilter] = useState('')
  const [page, setPage] = useState(1)
  const limit = 20

  const fetchTickets = useCallback(async () => {
    setLoading(true)
    try {
      const res = await ticketService.getTickets({
        page,
        limit,
        search: search || undefined,
        status: statusFilter || undefined,
        priority: priorityFilter || undefined,
      })
      const data = res.data.data
      setTickets(data.items ?? [])
      setMeta(data)
    } catch {
      showToast('Failed to load tickets', 'error')
    } finally {
      setLoading(false)
    }
  }, [page, search, statusFilter, priorityFilter]) // eslint-disable-line

  useEffect(() => { fetchTickets() }, [fetchTickets])

  const handleSearchSubmit = (e) => {
    e.preventDefault()
    setPage(1)
    fetchTickets()
  }

  const handleFilterChange = (setter) => (e) => {
    setter(e.target.value)
    setPage(1)
  }

  return (
    <>
      <Toast toast={toast} />
      <main className="max-w-7xl mx-auto px-6 py-6">
        <h1 className="font-bold text-gray-800 dark:text-gray-100 text-xl mb-4">All Tickets</h1>
        {/* Filters + Action */}
        <div className="flex flex-wrap gap-3 items-center justify-between mb-4">
          <div className="flex flex-wrap gap-3 items-center">
            <form onSubmit={handleSearchSubmit} className="flex gap-2 items-center">
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search tickets…"
              className="rounded-lg border border-gray-200 dark:border-gray-600 px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-blue-400 w-56"
            />
            <button type="submit" className="rounded-lg bg-blue-600 px-3 py-2 text-sm text-white hover:bg-blue-700 transition">Search</button>
          </form>

          <select
            value={statusFilter}
            onChange={handleFilterChange(setStatusFilter)}
            className="select-field"
          >
            <option value="">All Statuses</option>
            {STATUSES.map((s) => <option key={s} value={s}>{s.replace('_', ' ')}</option>)}
          </select>

          <select
            value={priorityFilter}
            onChange={handleFilterChange(setPriorityFilter)}
            className="select-field"
          >
            <option value="">All Priorities</option>
            {PRIORITIES.map((p) => <option key={p} value={p}>{p}</option>)}
          </select>

          {(search || statusFilter || priorityFilter) && (
            <button
              onClick={() => { setSearch(''); setStatusFilter(''); setPriorityFilter(''); setPage(1) }}
              className="text-sm text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:text-gray-300 underline"
            >
              Clear
            </button>
          )}
          </div>
          <Link
            to="/tickets/new"
            className="px-4 py-2 bg-emerald-600 text-white text-sm font-medium rounded-lg hover:bg-emerald-700 transition shrink-0 self-center"
          >
            + New Ticket
          </Link>
        </div>

        {/* Table */}
        <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 overflow-hidden">
          {loading ? (
            <div className="py-16 text-center text-sm text-gray-400 dark:text-gray-500 animate-pulse">Loading tickets…</div>
          ) : tickets.length === 0 ? (
            <div className="py-16 text-center text-sm text-gray-400 dark:text-gray-500">No tickets found.</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-900 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 dark:text-gray-500 uppercase tracking-wide">
                    <th className="px-4 py-3">Ticket #</th>
                    <th className="px-4 py-3">Subject</th>
                    <th className="px-4 py-3">Customer</th>
                    <th className="px-4 py-3">Priority</th>
                    <th className="px-4 py-3">Status</th>
                    <th className="px-4 py-3">Assigned To</th>
                    <th className="px-4 py-3">Created</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-50">
                  {tickets.map((t) => (
                    <tr
                      key={t.id}
                      onClick={() => navigate(`/tickets/${t.id}`)}
                      className="hover:bg-gray-50 dark:bg-gray-900 cursor-pointer transition"
                    >
                      <td className="px-4 py-3 font-mono text-xs text-blue-600 font-medium">{t.ticket_number}</td>
                      <td className="px-4 py-3 max-w-[200px] truncate text-gray-800 dark:text-gray-100 font-medium">{t.subject}</td>
                      <td className="px-4 py-3 text-gray-600 dark:text-gray-300 dark:text-gray-600">
                        <div>{t.customer_name}</div>
                        <div className="text-xs text-gray-400 dark:text-gray-500">{t.customer_email}</div>
                      </td>
                      <td className="px-4 py-3"><PriorityBadge priority={t.priority} /></td>
                      <td className="px-4 py-3"><StatusBadge status={t.status} /></td>
                      <td className="px-4 py-3 text-gray-500 dark:text-gray-400 dark:text-gray-500 text-xs">{t.assignee?.name ?? '—'}</td>
                      <td className="px-4 py-3 text-gray-400 dark:text-gray-500 text-xs">{formatDate(t.created_at)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Pagination */}
        {meta.total_pages > 1 && (
          <div className="mt-4 flex items-center justify-between text-sm text-gray-500 dark:text-gray-400 dark:text-gray-500">
            <span>{meta.total_count} total tickets</span>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:bg-gray-900 disabled:opacity-40 disabled:cursor-not-allowed transition"
              >
                ← Prev
              </button>
              <span className="px-2">Page {page} of {meta.total_pages}</span>
              <button
                onClick={() => setPage((p) => Math.min(meta.total_pages, p + 1))}
                disabled={page === meta.total_pages}
                className="px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:bg-gray-900 disabled:opacity-40 disabled:cursor-not-allowed transition"
              >
                Next →
              </button>
            </div>
          </div>
        )}
      </main>
    </>
  )
}

export default TicketList
