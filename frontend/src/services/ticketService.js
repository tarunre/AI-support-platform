import apiClient from './api'

export const ticketService = {
  createTicket: (data) =>
    apiClient.post('/api/v1/tickets', data),

  getTickets: (params = {}) =>
    apiClient.get('/api/v1/tickets', { params }),

  getTicket: (id) =>
    apiClient.get(`/api/v1/tickets/${id}`),

  updateTicket: (id, data) =>
    apiClient.put(`/api/v1/tickets/${id}`, data),

  updateStatus: (id, status) =>
    apiClient.patch(`/api/v1/tickets/${id}/status`, { status }),

  assignTicket: (id, assignedTo) =>
    apiClient.patch(`/api/v1/tickets/${id}/assign`, { assignedTo }),

  takeOwnership: (id) =>
    apiClient.patch(`/api/v1/tickets/${id}/take-ownership`),

  deleteTicket: (id) =>
    apiClient.delete(`/api/v1/tickets/${id}`),

  getAgents: () =>
    apiClient.get('/api/v1/users/agents'),

  getMyTickets: (params = {}) =>
    apiClient.get('/api/v1/my-tickets', { params }),

  getTeamTickets: (params = {}) =>
    apiClient.get('/api/v1/team-tickets', { params }),

  getUnassigned: (params = {}) =>
    apiClient.get('/api/v1/tickets/unassigned', { params }),
}
