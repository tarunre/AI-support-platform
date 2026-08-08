import api from './api'

const portalService = {
  getConversation(token) {
    return api.get(`/api/v1/portal/conversation?token=${encodeURIComponent(token)}`)
  },
  sendReply(token, message) {
    return api.post('/api/v1/portal/reply', { token, message })
  },
}

export default portalService
