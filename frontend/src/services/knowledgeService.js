import apiClient from './api'

export const knowledgeService = {
  list: (params = {}) =>
    apiClient.get('/api/v1/knowledge-base', { params }),

  create: (data) =>
    apiClient.post('/api/v1/knowledge-base', data),

  update: (id, data) =>
    apiClient.put(`/api/v1/knowledge-base/${id}`, data),

  delete: (id) =>
    apiClient.delete(`/api/v1/knowledge-base/${id}`),
}
