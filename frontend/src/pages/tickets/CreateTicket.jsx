import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { ticketService } from '../../services/ticketService'
import Toast, { useToast } from '../../components/Toast'
import { TICKET_CATEGORIES, TICKET_PRIORITIES } from '../../utils/categories'

const PRIORITIES = TICKET_PRIORITIES
const CATEGORIES = TICKET_CATEGORIES

const inputCls = (err) =>
  `w-full px-4 py-2.5 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent dark:bg-gray-700 dark:text-gray-100 dark:placeholder-gray-400 transition ${
    err ? 'border-red-400' : 'border-gray-300 dark:border-gray-600'
  }`

function CreateTicket() {
  const navigate = useNavigate()
  const { toast, showToast } = useToast()

  const [form, setForm] = useState({
    subject: '',
    description: '',
    customerName: '',
    customerEmail: '',
    priority: 'MEDIUM',
    category: 'GENERAL',
  })
  const [errors, setErrors] = useState({})
  const [submitting, setSubmitting] = useState(false)

  const handleChange = (e) => {
    setForm((prev) => ({ ...prev, [e.target.name]: e.target.value }))
    setErrors((prev) => ({ ...prev, [e.target.name]: '' }))
  }

  const validate = () => {
    const next = {}
    if (!form.subject.trim() || form.subject.trim().length < 5)
      next.subject = 'Subject must be at least 5 characters.'
    if (form.subject.length > 150)
      next.subject = 'Subject must not exceed 150 characters.'
    if (!form.description.trim())
      next.description = 'Description is required.'
    if (!form.customerName.trim())
      next.customerName = 'Customer name is required.'
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.customerEmail))
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
      const res = await ticketService.createTicket(form)
      const ticket = res.data.data
      showToast(`Ticket ${ticket.ticket_number} created! AI analysis started…`, 'success')
      setTimeout(() => navigate(`/tickets/${ticket.id}`), 1200)
    } catch (err) {
      showToast(err.response?.data?.message ?? 'Failed to create ticket', 'error')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <>
      <Toast toast={toast} />
      <main className="max-w-2xl mx-auto px-6 py-8 animate-fade-up">
        {/* Header */}
        <div className="mb-6 flex items-center justify-between animate-fade-in">
          <div>
            <h1 className="text-xl font-bold text-gray-900 dark:text-white">Create New Ticket</h1>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">AI will automatically analyze and categorize this ticket</p>
          </div>
          <Link to="/tickets" className="text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 flex items-center gap-1 transition-colors">
            ← Back to Tickets
          </Link>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-2xl border border-gray-100 dark:border-gray-700 p-8 shadow-sm animate-scale-in">
          <form onSubmit={handleSubmit} className="space-y-5">

            {/* Subject */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1.5">
                Subject <span className="text-red-500">*</span>
              </label>
              <input
                name="subject"
                value={form.subject}
                onChange={handleChange}
                placeholder="Brief description of the issue"
                className={inputCls(errors.subject)}
              />
              {errors.subject && <p className="mt-1 text-xs text-red-500">{errors.subject}</p>}
            </div>

            {/* Description */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1.5">
                Description <span className="text-red-500">*</span>
              </label>
              <textarea
                name="description"
                value={form.description}
                onChange={handleChange}
                rows={5}
                placeholder="Detailed description of the issue — more detail helps AI give better analysis"
                className={`${inputCls(errors.description)} resize-none`}
              />
              {errors.description && <p className="mt-1 text-xs text-red-500">{errors.description}</p>}
            </div>

            {/* Priority + Category */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1.5">Priority</label>
                <select
                  name="priority"
                  value={form.priority}
                  onChange={handleChange}
                  className={inputCls(false)}
                >
                  {PRIORITIES.map((p) => (
                    <option key={p.value} value={p.value}>{p.label}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1.5">Category</label>
                <select
                  name="category"
                  value={form.category}
                  onChange={handleChange}
                  className={inputCls(false)}
                >
                  {CATEGORIES.map((c) => (
                    <option key={c.value} value={c.value}>{c.label}</option>
                  ))}
                </select>
              </div>
            </div>

            {/* Customer Name + Email */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1.5">
                  Customer Name <span className="text-red-500">*</span>
                </label>
                <input
                  name="customerName"
                  value={form.customerName}
                  onChange={handleChange}
                  placeholder="Rahul Sharma"
                  className={inputCls(errors.customerName)}
                />
                {errors.customerName && <p className="mt-1 text-xs text-red-500">{errors.customerName}</p>}
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1.5">
                  Customer Email <span className="text-red-500">*</span>
                </label>
                <input
                  name="customerEmail"
                  type="email"
                  value={form.customerEmail}
                  onChange={handleChange}
                  placeholder="rahul@example.com"
                  className={inputCls(errors.customerEmail)}
                />
                {errors.customerEmail && <p className="mt-1 text-xs text-red-500">{errors.customerEmail}</p>}
              </div>
            </div>

            {/* AI notice */}
            <div className="flex items-start gap-2.5 bg-blue-50 dark:bg-blue-900/20 border border-blue-100 dark:border-blue-800 rounded-lg px-4 py-3">
              <svg className="w-4 h-4 text-blue-500 mt-0.5 shrink-0" fill="currentColor" viewBox="0 0 20 20">
                <path d="M13 6a3 3 0 11-6 0 3 3 0 016 0zM18 8a2 2 0 11-4 0 2 2 0 014 0zM14 15a4 4 0 00-8 0v3h8v-3z"/>
              </svg>
              <p className="text-xs text-blue-700 dark:text-blue-300">
                After creating, AI will automatically analyze this ticket — detecting priority, sentiment, and recommending which team should handle it.
              </p>
            </div>

            {/* Actions */}
            <div className="flex gap-3 pt-1">
              <button
                type="submit"
                disabled={submitting}
                className="flex-1 py-2.5 bg-blue-600 text-white font-semibold rounded-lg hover:bg-blue-700 transition disabled:opacity-50 disabled:cursor-not-allowed text-sm"
              >
                {submitting ? 'Creating…' : 'Create Ticket'}
              </button>
              <Link
                to="/tickets"
                className="px-6 py-2.5 border border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 font-medium rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition text-sm text-center"
              >
                Cancel
              </Link>
            </div>
          </form>
        </div>
      </main>
    </>
  )
}

export default CreateTicket
