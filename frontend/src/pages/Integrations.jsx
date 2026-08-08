import { useState, useEffect } from 'react';
import { integrationService } from '../services/integrationService';

const PROVIDER_META = {
  slack:       { label: 'Slack',            icon: '💬', fields: [
    { key: 'webhook_url',        label: 'Outbound Webhook URL', required: true, hint: 'For sending ticket notifications TO Slack' },
    { key: 'signing_secret',     label: 'Signing Secret', secret: true, hint: 'From Slack App → Basic Information → Signing Secret' },
    { key: 'bot_token',          label: 'Bot Token (xoxb-...)', secret: true, hint: 'From Slack App → OAuth & Permissions → Bot User OAuth Token' },
    { key: 'support_channel_id', label: 'Support Channel ID', hint: 'Channel where users post support requests (e.g. C0123456789)' },
  ]},
  teams:       { label: 'Microsoft Teams',  icon: '🟦', fields: [{ key: 'webhook_url', label: 'Webhook URL', required: true }] },
  discord:     { label: 'Discord',          icon: '🎮', fields: [{ key: 'webhook_url', label: 'Webhook URL', required: true }, { key: 'username', label: 'Bot Username' }] },
  jira:        { label: 'Jira',             icon: '🔵', fields: [
    { key: 'base_url',     label: 'Base URL (e.g. https://org.atlassian.net)', required: true },
    { key: 'email',        label: 'Email',        required: true },
    { key: 'api_token',    label: 'API Token',    required: true, secret: true },
    { key: 'project_key',  label: 'Project Key',  required: true },
    { key: 'issue_type',   label: 'Issue Type (default: Task)' },
  ]},
  linear:      { label: 'Linear',           icon: '⚡', fields: [
    { key: 'api_key',  label: 'API Key',  required: true, secret: true },
    { key: 'team_id',  label: 'Team ID',  required: true },
  ]},
  github:      { label: 'GitHub',           icon: '🐙', fields: [
    { key: 'token', label: 'Personal Access Token', required: true, secret: true },
    { key: 'owner', label: 'Repository Owner',      required: true },
    { key: 'repo',  label: 'Repository Name',       required: true },
  ]},
  webhook:     { label: 'Webhook',          icon: '🔗', fields: [
    { key: 'url',             label: 'Endpoint URL', required: true },
    { key: 'secret',          label: 'Signing Secret', secret: true },
    { key: 'timeout_seconds', label: 'Timeout (seconds)' },
  ]},
  salesforce:  { label: 'Salesforce',       icon: '☁️', fields: [
    { key: 'instance_url',  label: 'Instance URL',  required: true },
    { key: 'access_token',  label: 'Access Token',  required: true, secret: true },
  ]},
  hubspot:     { label: 'HubSpot',          icon: '🟠', fields: [
    { key: 'access_token', label: 'Access Token', required: true, secret: true },
  ]},
  gcal:        { label: 'Google Calendar',  icon: '📅', fields: [
    { key: 'access_token', label: 'Access Token',  required: true, secret: true },
    { key: 'calendar_id',  label: 'Calendar ID (default: primary)' },
  ]},
};

const STATUS_COLORS = {
  ACTIVE:   'bg-green-100 text-green-800',
  ERROR:    'bg-red-100 text-red-800',
  INACTIVE: 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-200',
};

