import { useState, useEffect } from 'react'
import emailService from '../services/emailService'

function StatCard({ label, value, color, sub }) {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 shadow-sm p-5">
      <p className="text-sm text-gray-500 dark:text-gray-400 dark:text-gray-500">{label}</p>
      <p className={`text-3xl font-bold mt-1 ${color}`}>{value ?? 0}</p>
      {sub && <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">{sub}</p>}
    </div>
  )
}

const STATUS_COLORS = {
  RECEIVED:  'bg-blue-100 text-blue-700',
  QUEUED:    'bg-yellow-100 text-yellow-700',
  SENT:      'bg-green-100 text-green-700',
  FAILED:    'bg-red-100 text-red-700',
  DELIVERED: 'bg-emerald-100 text-emerald-700',
  READ:      'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 dark:text-gray-600',
}

export default function EmailMonitor() {
  const [stats, setStats] = useState(null)
  const [loading, setLoading] = useState(true)
  const [syncing, setSyncing] = useState(false)

  const load = async () => {
    try {
      const res = await emailService.getMonitor()
      setStats(res.data?.data)
    } catch {
      // stats stay null
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
    const interval = setInterval(load, 30000)
    return () => clearInterval(interval)
  }, [])

  const handleSync = async () => {
    setSyncing(true)
    try {
      await emailService.triggerSync()
    } finally {
      setSyncing(false)
      setTimeout(load, 2000)
    }
  }

  return (
    <div className="max-w-6xl mx-auto px-6 py-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Email Monitor</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 dark:text-gray-500 mt-1">
            {stats?.last_sync_at
              ? `Last sync: ${new Date(stats.last_sync_at).toLocaleString()}`
              : 'No sync recorded yet'}
          </p>
        </div>
        <button
          onClick={handleSync}
          disabled={syncing}
          className="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700 disabled:opacity-50"
        >
          {syncing ? 'Syncing…' : 'Sync Now'}
        </button>
      </div>

      {loading ? (
        <div className="text-center py-16 text-gray-400 dark:text-gray-500">Loading…</div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4 mb-8">
            <StatCard label="Total Accounts"   value={stats?.total_accounts}   color="text-gray-800 dark:text-gray-100" />
            <StatCard label="Active Accounts"  value={stats?.active_accounts}  color="text-green-600" />
            <StatCard label="Queued Emails"    value={stats?.queued_count}     color="text-yellow-600" />
            <StatCard label="Failed Emails"    value={stats?.failed_count}     color="text-red-600" />
            <StatCard label="Sent Today"       value={stats?.sent_today}       color="text-emerald-600" />
            <StatCard label="Received Today"   value={stats?.received_today}   color="text-blue-600" />
          </div>

          <h2 className="text-sm font-semibold text-gray-500 dark:text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-3">Connected Accounts</h2>
          {!stats?.accounts?.length ? (
            <p className="text-sm text-gray-400 dark:text-gray-500">No email accounts configured.</p>
          ) : (
            <div className="grid gap-3">
              {stats.accounts.map(a => (
                <div key={a.id} className="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 shadow-sm p-4 flex items-center justify-between">
                  <div>
                    <div className="flex items-center gap-3">
                      <span className="font-medium text-gray-900 dark:text-white">{a.email_address}</span>
                      <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${a.is_active ? 'bg-green-100 text-green-700' : 'bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 dark:text-gray-500'}`}>
                        {a.is_active ? 'Active' : 'Inactive'}
                      </span>
                      <span className="text-xs text-gray-400 dark:text-gray-500">{a.provider}</span>
                    </div>
                    <div className="flex gap-4 mt-1 text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500">
                      {a.imap_host && <span>IMAP: {a.imap_host}:{a.imap_port}</span>}
                      {a.smtp_host && <span>SMTP: {a.smtp_host}:{a.smtp_port}</span>}
                    </div>
                  </div>
                  <div className="text-xs text-gray-400 dark:text-gray-500">
                    {a.last_sync_at ? `Synced ${new Date(a.last_sync_at).toLocaleString()}` : 'Never synced'}
                  </div>
                </div>
              ))}
            </div>
          )}
        </>
      )}
    </div>
  )
}
