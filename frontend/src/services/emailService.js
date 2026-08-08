import api from './api';

const emailService = {
  // Account management (admin)
  listAccounts() {
    return api.get('/api/v1/email/accounts');
  },
  createAccount(data) {
    return api.post('/api/v1/email/accounts', data);
  },
  updateAccount(id, data) {
    return api.put(`/api/v1/email/accounts/${id}`, data);
  },
  deleteAccount(id) {
    return api.delete(`/api/v1/email/accounts/${id}`);
  },
  testConnection(id, protocol = 'smtp') {
    return api.post(`/api/v1/email/accounts/${id}/test?protocol=${protocol}`);
  },

  // Monitor
  getMonitor() {
    return api.get('/api/v1/email/monitor');
  },

  // Sync
  triggerSync() {
    return api.post('/api/v1/email/sync');
  },

  // Ticket emails
  getTicketEmails(ticketId) {
    return api.get(`/api/v1/tickets/${ticketId}/emails`);
  },
  sendEmail(ticketId, data) {
    return api.post(`/api/v1/tickets/${ticketId}/send-email`, data);
  },
};

export default emailService;
