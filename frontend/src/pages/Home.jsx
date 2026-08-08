import { useState, useEffect } from 'react'
import { Link, Navigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { healthService } from '../services/api'
import DarkModeToggle from '../components/DarkModeToggle'

// Possible values: 'loading' | 'online' | 'offline'
const STATUS = {
  LOADING: 'loading',
  ONLINE: 'online',
  OFFLINE: 'offline',
}

function Home() {
  const { user, loading: authLoading } = useAuth()
  const [backendStatus, setBackendStatus] = useState(STATUS.LOADING)

  useEffect(() => {
    const checkBackend = async () => {
      try {
        await healthService.check()
        setBackendStatus(STATUS.ONLINE)
      } catch {
        setBackendStatus(STATUS.OFFLINE)
      }
    }

    checkBackend()
  }, [])

  // Redirect already-authenticated users straight to the dashboard
  if (!authLoading && user) return <Navigate to="/dashboard" replace />

  return (
    <div className="flex flex-col items-center justify-center min-h-screen gap-8 bg-gray-50 dark:bg-gray-900">
      <div className="absolute top-4 right-4"><DarkModeToggle /></div>
      <h1 className="text-4xl font-bold text-gray-800 dark:text-gray-100 tracking-tight">
        SupportIQ
      </h1>

      <div className="bg-white dark:bg-gray-800 rounded-xl shadow-md px-8 py-5 flex items-center gap-3">
        <span className="text-base font-medium text-gray-500 dark:text-gray-400 dark:text-gray-500">Backend Status</span>
        <span className="text-gray-300 dark:text-gray-600">|</span>

        {backendStatus === STATUS.LOADING && (
          <span className="text-gray-400 dark:text-gray-500 text-sm animate-pulse">Checking...</span>
        )}

        {backendStatus === STATUS.ONLINE && (
          <span className="text-green-600 font-semibold">🟢 Backend Connected</span>
        )}

        {backendStatus === STATUS.OFFLINE && (
          <span className="text-red-500 font-semibold">🔴 Backend Offline</span>
        )}
      </div>

      <div className="flex gap-4">
        <Link
          to="/login"
          className="px-6 py-3 bg-blue-600 text-white font-semibold rounded-lg shadow hover:bg-blue-700 transition"
        >
          Login
        </Link>
        <Link
          to="/register"
          className="px-6 py-3 bg-white dark:bg-gray-800 text-blue-600 font-semibold rounded-lg shadow border border-blue-600 hover:bg-blue-50 transition"
        >
          Register
        </Link>
      </div>
    </div>
  )
}

export default Home
