import { useState, useEffect } from 'react'
import emailService from '../../services/emailService'
import { useToast } from '../Toast'

const DIRECTION_STYLE = {
  INBOUND:  'bg-blue-50 border-blue-100',
  OUTBOUND: 'bg-emerald-50 border-emerald-100',
}

const STATUS_BADGE = {
  RECEIVED:  'bg-blue-100 text-blue-700',
  QUEUED:    'bg-yellow-100 text-yellow-700',
  SENT:      'bg-green-100 text-green-700',
  FAILED:    'bg-red-100 text-red-700',
  DELIVERED: 'bg-emerald-100 text-emerald-700',
  READ:      'bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 dark:text-gray-500',
}

function EmailCard({ msg, onRetry }) {
  const [expanded, setExpanded] = useState(false)
  const isInbound = msg.direction === 'INBOUND'

  return (
    <div className={`rounded-xl border p-4 ${DIRECTION_STYLE[msg.direction] || 'bg-gray-50 dark:bg-gray-900 border-gray-100 dark:border-gray-700'}`}>
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className={`text-xs font-semibold px-2 py-0.5 rounded-full ${isInbound ? 'bg-blue-200 text-blue-800' : 'bg-emerald-200 text-emerald-800'}`}>
              {isInbound ? '← Inbound' : '→ Outbound'}
            </span>
            <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${STATUS_BADGE[msg.status] || 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 dark:text-gray-600'}`}>
              {msg.status}
            </span>
            <span className="text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500">
              {isInbound ? `From: ${msg.sender}` : `To: ${msg.recipient}`}
            </span>
          </div>

          <p className="text-sm font-medium text-gray-900 dark:text-white mt-2 truncate">{msg.subject || '(No subject)'}</p>

          <p className="text-sm text-gray-600 dark:text-gray-300 dark:text-gray-600 mt-1 line-clamp-2">
            {msg.body || '(Empty body)'}
          </p>

          {expanded && (
            <div className="mt-3">
              <p className="text-sm text-gray-700 dark:text-gray-200 whitespace-pre-wrap">{msg.body}</p>
              {msg.attachments_count > 0 && (
                <div className="mt-2 flex flex-wrap gap-2">
                  {(msg.attachments || []).map((att, i) => (
                    <span key={i} className="text-xs bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded px-2 py-1 text-gray-600 dark:text-gray-300 dark:text-gray-600">
                      📎 {att.filename} ({att.content_type})
                    </span>
                  ))}
                </div>
              )}
            </div>
          )}

          {msg.error_message && (
            <p className="text-xs text-red-500 mt-2">Error: {msg.error_message}</p>
          )}
        </div>

        <div className="flex flex-col items-end gap-1 ml-3 shrink-0">
          <span className="text-xs text-gray-400 dark:text-gray-500">
            {msg.received_at
              ? new Date(msg.received_at).toLocaleString()
              : msg.sent_at
              ? new Date(msg.sent_at).toLocaleString()
              : new Date(msg.created_at).toLocaleString()}
          </span>
          {msg.attachments_count > 0 && (
            <span className="text-xs text-gray-400 dark:text-gray-500">📎 {msg.attachments_count}</span>
          )}
        </div>
      </div>

      <div className="flex gap-3 mt-3 pt-3 border-t border-black/5">
        <button onClick={() => setExpanded(e => !e)} className="text-xs text-indigo-600 hover:text-indigo-800">
          {expanded ? 'Collapse' : 'Read full'}
        </button>
        {msg.status === 'FAILED' && onRetry && (
          <button onClick={() => onRetry(msg.id)} className="text-xs text-red-600 hover:text-red-800">
            Retry
          </button>
        )}
      </div>
    </div>
  )
}

export default function EmailConversationPanel({ ticketId, customerEmail }) {
  const [emails, setEmails] = useState([])
  const [loading, setLoading] = useState(true)
  const [showCompose, setShowCompose] = useState(false)
  const [compose, setCompose] = useState({ to: customerEmail || '', subject: '', body: '' })
  const [sending, setSending] = useState(false)
  const { showToast } = useToast()

  const load = async () => {
    try {
      const res = await emailService.getTicketEmails(ticketId)
      setEmails(res.data?.data || [])
    } catch {
      showToast('Failed to load email thread', 'error')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (ticketId) load()
  }, [ticketId])

  const handleSend = async () => {
    if (!compose.to || !compose.subject || !compose.body) {
      showToast('To, Subject and Body are required', 'error')
      return
    }
    setSending(true)
    try {
      await emailService.sendEmail(ticketId, compose)
      showToast('Email sent')
      setShowCompose(false)
      setCompose({ to: customerEmail || '', subject: '', body: '' })
      load()
    } catch (err) {
      showToast(err.response?.data?.message || 'Failed to send email', 'error')
    } finally {
      setSending(false)
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-200">Email Conversation</h3>
        <button
          onClick={() => setShowCompose(s => !s)}
          className="text-sm text-indigo-600 hover:text-indigo-800 font-medium"
        >
          {showCompose ? 'Cancel' : '+ Compose Email'}
        </button>
      </div>

      {showCompose && (
        <div className="mb-4 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-xl p-4 shadow-sm">
          <p className="text-sm font-semibold text-gray-700 dark:text-gray-200 mb-3">New Email</p>
          <div className="space-y-3">
            <div>
              <label className="text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500">To</label>
              <input
                value={compose.to}
                onChange={e => setCompose(c => ({ ...c, to: e.target.value }))}
                className="w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm mt-1"
                placeholder="recipient@example.com"
              />
            </div>
            <div>
              <label className="text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500">Subject</label>
              <input
                value={compose.subject}
                onChange={e => setCompose(c => ({ ...c, subject: e.target.value }))}
                className="w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm mt-1"
              />
            </div>
            <div>
              <label className="text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500">Message</label>
              <textarea
                value={compose.body}
                onChange={e => setCompose(c => ({ ...c, body: e.target.value }))}
                rows={5}
                className="w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm mt-1 resize-none"
              />
            </div>
            <div className="flex justify-end">
              <button
                onClick={handleSend}
                disabled={sending}
                className="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700 disabled:opacity-50"
              >
                {sending ? 'Sending…' : 'Send Email'}
              </button>
            </div>
          </div>
        </div>
      )}

      {loading ? (
        <p className="text-sm text-gray-400 dark:text-gray-500 py-8 text-center">Loading email thread…</p>
      ) : emails.length === 0 ? (
        <p className="text-sm text-gray-400 dark:text-gray-500 py-8 text-center">No emails in this conversation yet.</p>
      ) : (
        <div className="space-y-3">
          {emails.map(msg => (
            <EmailCard key={msg.id} msg={msg} />
          ))}
        </div>
      )}
    </div>
  )
}
