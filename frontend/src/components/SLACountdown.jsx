import { useState, useEffect, useRef } from 'react'
import SLABadge from './SLABadge'

/**
 * SLACountdown — live countdown timer that updates every second.
 * Props:
 *   resolutionDueAt  (string | null) — ISO timestamp of the deadline
 *   firstResponseDueAt (string | null)
 *   slaStatus        (string) — current SLA status
 *   onStatusChange   (fn, optional) — called when local countdown hits 0
 */
function formatDuration(totalSeconds) {
  const abs = Math.abs(totalSeconds)
  const h = Math.floor(abs / 3600)
  const m = Math.floor((abs % 3600) / 60)
  const s = abs % 60
  const overdue = totalSeconds < 0 ? '-' : ''
  if (h > 0) return `${overdue}${h}h ${m}m`
  if (m > 0) return `${overdue}${m}m ${s}s`
  return `${overdue}${s}s`
}

export default function SLACountdown({ resolutionDueAt, firstResponseDueAt, slaStatus, compact = false }) {
  const [secondsLeft, setSecondsLeft] = useState(null)
  const [frSecondsLeft, setFrSecondsLeft] = useState(null)
  const timerRef = useRef(null)

  useEffect(() => {
    if (!resolutionDueAt) return

    const update = () => {
      const due = new Date(resolutionDueAt).getTime()
      const now = Date.now()
      setSecondsLeft(Math.round((due - now) / 1000))

      if (firstResponseDueAt) {
        const frDue = new Date(firstResponseDueAt).getTime()
        setFrSecondsLeft(Math.round((frDue - now) / 1000))
      }
    }

    update()
    timerRef.current = setInterval(update, 1000)
    return () => clearInterval(timerRef.current)
  }, [resolutionDueAt, firstResponseDueAt])

  if (!resolutionDueAt || secondsLeft === null) return null

  const isBreached = secondsLeft < 0
  const isAtRisk = secondsLeft >= 0 && secondsLeft < 7200 // < 2 hours

  const resColor = isBreached
    ? 'text-red-600'
    : isAtRisk
    ? 'text-yellow-600'
    : 'text-green-600'

  if (compact) {
    return (
      <span className={`text-xs font-mono font-semibold ${resColor}`}>
        {isBreached ? '⚠ ' : ''}{formatDuration(secondsLeft)}
      </span>
    )
  }

  return (
    <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 p-4 shadow-sm space-y-3">
      <div className="flex items-center justify-between">
        <h3 className="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wide">SLA Status</h3>
        <SLABadge status={slaStatus} />
      </div>

      <div className="space-y-2 text-sm">
        <div className="flex justify-between items-center">
          <span className="text-gray-500 dark:text-gray-400 dark:text-gray-500">Resolution due</span>
          <span className={`font-mono font-semibold tabular-nums ${resColor}`}>
            {isBreached ? '⚠ OVERDUE ' : ''}{formatDuration(secondsLeft)}
          </span>
        </div>

        {frSecondsLeft !== null && (
          <div className="flex justify-between items-center">
            <span className="text-gray-500 dark:text-gray-400 dark:text-gray-500">First response</span>
            <span className={`font-mono text-xs tabular-nums ${frSecondsLeft < 0 ? 'text-red-500' : 'text-gray-600 dark:text-gray-300 dark:text-gray-600'}`}>
              {frSecondsLeft < 0 ? '⚠ ' : ''}{formatDuration(frSecondsLeft)}
            </span>
          </div>
        )}
      </div>

      {/* Progress bar */}
      {(() => {
        const due = new Date(resolutionDueAt).getTime()
        const now = Date.now()
        const totalMs = due - (due - Math.abs(secondsLeft) * 1000 - (secondsLeft < 0 ? 0 : 0))
        // Approximate: show pct based on status
        const pct = slaStatus === 'BREACHED' ? 100
          : slaStatus === 'COMPLETED' ? 100
          : slaStatus === 'AT_RISK' ? 85
          : 50
        const barColor = isBreached ? 'bg-red-500' : isAtRisk ? 'bg-yellow-400' : 'bg-green-500'
        return (
          <div className="w-full bg-gray-100 dark:bg-gray-800 rounded-full h-1.5">
            <div className={`h-1.5 rounded-full transition-all ${barColor}`} style={{ width: `${Math.min(pct, 100)}%` }} />
          </div>
        )
      })()}

      <p className="text-xs text-gray-400 dark:text-gray-500">
        Due: {new Date(resolutionDueAt).toLocaleString()}
      </p>
    </div>
  )
}
