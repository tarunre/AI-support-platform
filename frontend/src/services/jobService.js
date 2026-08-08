import api from './api';

const jobService = {
  list(params = {}) {
    return api.get('/api/v1/jobs', { params });
  },

  getById(id) {
    return api.get(`/api/v1/jobs/${id}`);
  },

  retry(id) {
    return api.post(`/api/v1/jobs/${id}/retry`);
  },

  statistics() {
    return api.get('/api/v1/jobs/statistics');
  },
};

export default jobService;
