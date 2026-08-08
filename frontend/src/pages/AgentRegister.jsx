import { useState } from 'react'
import { Link, useNavigate, Navigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import DarkModeToggle from '../components/DarkModeToggle'

const TEAMS = ['Support', 'Engineering', 'Billing', 'Sales', 'Operations']

function TeamPicker({ selected, onSelect, onClose }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4">
      <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-xl border border-gray-100 dark:border-gray-700 p-6 w-full max-w-sm">
        <h2 className="text-base font-semibold text-gray-900 dark:text-white mb-1">Select team type</h2>
        <p className="text-xs text-gray-400 dark:text-gray-500 mb-4">All members of this team will share this login.</p>
        <div className="space-y-2">
          {TEAMS.map((team) => (
            <button
              key={team}
              type="button"
              onClick={() => { onSelect(team); onClose() }}
              className={`w-full text-left px-4 py-3 rounded-lg border text-sm font-medium transition ${
                selected === team
                  ? 'border-emerald-500 bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-300'
                  : 'border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700'
              }`}
            >
              {team}
            </button>
          ))}
        </div>
        <button
          type="button"
          onClick={onClose}
          className="mt-4 w-full text-sm text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 transition"
        >
          Cancel
        </button>
      </div>
    </div>
  )
}

export default function AgentRegister() {
  const { user, loading, agentJoin } = useAuth()
  const navigate = useNavigate()

  const [form, setForm] = useState({ name: '', email: '', password: '', company_slug: '', team: '' })
  const [errors, setErrors] = useState({})
  const [apiError, setApiError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [showTeamPicker, setShowTeamPicker] = useState(false)
  const [showPassword, setShowPassword] = useState(false)

  if (loading) return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
      <p className="text-sm text-gray-400 animate-pulse">Loading…</p>
    </div>
  )
  if (user) return <Navigate to={user.role === 'SupportAgent' ? '/agent' : '/dashboard'} replace />

  const handleChange = (e) => {
    setForm((p) => ({ ...p, [e.target.name]: e.target.value }))
    setErrors((p) => ({ ...p, [e.target.name]: '' }))
  }

  const validate = () => {
    const next = {}
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email))    next.email = 'Enter a valid email address.'
    if (form.password.length < 8)                           next.password = 'Password must be at least 8 characters.'
    if (!form.company_slug.trim())                          next.company_slug = 'Company slug is required.'
    if (!form.team)                                         next.team = 'Please select a team.'
    return next
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setApiError('')
    const fieldErrors = validate()
    if (Object.keys(fieldErrors).length > 0) { setErrors(fieldErrors); return }

    setSubmitting(true)
    try {
      await agentJoin({ ...form, name: form.team + ' Team' })
      navigate('/agent')
    } catch (err) {
      setApiError(err.response?.data?.message || 'Registration failed. Check the company slug and try again.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center px-4">
      {showTeamPicker && (
        <TeamPicker
          selected={form.team}
          onSelect={(t) => setForm((p) => ({ ...p, team: t }))}
          onClose={() => setShowTeamPicker(false)}
        />
      )}

      <div className="absolute top-4 right-4"><DarkModeToggle /></div>

      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">SupportIQ</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">Create a Team Account</p>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-100 dark:border-gray-700 p-8">
          <div className="mb-5 px-4 py-3 bg-blue-50 dark:bg-blue-900/20 border border-blue-100 dark:border-blue-800 rounded-lg">
            <p className="text-xs text-blue-700 dark:text-blue-300 font-medium">👥 Shared Team Account</p>
            <p className="text-xs text-blue-600/80 dark:text-blue-400 mt-0.5">One login for the whole team. All members use the same email and password to access the team dashboard.</p>
          </div>
          {apiError && (
            <div className="mb-5 px-4 py-3 bg-red-50 border border-red-200 text-red-600 rounded-lg text-sm">{apiError}</div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            {/* Email */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1.5">Email</label>
              <input
                name="email" type="email" value={form.email} onChange={handleChange} autoComplete="email"
                className={`w-full px-4 py-2.5 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500 dark:bg-gray-700 dark:text-gray-100 transition ${errors.email ? 'border-red-400' : 'border-gray-300 dark:border-gray-600'}`}
              />
              {errors.email && <p className="mt-1 text-xs text-red-500">{errors.email}</p>}
            </div>

            {/* Password */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1.5">Password</label>
              <div className="relative">
                <input
                  name="password" type={showPassword ? 'text' : 'password'} value={form.password} onChange={handleChange}
                  className={`w-full px-4 py-2.5 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500 dark:bg-gray-700 dark:text-gray-100 transition pr-10 ${errors.password ? 'border-red-400' : 'border-gray-300 dark:border-gray-600'}`}
                />
                <button type="button" tabIndex={-1} onClick={() => setShowPassword((v) => !v)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200">
                  {showPassword
                    ? <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" /></svg>
                    : <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
                  }
                </button>
              </div>
              {errors.password && <p className="mt-1 text-xs text-red-500">{errors.password}</p>}
            </div>

            {/* Company Slug */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1.5">Company Slug</label>
              <input
                name="company_slug" value={form.company_slug} onChange={handleChange}
                placeholder="e.g. acme-corp  (ask your admin)"
                className={`w-full px-4 py-2.5 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500 dark:bg-gray-700 dark:text-gray-100 transition ${errors.company_slug ? 'border-red-400' : 'border-gray-300 dark:border-gray-600'}`}
              />
              {errors.company_slug
                ? <p className="mt-1 text-xs text-red-500">{errors.company_slug}</p>
                : <p className="mt-1 text-xs text-gray-400 dark:text-gray-500">Your admin can find this in Settings.</p>
              }
            </div>

            {/* Team */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1.5">Team</label>
              <button
                type="button"
                onClick={() => setShowTeamPicker(true)}
                className={`w-full text-left px-4 py-2.5 border rounded-lg text-sm transition ${
                  errors.team ? 'border-red-400' : 'border-gray-300 dark:border-gray-600'
                } ${form.team ? 'text-gray-900 dark:text-gray-100' : 'text-gray-400 dark:text-gray-500'} bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 flex items-center justify-between`}
              >
                <span>{form.team || 'Select your team…'}</span>
                <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              </button>
              {errors.team && <p className="mt-1 text-xs text-red-500">{errors.team}</p>}
            </div>

            <button
              type="submit" disabled={submitting}
              className="w-full py-2.5 bg-emerald-600 text-white font-semibold rounded-lg hover:bg-emerald-700 transition disabled:opacity-50 text-sm mt-2"
            >
              {submitting ? 'Creating team account…' : 'Create Team Account'}
            </button>
          </form>

          <p className="mt-6 text-center text-sm text-gray-500 dark:text-gray-400">
            Already have an account?{' '}
            <Link to="/login" className="text-emerald-600 font-medium hover:underline">Sign in</Link>
          </p>
          <p className="mt-2 text-center text-sm text-gray-500 dark:text-gray-400">
            Setting up a new company?{' '}
            <Link to="/register" className="text-blue-600 font-medium hover:underline">Register here</Link>
          </p>
        </div>
      </div>
    </div>
  )
}
