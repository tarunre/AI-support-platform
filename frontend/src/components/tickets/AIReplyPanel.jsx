import { useState, useEffect, useRef } from 'react'
import { replyService } from '../../services/replyService'

const STATUS_COLORS = {
  GENERATED:   'bg-blue-100 text-blue-700',
  APPROVED:    'bg-green-100 text-green-700',
  REJECTED:    'bg-red-100 text-red-700',
  REGENERATED: 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 dark:text-gray-600',
  SENT:        'bg-purple-100 text-purple-700',
}

function ConfidenceBar({ value }) {
  const color = value >= 85 ? 'bg-green-500' : value >= 60 ? 'bg-amber-400' : 'bg-red-400'
  return (
    <div className="space-y-1">
      <div className="flex justify-between text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500">
        <span>Confidence</span>
        <span className="font-semibold">{value}%</span>
      </div>
      <div className="h-2 w-full rounded-full bg-gray-100 dark:bg-gray-800">
        <div className={`h-2 rounded-full transition-all ${color}`} style={{ width: `${value}%` }} />
      </div>
    </div>
  )
}

export default function AIReplyPanel({ ticketId }) {
  const [reply, setReply] = useState(null)
  const [loading, setLoading] = useState(true)
  const [actionLoading, setActionLoading] = useState(false)
  const [error, setError] = useState(null)
  const [editMode, setEditMode] = useState(false)
  const [editText, setEditText] = useState('')
  const [copied, setCopied] = useState(false)
  const textareaRef = useRef(null)

  const load = async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await replyService.getReply(ticketId)
      setReply(res.data.data)
    } catch (err) {
      if (err.response?.status !== 404) {
        setError('Failed to load reply.')
      }
      setReply(null)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [ticketId])

  const handleGenerate = async () => {
    setActionLoading(true)
    setError(null)
    try {
      const res = await replyService.generateReply(ticketId)
      setReply(res.data.data)
    } catch (err) {
      setError(err.response?.data?.message || 'Reply generation failed. Make sure knowledge base articles exist.')
    } finally {
      setActionLoading(false)
    }
  }

  const handleRegenerate = async () => {
    setActionLoading(true)
    setError(null)
    setEditMode(false)
    try {
      const res = await replyService.regenerateReply(ticketId)
      setReply(res.data.data)
    } catch (err) {
      setError(err.response?.data?.message || 'Regeneration failed.')
    } finally {
      setActionLoading(false)
    }
  }

  const handleApprove = async () => {
    setActionLoading(true)
    try {
      const res = await replyService.approveReply(ticketId)
      setReply(res.data.data)
    } catch (err) {
      setError(err.response?.data?.message || 'Approval failed.')
    } finally {
      setActionLoading(false)
    }
  }

  const handleReject = async () => {
    setActionLoading(true)
    try {
      const res = await replyService.rejectReply(ticketId)
      setReply(res.data.data)
    } catch (err) {
      setError(err.response?.data?.message || 'Rejection failed.')
    } finally {
      setActionLoading(false)
    }
  }

  const handleEditSave = async () => {
    if (!editText.trim()) return
    setActionLoading(true)
    try {
      const res = await replyService.editReply(ticketId, editText)
      setReply(res.data.data)
      setEditMode(false)
    } catch (err) {
      setError(err.response?.data?.message || 'Edit failed.')
    } finally {
      setActionLoading(false)
    }
  }

  const handleCopy = () => {
    const text = reply?.edited_reply || reply?.generated_reply || ''
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }

  const startEdit = () => {
    setEditText(reply?.edited_reply || reply?.generated_reply || '')
    setEditMode(true)
    setTimeout(() => textareaRef.current?.focus(), 50)
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-16">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        <span className="ml-3 text-sm text-gray-400 dark:text-gray-500">Loading reply…</span>
      </div>
    )
  }

  // No reply yet
  if (!reply) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-6">
        {/* Icon */}
        <div className="w-20 h-20 rounded-3xl bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-blue-900/30 dark:to-indigo-900/30 flex items-center justify-center text-4xl shadow-sm">
          🤖
        </div>

        {/* Text */}
        <div className="text-center">
          <p className="text-base font-semibold text-gray-800 dark:text-gray-100">No AI Reply Generated</p>
          <p className="text-sm text-gray-400 dark:text-gray-500 mt-1 max-w-xs">
            Generate a reply grounded in your knowledge base. The agent must approve before it becomes official.
          </p>
        </div>

        {/* Error */}
        {error && (
          <div className="w-full max-w-sm rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-100 dark:border-red-800 px-4 py-3 text-sm text-red-600 dark:text-red-400 text-center">
            {error}
          </div>
        )}

        {/* Button */}
        <button
          onClick={handleGenerate}
          disabled={actionLoading}
          className="inline-flex items-center gap-2 rounded-xl bg-emerald-600 px-6 py-3 text-sm font-semibold text-white hover:bg-emerald-700 disabled:opacity-50 transition shadow-sm"
        >
          {actionLoading ? (
            <>
              <div className="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent" />
              Generating…
            </>
          ) : (
            <>
              <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path strokeLinecap="round" strokeLinejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09z" />
              </svg>
              Generate AI Reply
            </>
          )}
        </button>

        {/* Info hint */}
        <p className="text-xs text-gray-400 dark:text-gray-500">
          Grounded in your knowledge base · Requires agent approval
        </p>
      </div>
    )
  }

  const displayReply = reply.edited_reply || reply.generated_reply
  const isGenerated = reply.status === 'GENERATED'
  const isApproved = reply.status === 'APPROVED'
  const isRejected = reply.status === 'REJECTED'

  return (
    <div className="space-y-4">
      {/* Header card */}
      <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 p-5 shadow-sm">
        <div className="flex items-start justify-between gap-4 mb-4">
          <div>
            <h3 className="text-sm font-semibold text-gray-800 dark:text-gray-100">AI Suggested Reply</h3>
            <p className="text-xs text-gray-400 dark:text-gray-500 mt-0.5">
              Generated {new Date(reply.created_at).toLocaleString()} · Model: {reply.model || 'AI'} · Prompt {reply.prompt_version}
            </p>
          </div>
          <span className={`shrink-0 rounded-full px-3 py-1 text-xs font-semibold ${STATUS_COLORS[reply.status] || 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 dark:text-gray-600'}`}>
            {reply.status}
          </span>
        </div>

        <ConfidenceBar value={reply.confidence} />

        {isApproved && reply.approver && (
          <p className="mt-3 text-xs text-green-600">
            ✓ Approved by <strong>{reply.approver.name}</strong> on {new Date(reply.approved_at).toLocaleString()}
          </p>
        )}
      </div>

      {/* Reply content */}
      <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 p-5 shadow-sm">
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wide">
            {reply.edited_reply ? 'Edited Reply' : 'Generated Reply'}
          </h3>
          <button
            onClick={handleCopy}
            className="text-xs text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:text-gray-300 dark:text-gray-600 transition"
          >
            {copied ? '✓ Copied' : '📋 Copy'}
          </button>
        </div>

        {editMode ? (
          <div className="space-y-3">
            <textarea
              ref={textareaRef}
              value={editText}
              onChange={(e) => setEditText(e.target.value)}
              rows={8}
              className="w-full rounded-lg border border-blue-300 p-3 text-sm text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-400 resize-y"
            />
            <div className="flex gap-2">
              <button
                onClick={handleEditSave}
                disabled={actionLoading || !editText.trim()}
                className="rounded-lg bg-blue-600 px-4 py-1.5 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50 transition"
              >
                {actionLoading ? 'Saving…' : 'Save Changes'}
              </button>
              <button
                onClick={() => { setEditMode(false); setError(null) }}
                className="rounded-lg border border-gray-200 dark:border-gray-600 px-4 py-1.5 text-sm font-medium text-gray-600 dark:text-gray-300 dark:text-gray-600 hover:bg-gray-50 dark:bg-gray-900 transition"
              >
                Cancel
              </button>
            </div>
          </div>
        ) : (
          <p className="text-sm text-gray-800 dark:text-gray-100 whitespace-pre-wrap leading-relaxed">{displayReply}</p>
        )}

        {reply.edited_reply && !editMode && (
          <div className="mt-4 border-t border-gray-50 dark:border-gray-700 pt-4">
            <p className="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-2">Original Generated Reply</p>
            <p className="text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500 whitespace-pre-wrap leading-relaxed">{reply.generated_reply}</p>
          </div>
        )}
      </div>

      {/* Error banner */}
      {error && (
        <div className="rounded-lg bg-red-50 border border-red-100 px-4 py-3 text-sm text-red-600">
          {error}
        </div>
      )}

      {/* Action buttons */}
      {!editMode && (
        <div className="flex flex-wrap gap-2">
          {isGenerated && (
            <>
              <button
                onClick={handleApprove}
                disabled={actionLoading}
                className="rounded-lg bg-green-600 px-4 py-2 text-sm font-medium text-white hover:bg-green-700 disabled:opacity-50 transition"
              >
                ✓ Approve
              </button>
              <button
                onClick={startEdit}
                disabled={actionLoading}
                className="rounded-lg border border-blue-200 px-4 py-2 text-sm font-medium text-blue-600 hover:bg-blue-50 disabled:opacity-50 transition"
              >
                ✏️ Edit
              </button>
              <button
                onClick={handleReject}
                disabled={actionLoading}
                className="rounded-lg border border-red-200 px-4 py-2 text-sm font-medium text-red-500 hover:bg-red-50 disabled:opacity-50 transition"
              >
                ✗ Reject
              </button>
            </>
          )}
          <button
            onClick={handleRegenerate}
            disabled={actionLoading}
            className="rounded-lg border border-gray-200 dark:border-gray-600 px-4 py-2 text-sm font-medium text-gray-600 dark:text-gray-300 dark:text-gray-600 hover:bg-gray-50 dark:bg-gray-900 disabled:opacity-50 transition"
          >
            {actionLoading ? (
              <span className="flex items-center gap-2">
                <div className="h-3 w-3 animate-spin rounded-full border-2 border-gray-400 border-t-transparent" />
                Working…
              </span>
            ) : '🔄 Regenerate'}
          </button>
        </div>
      )}

      {isApproved && (
        <div className="rounded-xl bg-green-50 border border-green-100 px-5 py-4 text-sm text-green-700">
          <strong>Reply Approved.</strong> This reply has been approved and is ready to send to the customer.
        </div>
      )}

      {isRejected && (
        <div className="rounded-xl bg-red-50 border border-red-100 px-5 py-4 text-sm text-red-600">
          <strong>Reply Rejected.</strong> Use Regenerate to create a new reply draft.
        </div>
      )}
    </div>
  )
}
