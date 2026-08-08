import { useState, useEffect, useCallback } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, Cell } from 'recharts'
import analyticsService from '../../services/analyticsService'

const COLORS = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#06B6D4', '#F97316']

function Medal({ rank }) {
  if (rank === 1) return <span className="text-lg">🥇</span>
  if (rank === 2) return <span className="text-lg">🥈</span>
  if (rank === 3) return <span className="text-lg">🥉</span>
  return <span className="text-sm text-gray-400 dark:text-gray-500 font-mono w-5 text-center">{rank}</span>
}

function Bar2({ label, value, max, color = '#3B82F6' }) {
  const pct = max > 0 ? Math.min((value / max) * 100, 100) : 0
  return (
    <div className="flex items-center gap-3 text-sm">
      <span className="w-24 truncate text-gray-600 dark:text-gray-300 dark:text-gray-600">{label}</span>
      <div className="flex-1 h-2 rounded-full bg-gray-100 dark:bg-gray-800">
        <div className="h-2 rounded-full" style={{ width: `${pct}%`, backgroundColor: color }} />
      </div>
      <span className="w-8 text-right text-gray-700 dark:text-gray-200 font-medium">{value}</span>
    </div>
  )
}

export default function AgentPerformance() {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await analyticsService.getAgentStats()
      setData(res.data.data)
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const agents = data?.agents ?? []
  const leaderboard = data?.leaderboard ?? []
  const chartData = leaderboard.map(a => ({
    name: a.name?.split(' ')[0] ?? 'Agent',
    assigned: a.tickets_assigned,
    resolved: a.tickets_resolved,
    active: Number(a.active_tickets),
  }))

  const maxResolved = agents.reduce((m, a) => Math.max(m, a.tickets_resolved), 0)

  return (
    <div className="max-w-6xl mx-auto px-6 py-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">Agent Performance</h1>
        <button
          onClick={load}
          className="rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-800"
        >
          Refresh
        </button>
      </div>

      {loading ? (
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
        </div>
      ) : agents.length === 0 ? (
        <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 p-12 text-center text-gray-400 dark:text-gray-500">
          <p className="text-lg mb-1">No agent data yet</p>
          <p className="text-sm">Run aggregation to populate agent metrics.</p>
        </div>
      ) : (
        <>
          {/* Leaderboard chart */}
          <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 p-5 mb-6">
            <h2 className="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">Agent Activity Overview</h2>
            {chartData.length > 0 ? (
              <ResponsiveContainer width="100%" height={240}>
                <BarChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" vertical={false} />
                  <XAxis dataKey="name" tick={{ fontSize: 11 }} />
                  <YAxis tick={{ fontSize: 11 }} allowDecimals={false} />
                  <Tooltip />
                  <Legend />
                  <Bar dataKey="assigned" name="Assigned" fill="#3B82F6" radius={[4, 4, 0, 0]} />
                  <Bar dataKey="resolved" name="Resolved" fill="#10B981" radius={[4, 4, 0, 0]} />
                  <Bar dataKey="active" name="Active" fill="#F59E0B" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            ) : <p className="py-12 text-center text-sm text-gray-400 dark:text-gray-500">No data</p>}
          </div>

          {/* Resolution leaderboard */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
            <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 p-5">
              <h2 className="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">Resolution Leaderboard</h2>
              <div className="space-y-3">
                {agents.slice(0, 10).map((a, i) => (
                  <Bar2
                    key={a.user_id}
                    label={a.name || `Agent ${a.user_id}`}
                    value={a.tickets_resolved}
                    max={maxResolved}
                    color={COLORS[i % COLORS.length]}
                  />
                ))}
              </div>
            </div>

            <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 p-5">
              <h2 className="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">Avg Resolution Time (hours)</h2>
              <div className="space-y-3">
                {[...agents]
                  .sort((a, b) => a.average_resolution_time - b.average_resolution_time)
                  .slice(0, 10)
                  .map((a) => {
                    const maxTime = Math.max(...agents.map(x => x.average_resolution_time))
                    return (
                      <Bar2
                        key={a.user_id}
                        label={a.name || `Agent ${a.user_id}`}
                        value={a.average_resolution_time}
                        max={maxTime}
                        color="#8B5CF6"
                      />
                    )
                  })
                }
              </div>
            </div>
          </div>

          {/* Full agent table */}
          <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 overflow-hidden">
            <div className="px-5 py-3 border-b border-gray-100 dark:border-gray-700">
              <h2 className="text-sm font-semibold text-gray-700 dark:text-gray-200">All Agents</h2>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-900 text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500">
                    <th className="px-4 py-3 text-left font-medium">#</th>
                    <th className="px-4 py-3 text-left font-medium">Agent</th>
                    <th className="px-4 py-3 text-right font-medium">Assigned</th>
                    <th className="px-4 py-3 text-right font-medium">Resolved</th>
                    <th className="px-4 py-3 text-right font-medium">Active</th>
                    <th className="px-4 py-3 text-right font-medium">Avg Res. (h)</th>
                    <th className="px-4 py-3 text-right font-medium">Avg Reply (h)</th>
                    <th className="px-4 py-3 text-right font-medium">Last Updated</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-50">
                  {agents.map((a, i) => (
                    <tr key={a.user_id} className="hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
                      <td className="px-4 py-3"><Medal rank={i + 1} /></td>
                      <td className="px-4 py-3">
                        <div>
                          <p className="font-medium text-gray-800 dark:text-gray-100">{a.name || `Agent ${a.user_id}`}</p>
                          <p className="text-xs text-gray-400 dark:text-gray-500">{a.email}</p>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-right text-gray-700 dark:text-gray-200">{a.tickets_assigned}</td>
                      <td className="px-4 py-3 text-right font-semibold text-green-700">{a.tickets_resolved}</td>
                      <td className="px-4 py-3 text-right text-yellow-700">{a.active_tickets}</td>
                      <td className="px-4 py-3 text-right text-gray-700 dark:text-gray-200">{a.average_resolution_time}h</td>
                      <td className="px-4 py-3 text-right text-gray-700 dark:text-gray-200">{a.average_reply_time}h</td>
                      <td className="px-4 py-3 text-right text-gray-400 dark:text-gray-500 text-xs">
                        {a.last_calculated ? new Date(a.last_calculated).toLocaleDateString() : '—'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
