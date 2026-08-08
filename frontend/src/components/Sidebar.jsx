import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import DarkModeToggle from './DarkModeToggle'

const ICONS = {
  home:    <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor"><path d="M10.707 2.293a1 1 0 00-1.414 0l-7 7a1 1 0 001.414 1.414L4 10.414V17a1 1 0 001 1h2a1 1 0 001-1v-2a1 1 0 011-1h2a1 1 0 011 1v2a1 1 0 001 1h2a1 1 0 001-1v-6.586l.293.293a1 1 0 001.414-1.414l-7-7z" /></svg>,
  inbox:   <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor"><path d="M8.707 7.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l2-2a1 1 0 00-1.414-1.414L11 7.586V3a1 1 0 10-2 0v4.586l-.293-.293z"/><path d="M3 5a2 2 0 012-2h1a1 1 0 010 2H5v7h2l1 2h4l1-2h2V5h-1a1 1 0 110-2h1a2 2 0 012 2v10a2 2 0 01-2 2H5a2 2 0 01-2-2V5z"/></svg>,
  user:    <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clipRule="evenodd" /></svg>,
  tickets: <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor"><path d="M2 6a2 2 0 012-2h12a2 2 0 012 2v2a1 1 0 01-1 1 1 1 0 100 2 1 1 0 011 1v2a2 2 0 01-2 2H4a2 2 0 01-2-2v-2a1 1 0 011-1 1 1 0 100-2 1 1 0 01-1-1V6z" /></svg>,
  book:    <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor"><path d="M9 4.804A7.968 7.968 0 005.5 4c-1.255 0-2.443.29-3.5.804v10A7.969 7.969 0 015.5 14c1.669 0 3.218.51 4.5 1.385A7.962 7.962 0 0114.5 14c1.255 0 2.443.29 3.5.804v-10A7.968 7.968 0 0014.5 4c-1.255 0-2.443.29-3.5.804V12a1 1 0 11-2 0V4.804z" /></svg>,
  cpu:     <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor"><path d="M13 7H7v6h6V7z"/><path fillRule="evenodd" d="M7 2a1 1 0 012 0v1h2V2a1 1 0 112 0v1h2a2 2 0 012 2v2h1a1 1 0 110 2h-1v2h1a1 1 0 110 2h-1v2a2 2 0 01-2 2h-2v1a1 1 0 11-2 0v-1H9v1a1 1 0 11-2 0v-1H5a2 2 0 01-2-2v-2H2a1 1 0 110-2h1V9H2a1 1 0 010-2h1V5a2 2 0 012-2h2V2zM5 5h10v10H5V5z" clipRule="evenodd"/></svg>,
  email:   <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor"><path d="M2.003 5.884L10 9.882l7.997-3.998A2 2 0 0016 4H4a2 2 0 00-1.997 1.884z"/><path d="M18 8.118l-8 4-8-4V14a2 2 0 002 2h12a2 2 0 002-2V8.118z"/></svg>,
  chart:   <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor"><path d="M2 11a1 1 0 011-1h2a1 1 0 011 1v5a1 1 0 01-1 1H3a1 1 0 01-1-1v-5zM8 7a1 1 0 011-1h2a1 1 0 011 1v9a1 1 0 01-1 1H9a1 1 0 01-1-1V7zM14 4a1 1 0 011-1h2a1 1 0 011 1v12a1 1 0 01-1 1h-2a1 1 0 01-1-1V4z" /></svg>,
  ai:      <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor"><path d="M13 6a3 3 0 11-6 0 3 3 0 016 0zM18 8a2 2 0 11-4 0 2 2 0 014 0zM14 15a4 4 0 00-8 0v3h8v-3zM6 8a2 2 0 11-4 0 2 2 0 014 0zM16 18v-3a5.972 5.972 0 00-.75-2.906A3.005 3.005 0 0119 15v3h-3zM4.75 12.094A5.973 5.973 0 004 15v3H1v-3a3 3 0 013.75-2.906z" /></svg>,
  queue:   <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor"><path d="M7 3a1 1 0 000 2h6a1 1 0 000-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zM2 11a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z" /></svg>,
  puzzle:  <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor"><path d="M10 3.5a1.5 1.5 0 013 0V4a1 1 0 001 1h3a1 1 0 011 1v3a1 1 0 01-1 1h-.5a1.5 1.5 0 000 3h.5a1 1 0 011 1v3a1 1 0 01-1 1h-3a1 1 0 01-1-1v-.5a1.5 1.5 0 00-3 0v.5a1 1 0 01-1 1H6a1 1 0 01-1-1v-3a1 1 0 00-1-1h-.5a1.5 1.5 0 010-3H4a1 1 0 001-1V6a1 1 0 011-1h3a1 1 0 001-1v-.5z" /></svg>,
  gear:    <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z" clipRule="evenodd" /></svg>,
  shield:  <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M2.166 4.999A11.954 11.954 0 0010 1.944 11.954 11.954 0 0017.834 5c.11.65.166 1.32.166 2.001 0 5.225-3.34 9.67-8 11.317C5.34 16.67 2 12.225 2 7c0-.682.057-1.35.166-2.001zm11.541 3.708a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd"/></svg>,
  clock:   <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clipRule="evenodd"/></svg>,
  logout:  <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M3 3a1 1 0 00-1 1v12a1 1 0 102 0V4a1 1 0 00-1-1zm10.293 9.293a1 1 0 001.414 1.414l3-3a1 1 0 000-1.414l-3-3a1 1 0 10-1.414 1.414L14.586 9H7a1 1 0 100 2h7.586l-1.293 1.293z" clipRule="evenodd" /></svg>,
  reports: <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z" clipRule="evenodd" /></svg>,
}

