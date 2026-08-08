import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { ticketService } from '../../services/ticketService'
import Toast, { useToast } from '../../components/Toast'
import { TICKET_CATEGORIES, TICKET_PRIORITIES } from '../../utils/categories'

const CATEGORIES = TICKET_CATEGORIES
const PRIORITIES = TICKET_PRIORITIES

function EditTicket() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { toast, showToast } = useToast()

  const [form, setForm] = useState({
    subject: '',
    description: '',
    priority: '',
    category: '',
    customerName: '',
    customerEmail: '',
  })
  const [errors, setErrors] = useState({})
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    ticketService.getTicket(id)
      .then((res) => {
        const t = res.data.data
        setForm({
          subject: t.subject,
          description: t.description,
          priority: t.priority,
          category: t.category,
          customerName: t.customer_name,
          customerEmail: t.customer_email,
        })
      })
      .catch(() => showToast('Failed to load ticket', 'error'))
      .finally(() => setLoading(false))
  }, [id]) // eslint-disable-line

  const handleChange = (e) => {
    setForm((prev) => ({ ...prev, [e.target.name]: e.target.value }))
    setErrors((prev) => ({ ...prev, [e.target.name]: '' }))
  }

  const validate = () => {
    const next = {}
    if (form.subject && form.subject.trim().length < 5)
      next.subject = 'Subject must be at least 5 characters.'
    if (form.customerEmail && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.customerEmail))
      next.customerEmail = 'Please enter a valid email address.'
    return next
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    const fieldErrors = validate()
    if (Object.keys(fieldErrors).length > 0) {
      setErrors(fieldErrors)
      return
    }

    setSubmitting(true)
    try {
      await ticketService.updateTicket(id, form)
      showToast('Ticket updated successfully')
      setTimeout(() => navigate(`/tickets/${id}`), 1000)
    } catch (err) {
      showToast(err.response?.data?.message ?? 'Failed to update ticket', 'error')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-24">
        <p className="text-sm text-gray-400 dark:text-gray-500 animate-pulse">Loading…</p>
      </div>
    )
  }

  return (
    <>
      <Toast toast={toast} />
      <main className="flex items-center justify-center py-10 px-6">
        <div className="relative w-full max-w-2xl bg-white dark:bg-gray-800 rounded-2xl border border-gray-100 dark:border-gray-700 p-8 shadow-sm">

          {/* ✕ close button */}
          <Link
            to={`/tickets/${id}`}
            className="absolute top-4 right-4 w-8 h-8 flex items-center justify-center rounded-lg text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 transition"
            title="Cancel"
          >
            <svg xmlns="http://www.w3.org/2000/svg" className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </Link>

          <h1 className="text-lg font-semibold text-gray-900 dark:text-white mb-6">Edit Ticket</h1>

          <form onSubmit={handleSubmit} className="space-y-5">
            {/* Subject */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1.5">Subject</label>
              <input
                name="subject"
                value={form.subject}
                onChange={handleChange}
                className={`w-full px-4 py-2.5 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 transition ${
                  errors.subject ? 'border-red-400' : 'border-gray-300 dark:border-gray-600'
                }`}
              />
              {errors.subject && <p className="mt-1 text-xs text-red-500">{errors.subject}</p>}
            </div>

            {/* Description */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1.5">Description</label>
              <textarea
                name="description"
                value={form.description}
                onChange={handleChange}
                rows={5}
                className="w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 transition resize-y min-h-[80px]"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              {/* Priority */}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1.5">Priority</label>
                <select name="priority" value={form.priority} onChange={handleChange} className="select-field w-full">
                  {PRIORITIES.map((p) => <option key={p.value} value={p.value}>{p.label}</option>)}
                </select>
              </div>
              {/* Category */}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1.5">Category</label>
                <select name="category" value={form.category} onChange={handleChange} className="select-field w-full">
                  {CATEGORIES.map((c) => <option key={c.value} value={c.value}>{c.label}</option>)}
                </select>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              {/* Customer Name */}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1.5">Customer Name</label>
                <input
                  name="customerName"
                  value={form.customerName}
                  onChange={handleChange}
                  className="w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 transition"
                />
              </div>
              {/* Customer Email */}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1.5">Customer Email</label>
                <input
                  name="customerEmail"
                  type="email"
                  value={form.customerEmail}
                  onChange={handleChange}
                  className={`w-full px-4 py-2.5 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 transition ${
                    errors.customerEmail ? 'border-red-400' : 'border-gray-300 dark:border-gray-600'
                  }`}
                />
                {errors.customerEmail && <p className="mt-1 text-xs text-red-500">{errors.customerEmail}</p>}
              </div>
            </div>

            <div className="pt-2">
              <button
                type="submit"
                disabled={submitting}
                className="w-full py-2.5 bg-emerald-600 text-white font-semibold rounded-lg hover:bg-emerald-700 transition disabled:opacity-50 text-sm"
              >
                {submitting ? 'Saving…' : 'Save Changes'}
              </button>
            </div>
          </form>
        </div>
      </main>
    </>
  )
}

export default EditTicket
