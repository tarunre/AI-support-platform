import apiClient from './api'

export const commentService = {
  create: (ticketId, data) =>
    apiClient.post(`/api/v1/tickets/${ticketId}/comments`, data),

  list: (ticketId) =>
    apiClient.get(`/api/v1/tickets/${ticketId}/comments`),
}
