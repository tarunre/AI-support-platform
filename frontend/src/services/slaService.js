import apiClient from './api'

export const slaService = {
  // SLA Policies
  listPolicies: () => apiClient.get('/api/v1/sla-policies'),
  getPolicy: (id) => apiClient.get(`/api/v1/sla-policies/${id}`),
  createPolicy: (data) => apiClient.post('/api/v1/sla-policies', data),
  updatePolicy: (id, data) => apiClient.put(`/api/v1/sla-policies/${id}`, data),
  deletePolicy: (id) => apiClient.delete(`/api/v1/sla-policies/${id}`),

  // SLA Dashboard
  getDashboard: () => apiClient.get('/api/v1/tickets/sla'),
}
