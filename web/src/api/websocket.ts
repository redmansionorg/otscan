import { useEffect, useRef, useCallback, useState } from 'react';

export interface WSEvent {
  type: string;
  timestamp: number;
  data: unknown;
}

export type WSEventHandler = (event: WSEvent) => void;

export function useWebSocket(onEvent?: WSEventHandler) {
  const wsRef = useRef<WebSocket | null>(null);
  const [connected, setConnected] = useState(false);
  const handlersRef = useRef<WSEventHandler | undefined>(onEvent);
  handlersRef.current = onEvent;

  const connect = useCallback(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const url = `${protocol}//${window.location.host}/api/v1/ws`;

    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      setConnected(true);
    };

    ws.onmessage = (evt) => {
      try {
        const event: WSEvent = JSON.parse(evt.data);
        handlersRef.current?.(event);
      } catch {
        // ignore parse errors
      }
    };

    ws.onclose = () => {
      setConnected(false);
      // Reconnect after 3 seconds
      setTimeout(connect, 3000);
    };

    ws.onerror = () => {
      ws.close();
    };
  }, []);

  useEffect(() => {
    connect();
    return () => {
      wsRef.current?.close();
    };
  }, [connect]);

  return { connected };
}
