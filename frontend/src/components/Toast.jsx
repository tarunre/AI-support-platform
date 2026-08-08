import { useState, useEffect } from 'react'

export function useToast() {
  const [toast, setToast] = useState(null)

  const showToast = (message, type = 'success') => {
    setToast({ message, type })
  }

  useEffect(() => {
    if (!toast) return
    const t = setTimeout(() => setToast(null), 3500)
    return () => clearTimeout(t)
  }, [toast])

  return { toast, showToast }
}

function Toast({ toast }) {
  if (!toast) return null
  return (
    <div
      className={`fixed top-5 right-5 z-50 px-4 py-3 rounded-lg shadow-md text-sm font-medium border ${
        toast.type === 'success'
          ? 'bg-green-50 border-green-200 text-green-700'
          : 'bg-red-50 border-red-200 text-red-600'
      }`}
    >
      {toast.message}
    </div>
  )
}

export default Toast
