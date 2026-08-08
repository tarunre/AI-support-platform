import api from './api'

const BASE = '/api/v1/analytics'

export const analyticsService = {
  // Overview
  getOverview: () => api.get(`${BASE}/overview`),

  // Tickets
  getTicketStats: (params = {}) => api.get(`${BASE}/tickets`, { params }),

  // Agents
  getAgentStats: (params = {}) => api.get(`${BASE}/agents`, { params }),

  // AI
  getAIStats: (params = {}) => api.get(`${BASE}/ai`, { params }),

  // Queue
  getQueueStats: () => api.get(`${BASE}/queues`),

  // Email
  getEmailStats: (params = {}) => api.get(`${BASE}/email`, { params }),

  // Trends
  getTrends: (params = {}) => api.get(`${BASE}/trends`, { params }),

  // Manual aggregation trigger (admin only)
  triggerAggregation: () => api.post(`${BASE}/aggregate`),

  // Reports
  generateReport: (data) => api.post(`${BASE}/reports`, data),
  listReports: () => api.get(`${BASE}/reports`),
  getReport: (id) => api.get(`${BASE}/reports/${id}`),
  downloadReport: (id) => api.get(`${BASE}/reports/${id}/download`, { responseType: 'blob' }),
}

export default analyticsService
