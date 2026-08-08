import { useState, useEffect, useCallback } from 'react'
import {
  AreaChart, Area, BarChart, Bar, PieChart, Pie, Cell,
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip,
  Legend, ResponsiveContainer
} from 'recharts'
import analyticsService from '../../services/analyticsService'
import useWebSocket from '../../hooks/useWebSocket'

const PERIOD_OPTIONS = [
  { label: 'Today', value: 'today' },
  { label: 'Yesterday', value: 'yesterday' },
  { label: 'Last 7 days', value: 'last7' },
  { label: 'Last 30 days', value: 'last30' },
  { label: 'Last 90 days', value: 'last90' },
]

const COLORS = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#06B6D4', '#F97316', '#84CC16']

function StatCard({ label, value, sub, color = 'blue', icon }) {
  const colorMap = {
    blue: 'bg-blue-50 text-blue-700 border-blue-200',
    green: 'bg-green-50 text-green-700 border-green-200',
    yellow: 'bg-yellow-50 text-yellow-700 border-yellow-200',
    red: 'bg-red-50 text-red-700 border-red-200',
    purple: 'bg-purple-50 text-purple-700 border-purple-200',
    cyan: 'bg-cyan-50 text-cyan-700 border-cyan-200',
    orange: 'bg-orange-50 text-orange-700 border-orange-200',
    indigo: 'bg-indigo-50 text-indigo-700 border-indigo-200',
  }
  return (
    <div className={`rounded-xl border p-4 ${colorMap[color] || colorMap.blue}`}>
      <div className="flex items-start justify-between">
        <p className="text-xs font-semibold uppercase tracking-wide opacity-70">{label}</p>
        {icon && <span className="text-xl">{icon}</span>}
      </div>
      <p className="mt-2 text-3xl font-bold">{value ?? '—'}</p>
      {sub && <p className="mt-1 text-xs opacity-60">{sub}</p>}
    </div>
  )
}

