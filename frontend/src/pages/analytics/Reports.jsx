import { useState, useEffect, useCallback } from 'react'
import analyticsService from '../../services/analyticsService'

const PERIOD_OPTIONS = [
  { label: 'Today', value: 'today' },
  { label: 'Last 7 days', value: 'last7' },
  { label: 'Last 30 days', value: 'last30' },
  { label: 'Last 90 days', value: 'last90' },
]

const REPORT_TYPES = [
  { label: 'Tickets', value: 'tickets' },
  { label: 'Agent Performance', value: 'agents' },
  { label: 'AI Performance', value: 'ai' },
  { label: 'Email Activity', value: 'email' },
]

const FORMAT_OPTIONS = [
  { label: 'CSV', value: 'CSV' },
  { label: 'Excel (.xlsx)', value: 'EXCEL' },
  { label: 'HTML', value: 'HTML' },
]

const STATUS_COLORS = {
  PENDING: 'bg-yellow-100 text-yellow-700',
  COMPLETED: 'bg-green-100 text-green-700',
  FAILED: 'bg-red-100 text-red-700',
}

function StatusBadge({ status }) {
  return (
    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_COLORS[status] ?? 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 dark:text-gray-600'}`}>
      {status}
    </span>
  )
}

export default function Reports() {
  const [reports, setReports] = useState([])
  const [loading, setLoading] = useState(true)
  const [generating, setGenerating] = useState(false)
  const [form, setForm] = useState({
    name: '',
    report_type: 'tickets',
    format: 'CSV',
    period: 'last30',
  })
  const [toast, setToast] = useState(null)

  const showToast = (msg, type = 'success') => {
    setToast({ msg, type })
    setTimeout(() => setToast(null), 3500)
  }

  const loadReports = useCallback(async () => {
    setLoading(true)
    try {
      const res = await analyticsService.listReports()
      setReports(res.data.data ?? [])
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { loadReports() }, [loadReports])

  // Poll PENDING reports to detect completion
  useEffect(() => {
    const pending = reports.filter(r => r.status === 'PENDING')
    if (pending.length === 0) return
    const t = setTimeout(loadReports, 3000)
    return () => clearTimeout(t)
  }, [reports, loadReports])

  const handleGenerate = async (e) => {
    e.preventDefault()
    if (!form.name.trim()) return showToast('Report name is required', 'error')
    setGenerating(true)
    try {
      await analyticsService.generateReport(form)
      showToast('Report queued — it will appear below when ready')
      setForm(f => ({ ...f, name: '' }))
      setTimeout(loadReports, 800)
    } catch (err) {
      showToast(err?.response?.data?.error ?? 'Failed to generate report', 'error')
    } finally {
      setGenerating(false)
    }
  }

  const handleDownload = async (report) => {
    if (report.status !== 'COMPLETED') return
    try {
      const res = await analyticsService.downloadReport(report.id)
      const blob = res.data
      const ext = report.format === 'EXCEL' ? '.xlsx' : report.format === 'CSV' ? '.csv' : '.html'
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = report.name + ext
      document.body.appendChild(a)
      a.click()
      a.remove()
      URL.revokeObjectURL(url)
    } catch {
      showToast('Download failed', 'error')
    }
  }

  const formatSize = (bytes) => {
    if (!bytes || bytes === 0) return '—'
    if (bytes < 1024) return `${bytes}B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB`
    return `${(bytes / 1024 / 1024).toFixed(1)}MB`
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">Reports</h1>

      {/* Generator form */}
      <div className="rounded-xl border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 p-5 mb-6">
        <h2 className="text-sm font-semibold text-gray-700 dark:text-gray-200 mb-4">Generate New Report</h2>
        <form onSubmit={handleGenerate}>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
            <div className="lg:col-span-2">
              <label className="block text-xs font-medium text-gray-600 dark:text-gray-300 dark:text-gray-600 mb-1">Report Name</label>
              <input
                type="text"
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                placeholder="e.g. Q3 Ticket Report"
                className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 dark:text-gray-300 dark:text-gray-600 mb-1">Report Type</label>
              <select
                value={form.report_type}
                onChange={e => setForm(f => ({ ...f, report_type: e.target.value }))}
                className="select-field w-full"
              >
                {REPORT_TYPES.map(t => <option key={t.value} value={t.value}>{t.label}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 dark:text-gray-300 dark:text-gray-600 mb-1">Period</label>
              <select
                value={form.period}
                onChange={e => setForm(f => ({ ...f, period: e.target.value }))}
                className="select-field w-full"
              >
                {PERIOD_OPTIONS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 dark:text-gray-300 dark:text-gray-600 mb-1">Format</label>
              <select
                value={form.format}
                onChange={e => setForm(f => ({ ...f, format: e.target.value }))}
                className="select-field w-full"
              >
                {FORMAT_OPTIONS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
              </select>
            </div>
          </div>
          <div className="mt-4 flex justify-end">
            <button
              type="submit"
              disabled={generating}
              className="rounded-lg bg-blue-600 px-5 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {generating ? 'Queuing…' : 'Generate Report'}
            </button>
          </div>
        </form>
      </div>

      {/* Reports table */}
      <div className="rounded-xl border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 overflow-hidden">
        <div className="flex items-center justify-between px-5 py-3 border-b border-gray-100 dark:border-gray-700">
          <h2 className="text-sm font-semibold text-gray-700 dark:text-gray-200">Report History</h2>
          <button onClick={loadReports} className="text-xs text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:text-gray-300 dark:text-gray-600">↻ Refresh</button>
        </div>
        {loading ? (
          <div className="flex h-40 items-center justify-center">
            <div className="h-6 w-6 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
          </div>
        ) : reports.length === 0 ? (
          <div className="py-16 text-center text-sm text-gray-400 dark:text-gray-500">No reports generated yet.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-900 text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500">
                  <th className="px-4 py-3 text-left font-medium">Name</th>
                  <th className="px-4 py-3 text-left font-medium">Type</th>
                  <th className="px-4 py-3 text-left font-medium">Format</th>
                  <th className="px-4 py-3 text-left font-medium">Status</th>
                  <th className="px-4 py-3 text-right font-medium">Size</th>
                  <th className="px-4 py-3 text-right font-medium">Generated</th>
                  <th className="px-4 py-3 text-right font-medium">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-50">
                {reports.map(r => (
                  <tr key={r.id} className="hover:bg-gray-50 dark:bg-gray-900 transition-colors">
                    <td className="px-4 py-3 font-medium text-gray-900 dark:text-white">{r.name}</td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400 dark:text-gray-500 capitalize">{r.report_type}</td>
                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400 dark:text-gray-500">{r.format}</td>
                    <td className="px-4 py-3">
                      <StatusBadge status={r.status} />
                      {r.status === 'PENDING' && (
                        <span className="ml-2 inline-block h-3 w-3 animate-pulse rounded-full bg-yellow-400" />
                      )}
                    </td>
                    <td className="px-4 py-3 text-right text-gray-500 dark:text-gray-400 dark:text-gray-500">{formatSize(r.file_size)}</td>
                    <td className="px-4 py-3 text-right text-gray-400 dark:text-gray-500 text-xs">
                      {new Date(r.created_at).toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-right">
                      {r.status === 'COMPLETED' ? (
                        <button
                          onClick={() => handleDownload(r)}
                          className="text-blue-600 hover:text-blue-700 font-medium text-xs"
                        >
                          Download ↓
                        </button>
                      ) : r.status === 'FAILED' ? (
                        <span className="text-xs text-red-500" title={r.error_message}>Failed</span>
                      ) : (
                        <span className="text-xs text-gray-400 dark:text-gray-500">Processing…</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Toast */}
      {toast && (
        <div className={`fixed bottom-4 right-4 z-50 rounded-lg px-4 py-3 text-sm font-medium shadow-lg ${
          toast.type === 'error' ? 'bg-red-500 text-white' : 'bg-gray-900 text-white'
        }`}>
          {toast.msg}
        </div>
      )}
    </div>
  )
}
