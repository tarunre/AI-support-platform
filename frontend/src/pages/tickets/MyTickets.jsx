import { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { ticketService } from '../../services/ticketService'
import StatusBadge from '../../components/tickets/StatusBadge'
import PriorityBadge from '../../components/tickets/PriorityBadge'
import { formatDate } from '../../utils/format'

export default function MyTickets() {
  const navigate = useNavigate()
  const [tickets, setTickets] = useState([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [search, setSearch] = useState('')
  const [searchInput, setSearchInput] = useState('')
  const [status, setStatus] = useState('')
  const [priority, setPriority] = useState('')
  const [loading, setLoading] = useState(true)

  const load = () => {
    setLoading(true)
    ticketService
      .getMyTickets({ page, limit: 20, search, status, priority })
      .then((r) => {
        const d = r.data.data
        setTickets(d.items || [])
        setTotal(d.total_count || 0)
        setTotalPages(d.total_pages || 1)
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [page, search, status, priority])

  const handleSearch = (e) => {
    e.preventDefault()
    setPage(1)
    setSearch(searchInput)
  }

  const clearFilters = () => {
    setSearch('')
    setSearchInput('')
    setStatus('')
    setPriority('')
    setPage(1)
  }

  return (
    <main className="max-w-6xl mx-auto px-6 py-6 space-y-4">
      <h1 className="font-bold text-gray-800 dark:text-gray-100 text-xl">My Tickets</h1>
        {/* Filters + count */}
        <div className="flex flex-wrap gap-3 items-center justify-between">
          <div className="flex flex-wrap gap-3 items-center">
          <form onSubmit={handleSearch} className="flex gap-2">
            <input
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              placeholder="Search tickets…"
              className="rounded-lg border border-gray-200 dark:border-gray-600 px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-blue-400 w-56"
            />
            <button type="submit" className="rounded-lg bg-blue-600 px-3 py-2 text-sm text-white hover:bg-blue-700 transition">Search</button>
          </form>

          <select
            value={status}
            onChange={(e) => { setStatus(e.target.value); setPage(1) }}
            className="select-field"
          >
            <option value="">All Statuses</option>
            {['OPEN', 'IN_PROGRESS', 'RESOLVED', 'CLOSED'].map((s) => (
              <option key={s} value={s}>{s.replace('_', ' ')}</option>
            ))}
          </select>

          <select
            value={priority}
            onChange={(e) => { setPriority(e.target.value); setPage(1) }}
            className="select-field"
          >
            <option value="">All Priorities</option>
            {['LOW', 'MEDIUM', 'HIGH', 'URGENT'].map((p) => (
              <option key={p} value={p}>{p}</option>
            ))}
          </select>

          {(search || status || priority) && (
            <button onClick={clearFilters} className="text-sm text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:text-gray-300 dark:text-gray-600 underline">
              Clear
            </button>
          )}
          </div>
          <span className="text-xs text-gray-400 dark:text-gray-500 shrink-0">{total} ticket{total !== 1 ? 's' : ''}</span>
        </div>

        {/* Table */}
        <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-sm overflow-hidden">
          {loading ? (
            <p className="p-6 text-sm text-gray-400 dark:text-gray-500 animate-pulse text-center">Loading…</p>
          ) : tickets.length === 0 ? (
            <p className="p-6 text-sm text-gray-400 dark:text-gray-500 text-center">No tickets found.</p>
          ) : (
            <table className="w-full text-sm">
              <thead className="bg-gray-50 dark:bg-gray-900 border-b border-gray-100 dark:border-gray-700">
                <tr>
                  {['Ticket #', 'Subject', 'Customer', 'Priority', 'Status', 'Updated'].map((h) => (
                    <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 dark:text-gray-500 uppercase tracking-wide">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-50">
                {tickets.map((t) => (
                  <tr
                    key={t.id}
                    onClick={() => navigate(`/tickets/${t.id}`)}
                    className="hover:bg-gray-50 dark:bg-gray-900 cursor-pointer transition"
                  >
                    <td className="px-4 py-3 font-mono text-xs text-blue-600">{t.ticket_number}</td>
                    <td className="px-4 py-3 font-medium text-gray-800 dark:text-gray-100 max-w-xs truncate">{t.subject}</td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400 dark:text-gray-500">{t.customer_name}</td>
                    <td className="px-4 py-3"><PriorityBadge priority={t.priority} /></td>
                    <td className="px-4 py-3"><StatusBadge status={t.status} /></td>
                    <td className="px-4 py-3 text-gray-400 dark:text-gray-500 text-xs">{formatDate(t.updated_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-between">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
              className="rounded-lg border border-gray-200 dark:border-gray-600 px-3 py-1.5 text-sm text-gray-600 dark:text-gray-300 dark:text-gray-600 hover:bg-gray-50 dark:bg-gray-900 disabled:opacity-40 transition"
            >
              ← Prev
            </button>
            <span className="text-sm text-gray-400 dark:text-gray-500">Page {page} of {totalPages}</span>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="rounded-lg border border-gray-200 dark:border-gray-600 px-3 py-1.5 text-sm text-gray-600 dark:text-gray-300 dark:text-gray-600 hover:bg-gray-50 dark:bg-gray-900 disabled:opacity-40 transition"
            >
              Next →
            </button>
          </div>
        )}
      </main>
  )
}
