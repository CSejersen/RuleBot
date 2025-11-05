import React from "react";

type ConnectionState = "disconnected" | "connecting" | "connected" | "error";

interface SocketConfig {
  url: string;
  maxQueueSize: number;
  reconnectDelayMs: number;
  maxReconnectDelayMs: number;
}

const DEFAULT_CONFIG: SocketConfig = {
  url: process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080/ws",
  maxQueueSize: 100,
  reconnectDelayMs: 1000,
  maxReconnectDelayMs: 30000,
};

class EngineSocketManager {
  private ws: WebSocket | null = null;
  private state: ConnectionState = "disconnected";
  private reconnectTimeoutId: NodeJS.Timeout | null = null;
  private reconnectAttempts = 0;
  private messageQueue: any[] = [];
  private messageListeners = new Set<(event: MessageEvent) => void>();
  private stateListeners = new Set<(state: ConnectionState) => void>();
  private config: SocketConfig;

  constructor(config: Partial<SocketConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
  }

  getState(): ConnectionState {
    return this.state;
  }

  private setState(newState: ConnectionState) {
    if (this.state !== newState) {
      this.state = newState;
      this.stateListeners.forEach((listener) => listener(newState));
    }
  }

  onStateChange(listener: (state: ConnectionState) => void): () => void {
    this.stateListeners.add(listener);
    // Immediately call with current state
    listener(this.state);
    // Return unsubscribe function
    return () => {
      this.stateListeners.delete(listener);
    };
  }

  connect(): void {
    // Prevent multiple simultaneous connection attempts
    if (this.ws && (this.ws.readyState === WebSocket.CONNECTING || this.ws.readyState === WebSocket.OPEN)) {
      return;
    }

    // Clear any existing reconnect timeout
    if (this.reconnectTimeoutId) {
      clearTimeout(this.reconnectTimeoutId);
      this.reconnectTimeoutId = null;
    }

    this.setState("connecting");

    try {
      this.ws = new WebSocket(this.config.url);

      this.ws.onopen = () => {
        this.setState("connected");
        this.reconnectAttempts = 0;
        
        // Send all queued messages
        this.flushMessageQueue();
        
        console.log("[EngineSocket] Connected to engine");
      };

      this.ws.onmessage = (event: MessageEvent) => {
        // Notify all message listeners
        this.messageListeners.forEach((listener) => {
          try {
            listener(event);
          } catch (err) {
            console.error("[EngineSocket] Error in message listener:", err);
          }
        });
      };

      this.ws.onerror = (error) => {
        this.setState("error");
        console.error("[EngineSocket] WebSocket error:", error);
      };

      this.ws.onclose = (event) => {
        this.setState("disconnected");
        this.ws = null;
        
        // Don't reconnect if it was a normal closure (code 1000)
        if (event.code !== 1000) {
          this.scheduleReconnect();
        } else {
          console.log("[EngineSocket] Connection closed normally");
        }
      };
    } catch (error) {
      this.setState("error");
      console.error("[EngineSocket] Failed to create WebSocket:", error);
      this.scheduleReconnect();
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimeoutId) {
      return; // Already scheduled
    }

    // Exponential backoff: start at 1s, max at 30s
    const delay = Math.min(
      this.config.reconnectDelayMs * Math.pow(2, this.reconnectAttempts),
      this.config.maxReconnectDelayMs
    );

    this.reconnectAttempts++;
    console.log(`[EngineSocket] Scheduling reconnect attempt ${this.reconnectAttempts} in ${delay}ms`);

    this.reconnectTimeoutId = setTimeout(() => {
      this.reconnectTimeoutId = null;
      this.connect();
    }, delay);
  }

  private flushMessageQueue(): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      return;
    }

    while (this.messageQueue.length > 0 && this.ws.readyState === WebSocket.OPEN) {
      const message = this.messageQueue.shift();
      try {
        this.ws.send(JSON.stringify(message));
      } catch (err) {
        console.error("[EngineSocket] Error sending queued message:", err);
        // Put message back at front of queue if send failed
        this.messageQueue.unshift(message);
        break;
      }
    }
  }

  sendMessage(message: any): void {
    // Ensure connection is established
    if (!this.ws || this.ws.readyState === WebSocket.CLOSED) {
      this.connect();
    }

    // If connected, send immediately
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      try {
        this.ws.send(JSON.stringify(message));
        return;
      } catch (err) {
        console.error("[EngineSocket] Error sending message:", err);
      }
    }

    // Otherwise, queue the message
    if (this.messageQueue.length >= this.config.maxQueueSize) {
      console.warn("[EngineSocket] Message queue full, dropping oldest message");
      this.messageQueue.shift();
    }
    this.messageQueue.push(message);

    // If connecting, wait for connection
    if (this.ws && this.ws.readyState === WebSocket.CONNECTING) {
      const onceOpen = () => {
        this.flushMessageQueue();
        this.ws?.removeEventListener("open", onceOpen);
      };
      this.ws.addEventListener("open", onceOpen, { once: true });
    }
  }

  onMessage(listener: (event: MessageEvent) => void): () => void {
    this.messageListeners.add(listener);
    
    // Ensure connection is established
    if (!this.ws || this.ws.readyState === WebSocket.CLOSED) {
      this.connect();
    }
    
    // Return unsubscribe function
    return () => {
      this.messageListeners.delete(listener);
    };
  }

  getSocket(): WebSocket | null {
    // Ensure connection is established
    if (!this.ws || this.ws.readyState === WebSocket.CLOSED) {
      this.connect();
    }
    return this.ws;
  }

  disconnect(): void {
    if (this.reconnectTimeoutId) {
      clearTimeout(this.reconnectTimeoutId);
      this.reconnectTimeoutId = null;
    }

    if (this.ws) {
      this.ws.close(1000); // Normal closure
      this.ws = null;
    }

    this.setState("disconnected");
    this.messageQueue = [];
    this.reconnectAttempts = 0;
  }
}

// Singleton instance
const socketManager = new EngineSocketManager();

// Initialize connection on module load (client-side only)
if (typeof window !== "undefined") {
  socketManager.connect();
}

// API exports
export function getConnectionState(): ConnectionState {
  return socketManager.getState();
}

export function sendSocketMessage(message: any): void {
  socketManager.sendMessage(message);
}

export function onConnectionStateChange(listener: (state: ConnectionState) => void): () => void {
  return socketManager.onStateChange(listener);
}

export function onSocketMessage(listener: (event: MessageEvent) => void): () => void {
  return socketManager.onMessage(listener);
}

export function connectSocket(): void {
  socketManager.connect();
}

export function disconnectSocket(): void {
  socketManager.disconnect();
}

// React hook for connection state 
export function useConnectionStatus(): ConnectionState {
  if (typeof window === "undefined") {
    return "disconnected";
  }

  const [state, setState] = React.useState<ConnectionState>(socketManager.getState());

  React.useEffect(() => {
    return socketManager.onStateChange(setState);
  }, []);

  return state;
}
