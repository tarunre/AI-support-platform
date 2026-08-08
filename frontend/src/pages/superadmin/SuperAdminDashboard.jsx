import { useState, useEffect } from 'react'
import { tenantService } from '../../services/tenantService'

export default function SuperAdminDashboard() {
  const [overview, setOverview] = useState(null)
  const [tenants, setTenants] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([tenantService.getOverview(), tenantService.list()])
      .then(([ovRes, tenantsRes]) => {
        setOverview(ovRes.data.data || ovRes.data)
        setTenants((tenantsRes.data.data || tenantsRes.data?.tenants || []))
      })
      .catch(() => setError('Failed to load platform data'))
      .finally(() => setLoading(false))
  }, [])

  const handleDelete = async (id) => {
    if (!window.confirm('Delete this tenant? This cannot be undone.')) return
    try {
      await tenantService.remove(id)
      setTenants((prev) => prev.filter((t) => t.id !== id))
    } catch {
      alert('Failed to delete tenant.')
    }
  }

  if (loading) return <div className="p-8 text-gray-500 dark:text-gray-400 dark:text-gray-500">Loading platform data…</div>
  if (error) return <div className="p-8 text-red-600">{error}</div>

  return (
    <div className="max-w-6xl mx-auto px-6 py-6 space-y-6">
      <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100 mb-6">SuperAdmin Dashboard</h1>

      {overview && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
          {[
            { label: 'Total Tenants', value: overview.total_tenants ?? overview.totalTenants },
            { label: 'Active Tenants', value: overview.active_tenants ?? overview.activeTenants },
            { label: 'Total Users', value: overview.total_users ?? overview.totalUsers },
            { label: 'Total Tickets', value: overview.total_tickets ?? overview.totalTickets },
          ].map((stat) => (
            <div key={stat.label} className="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 shadow-sm p-6 text-center">
              <p className="text-3xl font-bold text-blue-600">{stat.value ?? '—'}</p>
              <p className="text-sm text-gray-500 dark:text-gray-400 dark:text-gray-500 mt-1">{stat.label}</p>
            </div>
          ))}
        </div>
      )}

      <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-100 dark:border-gray-700 overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-100 dark:border-gray-700">
          <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100">All Tenants</h2>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 dark:bg-gray-900 text-gray-600 dark:text-gray-300 dark:text-gray-600 text-xs uppercase">
              <tr>
                <th className="px-6 py-3 text-left">Name</th>
                <th className="px-6 py-3 text-left">Domain</th>
                <th className="px-6 py-3 text-left">Status</th>
                <th className="px-6 py-3 text-left">Plan</th>
                <th className="px-6 py-3 text-left">Created</th>
                <th className="px-6 py-3"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {tenants.map((t) => (
                <tr key={t.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                  <td className="px-6 py-4 font-medium text-gray-800 dark:text-gray-100">{t.name}</td>
                  <td className="px-6 py-4 text-gray-500 dark:text-gray-400 dark:text-gray-500">{t.domain || '—'}</td>
                  <td className="px-6 py-4">
                    <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${
                      t.status === 'active' ? 'bg-green-100 text-green-700' : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 dark:text-gray-600'
                    }`}>
                      {t.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-gray-500 dark:text-gray-400 dark:text-gray-500 capitalize">{t.plan}</td>
                  <td className="px-6 py-4 text-gray-400 dark:text-gray-500">
                    {t.created_at ? new Date(t.created_at).toLocaleDateString() : '—'}
                  </td>
                  <td className="px-6 py-4 text-right">
                    <button
                      onClick={() => handleDelete(t.id)}
                      className="text-red-500 hover:text-red-700 text-xs font-medium"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
              {tenants.length === 0 && (
                <tr>
                  <td colSpan={6} className="px-6 py-8 text-center text-gray-400 dark:text-gray-500">No tenants found.</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
