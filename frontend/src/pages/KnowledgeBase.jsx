import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { knowledgeService } from '../services/knowledgeService'
import { useNavigate } from 'react-router-dom'
import { formatDate } from '../utils/format'
import { KB_CATEGORIES } from '../utils/categories'

const CATEGORIES = KB_CATEGORIES

const EMPTY_FORM = { title: '', category: 'FAQ', content: '', is_active: true }

function Modal({ title, onClose, children }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div className="w-full max-w-2xl rounded-2xl bg-white dark:bg-gray-800 shadow-2xl overflow-hidden">
        <div className="flex items-center justify-between border-b border-gray-100 dark:border-gray-700 px-6 py-4">
          <h2 className="text-sm font-bold text-gray-800 dark:text-gray-100">{title}</h2>
          <button onClick={onClose} className="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:text-gray-300 dark:text-gray-600 text-lg leading-none">×</button>
        </div>
        <div className="px-6 py-5">{children}</div>
      </div>
    </div>
  )
}

function DocForm({ initial, onSubmit, onCancel, saving }) {
  const [form, setForm] = useState(initial || EMPTY_FORM)

  const set = (k, v) => setForm((f) => ({ ...f, [k]: v }))

  const handleSubmit = (e) => {
    e.preventDefault()
    onSubmit(form)
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label className="block text-xs font-medium text-gray-600 dark:text-gray-300 dark:text-gray-600 mb-1">Title *</label>
        <input
          value={form.title}
          onChange={(e) => set('title', e.target.value)}
          required
          minLength={3}
          maxLength={255}
          className="w-full rounded-lg border border-gray-200 dark:border-gray-600 px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-blue-400"
          placeholder="e.g. Standard Refund Policy"
        />
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-600 dark:text-gray-300 dark:text-gray-600 mb-1">Category *</label>
        <select
          value={form.category}
          onChange={(e) => set('category', e.target.value)}
          className="select-field w-full"
        >
          {CATEGORIES.map((c) => (
            <option key={c} value={c}>{c}</option>
          ))}
        </select>
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-600 dark:text-gray-300 dark:text-gray-600 mb-1">Content *</label>
        <textarea
          value={form.content}
          onChange={(e) => set('content', e.target.value)}
          required
          minLength={10}
          rows={8}
          className="w-full rounded-lg border border-gray-200 dark:border-gray-600 px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-blue-400 resize-y"
          placeholder="Write the full document content here. This will be used by the AI to answer customer queries."
        />
      </div>

      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          id="is_active"
          checked={form.is_active}
          onChange={(e) => set('is_active', e.target.checked)}
          className="rounded"
        />
        <label htmlFor="is_active" className="text-sm text-gray-600 dark:text-gray-300 dark:text-gray-600">Active (used by AI for reply generation)</label>
      </div>

      <div className="flex gap-2 pt-2">
        <button
          type="submit"
          disabled={saving}
          className="rounded-lg bg-blue-600 px-5 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50 transition"
        >
          {saving ? 'Saving…' : 'Save Document'}
        </button>
        <button
          type="button"
          onClick={onCancel}
          className="rounded-lg border border-gray-200 dark:border-gray-600 px-5 py-2 text-sm font-medium text-gray-600 dark:text-gray-300 dark:text-gray-600 hover:bg-gray-50 dark:bg-gray-900 transition"
        >
          Cancel
        </button>
      </div>
    </form>
  )
}

