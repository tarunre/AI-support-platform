import { useState, useEffect, useCallback } from 'react'
import {
  LineChart, Line, BarChart, Bar, PieChart, Pie, Cell,
  XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer
} from 'recharts'
import analyticsService from '../../services/analyticsService'

const PERIOD_OPTIONS = [
  { label: 'Today', value: 'today' },
  { label: 'Last 7 days', value: 'last7' },
  { label: 'Last 30 days', value: 'last30' },
  { label: 'Last 90 days', value: 'last90' },
]
const COLORS = ['#8B5CF6', '#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#06B6D4']

function Metric({ label, value, sub, highlight }) {
  return (
    <div className={`rounded-xl border p-4 ${highlight ? 'bg-purple-50 border-purple-200' : 'bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-600'}`}>
      <p className="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400 dark:text-gray-500">{label}</p>
      <p className="mt-2 text-3xl font-bold text-gray-900 dark:text-white">{value ?? '—'}</p>
      {sub && <p className="mt-1 text-xs text-gray-400 dark:text-gray-500">{sub}</p>}
    </div>
  )
}

export default function AIInsights() {
  const [period, setPeriod] = useState('last30')
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await analyticsService.getAIStats({ period })
      setData(res.data.data)
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }, [period])

  useEffect(() => { load() }, [load])

  const dailyTrend = (data?.daily_trend ?? []).map(m => ({
    date: m.date?.slice(0, 10),
    analyses: m.analysis_generated,
    replies: m.replies_generated,
    confidence: m.average_confidence,
    approval: m.approval_rate,
  }))

  const rateData = data ? [
    { name: 'Approved', value: Number(data.approval_rate) },
    { name: 'Edited', value: Number(data.edit_rate) },
    { name: 'Rejected', value: Number(data.rejection_rate) },
    { name: 'Retried', value: Number(data.retry_rate) },
  ] : []

  const catData = (data?.top_categories ?? []).map(c => ({ name: c.label, count: Number(c.count) }))
  const sentData = (data?.top_sentiments ?? []).map(c => ({ name: c.label, count: Number(c.count) }))

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">AI Insights</h1>
        <select
          value={period}
          onChange={e => setPeriod(e.target.value)}
          className="select-field"
        >
          {PERIOD_OPTIONS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
        </select>
      </div>

      {loading ? (
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-purple-600 border-t-transparent" />
        </div>
      ) : (
        <>
          {/* Metrics */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <Metric label="Analyses Generated" value={data?.total_analyses?.toLocaleString()} />
            <Metric label="Replies Generated" value={data?.total_replies?.toLocaleString()} />
            <Metric label="Avg Confidence" value={data?.avg_confidence != null ? `${data.avg_confidence}%` : '—'} highlight />
            <Metric label="Avg Generation" value={data?.avg_generation_ms != null ? `${data.avg_generation_ms}ms` : '—'} />
          </div>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <Metric label="Approval Rate" value={data?.approval_rate != null ? `${data.approval_rate}%` : '—'} sub="of generated replies" />
            <Metric label="Edit Rate" value={data?.edit_rate != null ? `${data.edit_rate}%` : '—'} sub="approved with edits" />
            <Metric label="Rejection Rate" value={data?.rejection_rate != null ? `${data.rejection_rate}%` : '—'} sub="rejected by agents" />
            <Metric label="Retry Count" value={data?.retry_count?.toLocaleString()} sub={`${data?.failure_count ?? 0} failures`} />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
            {/* Approval rate donut */}
            <div className="rounded-xl border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 p-5">
              <h2 className="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">Reply Outcome Distribution</h2>
              {rateData.some(r => r.value > 0) ? (
                <ResponsiveContainer width="100%" height={220}>
                  <PieChart>
                    <Pie data={rateData} cx="50%" cy="50%" innerRadius={55} outerRadius={85}
                      dataKey="value" nameKey="name"
                      label={({ name, value }) => value > 0 ? `${name} ${value}%` : ''}
                      labelLine={false} fontSize={11}>
                      {rateData.map((_, i) => <Cell key={i} fill={COLORS[i % COLORS.length]} />)}
                    </Pie>
                    <Tooltip formatter={(v) => `${v}%`} />
                    <Legend />
                  </PieChart>
                </ResponsiveContainer>
              ) : <p className="py-16 text-center text-sm text-gray-400 dark:text-gray-500">No AI reply data for this period</p>}
            </div>

            {/* Confidence + approval trend */}
            <div className="rounded-xl border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 p-5">
              <h2 className="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">Confidence & Approval Trend</h2>
              {dailyTrend.length > 0 ? (
                <ResponsiveContainer width="100%" height={220}>
                  <LineChart data={dailyTrend}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                    <XAxis dataKey="date" tick={{ fontSize: 10 }} />
                    <YAxis tick={{ fontSize: 10 }} domain={[0, 100]} />
                    <Tooltip />
                    <Legend />
                    <Line type="monotone" dataKey="confidence" name="Confidence %" stroke="#8B5CF6" strokeWidth={2} dot={false} />
                    <Line type="monotone" dataKey="approval" name="Approval %" stroke="#10B981" strokeWidth={2} dot={false} />
                  </LineChart>
                </ResponsiveContainer>
              ) : <p className="py-16 text-center text-sm text-gray-400 dark:text-gray-500">No trend data yet</p>}
            </div>
          </div>

          {/* Daily activity */}
          <div className="rounded-xl border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 p-5 mb-6">
            <h2 className="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">Daily AI Activity</h2>
            {dailyTrend.length > 0 ? (
              <ResponsiveContainer width="100%" height={200}>
                <BarChart data={dailyTrend}>
                  <CartesianGrid strokeDasharray="3 3" vertical={false} />
                  <XAxis dataKey="date" tick={{ fontSize: 10 }} />
                  <YAxis tick={{ fontSize: 10 }} allowDecimals={false} />
                  <Tooltip />
                  <Legend />
                  <Bar dataKey="analyses" name="Analyses" fill="#8B5CF6" radius={[4, 4, 0, 0]} />
                  <Bar dataKey="replies" name="Replies" fill="#3B82F6" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            ) : <p className="py-8 text-center text-sm text-gray-400 dark:text-gray-500">No data — aggregation not yet run</p>}
          </div>

          {/* Top categories and sentiments */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="rounded-xl border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 p-5">
              <h2 className="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">Top AI Categories</h2>
              {catData.length > 0 ? (
                <ResponsiveContainer width="100%" height={200}>
                  <BarChart data={catData} layout="vertical">
                    <CartesianGrid strokeDasharray="3 3" horizontal={false} />
                    <XAxis type="number" tick={{ fontSize: 11 }} />
                    <YAxis type="category" dataKey="name" tick={{ fontSize: 10 }} width={90} />
                    <Tooltip />
                    <Bar dataKey="count" fill="#8B5CF6" radius={[0, 4, 4, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              ) : <p className="py-12 text-center text-sm text-gray-400 dark:text-gray-500">No category data</p>}
            </div>

            <div className="rounded-xl border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 p-5">
              <h2 className="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">Sentiment Distribution</h2>
              {sentData.length > 0 ? (
                <ResponsiveContainer width="100%" height={200}>
                  <PieChart>
                    <Pie data={sentData} cx="50%" cy="50%" outerRadius={80}
                      dataKey="count" nameKey="name"
                      label={({ name }) => name} labelLine={false} fontSize={11}>
                      {sentData.map((_, i) => <Cell key={i} fill={COLORS[i % COLORS.length]} />)}
                    </Pie>
                    <Tooltip />
                    <Legend />
                  </PieChart>
                </ResponsiveContainer>
              ) : <p className="py-12 text-center text-sm text-gray-400 dark:text-gray-500">No sentiment data</p>}
            </div>
          </div>
        </>
      )}
    </div>
  )
}
