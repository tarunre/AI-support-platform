/**
 * SLABadge — displays the current SLA status with colour coding.
 * Props: status (string) — 'ON_TRACK' | 'AT_RISK' | 'BREACHED' | 'COMPLETED' | ''
 */
const STATUS_CONFIG = {
  ON_TRACK:  { label: 'On Track',  className: 'bg-green-100 text-green-700' },
  AT_RISK:   { label: 'At Risk',   className: 'bg-yellow-100 text-yellow-700' },
  BREACHED:  { label: 'Breached',  className: 'bg-red-100 text-red-700' },
  COMPLETED: { label: 'Completed', className: 'bg-blue-100 text-blue-700' },
}

export default function SLABadge({ status }) {
  if (!status) return null
  const cfg = STATUS_CONFIG[status] || { label: status, className: 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 dark:text-gray-600' }
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold ${cfg.className}`}>
      {cfg.label}
    </span>
  )
}
