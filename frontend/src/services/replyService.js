import apiClient from './api'

export const replyService = {
  getReply: (ticketId) =>
    apiClient.get(`/api/v1/tickets/${ticketId}/reply`),

  generateReply: (ticketId) =>
    apiClient.post(`/api/v1/tickets/${ticketId}/reply/generate`),

  regenerateReply: (ticketId) =>
    apiClient.post(`/api/v1/tickets/${ticketId}/reply/regenerate`),

  editReply: (ticketId, reply) =>
    apiClient.patch(`/api/v1/tickets/${ticketId}/reply/edit`, { reply }),

  approveReply: (ticketId) =>
    apiClient.post(`/api/v1/tickets/${ticketId}/reply/approve`),

  rejectReply: (ticketId) =>
    apiClient.post(`/api/v1/tickets/${ticketId}/reply/reject`),
}
