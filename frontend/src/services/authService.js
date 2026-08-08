import apiClient from './api'

export const authService = {
  register:   (data) => apiClient.post('/api/v1/auth/register', data),
  agentJoin:  (data) => apiClient.post('/api/v1/auth/agent-join', data),
  login:      (data) => apiClient.post('/api/v1/auth/login', data),
  logout:     ()     => apiClient.post('/api/v1/auth/logout'),
  getMe:      ()     => apiClient.get('/api/v1/auth/me'),
  getTeams:   ()     => apiClient.get('/api/v1/teams'),
}
