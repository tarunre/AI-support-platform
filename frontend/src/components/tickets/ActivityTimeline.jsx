import { useState, useEffect } from 'react'
import { activityService } from '../../services/activityService'
import { formatDate } from '../../utils/format'

const ICON_MAP = {
  CREATE_TICKET: '🎫',
  UPDATE_TICKET: '✏️',
  ASSIGN_TICKET: '👤',
  TAKE_OWNERSHIP: '✋',
  STATUS_CHANGED: '🔄',
  PRIORITY_CHANGED: '⚡',
  CATEGORY_CHANGED: '🏷️',
  COMMENT_ADDED: '💬',
  INTERNAL_NOTE_ADDED: '📝',
  TICKET_CLOSED: '✅',
  TICKET_REOPENED: '🔓',
}

const COLOR_MAP = {
  CREATE_TICKET: 'bg-blue-100 text-blue-700',
  UPDATE_TICKET: 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 dark:text-gray-600',
  ASSIGN_TICKET: 'bg-purple-100 text-purple-700',
  TAKE_OWNERSHIP: 'bg-indigo-100 text-indigo-700',
  STATUS_CHANGED: 'bg-amber-100 text-amber-700',
  PRIORITY_CHANGED: 'bg-orange-100 text-orange-700',
  CATEGORY_CHANGED: 'bg-teal-100 text-teal-700',
  COMMENT_ADDED: 'bg-green-100 text-green-700',
  INTERNAL_NOTE_ADDED: 'bg-yellow-100 text-yellow-700',
  TICKET_CLOSED: 'bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 dark:text-gray-500',
  TICKET_REOPENED: 'bg-blue-100 text-blue-600',
}

export default function ActivityTimeline({ ticketId }) {
  const [activities, setActivities] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    activityService
      .listByTicket(ticketId)
      .then((r) => setActivities(r.data.data || []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [ticketId])

  if (loading) {
    return <p className="p-4 text-sm text-gray-400 dark:text-gray-500 animate-pulse">Loading timeline…</p>
  }

  if (activities.length === 0) {
    return <p className="p-4 text-sm text-gray-400 dark:text-gray-500">No activity recorded yet.</p>
  }

  return (
    <div className="p-4">
      <ol className="relative border-l border-gray-200 dark:border-gray-600 ml-3 space-y-6">
        {activities.map((a) => (
          <li key={a.id} className="ml-6">
            {/* Dot */}
            <span
              className={`absolute -left-3 flex h-6 w-6 items-center justify-center rounded-full ring-4 ring-white text-xs ${
                COLOR_MAP[a.activity_type] || 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 dark:text-gray-600'
              }`}
            >
              {ICON_MAP[a.activity_type] || '●'}
            </span>

            <div className="rounded-lg border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 p-3 shadow-sm">
              <p className="text-sm font-medium text-gray-800 dark:text-gray-100">{a.description}</p>

              {(a.old_value || a.new_value) && (
                <div className="mt-1 flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500">
                  {a.old_value && (
                    <span className="rounded bg-red-50 px-1.5 py-0.5 text-red-600">
                      {a.old_value}
                    </span>
                  )}
                  {a.old_value && a.new_value && <span>→</span>}
                  {a.new_value && (
                    <span className="rounded bg-green-50 px-1.5 py-0.5 text-green-600">
                      {a.new_value}
                    </span>
                  )}
                </div>
              )}

              <p className="mt-1.5 text-xs text-gray-400 dark:text-gray-500">
                {a.user?.name || 'System'} · {formatDate(a.created_at)}
              </p>
            </div>
          </li>
        ))}
      </ol>
    </div>
  )
}
