import apiClient from './api'

export const activityService = {
  listByTicket: (ticketId) =>
    apiClient.get(`/api/v1/tickets/${ticketId}/activity`),

  getRecent: () =>
    apiClient.get('/api/v1/activities'),
}
