import { Navigate, Outlet, NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'

function NavItem({ to, icon, label, end }) {
  return (
    <NavLink
      to={to}
      end={end}
      className={({ isActive }) =>
        `flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
          isActive
            ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
            : 'text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
        }`
      }
    >
      <span className="text-base">{icon}</span>
      {label}
    </NavLink>
  )
}

export default function AgentLayout() {
  const { user, loading, logout } = useAuth()
  const navigate = useNavigate()

  if (loading) return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
      <p className="text-sm text-gray-400 animate-pulse">Loading…</p>
    </div>
  )

  if (!user) return <Navigate to="/login" replace />
  if (user.role !== 'SupportAgent') return <Navigate to="/dashboard" replace />

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  return (
    <div className="flex h-screen overflow-hidden bg-gray-50 dark:bg-gray-900">
      {/* Sidebar */}
      <aside className="w-56 shrink-0 flex flex-col bg-white dark:bg-gray-800 border-r border-gray-100 dark:border-gray-700">
        {/* Logo */}
        <div className="px-5 py-5 border-b border-gray-100 dark:border-gray-700">
          <span className="text-lg font-bold text-gray-900 dark:text-white">SupportIQ</span>
          <div className="flex items-center gap-1.5 mt-1">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
            <span className="text-xs text-emerald-600 dark:text-emerald-400 font-medium">Agent Portal</span>
          </div>
        </div>

        {/* Nav */}
        <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
          <NavItem to="/agent"           end   icon="🏠" label="Dashboard" />
          <NavItem to="/agent/queue"           icon="📋" label="Team Queue" />
          <NavItem to="/agent/available"       icon="📥" label="Pick Up" />
        </nav>

        {/* Agent info footer */}
        <div className="px-4 py-4 border-t border-gray-100 dark:border-gray-700">
          <div className="mb-3">
            <p className="text-xs font-semibold text-gray-900 dark:text-white truncate">{user.name}</p>
            <p className="text-xs text-gray-400 dark:text-gray-500 truncate">{user.team || 'Support'} Team</p>
          </div>
          <button
            onClick={handleLogout}
            className="flex items-center gap-2 text-xs text-red-500 hover:text-red-600 transition"
          >
            <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
            Sign out
          </button>
        </div>
      </aside>

      {/* Content */}
      <div className="flex-1 min-w-0 overflow-y-auto">
        <Outlet />
      </div>
    </div>
  )
}
