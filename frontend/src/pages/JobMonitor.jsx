import { useEffect, useState } from 'react';
import jobService from '../services/jobService';

const STATUS_COLORS = {
  QUEUED:     'bg-yellow-100 text-yellow-800',
  PROCESSING: 'bg-blue-100 text-blue-800',
  COMPLETED:  'bg-green-100 text-green-800',
  FAILED:     'bg-red-100 text-red-800',
  RETRYING:   'bg-orange-100 text-orange-800',
  DEAD:       'bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-100',
};

const JOB_TYPE_LABELS = {
  AI_ANALYSIS:      'AI Analysis',
  GENERATE_REPLY:   'Generate Reply',
  REGENERATE_REPLY: 'Regenerate Reply',
  RETRY_AI:         'Retry AI',
  RETRY_REPLY:      'Retry Reply',
};

function StatCard({ label, value, color }) {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-5">
      <p className="text-sm text-gray-500 dark:text-gray-400 dark:text-gray-500">{label}</p>
      <p className={`text-3xl font-bold mt-1 ${color}`}>{value}</p>
    </div>
  );
}

function duration(startedAt, completedAt) {
  if (!startedAt || !completedAt) return '—';
  const ms = new Date(completedAt) - new Date(startedAt);
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}

export default function JobMonitor() {
  const [jobs, setJobs]   = useState([]);
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter]   = useState('');
  const [page, setPage]       = useState(1);
  const [retrying, setRetrying] = useState({});
  const PER_PAGE = 20;

  const fetchAll = async () => {
    try {
      const [jobsRes, statsRes] = await Promise.all([
        jobService.list({ page: 1, per_page: 200 }),
        jobService.statistics(),
      ]);
      setJobs(jobsRes.data?.jobs || []);
      setStats(statsRes.data);
    } catch {
      // errors shown inline
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAll();
    const interval = setInterval(fetchAll, 10000); // auto-refresh every 10 s
    return () => clearInterval(interval);
  }, []);

  const handleRetry = async (id) => {
    setRetrying((prev) => ({ ...prev, [id]: true }));
    try {
      await jobService.retry(id);
      await fetchAll();
    } finally {
      setRetrying((prev) => ({ ...prev, [id]: false }));
    }
  };

  const filtered = jobs.filter((j) => {
    if (!filter) return true;
    return (
      j.status === filter ||
      j.job_type === filter ||
      j.id?.toString().includes(filter) ||
      j.reference_id?.includes(filter)
    );
  });

  const paged = filtered.slice((page - 1) * PER_PAGE, page * PER_PAGE);
  const totalPages = Math.ceil(filtered.length / PER_PAGE);

  return (
    <div className="max-w-6xl mx-auto px-6 py-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Job Monitor</h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 dark:text-gray-500 mt-1">Background processing queue status</p>
      </div>

      {/* Stats */}
      {stats && (
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4 mb-8">
          <StatCard label="Queued"     value={stats.queued     ?? 0} color="text-yellow-600" />
          <StatCard label="Processing" value={stats.processing ?? 0} color="text-blue-600" />
          <StatCard label="Completed"  value={stats.completed  ?? 0} color="text-green-600" />
          <StatCard label="Failed"     value={stats.failed     ?? 0} color="text-red-600" />
          <StatCard label="Retrying"   value={stats.retrying   ?? 0} color="text-orange-600" />
          <StatCard label="Dead"       value={stats.dead       ?? 0} color="text-gray-600 dark:text-gray-300 dark:text-gray-600" />
        </div>
      )}

      {/* Filters */}
      <div className="flex gap-3 mb-4">
        <input
          type="text"
          placeholder="Filter by status, type, ID, or ticket ID…"
          value={filter}
          onChange={(e) => { setFilter(e.target.value); setPage(1); }}
          className="flex-1 border border-gray-300 dark:border-gray-600 rounded-lg px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
        <button
          onClick={fetchAll}
          className="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700"
        >
          Refresh
        </button>
      </div>

      {/* Table */}
      <div className="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-900 border-b border-gray-100 dark:border-gray-700">
            <tr>
              <th className="text-left px-4 py-3 text-gray-500 dark:text-gray-400 dark:text-gray-500 font-medium">ID</th>
              <th className="text-left px-4 py-3 text-gray-500 dark:text-gray-400 dark:text-gray-500 font-medium">Type</th>
              <th className="text-left px-4 py-3 text-gray-500 dark:text-gray-400 dark:text-gray-500 font-medium">Status</th>
              <th className="text-left px-4 py-3 text-gray-500 dark:text-gray-400 dark:text-gray-500 font-medium">Ticket</th>
              <th className="text-left px-4 py-3 text-gray-500 dark:text-gray-400 dark:text-gray-500 font-medium">Retries</th>
              <th className="text-left px-4 py-3 text-gray-500 dark:text-gray-400 dark:text-gray-500 font-medium">Created</th>
              <th className="text-left px-4 py-3 text-gray-500 dark:text-gray-400 dark:text-gray-500 font-medium">Duration</th>
              <th className="text-left px-4 py-3 text-gray-500 dark:text-gray-400 dark:text-gray-500 font-medium">Error</th>
              <th className="px-4 py-3" />
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50">
            {loading ? (
              <tr>
                <td colSpan={9} className="text-center py-12 text-gray-400 dark:text-gray-500">
                  Loading…
                </td>
              </tr>
            ) : paged.length === 0 ? (
              <tr>
                <td colSpan={9} className="text-center py-12 text-gray-400 dark:text-gray-500">
                  No jobs found.
                </td>
              </tr>
            ) : (
              paged.map((job) => (
                <tr key={job.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                  <td className="px-4 py-3 font-mono text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500">#{job.id}</td>
                  <td className="px-4 py-3 text-gray-700 dark:text-gray-200">
                    {JOB_TYPE_LABELS[job.job_type] || job.job_type}
                  </td>
                  <td className="px-4 py-3">
                    <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${STATUS_COLORS[job.status] || 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 dark:text-gray-600'}`}>
                      {job.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 font-mono text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500">
                    {job.reference_id?.slice(0, 8) ?? '—'}
                  </td>
                  <td className="px-4 py-3 text-gray-600 dark:text-gray-300 dark:text-gray-600 text-center">
                    {job.retry_count ?? 0}
                  </td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400 dark:text-gray-500 text-xs">
                    {job.created_at ? new Date(job.created_at).toLocaleString() : '—'}
                  </td>
                  <td className="px-4 py-3 text-gray-500 dark:text-gray-400 dark:text-gray-500 text-xs">
                    {duration(job.started_at, job.completed_at)}
                  </td>
                  <td className="px-4 py-3 text-red-500 text-xs max-w-xs truncate">
                    {job.error_message || ''}
                  </td>
                  <td className="px-4 py-3">
                    {(job.status === 'FAILED' || job.status === 'DEAD') && (
                      <button
                        onClick={() => handleRetry(job.id)}
                        disabled={retrying[job.id]}
                        className="text-xs text-indigo-600 hover:text-indigo-800 disabled:opacity-50"
                      >
                        {retrying[job.id] ? 'Retrying…' : 'Retry'}
                      </button>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-between px-4 py-3 border-t border-gray-100 dark:border-gray-700">
            <span className="text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500">
              {filtered.length} jobs — page {page} of {totalPages}
            </span>
            <div className="flex gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-3 py-1 text-xs border rounded-lg disabled:opacity-40"
              >
                Prev
              </button>
              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="px-3 py-1 text-xs border rounded-lg disabled:opacity-40"
              >
                Next
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