export default function KnowledgeBase() {
  const { user } = useAuth()
  const navigate = useNavigate()

  const [docs, setDocs] = useState([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [categoryFilter, setCategoryFilter] = useState('')
  const [showModal, setShowModal] = useState(false)
  const [editDoc, setEditDoc] = useState(null)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState(null)
  const [toast, setToast] = useState(null)

  const showToast = (msg, type = 'success') => {
    setToast({ msg, type })
    setTimeout(() => setToast(null), 3000)
  }

  const load = async () => {
    setLoading(true)
    try {
      const res = await knowledgeService.list({ page, limit: 20, search, category: categoryFilter })
      const data = res.data.data
      setDocs(data.items || [])
      setTotal(data.total_count || 0)
    } catch {
      setError('Failed to load knowledge base.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [page, search, categoryFilter])

  const handleCreate = async (form) => {
    setSaving(true)
    try {
      await knowledgeService.create(form)
      setShowModal(false)
      showToast('Document created')
      load()
    } catch (err) {
      showToast(err.response?.data?.message || 'Create failed', 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleUpdate = async (form) => {
    setSaving(true)
    try {
      await knowledgeService.update(editDoc.id, form)
      setEditDoc(null)
      showToast('Document updated')
      load()
    } catch (err) {
      showToast(err.response?.data?.message || 'Update failed', 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id) => {
    if (!window.confirm('Delete this document? This cannot be undone.')) return
    try {
      await knowledgeService.delete(id)
      showToast('Document deleted')
      load()
    } catch {
      showToast('Delete failed', 'error')
    }
  }

  const totalPages = Math.ceil(total / 20)

  return (
    <div className="max-w-6xl mx-auto px-6 py-6 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="font-bold text-gray-800 dark:text-gray-100 text-xl">Knowledge Base</h1>
        <button
          onClick={() => setShowModal(true)}
          className="rounded-lg bg-emerald-600 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-700 transition"
        >
          + Add Document
        </button>
      </div>
        {/* Stats */}
        <div className="grid grid-cols-3 gap-4">
          <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 p-4 shadow-sm">
            <p className="text-2xl font-bold text-gray-800 dark:text-gray-100">{total}</p>
            <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">Total Documents</p>
          </div>
          <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 p-4 shadow-sm">
            <p className="text-2xl font-bold text-green-600">{docs.filter((d) => d.is_active).length}</p>
            <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">Active (used by AI)</p>
          </div>
          <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 p-4 shadow-sm">
            <p className="text-2xl font-bold text-gray-400 dark:text-gray-500">{CATEGORIES.length}</p>
            <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">Categories</p>
          </div>
        </div>

        {/* Filters */}
        <div className="flex gap-3">
          <input
            type="text"
            placeholder="Search documents…"
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
            className="flex-1 rounded-lg border border-gray-200 dark:border-gray-600 px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-blue-400"
          />
          <select
            value={categoryFilter}
            onChange={(e) => { setCategoryFilter(e.target.value); setPage(1) }}
            className="select-field"
          >
            <option value="">All Categories</option>
            {CATEGORIES.map((c) => <option key={c} value={c}>{c}</option>)}
          </select>
        </div>

        {/* Error */}
        {error && (
          <div className="rounded-lg bg-red-50 border border-red-100 px-4 py-3 text-sm text-red-600">{error}</div>
        )}

        {/* Table */}
        <div className="rounded-xl border border-gray-100 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-sm overflow-hidden">
          {loading ? (
            <div className="py-16 text-center text-sm text-gray-400 dark:text-gray-500 animate-pulse">Loading…</div>
          ) : docs.length === 0 ? (
            <div className="py-16 text-center">
              <p className="text-2xl mb-2">📚</p>
              <p className="text-sm text-gray-500 dark:text-gray-400 dark:text-gray-500 font-medium">No documents found</p>
              <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">Add knowledge base articles to enable AI reply generation</p>
            </div>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-900">
                  <th className="px-5 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 dark:text-gray-500 uppercase tracking-wide">Title</th>
                  <th className="px-5 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 dark:text-gray-500 uppercase tracking-wide">Category</th>
                  <th className="px-5 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 dark:text-gray-500 uppercase tracking-wide">Status</th>
                  <th className="px-5 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 dark:text-gray-500 uppercase tracking-wide">Updated</th>
                  <th className="px-5 py-3 text-right text-xs font-semibold text-gray-500 dark:text-gray-400 dark:text-gray-500 uppercase tracking-wide">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-50">
                {docs.map((doc) => (
                  <tr key={doc.id} className="hover:bg-gray-50 dark:bg-gray-900 transition">
                    <td className="px-5 py-4">
                      <div className="font-medium text-gray-800 dark:text-gray-100">{doc.title}</div>
                      <div className="text-xs text-gray-400 dark:text-gray-500 mt-0.5 line-clamp-1">{doc.content.slice(0, 100)}…</div>
                    </td>
                    <td className="px-5 py-4">
                      <span className="rounded-full bg-blue-50 text-blue-700 px-2 py-0.5 text-xs font-medium">{doc.category}</span>
                    </td>
                    <td className="px-5 py-4">
                      <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${doc.is_active ? 'bg-green-50 text-green-700' : 'bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 dark:text-gray-500'}`}>
                        {doc.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                    <td className="px-5 py-4 text-xs text-gray-400 dark:text-gray-500">{formatDate(doc.updated_at)}</td>
                    <td className="px-5 py-4 text-right">
                      <div className="flex justify-end gap-2">
                        <button
                          onClick={() => setEditDoc(doc)}
                          className="text-xs text-blue-500 hover:text-blue-700 font-medium transition"
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => handleDelete(doc.id)}
                          className="text-xs text-red-400 hover:text-red-600 font-medium transition"
                        >
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex justify-center gap-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
              className="rounded-lg border border-gray-200 dark:border-gray-600 px-3 py-1.5 text-xs font-medium text-gray-600 dark:text-gray-300 dark:text-gray-600 disabled:opacity-40 hover:bg-gray-50 dark:bg-gray-900 transition"
            >
              Previous
            </button>
            <span className="px-3 py-1.5 text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500">Page {page} of {totalPages}</span>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="rounded-lg border border-gray-200 dark:border-gray-600 px-3 py-1.5 text-xs font-medium text-gray-600 dark:text-gray-300 dark:text-gray-600 disabled:opacity-40 hover:bg-gray-50 dark:bg-gray-900 transition"
            >
              Next
            </button>
          </div>
        )}

      {/* Create modal */}
      {showModal && (
        <Modal title="Add Knowledge Base Document" onClose={() => setShowModal(false)}>
          <DocForm onSubmit={handleCreate} onCancel={() => setShowModal(false)} saving={saving} />
        </Modal>
      )}

      {/* Edit modal */}
      {editDoc && (
        <Modal title="Edit Knowledge Base Document" onClose={() => setEditDoc(null)}>
          <DocForm
            initial={{ title: editDoc.title, category: editDoc.category, content: editDoc.content, is_active: editDoc.is_active }}
            onSubmit={handleUpdate}
            onCancel={() => setEditDoc(null)}
            saving={saving}
          />
        </Modal>
      )}

      {/* Toast */}
      {toast && (
        <div className={`fixed top-4 right-4 z-50 rounded-lg px-4 py-3 text-sm font-medium shadow-lg ${
          toast.type === 'error' ? 'bg-red-500 text-white' : 'bg-gray-900 text-white'
        }`}>
          {toast.msg}
        </div>
      )}
    </div>
  )
}
