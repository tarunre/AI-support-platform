// Single source of truth for all category constants.
// Import from here instead of defining locally in each page.

/** Ticket categories — used in CreateTicket, EditTicket, filters */
export const TICKET_CATEGORIES = [
  { label: 'General',                value: 'GENERAL' },
  { label: 'Technical Issue',        value: 'TECHNICAL_ISSUE' },
  { label: 'Engineering / Tech Team',value: 'ENGINEERING' },
  { label: 'Payment',                value: 'PAYMENT' },
  { label: 'Authentication',         value: 'AUTHENTICATION' },
  { label: 'Refund',                 value: 'REFUND' },
  { label: 'Account',                value: 'ACCOUNT' },
  { label: 'Subscription',           value: 'SUBSCRIPTION' },
]

/** Knowledge base categories — used in KnowledgeBase page */
export const KB_CATEGORIES = [
  'FAQ',
  'Refund Policy',
  'Shipping Policy',
  'Subscription Policy',
  'Account Policy',
  'Payment Policy',
  'General Documentation',
  'Engineering Team',
]

/** Ticket priorities — used in CreateTicket, EditTicket */
export const TICKET_PRIORITIES = [
  { label: 'Low',    value: 'LOW' },
  { label: 'Medium', value: 'MEDIUM' },
  { label: 'High',   value: 'HIGH' },
  { label: 'Urgent', value: 'URGENT' },
]
