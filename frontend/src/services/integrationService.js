import api from './api';

const BASE = '/api/v1/integrations';

export const integrationService = {
  list: () => api.get(BASE),

  create: (data) => api.post(BASE, data),

  update: (id, data) => api.put(`${BASE}/${id}`, data),

  delete: (id) => api.delete(`${BASE}/${id}`),

  test: (id) => api.post(`${BASE}/${id}/test`),

  listEvents: (id) => api.get(`${BASE}/${id}/events`),

  getTicketIntegrations: (ticketId) => api.get(`/api/v1/tickets/${ticketId}/integrations`),

  createJira: (ticketId) => api.post(`/api/v1/tickets/${ticketId}/create-jira`),

  createLinear: (ticketId) => api.post(`/api/v1/tickets/${ticketId}/create-linear`),

  createGitHub: (ticketId) => api.post(`/api/v1/tickets/${ticketId}/create-github-issue`),
};

export default integrationService;
