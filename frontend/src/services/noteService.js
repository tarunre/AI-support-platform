import apiClient from './api'

export const noteService = {
  create: (ticketId, data) =>
    apiClient.post(`/api/v1/tickets/${ticketId}/notes`, data),

  list: (ticketId) =>
    apiClient.get(`/api/v1/tickets/${ticketId}/notes`),
}
