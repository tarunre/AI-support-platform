import { useState, useEffect, useRef } from 'react'
import { aiService } from '../../services/aiService'

// ─── Utility badges ──────────────────────────────────────────────────────────

const PRIORITY_COLOR = {
  Urgent: 'bg-red-100 text-red-700 border-red-200',
  High:   'bg-orange-100 text-orange-700 border-orange-200',
  Medium: 'bg-amber-100 text-amber-700 border-amber-200',
  Low:    'bg-green-100 text-green-700 border-green-200',
}

const SENTIMENT_COLOR = {
  Positive:   'bg-green-100 text-green-700',
  Neutral:    'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 dark:text-gray-600',
  Frustrated: 'bg-orange-100 text-orange-700',
  Angry:      'bg-red-100 text-red-600',
  Confused:   'bg-purple-100 text-purple-700',
}

function Badge({ label, colorClass }) {
  return (
    <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium border ${colorClass}`}>
      {label}
    </span>
  )
}

function ConfidenceBar({ value }) {
  const color = value >= 85 ? 'bg-green-500' : value >= 60 ? 'bg-amber-500' : 'bg-red-500'
  return (
    <div className="space-y-1">
      <div className="flex justify-between text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500">
        <span>Confidence</span>
        <span className="font-semibold text-gray-700 dark:text-gray-200">{value}%</span>
      </div>
      <div className="h-2 w-full rounded-full bg-gray-100 dark:bg-gray-800 overflow-hidden">
        <div
          className={`h-full rounded-full transition-all duration-500 ${color}`}
          style={{ width: `${value}%` }}
        />
      </div>
    </div>
  )
}

// ─── Main component ──────────────────────────────────────────────────────────

const POLL_INTERVAL_MS = 3500

export default function AIAnalysisPanel({ ticketId, ticketCreatedAt }) {
  const [analysis, setAnalysis] = useState(null)
  const [loading, setLoading] = useState(true)
  const [retrying, setRetrying] = useState(false)
  const intervalRef = useRef(null)

  const fetch = async () => {
    try {
      const r = await aiService.getAnalysis(ticketId)
      const data = r.data.data
      setAnalysis(data)
      // Stop polling once we have a terminal status
      if (data.processing_status === 'COMPLETED' || data.processing_status === 'FAILED') {
        clearInterval(intervalRef.current)
      }
    } catch {
      // Silently ignore network errors during polling
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetch()
    // Start polling; will clear itself on COMPLETED / FAILED
    intervalRef.current = setInterval(fetch, POLL_INTERVAL_MS)
    return () => clearInterval(intervalRef.current)
  }, [ticketId])

  const handleRetry = async () => {
    setRetrying(true)
    try {
      await aiService.retryAnalysis(ticketId)
      setAnalysis((prev) => ({ ...prev, processing_status: 'PROCESSING' }))
      // Restart polling
      clearInterval(intervalRef.current)
      intervalRef.current = setInterval(fetch, POLL_INTERVAL_MS)
    } catch {
      // ignore
    } finally {
      setRetrying(false)
    }
  }

  if (loading) {
    return <p className="p-4 text-sm text-gray-400 dark:text-gray-500 animate-pulse">Loading AI analysis…</p>
  }

  const status = analysis?.processing_status

  // ── Pending / Processing ──────────────────────────────────────────────────
  if (status === 'PENDING' || status === 'PROCESSING') {
    return (
      <div className="p-8 flex flex-col items-center justify-center gap-6 text-center select-none">
        {/* Animated AI orb */}
        <div className="relative flex h-24 w-24 items-center justify-center">
          {/* Outer glow rings */}
          <div className="absolute inset-0 rounded-full bg-gradient-to-br from-blue-400/20 to-indigo-600/20 animate-ping" style={{ animationDuration: '2s' }} />
          <div className="absolute inset-2 rounded-full bg-gradient-to-br from-blue-400/30 to-indigo-600/30 animate-ping" style={{ animationDuration: '2s', animationDelay: '0.4s' }} />
          {/* Core orb */}
          <div className="relative h-16 w-16 rounded-full bg-gradient-to-br from-blue-500 to-indigo-700 shadow-lg shadow-blue-500/40 flex items-center justify-center">
            {/* Rotating scanner ring */}
            <div className="absolute inset-0 rounded-full border-2 border-transparent border-t-white/60 animate-spin" style={{ animationDuration: '1.2s' }} />
            {/* Spark dots orbiting */}
            <div className="absolute inset-0 rounded-full animate-spin" style={{ animationDuration: '2.5s' }}>
              <div className="absolute top-0 left-1/2 -translate-x-1/2 -translate-y-1 h-2 w-2 rounded-full bg-white/80 shadow-sm" />
            </div>
            <div className="absolute inset-0 rounded-full animate-spin" style={{ animationDuration: '3.5s', animationDirection: 'reverse' }}>
              <div className="absolute bottom-0 left-1/2 -translate-x-1/2 translate-y-1 h-1.5 w-1.5 rounded-full bg-blue-200/90" />
            </div>
            {/* AI icon */}
            <svg className="h-6 w-6 text-white drop-shadow" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.8}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09z" />
            </svg>
          </div>
        </div>

        {/* Text with animated dots */}
        <div className="space-y-2">
          <p className="text-base font-bold text-gray-800 dark:text-white tracking-tight">
            AI is analyzing this ticket
            <span className="inline-flex gap-0.5 ml-1 translate-y-0.5">
              {[0, 0.25, 0.5].map((d) => (
                <span
                  key={d}
                  className="inline-block h-1.5 w-1.5 rounded-full bg-blue-500 animate-bounce"
                  style={{ animationDelay: `${d}s`, animationDuration: '1.2s' }}
                />
              ))}
            </span>
          </p>
          <p className="text-sm text-gray-400 dark:text-gray-500">
            Detecting category · priority · sentiment · recommended team
          </p>
        </div>

        {/* Animated scan bars */}
        <div className="w-full max-w-xs space-y-2">
          {['Category & Tags', 'Priority Assessment', 'Sentiment Analysis', 'Team Routing'].map((label, i) => (
            <div key={label} className="flex items-center gap-3">
              <span className="w-32 shrink-0 text-right text-[10px] text-gray-400 dark:text-gray-500">{label}</span>
              <div className="flex-1 h-1.5 rounded-full bg-gray-100 dark:bg-gray-700 overflow-hidden">
                <div
                  className="h-full rounded-full bg-gradient-to-r from-blue-400 to-indigo-500 animate-pulse"
                  style={{
                    width: `${60 + i * 10}%`,
                    animationDelay: `${i * 0.15}s`,
                    animationDuration: '1.5s',
                  }}
                />
              </div>
            </div>
          ))}
        </div>
      </div>
    )
  }

  // ── Failed ────────────────────────────────────────────────────────────────
  if (status === 'FAILED') {
    // If ticket was created < 5 minutes ago, the worker is likely still retrying
    const isRecent = ticketCreatedAt && (Date.now() - new Date(ticketCreatedAt).getTime()) < 5 * 60 * 1000
    if (isRecent) {
      return (
        <div className="p-6 flex flex-col items-center justify-center gap-3 text-center">
          <div className="relative w-10 h-10">
            <div className="absolute inset-0 rounded-full border-4 border-blue-100" />
            <div className="absolute inset-0 rounded-full border-4 border-blue-500 border-t-transparent animate-spin" />
          </div>
          <p className="text-sm font-semibold text-gray-700 dark:text-gray-200">AI is still working…</p>
          <p className="text-xs text-gray-400 dark:text-gray-500">Retrying analysis, please wait a moment</p>
        </div>
      )
    }
    return (
      <div className="p-6 flex flex-col items-center justify-center gap-4 text-center">
        <span className="text-3xl">⚠️</span>
        <div>
          <p className="text-sm font-semibold text-gray-700 dark:text-gray-200">AI analysis failed</p>
          <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">
            The AI provider was unable to process this ticket.
          </p>
        </div>
        <button
          onClick={handleRetry}
          disabled={retrying}
          className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50 transition"
        >
          {retrying ? 'Queuing retry…' : 'Retry Analysis'}
        </button>
      </div>
    )
  }

  // ── Completed ─────────────────────────────────────────────────────────────
  if (status !== 'COMPLETED') {
    return <p className="p-4 text-sm text-gray-400 dark:text-gray-500">No AI analysis available.</p>
  }

  return (
    <div className="p-5 space-y-6 animate-fade-up">
      {/* Summary */}
      <div className="rounded-xl bg-blue-50 border border-blue-100 p-4 animate-fade-up delay-50">
        <p className="text-xs font-semibold text-blue-500 uppercase tracking-wide mb-1">AI Summary</p>
        <p className="text-sm text-gray-800 dark:text-gray-100 leading-relaxed">{analysis.summary}</p>
      </div>

      {/* Confidence */}
      {analysis.confidence != null && (
        <div className="animate-fade-up delay-100">
          <ConfidenceBar value={analysis.confidence} />
        </div>
      )}

      {/* Classification grid */}
      <div className="grid grid-cols-2 gap-4">
        <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 p-4 shadow-sm animate-fade-up delay-100">
          <p className="text-xs text-gray-400 dark:text-gray-500 mb-2">Detected Category</p>
          <Badge
            label={analysis.category}
            colorClass="bg-indigo-50 text-indigo-700 border-indigo-100"
          />
        </div>

        <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 p-4 shadow-sm animate-fade-up delay-150">
          <p className="text-xs text-gray-400 dark:text-gray-500 mb-2">Suggested Priority</p>
          <Badge
            label={analysis.priority}
            colorClass={PRIORITY_COLOR[analysis.priority] || 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 dark:text-gray-600 border-gray-200 dark:border-gray-600'}
          />
        </div>

        <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 p-4 shadow-sm animate-fade-up delay-200">
          <p className="text-xs text-gray-400 dark:text-gray-500 mb-2">Customer Sentiment</p>
          <Badge
            label={analysis.sentiment}
            colorClass={`${SENTIMENT_COLOR[analysis.sentiment] || 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 dark:text-gray-600'} border border-transparent`}
          />
        </div>

        <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 p-4 shadow-sm animate-fade-up delay-250">
          <p className="text-xs text-gray-400 dark:text-gray-500 mb-2">Recommended Team</p>
          <Badge
            label={analysis.recommended_team}
            colorClass="bg-teal-50 text-teal-700 border-teal-100"
          />
        </div>
      </div>

      {/* Tags */}
      {analysis.tags && analysis.tags.length > 0 && (
        <div>
          <p className="text-xs text-gray-400 dark:text-gray-500 mb-2">Tags</p>
          <div className="flex flex-wrap gap-2">
            {analysis.tags.map((tag) => (
              <span
                key={tag}
                className="rounded-full bg-gray-100 dark:bg-gray-800 px-3 py-1 text-xs font-medium text-gray-600 dark:text-gray-300 dark:text-gray-600"
              >
                #{tag}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Footer */}
      <div className="flex items-center justify-between pt-2 border-t border-gray-100 dark:border-gray-700">
        <p className="text-xs text-gray-400 dark:text-gray-500">
          {analysis.processed_at
            ? `Analyzed at ${new Date(analysis.processed_at).toLocaleString()}`
            : ''}
        </p>
        <button
          onClick={handleRetry}
          disabled={retrying}
          className="text-xs text-gray-400 dark:text-gray-500 hover:text-blue-600 underline disabled:opacity-50 transition"
        >
          {retrying ? 'Queuing…' : 'Re-analyze'}
        </button>
      </div>
    </div>
  )
}
