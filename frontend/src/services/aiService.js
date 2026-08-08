import apiClient from './api'

export const aiService = {
  getAnalysis: (ticketId) =>
    apiClient.get(`/api/v1/tickets/${ticketId}/ai-analysis`),

  retryAnalysis: (ticketId) =>
    apiClient.post(`/api/v1/tickets/${ticketId}/retry-ai`),
}
