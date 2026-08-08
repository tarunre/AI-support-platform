import { useEffect, useState } from 'react';
import { useWebSocket } from '../contexts/WebSocketContext';

const EVENT_LABELS = {
  'ticket.ai.completed':   '🤖 AI Analysis Complete',
  'ticket.reply.generated':'💬 Reply Generated',
  'ticket.reply.failed':   '❌ Reply Generation Failed',
  'ticket.updated':        '🔄 Ticket Updated',
  'job.completed':         '✅ Job Completed',
  'job.failed':            '❌ Job Failed',
};

let toastId = 0;

export default function RealtimeToast() {
  const { wsService } = useWebSocket();
  const [toasts, setToasts] = useState([]);

  useEffect(() => {
    const unsub = wsService.on('*', (event) => {
      const label = EVENT_LABELS[event.type];
      if (!label) return;

      const id = ++toastId;
      setToasts((prev) => [...prev, { id, label, type: event.type, data: event }]);

      // Auto-dismiss after 5 s
      setTimeout(() => {
        setToasts((prev) => prev.filter((t) => t.id !== id));
      }, 5000);
    });

    return unsub;
  }, [wsService]);

  if (toasts.length === 0) return null;

  return (
    <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-2">
      {toasts.map((toast) => (
        <div
          key={toast.id}
          className="flex items-center gap-3 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-lg shadow-lg px-4 py-3 min-w-[280px] animate-slide-in"
        >
          <div className="flex-1">
            <p className="text-sm font-medium text-gray-800 dark:text-gray-100">{toast.label}</p>
            {toast.data.ticket_id && (
              <p className="text-xs text-gray-500 dark:text-gray-400 dark:text-gray-500 mt-0.5">
                Ticket: {toast.data.ticket_id.slice(0, 8)}…
              </p>
            )}
          </div>
          <button
            onClick={() =>
              setToasts((prev) => prev.filter((t) => t.id !== toast.id))
            }
            className="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:text-gray-300 dark:text-gray-600 text-lg leading-none"
          >
            ×
          </button>
        </div>
      ))}
    </div>
  );
}
