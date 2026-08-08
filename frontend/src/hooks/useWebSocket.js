import { useEffect } from 'react'
import { useWebSocket as useWSContext } from '../contexts/WebSocketContext'

/**
 * useWebSocket(callback)
 * Subscribes to all incoming WebSocket messages and calls callback(parsedMessage).
 * The subscription is cleaned up on unmount.
 */
export default function useWebSocket(callback) {
  const { wsService } = useWSContext()

  useEffect(() => {
    if (!wsService || !callback) return
    const unsub = wsService.on('*', callback)
    return unsub
  }, [wsService, callback])
}
