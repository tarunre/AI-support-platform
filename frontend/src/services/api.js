import axios from 'axios'

// The key used to store the JWT access token in localStorage.
export const TOKEN_KEY = 'access_token'

// Single Axios instance — base URL from env, never hardcoded.
const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  timeout: 60000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor — automatically attach the Bearer token on every request.
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem(TOKEN_KEY)
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor — clear token on 401. Navigation is handled by
// ProtectedRoute / AuthContext so we don't force a full-page reload.
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem(TOKEN_KEY)
    }
    return Promise.reject(error)
  }
)

// --- Service modules ---

export const healthService = {
  check: () => apiClient.get('/api/v1/health'),
}

export default apiClient
