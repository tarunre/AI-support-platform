import { useState, useEffect } from 'react'
import emailService from '../services/emailService'
import { useToast } from '../components/Toast'

const PROVIDERS = ['SMTP_IMAP', 'GMAIL', 'OUTLOOK', 'SENDGRID', 'SES', 'MAILGUN']

const emptyForm = {
  provider: 'SMTP_IMAP',
  email_address: '',
  display_name: '',
  imap_host: '',
  imap_port: 993,
  imap_use_tls: true,
  smtp_host: '',
  smtp_port: 587,
  smtp_implicit_tls: false,
  username: '',
  password: '',
  is_active: true,
}

function Badge({ active }) {
  return (
    <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${active ? 'bg-green-100 text-green-700' : 'bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 dark:text-gray-500'}`}>
      {active ? 'Active' : 'Inactive'}
    </span>
  )
}

export default function EmailAccounts() {
  const [accounts, setAccounts] = useState([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [editing, setEditing] = useState(null)
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)
  const [testing, setTesting] = useState({})
  const { showToast } = useToast()

  const load = async () => {
    try {
      const res = await emailService.listAccounts()
      setAccounts(res.data?.data || [])
    } catch {
      showToast('Failed to load email accounts', 'error')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const openCreate = () => {
    setEditing(null)
    setForm(emptyForm)
    setShowModal(true)
  }

  const openEdit = (a) => {
    setEditing(a)
    setForm({
      provider: a.provider,
      email_address: a.email_address,
      display_name: a.display_name || '',
      imap_host: a.imap_host || '',
      imap_port: a.imap_port || 993,
      imap_use_tls: a.imap_use_tls,
      smtp_host: a.smtp_host || '',
      smtp_port: a.smtp_port || 587,
      smtp_implicit_tls: a.smtp_implicit_tls,
      username: a.username,
      password: '',
      is_active: a.is_active,
    })
    setShowModal(true)
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      if (editing) {
        const payload = { ...form }
        if (!payload.password) delete payload.password
        await emailService.updateAccount(editing.id, payload)
        showToast('Account updated')
      } else {
        await emailService.createAccount(form)
        showToast('Account created')
      }
      setShowModal(false)
      load()
    } catch (err) {
      showToast(err.response?.data?.message || 'Failed to save account', 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id) => {
    if (!confirm('Delete this email account?')) return
    try {
      await emailService.deleteAccount(id)
      showToast('Account deleted')
      load()
    } catch {
      showToast('Failed to delete account', 'error')
    }
  }

  const handleTest = async (id, proto) => {
    setTesting(prev => ({ ...prev, [`${id}_${proto}`]: true }))
    try {
      await emailService.testConnection(id, proto)
      showToast(`${proto.toUpperCase()} connection successful`, 'success')
    } catch (err) {
      showToast(err.response?.data?.message || `${proto.toUpperCase()} test failed`, 'error')
    } finally {
      setTesting(prev => ({ ...prev, [`${id}_${proto}`]: false }))
    }
  }

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Email Accounts</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 dark:text-gray-500 mt-1">Configure inbound and outbound email connections</p>
        </div>
        <button onClick={openCreate} className="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm font-medium hover:bg-indigo-700">
          + Add Account
        </button>
      </div>

      {loading ? (
        <div className="text-center py-16 text-gray-400 dark:text-gray-500">Loading…</div>
      ) : accounts.length === 0 ? (
        <div className="text-center py-16 text-gray-400 dark:text-gray-500">No email accounts configured yet.</div>
      ) : (
        <div className="grid gap-4">
          {accounts.map(a => (
            <div key={a.id} className="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 shadow-sm p-5">
              <div className="flex items-start justify-between">
                <div>
                  <div className="flex items-center gap-3">
                    <span className="font-semibold text-gray-900 dark:text-white">{a.email_address}</span>
                    <Badge active={a.is_active} />
                    <span className="text-xs text-gray-400 dark:text-gray-500 bg-gray-50 dark:bg-gray-900 px-2 py-0.5 rounded">{a.provider}</span>
                  </div>
                  {a.display_name && <p className="text-sm text-gray-500 dark:text-gray-400 dark:text-gray-500 mt-1">{a.display_name}</p>}
                  <div className="flex gap-6 mt-3 text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500">
                    {a.imap_host && (
                      <span>IMAP: {a.imap_host}:{a.imap_port} {a.imap_use_tls ? '(TLS)' : ''}</span>
                    )}
                    {a.smtp_host && (
                      <span>SMTP: {a.smtp_host}:{a.smtp_port} {a.smtp_implicit_tls ? '(TLS)' : '(STARTTLS)'}</span>
                    )}
                    {a.last_sync_at && (
                      <span>Last sync: {new Date(a.last_sync_at).toLocaleString()}</span>
                    )}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {a.imap_host && (
                    <button
                      onClick={() => handleTest(a.id, 'imap')}
                      disabled={testing[`${a.id}_imap`]}
                      className="text-xs text-blue-600 hover:text-blue-800 disabled:opacity-50"
                    >
                      {testing[`${a.id}_imap`] ? 'Testing…' : 'Test IMAP'}
                    </button>
                  )}
                  {a.smtp_host && (
                    <button
                      onClick={() => handleTest(a.id, 'smtp')}
                      disabled={testing[`${a.id}_smtp`]}
                      className="text-xs text-blue-600 hover:text-blue-800 disabled:opacity-50"
                    >
                      {testing[`${a.id}_smtp`] ? 'Testing…' : 'Test SMTP'}
                    </button>
                  )}
                  <button onClick={() => openEdit(a)} className="text-xs text-indigo-600 hover:text-indigo-800">Edit</button>
                  <button onClick={() => handleDelete(a.id)} className="text-xs text-red-500 hover:text-red-700">Delete</button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4">
          <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <div className="px-6 py-4 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between">
              <h2 className="text-lg font-semibold">{editing ? 'Edit' : 'Add'} Email Account</h2>
              <button onClick={() => setShowModal(false)} className="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:text-gray-300 dark:text-gray-600 text-xl">×</button>
            </div>

            <div className="p-6 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1">Provider</label>
                  <select
                    value={form.provider}
                    onChange={e => setForm(f => ({ ...f, provider: e.target.value }))}
                    className="select-field w-full"
                  >
                    {PROVIDERS.map(p => <option key={p}>{p}</option>)}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1">Display Name</label>
                  <input
                    value={form.display_name}
                    onChange={e => setForm(f => ({ ...f, display_name: e.target.value }))}
                    className="w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm"
                    placeholder="Support Team"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1">Email Address *</label>
                  <input
                    value={form.email_address}
                    onChange={e => setForm(f => ({ ...f, email_address: e.target.value }))}
                    className="w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm"
                    placeholder="support@example.com"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1">Username *</label>
                  <input
                    value={form.username}
                    onChange={e => setForm(f => ({ ...f, username: e.target.value }))}
                    className="w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm"
                    placeholder="support@example.com"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1">
                  Password {editing ? '(leave blank to keep existing)' : '*'}
                </label>
                <input
                  type="password"
                  value={form.password}
                  onChange={e => setForm(f => ({ ...f, password: e.target.value }))}
                  className="w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm"
                  autoComplete="new-password"
                />
              </div>

              <hr />
              <p className="text-xs font-semibold text-gray-500 dark:text-gray-400 dark:text-gray-500 uppercase tracking-wide">IMAP (Inbound)</p>
              <div className="grid grid-cols-3 gap-4">
                <div className="col-span-2">
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1">IMAP Host</label>
                  <input
                    value={form.imap_host}
                    onChange={e => setForm(f => ({ ...f, imap_host: e.target.value }))}
                    className="w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm"
                    placeholder="imap.example.com"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1">Port</label>
                  <input
                    type="number"
                    value={form.imap_port}
                    onChange={e => setForm(f => ({ ...f, imap_port: +e.target.value }))}
                    className="w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm"
                  />
                </div>
              </div>
              <label className="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-200">
                <input type="checkbox" checked={form.imap_use_tls} onChange={e => setForm(f => ({ ...f, imap_use_tls: e.target.checked }))} />
                Use TLS (recommended, port 993)
              </label>

              <hr />
              <p className="text-xs font-semibold text-gray-500 dark:text-gray-400 dark:text-gray-500 uppercase tracking-wide">SMTP (Outbound)</p>
              <div className="grid grid-cols-3 gap-4">
                <div className="col-span-2">
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1">SMTP Host</label>
                  <input
                    value={form.smtp_host}
                    onChange={e => setForm(f => ({ ...f, smtp_host: e.target.value }))}
                    className="w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm"
                    placeholder="smtp.example.com"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1">Port</label>
                  <input
                    type="number"
                    value={form.smtp_port}
                    onChange={e => setForm(f => ({ ...f, smtp_port: +e.target.value }))}
                    className="w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm"
                  />
                </div>
              </div>
              <label className="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-200">
                <input type="checkbox" checked={form.smtp_implicit_tls} onChange={e => setForm(f => ({ ...f, smtp_implicit_tls: e.target.checked }))} />
                Implicit TLS (port 465) — uncheck for STARTTLS (port 587)
              </label>

              <label className="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-200">
                <input type="checkbox" checked={form.is_active} onChange={e => setForm(f => ({ ...f, is_active: e.target.checked }))} />
                Active (enable IMAP polling for this account)
              </label>
            </div>

            <div className="px-6 py-4 border-t border-gray-100 dark:border-gray-700 flex justify-end gap-3">
              <button onClick={() => setShowModal(false)} className="px-4 py-2 text-gray-600 dark:text-gray-300 dark:text-gray-600 hover:text-gray-800 dark:text-gray-100 text-sm">Cancel</button>
              <button onClick={handleSave} disabled={saving} className="px-5 py-2 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700 disabled:opacity-50">
                {saving ? 'Saving…' : editing ? 'Update' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
