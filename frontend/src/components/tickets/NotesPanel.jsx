import { useState, useEffect, useCallback } from 'react'
import { noteService } from '../../services/noteService'
import { formatDate } from '../../utils/format'

export default function NotesPanel({ ticketId }) {
  const [notes, setNotes] = useState([])
  const [loading, setLoading] = useState(true)
  const [text, setText] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  const load = useCallback(() => {
    noteService
      .list(ticketId)
      .then((r) => setNotes(r.data.data || []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [ticketId])

  useEffect(() => { load() }, [load])

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (!text.trim()) return
    setSubmitting(true)
    setError('')
    try {
      await noteService.create(ticketId, { note: text.trim() })
      setText('')
      load()
    } catch (err) {
      setError(err.response?.data?.message || 'Failed to add note')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="p-4 space-y-6">
      {/* Notes list */}
      <div className="space-y-3">
        {loading ? (
          <p className="text-sm text-gray-400 dark:text-gray-500 animate-pulse">Loading notes…</p>
        ) : notes.length === 0 ? (
          <p className="text-sm text-gray-400 dark:text-gray-500">No internal notes yet.</p>
        ) : (
          notes.map((n) => (
            <div
              key={n.id}
              className="rounded-lg border border-yellow-200 bg-yellow-50 p-4"
            >
              <p className="text-sm text-gray-800 dark:text-gray-100 whitespace-pre-wrap">{n.note}</p>
              <p className="mt-2 text-xs text-gray-400 dark:text-gray-500">
                {n.user?.name || 'Unknown'} · {formatDate(n.created_at)}
              </p>
            </div>
          ))
        )}
      </div>

      {/* Add note form */}
      <form onSubmit={handleSubmit} className="space-y-2">
        <label className="block text-xs font-semibold text-gray-500 dark:text-gray-400 dark:text-gray-500 uppercase tracking-wide">
          Add Internal Note
        </label>
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="Write an internal note visible only to staff…"
          rows={4}
          className="w-full rounded-lg border border-gray-200 dark:border-gray-600 px-3 py-2 text-sm focus:border-yellow-400 focus:outline-none focus:ring-1 focus:ring-yellow-400"
        />
        {error && <p className="text-xs text-red-500">{error}</p>}
        <button
          type="submit"
          disabled={submitting || !text.trim()}
          className="rounded-lg bg-yellow-500 px-4 py-2 text-sm font-medium text-white hover:bg-yellow-600 disabled:opacity-50 transition"
        >
          {submitting ? 'Adding…' : 'Add Note'}
        </button>
      </form>
    </div>
  )
}
