import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'
import { ticketService } from '../../services/ticketService'
import { useToast } from '../../components/Toast'
import { useWebSocket } from '../../contexts/WebSocketContext'
import StatusBadge from '../../components/tickets/StatusBadge'
import PriorityBadge from '../../components/tickets/PriorityBadge'
import ActivityTimeline from '../../components/tickets/ActivityTimeline'
import NotesPanel from '../../components/tickets/NotesPanel'
import ConversationPanel from '../../components/tickets/ConversationPanel'
import AIAnalysisPanel from '../../components/tickets/AIAnalysisPanel'
import AIReplyPanel from '../../components/tickets/AIReplyPanel'
import EmailConversationPanel from '../../components/tickets/EmailConversationPanel'
import SLACountdown from '../../components/SLACountdown'
import { formatDate } from '../../utils/format'

const TABS = ['Overview', 'Conversation', 'Notes', 'Activity', 'AI Analysis', 'AI Reply', 'Email']

const AI_PROCESSING = ['PENDING', 'PROCESSING']

const NEXT_STATUS = {
  OPEN: 'IN_PROGRESS',
  IN_PROGRESS: 'RESOLVED',
  RESOLVED: 'CLOSED',
}

const NEXT_STATUS_LABEL = {
  OPEN: 'Start Progress',
  IN_PROGRESS: 'Mark Resolved',
  RESOLVED: 'Close Ticket',
}

