import { BrowserRouter } from 'react-router-dom'
import { AuthProvider } from './contexts/AuthContext'
import { WebSocketProvider } from './contexts/WebSocketContext'
import { ThemeProvider } from './contexts/ThemeContext'
import RealtimeToast from './components/RealtimeToast'
import AppRoutes from './routes'

function App() {
  return (
    <BrowserRouter>
      <ThemeProvider>
        <AuthProvider>
          <WebSocketProvider>
            <AppRoutes />
            <RealtimeToast />
          </WebSocketProvider>
        </AuthProvider>
      </ThemeProvider>
    </BrowserRouter>
  )
}

export default App
