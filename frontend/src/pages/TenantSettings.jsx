import { useState, useEffect } from 'react'
import { tenantService } from '../services/tenantService'

export default function TenantSettings() {
  const [form, setForm] = useState({ name: '', domain: '', plan: '' })
  const [slug, setSlug] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    tenantService.getSettings()
      .then((res) => {
        const t = res.data.data || res.data
        setForm({ name: t.name || '', domain: t.domain || '', plan: t.plan || '' })
        setSlug(t.slug || '')
      })
      .catch(() => setError('Failed to load settings'))
      .finally(() => setLoading(false))
  }, [])

  const handleChange = (e) => {
    setForm((prev) => ({ ...prev, [e.target.name]: e.target.value }))
    setMessage('')
    setError('')
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setSaving(true)
    setMessage('')
    setError('')
    try {
      await tenantService.updateSettings(form)
      setMessage('Settings saved.')
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to save settings')
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <div className="p-8 text-gray-500 dark:text-gray-400 dark:text-gray-500">Loading…</div>

  return (
    <div className="max-w-2xl mx-auto p-8">
      <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">Tenant Settings</h1>

      {/* Agent join info */}
      {slug && (
        <div className="mb-6 bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800 rounded-xl p-4">
          <p className="text-sm font-semibold text-emerald-700 dark:text-emerald-400 mb-1">🔗 Agent Join Code</p>
          <p className="text-xs text-emerald-600 dark:text-emerald-500 mb-2">Share this slug with agents so they can register at <strong>/agent-register</strong>:</p>
          <code className="block px-3 py-2 bg-white dark:bg-gray-800 border border-emerald-200 dark:border-emerald-700 rounded-lg text-sm font-mono text-emerald-700 dark:text-emerald-300 select-all">{slug}</code>
        </div>
      )}

      {message && (
        <div className="mb-4 px-4 py-3 bg-green-50 border border-green-200 text-green-700 rounded-lg text-sm">
          {message}
        </div>
      )}
      {error && (
        <div className="mb-4 px-4 py-3 bg-red-50 border border-red-200 text-red-600 rounded-lg text-sm">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-100 dark:border-gray-700 p-8 space-y-5">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1.5">Company name</label>
          <input
            type="text"
            name="name"
            value={form.name}
            onChange={handleChange}
            required
            className="w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1.5">Domain</label>
          <input
            type="text"
            name="domain"
            value={form.domain}
            onChange={handleChange}
            placeholder="acme.com"
            className="w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1.5">Subscription plan</label>
          <select
            name="plan"
            value={form.plan}
            onChange={handleChange}
            className="select-field w-full"
          >
            <option value="free">Free</option>
            <option value="starter">Starter</option>
            <option value="pro">Pro</option>
            <option value="enterprise">Enterprise</option>
          </select>
        </div>

        <button
          type="submit"
          disabled={saving}
          className="w-full py-2.5 bg-blue-600 text-white font-semibold rounded-lg hover:bg-blue-700 transition disabled:opacity-50 text-sm"
        >
          {saving ? 'Saving…' : 'Save settings'}
        </button>
      </form>
    </div>
  )
}
