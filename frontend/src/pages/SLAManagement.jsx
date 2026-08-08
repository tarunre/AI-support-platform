import { useState, useEffect } from 'react'
import { slaService } from '../services/slaService'
import SLABadge from '../components/SLABadge'

const PRIORITIES = ['LOW', 'MEDIUM', 'HIGH', 'URGENT']

const EMPTY_FORM = {
  name: '',
  priority: 'MEDIUM',
  first_response_minutes: 60,
  resolution_minutes: 480,
  is_default: false,
}

export default function SLAManagement() {
  const [policies, setPolicies] = useState([])
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState(EMPTY_FORM)
  const [editId, setEditId] = useState(null)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  const load = () => {
    slaService.listPolicies()
      .then((res) => setPolicies(res.data.data || []))
      .catch(() => setError('Failed to load SLA policies'))
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target
    setForm((prev) => ({ ...prev, [name]: type === 'checkbox' ? checked : value }))
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setSaving(true)
    setError('')
    setSuccess('')
    const payload = {
      ...form,
      first_response_minutes: parseInt(form.first_response_minutes, 10),
      resolution_minutes: parseInt(form.resolution_minutes, 10),
    }
    try {
      if (editId) {
        await slaService.updatePolicy(editId, payload)
        setSuccess('Policy updated.')
      } else {
        await slaService.createPolicy(payload)
        setSuccess('Policy created.')
      }
      setForm(EMPTY_FORM)
      setEditId(null)
      load()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to save policy')
    } finally {
      setSaving(false)
    }
  }

  const handleEdit = (p) => {
    setEditId(p.id)
    setForm({
      name: p.name,
      priority: p.priority,
      first_response_minutes: p.first_response_minutes,
      resolution_minutes: p.resolution_minutes,
      is_default: p.is_default,
    })
    setError('')
    setSuccess('')
  }

  const handleDelete = async (id) => {
    if (!window.confirm('Delete this SLA policy?')) return
    try {
      await slaService.deletePolicy(id)
      load()
    } catch {
      setError('Failed to delete policy')
    }
  }

  const handleCancel = () => {
    setEditId(null)
    setForm(EMPTY_FORM)
    setError('')
  }

  const fmtMinutes = (min) => {
    if (min < 60) return `${min}m`
    const h = Math.floor(min / 60)
    const m = min % 60
    return m ? `${h}h ${m}m` : `${h}h`
  }

  return (
    <div className="max-w-5xl mx-auto p-8">
      <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-1">SLA Management</h1>
      <p className="text-sm text-gray-500 dark:text-gray-400 dark:text-gray-500 mb-8">Define response and resolution time targets per ticket priority.</p>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Policy form */}
        <div className="lg:col-span-1">
          <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-100 dark:border-gray-700 p-6">
            <h2 className="text-sm font-semibold text-gray-700 dark:text-gray-200 mb-4">
              {editId ? 'Edit Policy' : 'New Policy'}
            </h2>

            {error && <p className="mb-3 text-xs text-red-600 bg-red-50 border border-red-200 rounded-lg px-3 py-2">{error}</p>}
            {success && <p className="mb-3 text-xs text-green-700 bg-green-50 border border-green-200 rounded-lg px-3 py-2">{success}</p>}

            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-gray-600 dark:text-gray-300 dark:text-gray-600 mb-1">Name</label>
                <input
                  name="name" value={form.name} onChange={handleChange} required
                  placeholder="e.g. High Priority SLA"
                  className="w-full border border-gray-200 dark:border-gray-600 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-600 dark:text-gray-300 dark:text-gray-600 mb-1">Priority</label>
                <select name="priority" value={form.priority} onChange={handleChange}
                  className="select-field w-full">
                  {PRIORITIES.map((p) => <option key={p}>{p}</option>)}
                </select>
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-600 dark:text-gray-300 dark:text-gray-600 mb-1">First Response (minutes)</label>
                <input
                  type="number" name="first_response_minutes" value={form.first_response_minutes}
                  onChange={handleChange} min={1} required
                  className="w-full border border-gray-200 dark:border-gray-600 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-600 dark:text-gray-300 dark:text-gray-600 mb-1">Resolution (minutes)</label>
                <input
                  type="number" name="resolution_minutes" value={form.resolution_minutes}
                  onChange={handleChange} min={1} required
                  className="w-full border border-gray-200 dark:border-gray-600 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" name="is_default" checked={form.is_default} onChange={handleChange}
                  className="w-4 h-4 text-blue-600 rounded" />
                <span className="text-sm text-gray-600 dark:text-gray-300 dark:text-gray-600">Default policy</span>
              </label>

              <div className="flex gap-2">
                <button type="submit" disabled={saving}
                  className="flex-1 py-2 bg-blue-600 text-white text-sm font-semibold rounded-lg hover:bg-blue-700 disabled:opacity-50 transition">
                  {saving ? 'Saving…' : editId ? 'Update' : 'Create'}
                </button>
                {editId && (
                  <button type="button" onClick={handleCancel}
                    className="px-4 py-2 border border-gray-200 dark:border-gray-600 text-sm text-gray-600 dark:text-gray-300 dark:text-gray-600 rounded-lg hover:bg-gray-50 dark:bg-gray-900">
                    Cancel
                  </button>
                )}
              </div>
            </form>
          </div>
        </div>

        {/* Policy list */}
        <div className="lg:col-span-2">
          {loading ? (
            <p className="text-gray-400 dark:text-gray-500 text-sm p-4">Loading…</p>
          ) : policies.length === 0 ? (
            <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-100 dark:border-gray-700 p-8 text-center text-gray-400 dark:text-gray-500 text-sm">
              No SLA policies yet. Create one to get started.
            </div>
          ) : (
            <div className="space-y-3">
              {policies.map((p) => (
                <div key={p.id} className="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 shadow-sm p-4 flex items-center justify-between">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-gray-900 dark:text-white text-sm">{p.name}</span>
                      {p.is_default && (
                        <span className="text-xs bg-blue-50 text-blue-600 font-medium px-2 py-0.5 rounded-full">Default</span>
                      )}
                    </div>
                    <div className="flex items-center gap-3 text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500">
                      <span className="capitalize font-medium text-gray-700 dark:text-gray-200">{p.priority}</span>
                      <span>·</span>
                      <span>First response: <strong className="text-gray-700 dark:text-gray-200">{fmtMinutes(p.first_response_minutes)}</strong></span>
                      <span>·</span>
                      <span>Resolution: <strong className="text-gray-700 dark:text-gray-200">{fmtMinutes(p.resolution_minutes)}</strong></span>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <button onClick={() => handleEdit(p)}
                      className="text-xs text-blue-600 hover:text-blue-800 font-medium px-2 py-1">
                      Edit
                    </button>
                    <button onClick={() => handleDelete(p.id)}
                      className="text-xs text-red-500 hover:text-red-700 font-medium px-2 py-1">
                      Delete
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
