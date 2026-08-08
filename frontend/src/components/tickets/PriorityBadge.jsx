const PRIORITY_STYLES = {
  LOW: 'bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 dark:text-gray-500',
  MEDIUM: 'bg-blue-100 text-blue-700',
  HIGH: 'bg-orange-100 text-orange-700',
  URGENT: 'bg-red-100 text-red-700',
}

function PriorityBadge({ priority }) {
  return (
    <span
      className={`inline-flex px-2.5 py-0.5 rounded-full text-xs font-medium ${
        PRIORITY_STYLES[priority] ?? 'bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 dark:text-gray-500'
      }`}
    >
      {priority}
    </span>
  )
}

export default PriorityBadge
