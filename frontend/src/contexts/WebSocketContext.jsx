import { createContext, useContext, useEffect, useRef, useState } from 'react';
import wsService from '../services/websocketService';

const WebSocketContext = createContext(null);

export function WebSocketProvider({ children }) {
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    wsService.connect();

    // Track connection state via heartbeat events from the server (or status updates)
    const unsub = wsService.on('*', () => setConnected(wsService.isConnected()));

    // Poll readyState until connection is established
    const interval = setInterval(() => {
      setConnected(wsService.isConnected());
    }, 2000);

    return () => {
      unsub();
      clearInterval(interval);
      wsService.disconnect();
    };
  }, []);

  return (
    <WebSocketContext.Provider value={{ connected, wsService }}>
      {children}
    </WebSocketContext.Provider>
  );
}

/** Hook to consume the WebSocket context. */
export function useWebSocket() {
  const ctx = useContext(WebSocketContext);
  if (!ctx) throw new Error('useWebSocket must be used inside WebSocketProvider');
  return ctx;
}