export default function TicketDetails() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { user } = useAuth()
  const { toast, showToast } = useToast()
  const { wsService } = useWebSocket()

  const [ticket, setTicket] = useState(null)
  const [agents, setAgents] = useState([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState('Overview')
  const [statusLoading, setStatusLoading] = useState(false)
  const [assignLoading, setAssignLoading] = useState(false)

  const isAdmin = user?.role === 'Admin'

  const load = () => {
    const calls = [ticketService.getTicket(id)]
    if (isAdmin) calls.push(ticketService.getAgents())

    Promise.all(calls)
      .then(([tRes, aRes]) => {
        setTicket(tRes.data.data)
        if (aRes) setAgents(aRes.data.data || [])
      })
      .catch(() => showToast('Failed to load ticket', 'error'))
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [id])

  // Auto-refresh when the worker completes AI analysis or reply generation for this ticket
  useEffect(() => {
    const REFRESH_EVENTS = [
      'ticket.ai.completed',
      'ticket.reply.generated',
      'ticket.updated',      'sla.updated',    ]
    const unsubs = REFRESH_EVENTS.map((eventType) =>
      wsService.on(eventType, (event) => {
        if (event.ticket_id === id) load()
      })
    )
    return () => unsubs.forEach((unsub) => unsub())
  }, [id, wsService])

  const handleStatusUpdate = async () => {
    const next = NEXT_STATUS[ticket.status]
    if (!next) return
    setStatusLoading(true)
    try {
      const r = await ticketService.updateStatus(id, next)
      setTicket(r.data.data)
      showToast(`Status updated to ${next}`)
    } catch (err) {
      showToast(err.response?.data?.message || 'Failed to update status', 'error')
    } finally {
      setStatusLoading(false)
    }
  }

  const handleAssign = async (e) => {
    const agentId = Number(e.target.value)
    if (!agentId) return
    setAssignLoading(true)
    try {
      const r = await ticketService.assignTicket(id, agentId)
      setTicket(r.data.data)
      showToast('Ticket assigned')
    } catch (err) {
      showToast(err.response?.data?.message || 'Failed to assign', 'error')
    } finally {
      setAssignLoading(false)
    }
  }

  const handleDelete = async () => {
    if (!window.confirm('Delete this ticket? This cannot be undone.')) return
    try {
      await ticketService.deleteTicket(id)
      showToast('Ticket deleted')
      setTimeout(() => navigate('/tickets'), 800)
    } catch (err) {
      showToast(err.response?.data?.message || 'Failed to delete', 'error')
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center p-12">
        <p className="text-gray-400 dark:text-gray-500 animate-pulse">Loading ticket…</p>
      </div>
    )
  }

  if (!ticket) {
    return (
      <div className="flex items-center justify-center p-12">
        <p className="text-gray-500 dark:text-gray-400">Ticket not found.</p>
      </div>
    )
  }

  return (
    <div>
      {/* Title bar */}
      <div className="bg-white dark:bg-gray-800 px-6 pt-4 pb-0">
        <p className="text-xs text-gray-400 dark:text-gray-500 mb-2">
          <Link to="/tickets" className="hover:text-gray-600 dark:hover:text-gray-300">← Tickets</Link>
          <span className="mx-1">/</span>
          <span className="font-mono">{ticket.ticket_number}</span>
        </p>
        <div className="flex items-center justify-between gap-4">
          <div>
            <h1 className="text-lg font-bold text-gray-900 dark:text-white">{ticket.subject}</h1>
            <div className="mt-1 flex items-center gap-2">
              <StatusBadge status={ticket.status} />
              <PriorityBadge priority={ticket.priority} />
            </div>
          </div>
          <div className="flex items-center gap-2 shrink-0">
            <Link
              to={`/tickets/${id}/edit`}
              className="rounded-lg border border-gray-200 dark:border-gray-600 px-3 py-1.5 text-sm font-medium text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition"
            >
              Edit
            </Link>
            {NEXT_STATUS[ticket.status] && (
              <button
                onClick={handleStatusUpdate}
                disabled={statusLoading}
                className="rounded-lg bg-blue-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50 transition"
              >
                {statusLoading ? 'Updating…' : NEXT_STATUS_LABEL[ticket.status]}
              </button>
            )}
            {isAdmin && (
              <button
                onClick={handleDelete}
                className="rounded-lg border border-red-200 px-3 py-1.5 text-sm font-medium text-red-500 hover:bg-red-50 transition"
              >
                Delete
              </button>
            )}
          </div>
        </div>

        {/* Tabs */}
        <div className="mt-4 flex gap-1">
          {TABS.map((tab) => {
            const isAITab = tab === 'AI Analysis'
            const aiProcessing = isAITab && AI_PROCESSING.includes(ticket.ai_processing_status)
            return (
              <button
                key={tab}
                onClick={() => setActiveTab(tab)}
                className={`relative px-4 py-2 text-sm font-medium border-b-2 transition ${
                  activeTab === tab
                    ? 'border-blue-600 text-blue-600'
                    : 'border-transparent text-gray-500 dark:text-gray-400 dark:text-gray-500 hover:text-gray-700 dark:text-gray-200'
                }`}
              >
                <span className="flex items-center gap-1.5">
                  {tab}
                  {aiProcessing && (
                    <span className="relative flex h-2 w-2">
                      <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75" />
                      <span className="relative inline-flex rounded-full h-2 w-2 bg-blue-500" />
                    </span>
                  )}
                </span>
              </button>
            )
          })}
        </div>
      </div>

      {/* Tab content */}
      <div className="max-w-5xl mx-auto px-6 py-6">
        {activeTab === 'Overview' && (
          <div className="grid grid-cols-3 gap-6">
            {/* Main content */}
            <div className="col-span-2 space-y-4">
              <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 p-5 shadow-sm">
                <h3 className="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-3">Description</h3>
                <p className="text-sm text-gray-800 dark:text-gray-100 whitespace-pre-wrap break-words">{ticket.description}</p>
              </div>

              <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 p-5 shadow-sm">
                <h3 className="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-3">Customer</h3>
                <dl className="space-y-2">
                  <div className="flex gap-3">
                    <dt className="w-16 text-xs text-gray-400 dark:text-gray-500 shrink-0">Name</dt>
                    <dd className="text-sm text-gray-800 dark:text-gray-100 font-medium">{ticket.customer_name}</dd>
                  </div>
                  <div className="flex gap-3">
                    <dt className="w-16 text-xs text-gray-400 dark:text-gray-500 shrink-0">Email</dt>
                    <dd className="text-sm text-gray-800 dark:text-gray-100">{ticket.customer_email}</dd>
                  </div>
                </dl>
              </div>
            </div>

            {/* Sidebar */}
            <div className="space-y-4">
              <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 p-5 shadow-sm space-y-3">
                <h3 className="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wide">Details</h3>
                <dl className="space-y-2.5 text-sm">
                  <div className="flex justify-between">
                    <dt className="text-gray-400 dark:text-gray-500">Priority</dt>
                    <dd><PriorityBadge priority={ticket.priority} /></dd>
                  </div>
                  <div className="flex justify-between">
                    <dt className="text-gray-400 dark:text-gray-500">Category</dt>
                    <dd className="text-gray-700 dark:text-gray-200 font-medium">{ticket.category}</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt className="text-gray-400 dark:text-gray-500">Source</dt>
                    <dd className="text-gray-700 dark:text-gray-200">{ticket.source}</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt className="text-gray-400 dark:text-gray-500">Created by</dt>
                    <dd className="text-gray-700 dark:text-gray-200">{ticket.creator?.name || '—'}</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt className="text-gray-400 dark:text-gray-500">Assigned to</dt>
                    <dd className="text-gray-700 dark:text-gray-200">{ticket.assignee?.name || 'Unassigned'}</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt className="text-gray-400 dark:text-gray-500">Created</dt>
                    <dd className="text-gray-700 dark:text-gray-200">{formatDate(ticket.created_at)}</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt className="text-gray-400 dark:text-gray-500">Updated</dt>
                    <dd className="text-gray-700 dark:text-gray-200">{formatDate(ticket.updated_at)}</dd>
                  </div>
                </dl>
              </div>

              {/* Admin assignment */}
              {isAdmin && (
                <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 p-5 shadow-sm space-y-3">
                  <h3 className="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wide">Assign</h3>

                  {/* Current assignee chip */}
                  {ticket.assignee ? (
                    <div className="flex items-center gap-2 rounded-lg bg-blue-50 dark:bg-blue-900/20 border border-blue-100 dark:border-blue-800 px-3 py-2">
                      <span className="flex h-6 w-6 items-center justify-center rounded-full bg-blue-500 text-[10px] font-bold text-white shrink-0">
                        {ticket.assignee.name?.charAt(0).toUpperCase()}
                      </span>
                      <div className="min-w-0">
                        <p className="text-xs font-semibold text-blue-700 dark:text-blue-300 truncate">{ticket.assignee.name}</p>
                        <p className="text-[10px] text-blue-500 dark:text-blue-400">{ticket.assignee.role}</p>
                      </div>
                    </div>
                  ) : (
                    <p className="text-xs text-gray-400 dark:text-gray-500 italic">Not yet assigned</p>
                  )}

                  {/* AI team suggestion */}
                  {ticket.ai_team && (
                    <p className="text-[10px] text-indigo-500 dark:text-indigo-400 font-medium">
                      🤖 AI suggests: <span className="font-semibold">{ticket.ai_team}</span> team
                    </p>
                  )}

                  {/* Dropdown — merges agents + current assignee so it never shows blank */}
                  <select
                    onChange={handleAssign}
                    disabled={assignLoading}
                    value={ticket.assigned_to || ''}
                    className="select-field w-full"
                  >
                    <option value="">Unassigned</option>
                    {[
                      // Always include current assignee even if they're an Admin (not in agents list)
                      ...(ticket.assignee && !agents.find(a => a.id === ticket.assignee.id)
                        ? [ticket.assignee]
                        : []),
                      ...agents,
                    ].map((a) => (
                      <option key={a.id} value={a.id}>{a.name}</option>
                    ))}
                  </select>
                </div>
              )}
              {/* SLA countdown widget */}
              {ticket.resolution_due_at && (
                <SLACountdown
                  resolutionDueAt={ticket.resolution_due_at}
                  firstResponseDueAt={ticket.first_response_due_at}
                  slaStatus={ticket.sla_status}
                />
              )}            </div>
          </div>
        )}

        {activeTab === 'Conversation' && (
          <div className="max-w-2xl">
            <ConversationPanel ticketId={id} customerName={ticket?.customer_name} />
          </div>
        )}

        {activeTab === 'Notes' && (
          <div className="max-w-2xl">
            <NotesPanel ticketId={id} />
          </div>
        )}

        {activeTab === 'Activity' && (
          <div className="max-w-2xl">
            <ActivityTimeline ticketId={id} />
          </div>
        )}

        {activeTab === 'AI Analysis' && (
          <div className="max-w-2xl">
            <AIAnalysisPanel ticketId={id} ticketCreatedAt={ticket?.created_at} />
          </div>
        )}

        {activeTab === 'AI Reply' && (
          <div className="w-full max-w-2xl mx-auto">
            <AIReplyPanel ticketId={id} />
          </div>
        )}

        {activeTab === 'Email' && (
          <div className="max-w-3xl">
            <EmailConversationPanel
              ticketId={id}
              customerEmail={ticket?.customer_email}
            />
          </div>
        )}
      </div>

      {toast && (
        <div className={`fixed top-4 right-4 z-50 rounded-lg px-4 py-3 text-sm font-medium shadow-lg ${
          toast.type === 'error' ? 'bg-red-500 text-white' : 'bg-gray-900 text-white'
        }`}>
          {toast.message}
        </div>
      )}
    </div>
  )
}