export default function Integrations() {
  const [integrations, setIntegrations] = useState([]);
  const [loading, setLoading]           = useState(true);
  const [error, setError]               = useState(null);
  const [modal, setModal]               = useState(null); // { mode: 'create'|'edit', data? }
  const [form, setForm]                 = useState({ provider: 'slack', name: '', config: {}, enabled: false });
  const [saving, setSaving]             = useState(false);
  const [testingId, setTestingId]       = useState(null);
  const [testResult, setTestResult]     = useState({});

  useEffect(() => { fetchIntegrations(); }, []);

  async function fetchIntegrations() {
    setLoading(true);
    try {
      const res = await integrationService.list();
      setIntegrations(res.data.integrations || []);
    } catch (e) {
      setError('Failed to load integrations');
    } finally {
      setLoading(false);
    }
  }

  function openCreate() {
    setForm({ provider: 'slack', name: '', config: {}, enabled: false });
    setModal({ mode: 'create' });
  }

  function openEdit(intg) {
    setForm({ provider: intg.provider, name: intg.name, config: {}, enabled: intg.enabled });
    setModal({ mode: 'edit', data: intg });
  }

  async function handleSave(e) {
    e.preventDefault();
    setSaving(true);
    try {
      if (modal.mode === 'create') {
        await integrationService.create({ provider: form.provider, name: form.name, config: form.config, enabled: form.enabled });
      } else {
        await integrationService.update(modal.data.id, { name: form.name, config: Object.keys(form.config).length ? form.config : undefined, enabled: form.enabled });
      }
      setModal(null);
      await fetchIntegrations();
    } catch (e) {
      alert(e.response?.data?.error || 'Save failed');
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(id) {
    if (!window.confirm('Delete this integration?')) return;
    try {
      await integrationService.delete(id);
      await fetchIntegrations();
    } catch {
      alert('Delete failed');
    }
  }

  async function handleTest(id) {
    setTestingId(id);
    setTestResult((prev) => ({ ...prev, [id]: null }));
    try {
      await integrationService.test(id);
      setTestResult((prev) => ({ ...prev, [id]: { success: true } }));
      await fetchIntegrations();
    } catch (e) {
      setTestResult((prev) => ({ ...prev, [id]: { success: false, msg: e.response?.data?.error || 'Test failed' } }));
    } finally {
      setTestingId(null);
    }
  }

  function setConfigField(key, val) {
    setForm((f) => ({ ...f, config: { ...f.config, [key]: val } }));
  }

  const providerFields = PROVIDER_META[form.provider]?.fields || [];

  if (loading) return <div className="p-8 text-gray-500 dark:text-gray-400 dark:text-gray-500">Loading integrations…</div>;
  if (error)   return <div className="p-8 text-red-500">{error}</div>;

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Enterprise Integrations</h1>
          <p className="text-gray-500 dark:text-gray-400 dark:text-gray-500 text-sm mt-1">Connect SupportIQ to your external tools and services.</p>
        </div>
        <button
          onClick={openCreate}
          className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition text-sm font-medium"
        >
          + Add Integration
        </button>
      </div>

      {integrations.length === 0 ? (
        <div className="text-center py-16 text-gray-400 dark:text-gray-500 border-2 border-dashed rounded-xl">
          <p className="text-4xl mb-3">🔌</p>
          <p className="font-medium">No integrations yet</p>
          <p className="text-sm">Click "Add Integration" to connect your first service.</p>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {integrations.map((intg) => (
            <div key={intg.id} className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-xl p-5 shadow-sm">
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <span className="text-2xl">{PROVIDER_META[intg.provider]?.icon || '🔗'}</span>
                  <div>
                    <h3 className="font-semibold text-gray-900 dark:text-white">{intg.name}</h3>
                    <p className="text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500">{PROVIDER_META[intg.provider]?.label || intg.provider}</p>
                  </div>
                </div>
                <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${STATUS_COLORS[intg.status] || 'bg-gray-100 dark:bg-gray-800'}`}>
                  {intg.status}
                </span>
              </div>

              {intg.error_message && (
                <p className="mt-2 text-xs text-red-600 bg-red-50 rounded p-2 truncate">{intg.error_message}</p>
              )}

              {testResult[intg.id] && (
                <p className={`mt-2 text-xs px-2 py-1 rounded ${testResult[intg.id].success ? 'text-green-700 bg-green-50' : 'text-red-700 bg-red-50'}`}>
                  {testResult[intg.id].success ? '✓ Connection successful' : `✗ ${testResult[intg.id].msg}`}
                </p>
              )}

              <div className="mt-4 flex items-center gap-2 flex-wrap">
                <button
                  onClick={() => handleTest(intg.id)}
                  disabled={testingId === intg.id}
                  className="text-xs px-3 py-1.5 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 rounded-lg transition disabled:opacity-50"
                >
                  {testingId === intg.id ? 'Testing…' : 'Test'}
                </button>
                <button
                  onClick={() => openEdit(intg)}
                  className="text-xs px-3 py-1.5 bg-blue-50 hover:bg-blue-100 text-blue-700 rounded-lg transition"
                >
                  Edit
                </button>
                <button
                  onClick={() => handleDelete(intg.id)}
                  className="text-xs px-3 py-1.5 bg-red-50 hover:bg-red-100 text-red-700 rounded-lg transition"
                >
                  Delete
                </button>
                <div className="ml-auto flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500">
                  {intg.last_sync_at && <span>Last sync: {new Date(intg.last_sync_at).toLocaleDateString()}</span>}
                  <span className={`w-2 h-2 rounded-full ${intg.enabled ? 'bg-green-400' : 'bg-gray-300'}`} />
                  <span>{intg.enabled ? 'Enabled' : 'Disabled'}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create / Edit Modal */}
      {modal && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-lg max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b">
              <h2 className="text-lg font-bold text-gray-900 dark:text-white">
                {modal.mode === 'create' ? 'Add Integration' : `Edit: ${modal.data.name}`}
              </h2>
            </div>
            <form onSubmit={handleSave} className="p-6 space-y-4">
              {modal.mode === 'create' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1">Provider</label>
                  <select
                    value={form.provider}
                    onChange={(e) => setForm((f) => ({ ...f, provider: e.target.value, config: {} }))}
                    className="select-field w-full"
                  >
                    {Object.entries(PROVIDER_META).map(([key, meta]) => (
                      <option key={key} value={key}>{meta.icon} {meta.label}</option>
                    ))}
                  </select>
                </div>
              )}

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1">Name</label>
                <input
                  required
                  type="text"
                  placeholder="e.g. Production Slack"
                  value={form.name}
                  onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                  className="w-full border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
                />
              </div>

              {providerFields.map((field) => (
                <div key={field.key}>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1">
                    {field.label}{field.required && <span className="text-red-500 ml-0.5">*</span>}
                  </label>
                  <input
                    required={field.required && modal.mode === 'create'}
                    type={field.secret ? 'password' : 'text'}
                    placeholder={field.secret ? '••••••••' : ''}
                    value={form.config[field.key] || ''}
                    onChange={(e) => setConfigField(field.key, e.target.value)}
                    className="w-full border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
                  />
                  {field.hint && (
                    <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">{field.hint}</p>
                  )}
                </div>
              ))}

              <div className="flex items-center gap-2">
                <input
                  id="enabled"
                  type="checkbox"
                  checked={form.enabled}
                  onChange={(e) => setForm((f) => ({ ...f, enabled: e.target.checked }))}
                  className="rounded"
                />
                <label htmlFor="enabled" className="text-sm text-gray-700 dark:text-gray-200">Enable this integration</label>
              </div>

              <div className="flex gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setModal(null)}
                  className="flex-1 px-4 py-2 border rounded-lg text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:bg-gray-900"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={saving}
                  className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
                >
                  {saving ? 'Saving…' : (modal.mode === 'create' ? 'Add Integration' : 'Save Changes')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
