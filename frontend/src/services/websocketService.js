/**
 * WebSocket service — manages a single persistent connection to the API.
 * Automatically reconnects with exponential backoff on unexpected disconnects.
 */

const WS_BASE = import.meta.env.VITE_WS_URL || 'ws://localhost:8080';
const MAX_RETRIES = 10;
const BASE_DELAY_MS = 1000;

class WebSocketService {
  constructor() {
    this.ws = null;
    this.listeners = new Map(); // eventType → Set<callback>
    this.retryCount = 0;
    this.retryTimer = null;
    this.intentionalClose = false;
  }

  /** Connect using the stored JWT access token. */
  connect() {
    const token = localStorage.getItem('access_token');
    if (!token) return;

    this.intentionalClose = false;
    this._openSocket(token);
  }

  _openSocket(token) {
    if (this.ws && this.ws.readyState < 2) return; // already open/connecting

    this.ws = new WebSocket(`${WS_BASE}/api/v1/ws?token=${token}`);

    this.ws.onopen = () => {
      this.retryCount = 0;
      clearTimeout(this.retryTimer);
    };

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        const eventType = data.type || 'message';
        const handlers = this.listeners.get(eventType) || new Set();
        const wildcards = this.listeners.get('*') || new Set();
        [...handlers, ...wildcards].forEach((cb) => cb(data));
      } catch {
        // ignore malformed frames
      }
    };

    this.ws.onclose = (ev) => {
      if (!this.intentionalClose) this._scheduleReconnect();
    };

    this.ws.onerror = () => {
      this.ws?.close();
    };
  }

  _scheduleReconnect() {
    if (this.retryCount >= MAX_RETRIES) return;
    const delay = BASE_DELAY_MS * Math.pow(2, this.retryCount);
    this.retryCount += 1;
    this.retryTimer = setTimeout(() => {
      const token = localStorage.getItem('access_token');
      if (token) this._openSocket(token);
    }, delay);
  }

  /** Cleanly close the connection. */
  disconnect() {
    this.intentionalClose = true;
    clearTimeout(this.retryTimer);
    this.ws?.close();
    this.ws = null;
  }

  /**
   * Subscribe to a specific event type (or '*' for all events).
   * Returns an unsubscribe function.
   */
  on(eventType, callback) {
    if (!this.listeners.has(eventType)) {
      this.listeners.set(eventType, new Set());
    }
    this.listeners.get(eventType).add(callback);
    return () => this.listeners.get(eventType)?.delete(callback);
  }

  isConnected() {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

// Singleton — imported once per app
const wsService = new WebSocketService();
export default wsService;
