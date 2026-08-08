import apiClient from './api'

export const tenantService = {
  // SuperAdmin
  list: () => apiClient.get('/api/v1/admin/tenants'),
  getById: (id) => apiClient.get(`/api/v1/admin/tenants/${id}`),
  create: (data) => apiClient.post('/api/v1/admin/tenants', data),
  update: (id, data) => apiClient.put(`/api/v1/admin/tenants/${id}`, data),
  remove: (id) => apiClient.delete(`/api/v1/admin/tenants/${id}`),
  getOverview: () => apiClient.get('/api/v1/admin/overview'),

  // Tenant settings (current tenant admin)
  getSettings: () => apiClient.get('/api/v1/settings'),
  updateSettings: (data) => apiClient.put('/api/v1/settings', data),
}