function NavItem({ to, label, icon, exact = false }) {
  const { pathname } = useLocation()
  const active = exact
    ? pathname === to
    : pathname === to || pathname.startsWith(to + '/')

  return (
    <Link
      to={to}
      className={`flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
        active
          ? 'bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400'
          : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-white'
      }`}
    >
      <span className="shrink-0 opacity-80">{icon}</span>
      {label}
    </Link>
  )
}

function SectionLabel({ children }) {
  return (
    <p className="px-3 pt-4 pb-1 text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest select-none">
      {children}
    </p>
  )
}

export default function Sidebar() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const isAdmin = user?.role === 'Admin' || user?.role === 'SuperAdmin'
  const isSuperAdmin = user?.role === 'SuperAdmin'

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  return (
    <aside className="w-60 shrink-0 flex flex-col bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700">
      {/* Logo */}
      <div className="px-5 py-4 border-b border-gray-100 dark:border-gray-700">
        <Link to="/dashboard">
          <span className="font-bold text-gray-900 dark:text-white text-sm">
            SupportIQ
          </span>
        </Link>
      </div>

      {/* Nav */}
      <nav className="flex-1 overflow-y-auto px-3 py-3 space-y-0.5">
        <NavItem to="/dashboard" exact label="Dashboard" icon={ICONS.home} />

        <SectionLabel>Tickets</SectionLabel>
        <NavItem to="/tickets/unassigned" exact label="Unassigned" icon={ICONS.inbox} />
        <NavItem to="/my-tickets" exact label="My Tickets" icon={ICONS.user} />
        <NavItem to="/tickets" exact label="All Tickets" icon={ICONS.tickets} />

        {isAdmin && (
          <>
            <SectionLabel>Content</SectionLabel>
            <NavItem to="/knowledge-base" label="Knowledge Base" icon={ICONS.book} />

            <SectionLabel>Monitoring</SectionLabel>
            <NavItem to="/jobs" label="Job Monitor" icon={ICONS.cpu} />
            <NavItem to="/email/accounts" label="Email Accounts" icon={ICONS.email} />
            <NavItem to="/email/monitor" label="Email Monitor" icon={ICONS.email} />

            <SectionLabel>Analytics</SectionLabel>
            <NavItem to="/analytics" exact label="Overview" icon={ICONS.chart} />
            <NavItem to="/analytics/agents" label="Agent Performance" icon={ICONS.user} />
            <NavItem to="/analytics/reports" label="Reports" icon={ICONS.reports} />

            <SectionLabel>System</SectionLabel>
            <NavItem to="/integrations" label="Integrations" icon={ICONS.puzzle} />
            <NavItem to="/settings" label="Settings" icon={ICONS.gear} />
            <NavItem to="/sla-management" label="SLA Management" icon={ICONS.shield} />
            <NavItem to="/sla" label="SLA Dashboard" icon={ICONS.clock} />
          </>
        )}

        {isSuperAdmin && (
          <>
            <SectionLabel>Super Admin</SectionLabel>
            <NavItem to="/admin" label="Admin Panel" icon={ICONS.shield} />
          </>
        )}

        {user?.role === 'SupportAgent' && (
          <>
            <SectionLabel>Analytics</SectionLabel>
            <NavItem to="/analytics/agents" label="My Performance" icon={ICONS.chart} />
          </>
        )}
      </nav>

      {/* Footer */}
      <div className="border-t border-gray-100 dark:border-gray-700 p-3 space-y-1">
        <div className="flex items-center justify-between px-3 py-2">
          <div className="min-w-0 flex-1">
            <p className="text-xs font-semibold text-gray-800 dark:text-gray-200 truncate">
              {user?.name}
            </p>
            <p className="text-[11px] text-gray-400 dark:text-gray-500 truncate">
              {user?.role}
            </p>
          </div>
          <DarkModeToggle />
        </div>
        <button
          onClick={handleLogout}
          className="w-full flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 hover:text-red-600 transition-colors"
        >
          {ICONS.logout}
          Sign out
        </button>
      </div>
    </aside>
  )
}
