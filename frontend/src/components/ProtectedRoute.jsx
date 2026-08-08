import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'

// Renders child routes only when the user is authenticated.
// Shows a loading screen while the session is being restored on first render.
function ProtectedRoute() {
  const { user, loading } = useAuth()

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
        <p className="text-sm text-gray-400 dark:text-gray-500 animate-pulse">Loading...</p>
      </div>
    )
  }
  return user ? <Outlet /> : <Navigate to="/login" replace />
}

export default ProtectedRoute