export default function AnalyticsDashboard() {
  const [period, setPeriod] = useState('last30')
  const [overview, setOverview] = useState(null)
  const [trends, setTrends] = useState(null)
  const [tickets, setTickets] = useState(null)
  const [loading, setLoading] = useState(true)
  const [lastRefresh, setLastRefresh] = useState(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [ov, tr, tk] = await Promise.all([
        analyticsService.getOverview(),
        analyticsService.getTrends({ period }),
        analyticsService.getTicketStats({ period }),
      ])
      setOverview(ov.data.data)
      setTrends(tr.data.data)
      setTickets(tk.data.data)
      setLastRefresh(new Date())
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }, [period])

  useEffect(() => { load() }, [load])

  // Listen for WebSocket analytics refresh events
  useWebSocket((msg) => {
    if (msg?.type === 'ANALYTICS_REFRESH') load()
  })

  const trendPoints = trends?.points ?? []
  const statusData = (tickets?.by_status ?? []).map(p => ({ name: p.label, value: Number(p.count) }))
  const priorityData = (tickets?.by_priority ?? []).map(p => ({ name: p.label, value: Number(p.count) }))
  const categoryData = (tickets?.by_category ?? []).slice(0, 8).map(p => ({ name: p.label, value: Number(p.count) }))

  return (
    <div className="p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Analytics Dashboard</h1>
          {lastRefresh && (
            <p className="text-xs text-gray-400 dark:text-gray-500 mt-0.5">Last updated: {lastRefresh.toLocaleTimeString()}</p>
          )}
        </div>
        <div className="flex items-center gap-3">
          <select
            value={period}
            onChange={e => setPeriod(e.target.value)}
            className="select-field"
          >
            {PERIOD_OPTIONS.map(o => (
              <option key={o.value} value={o.value}>{o.label}</option>
            ))}
          </select>
          <button
            onClick={load}
            className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
          >
            Refresh
          </button>
        </div>
      </div>

      {loading && !overview ? (
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
        </div>
      ) : (
        <>
          {/* Top Stat Cards */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <StatCard label="Total Tickets" value={overview?.total_tickets?.toLocaleString()} color="blue" icon="🎫" />
            <StatCard label="Open Tickets" value={overview?.open_tickets?.toLocaleString()} color="yellow" icon="📂"
              sub={`${overview?.in_progress_tickets ?? 0} in progress`} />
            <StatCard label="Resolved Today" value={overview?.resolved_today?.toLocaleString()} color="green" icon="✅"
              sub={`${overview?.created_today ?? 0} created today`} />
            <StatCard label="Avg Resolution" value={overview?.avg_resolution_hours != null ? `${overview.avg_resolution_hours}h` : '—'} color="indigo" icon="⏱️"
              sub="last 30 days" />
          </div>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <StatCard label="AI Confidence" value={overview?.avg_ai_confidence != null ? `${overview.avg_ai_confidence}%` : '—'} color="purple" icon="🤖" />
            <StatCard label="AI Approval Rate" value={overview?.ai_approval_rate != null ? `${overview.ai_approval_rate}%` : '—'} color="cyan" icon="👍" />
            <StatCard label="Queued Jobs" value={overview?.queued_jobs?.toLocaleString()} color="orange" icon="⚙️"
              sub={`${overview?.failed_jobs ?? 0} failed`} />
            <StatCard label="Emails Today" value={overview?.emails_processed_today?.toLocaleString()} color="green" icon="📧" />
          </div>

          {/* Priority alert bar */}
          {(overview?.urgent_tickets > 0 || overview?.high_priority_tickets > 0) && (
            <div className="mb-6 flex gap-3 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
              <span>⚠️</span>
              <span>
                <strong>{overview.urgent_tickets}</strong> urgent &amp; <strong>{overview.high_priority_tickets}</strong> high-priority tickets require attention.
              </span>
            </div>
          )}

          {/* Trend Chart */}
          <div className="mb-6 rounded-xl border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 p-5">
            <h2 className="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">Ticket Volume Trend</h2>
            {trendPoints.length > 0 ? (
              <ResponsiveContainer width="100%" height={240}>
                <AreaChart data={trendPoints}>
                  <defs>
                    <linearGradient id="created" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#3B82F6" stopOpacity={0.15} />
                      <stop offset="95%" stopColor="#3B82F6" stopOpacity={0} />
                    </linearGradient>
                    <linearGradient id="resolved" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#10B981" stopOpacity={0.15} />
                      <stop offset="95%" stopColor="#10B981" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                  <XAxis dataKey="date" tick={{ fontSize: 11 }} />
                  <YAxis tick={{ fontSize: 11 }} />
                  <Tooltip />
                  <Legend />
                  <Area type="monotone" dataKey="tickets_created" name="Created" stroke="#3B82F6" fill="url(#created)" strokeWidth={2} />
                  <Area type="monotone" dataKey="tickets_resolved" name="Resolved" stroke="#10B981" fill="url(#resolved)" strokeWidth={2} />
                  <Area type="monotone" dataKey="tickets_closed" name="Closed" stroke="#8B5CF6" fill="none" strokeWidth={2} strokeDasharray="4 2" />
                </AreaChart>
              </ResponsiveContainer>
            ) : (
              <p className="text-center text-sm text-gray-400 dark:text-gray-500 py-16">No trend data for this period. Run aggregation to populate.</p>
            )}
          </div>

          {/* Distribution charts */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
            {/* Status Pie */}
            <div className="rounded-xl border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 p-5">
              <h2 className="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">Status Distribution</h2>
              {statusData.length > 0 ? (
                <ResponsiveContainer width="100%" height={200}>
                  <PieChart>
                    <Pie data={statusData} cx="50%" cy="50%" innerRadius={50} outerRadius={80}
                      dataKey="value" nameKey="name" label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                      labelLine={false} fontSize={11}>
                      {statusData.map((_, i) => <Cell key={i} fill={COLORS[i % COLORS.length]} />)}
                    </Pie>
                    <Tooltip />
                  </PieChart>
                </ResponsiveContainer>
              ) : <p className="text-center text-sm text-gray-400 dark:text-gray-500 py-16">No data</p>}
            </div>

            {/* Priority Bar */}
            <div className="rounded-xl border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 p-5">
              <h2 className="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">Priority Distribution</h2>
              {priorityData.length > 0 ? (
                <ResponsiveContainer width="100%" height={200}>
                  <BarChart data={priorityData} layout="vertical">
                    <CartesianGrid strokeDasharray="3 3" horizontal={false} />
                    <XAxis type="number" tick={{ fontSize: 11 }} />
                    <YAxis type="category" dataKey="name" tick={{ fontSize: 11 }} width={70} />
                    <Tooltip />
                    <Bar dataKey="value" name="Tickets" radius={[0, 4, 4, 0]}>
                      {priorityData.map((_, i) => <Cell key={i} fill={COLORS[i % COLORS.length]} />)}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              ) : <p className="text-center text-sm text-gray-400 dark:text-gray-500 py-16">No data</p>}
            </div>

            {/* Category Bar */}
            <div className="rounded-xl border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 p-5">
              <h2 className="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">Top Categories</h2>
              {categoryData.length > 0 ? (
                <ResponsiveContainer width="100%" height={200}>
                  <BarChart data={categoryData} layout="vertical">
                    <CartesianGrid strokeDasharray="3 3" horizontal={false} />
                    <XAxis type="number" tick={{ fontSize: 11 }} />
                    <YAxis type="category" dataKey="name" tick={{ fontSize: 10 }} width={80} />
                    <Tooltip />
                    <Bar dataKey="value" name="Tickets" fill="#8B5CF6" radius={[0, 4, 4, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              ) : <p className="text-center text-sm text-gray-400 dark:text-gray-500 py-16">No data</p>}
            </div>
          </div>

          {/* Tickets by Hour */}
          <div className="rounded-xl border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 p-5">
            <h2 className="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">Tickets Created by Hour (Today)</h2>
            {(tickets?.by_hour?.length > 0) ? (
              <ResponsiveContainer width="100%" height={160}>
                <BarChart data={tickets.by_hour.map(p => ({ hour: p.label + ':00', count: Number(p.count) }))}>
                  <CartesianGrid strokeDasharray="3 3" vertical={false} />
                  <XAxis dataKey="hour" tick={{ fontSize: 10 }} />
                  <YAxis tick={{ fontSize: 10 }} allowDecimals={false} />
                  <Tooltip />
                  <Bar dataKey="count" name="Tickets" fill="#06B6D4" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            ) : <p className="text-center text-sm text-gray-400 dark:text-gray-500 py-8">No tickets created today yet.</p>}
          </div>
        </>
      )}
    </div>
  )
}
