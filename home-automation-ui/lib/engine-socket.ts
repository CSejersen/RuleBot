let ws: WebSocket | null = null;

export function getEngineSocket() {
  if (!ws || ws.readyState === WebSocket.CLOSED) {
    ws = new WebSocket("ws://localhost:8080/ws");
  }
  return ws;
}

export function engineWSsendMessage(message: any) {
  const socket = getEngineSocket();
  if (socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify(message));
  } else {
    socket.addEventListener("open", () => socket.send(JSON.stringify(message)), { once: true });
  }
}
