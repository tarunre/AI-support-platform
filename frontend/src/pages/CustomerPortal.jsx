import { useState, useEffect, useRef } from 'react'
import { useSearchParams } from 'react-router-dom'
import portalService from '../services/portalService'

/* ─── Helpers ────────────────────────────────────────────────────────────── */

function formatTime(iso) {
  const d = new Date(iso)
  return d.toLocaleString(undefined, {
    month: 'short', day: 'numeric',
    hour: '2-digit', minute: '2-digit',
  })
}

function StatusPill({ status }) {
  const map = {
    OPEN:        'bg-emerald-100 text-emerald-700',
    IN_PROGRESS: 'bg-blue-100 text-blue-700',
    RESOLVED:    'bg-gray-100 text-gray-600',
    CLOSED:      'bg-red-100 text-red-600',
  }
  return (
    <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold uppercase tracking-wide ${map[status] || 'bg-gray-100 text-gray-500'}`}>
      {status?.replace('_', ' ')}
    </span>
  )
}

/* ─── Message bubble ─────────────────────────────────────────────────────── */

function Bubble({ msg, customerName }) {
  const isCustomer = msg.direction === 'INBOUND'
  return (
    <div className={`flex gap-3 ${isCustomer ? 'flex-row-reverse' : 'flex-row'} animate-fade-up`}>
      {/* Avatar */}
      <div className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-bold text-white shadow-sm ${
        isCustomer ? 'bg-blue-500' : 'bg-indigo-600'
      }`}>
        {isCustomer
          ? (customerName?.charAt(0) || 'C').toUpperCase()
          : 'S'}
      </div>

      {/* Bubble */}
      <div className={`max-w-[75%] space-y-1 ${isCustomer ? 'items-end' : 'items-start'} flex flex-col`}>
        <div className={`rounded-2xl px-4 py-3 text-sm leading-relaxed shadow-sm ${
          isCustomer
            ? 'rounded-tr-sm bg-blue-600 text-white'
            : 'rounded-tl-sm bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-100 border border-gray-100 dark:border-gray-700'
        }`}>
          <p className="whitespace-pre-wrap break-words">{msg.body}</p>
        </div>
        <p className={`text-[10px] text-gray-400 px-1 ${isCustomer ? 'text-right' : 'text-left'}`}>
          {msg.sender} · {formatTime(msg.created_at)}
        </p>
      </div>
    </div>
  )
}

/* ─── Empty state ────────────────────────────────────────────────────────── */

function EmptyConversation() {
  return (
    <div className="flex flex-col items-center justify-center gap-3 py-16 text-center">
      <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-indigo-50 text-2xl">💬</div>
      <p className="text-sm font-medium text-gray-600 dark:text-gray-300">No messages yet</p>
      <p className="text-xs text-gray-400">Your conversation with our support team will appear here.</p>
    </div>
  )
}

/* ─── Loading skeleton ───────────────────────────────────────────────────── */

function Skeleton() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-slate-50 to-blue-50 dark:from-gray-950 dark:to-gray-900">
      <div className="flex flex-col items-center gap-4">
        <div className="h-10 w-10 animate-spin rounded-full border-4 border-blue-200 border-t-blue-600" />
        <p className="text-sm text-gray-500 dark:text-gray-400 animate-pulse">Loading your conversation…</p>
      </div>
    </div>
  )
}

/* ─── Error state ────────────────────────────────────────────────────────── */

function ErrorView({ message }) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-slate-50 to-red-50 dark:from-gray-950 dark:to-gray-900 px-4">
      <div className="max-w-sm w-full rounded-2xl border border-red-100 dark:border-red-900 bg-white dark:bg-gray-900 p-8 text-center shadow-xl">
        <div className="mb-4 text-4xl">🔒</div>
        <h2 className="text-lg font-bold text-gray-800 dark:text-white mb-2">Unable to load</h2>
        <p className="text-sm text-gray-500 dark:text-gray-400 mb-6">{message}</p>
        <p className="text-xs text-gray-400">If you believe this is a mistake, please contact support directly.</p>
      </div>
    </div>
  )
}

/* ─── Main portal page ───────────────────────────────────────────────────── */

export default function CustomerPortal() {
  const [params] = useSearchParams()
  const token = params.get('token') || ''

  const [ticket, setTicket]     = useState(null)
  const [messages, setMessages] = useState([])
  const [loading, setLoading]   = useState(true)
  const [error, setError]       = useState('')
  const [text, setText]         = useState('')
  const [sending, setSending]   = useState(false)
  const [aiTyping, setAiTyping] = useState(false)
  const [sendError, setSendError] = useState('')

  const bottomRef  = useRef(null)
  const pollRef    = useRef(null)
  const textareaRef = useRef(null)

  /* Load / poll conversation */
  const loadConversation = async () => {
    if (!token) return
    try {
      const res = await portalService.getConversation(token)
      const data = res.data?.data
      setTicket(data.ticket)
      const newMsgs = data.messages || []
      setMessages(prev => {
        // If we got a new OUTBOUND message, stop typing indicator
        if (newMsgs.length > prev.length) {
          const hasNewOutbound = newMsgs.slice(prev.length).some(m => m.direction === 'OUTBOUND')
          if (hasNewOutbound) setAiTyping(false)
        }
        return newMsgs
      })
    } catch (err) {
      if (loading) {
        setError(err.response?.data?.message || 'This link has expired or is invalid.')
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (!token) { setError('No access token found in the URL.'); setLoading(false); return }
    loadConversation()
    pollRef.current = setInterval(loadConversation, 5000)
    return () => clearInterval(pollRef.current)
  }, [token])

  /* Scroll to bottom on new messages */
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const handleSend = async () => {
    const msg = text.trim()
    if (!msg || sending) return
    setSending(true)
    setSendError('')
    try {
      const res = await portalService.sendReply(token, msg)
      setMessages(prev => [...prev, res.data.data])
      setText('')
      textareaRef.current?.focus()
      // Show typing indicator while AI generates a reply (~15s)
      setAiTyping(true)
      setTimeout(() => setAiTyping(false), 18000)
    } catch (err) {
      setSendError(err.response?.data?.message || 'Failed to send. Please try again.')
    } finally {
      setSending(false)
    }
  }

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  /* ── Render states ── */
  if (!token || (loading && !ticket)) return <Skeleton />
  if (error) return <ErrorView message={error} />

  return (
    <div className="flex min-h-screen flex-col bg-gradient-to-br from-slate-50 via-white to-blue-50 dark:from-gray-950 dark:via-gray-900 dark:to-indigo-950">

      {/* ── Header ── */}
      <header className="sticky top-0 z-10 border-b border-gray-100 dark:border-gray-800 bg-white/80 dark:bg-gray-900/80 backdrop-blur-md px-4 py-4 shadow-sm">
        <div className="mx-auto max-w-2xl">
          {/* Brand */}
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-indigo-600 text-white text-xs font-bold">S</div>
              <span className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-widest">SupportIQ</span>
            </div>
            {ticket && <StatusPill status={ticket.status} />}
          </div>

          {/* Ticket info */}
          {ticket && (
            <div>
              <h1 className="text-base font-bold text-gray-900 dark:text-white leading-snug">{ticket.subject}</h1>
              <p className="text-xs text-gray-400 dark:text-gray-500 mt-0.5">
                <span className="font-mono text-indigo-500">{ticket.ticket_number}</span>
                {' · '}Hi, {ticket.customer_name}
              </p>
            </div>
          )}
        </div>
      </header>

      {/* ── Messages ── */}
      <main className="flex-1 overflow-y-auto px-4 py-6">
        <div className="mx-auto max-w-2xl space-y-5">
          {messages.length === 0
            ? <EmptyConversation />
            : messages.map((m) => (
                <Bubble
                  key={m.id}
                  msg={m}
                  customerName={ticket?.customer_name}
                />
              ))
          }

          {/* AI typing indicator */}
          {aiTyping && (
            <div className="flex gap-3 flex-row animate-fade-up">
              <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-indigo-600 text-xs font-bold text-white shadow-sm">S</div>
              <div className="flex flex-col items-start space-y-1">
                <div className="rounded-2xl rounded-tl-sm bg-white dark:bg-gray-800 border border-gray-100 dark:border-gray-700 px-4 py-3 shadow-sm">
                  <div className="flex items-center gap-1.5">
                    <span className="text-xs text-gray-400 dark:text-gray-500 mr-1">Support Team is typing</span>
                    {[0, 0.2, 0.4].map((d) => (
                      <span
                        key={d}
                        className="inline-block h-1.5 w-1.5 rounded-full bg-indigo-400 animate-bounce"
                        style={{ animationDelay: `${d}s`, animationDuration: '1s' }}
                      />
                    ))}
                  </div>
                </div>
              </div>
            </div>
          )}

          <div ref={bottomRef} />
        </div>
      </main>

      {/* ── Input ── */}
      <div className="sticky bottom-0 border-t border-gray-100 dark:border-gray-800 bg-white/90 dark:bg-gray-900/90 backdrop-blur-md px-4 py-4 shadow-lg">
        <div className="mx-auto max-w-2xl">
          {sendError && (
            <p className="mb-2 text-xs text-red-500 font-medium">{sendError}</p>
          )}
          <div className="flex items-end gap-3">
            <textarea
              ref={textareaRef}
              value={text}
              onChange={(e) => setText(e.target.value)}
              onKeyDown={handleKeyDown}
              rows={1}
              placeholder="Write a reply… (Enter to send, Shift+Enter for new line)"
              disabled={sending}
              className="flex-1 resize-none rounded-2xl border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 px-4 py-3 text-sm text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 transition-all max-h-36 overflow-y-auto"
              style={{ fieldSizing: 'content' }}
            />
            <button
              onClick={handleSend}
              disabled={sending || !text.trim()}
              className="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-blue-600 text-white shadow-md hover:bg-blue-700 disabled:opacity-40 disabled:cursor-not-allowed transition-all active:scale-95"
              aria-label="Send message"
            >
              {sending ? (
                <span className="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent" />
              ) : (
                <svg viewBox="0 0 20 20" fill="currentColor" className="h-5 w-5">
                  <path d="M3.105 2.288a.75.75 0 00-.826.95l1.081 3.243a.75.75 0 00.652.515l9.444 1.214-9.444 1.214a.75.75 0 00-.652.515l-1.081 3.243a.75.75 0 00.826.95l16.5-6.75a.75.75 0 000-1.4l-16.5-6.75z" />
                </svg>
              )}
            </button>
          </div>
          <p className="mt-2 text-center text-[10px] text-gray-400">
            Powered by <span className="font-semibold text-indigo-500">SupportIQ</span> · Your messages are end-to-end encrypted
          </p>
        </div>
      </div>

    </div>
  )
}
