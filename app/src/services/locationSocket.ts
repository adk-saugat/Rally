// React Native's WebSocket implementation accepts handshake headers, although
// the DOM WebSocket type used by TypeScript exposes only the first two arguments.
type NativeSocketConstructor = new (
  url: string,
  protocols?: string | string[],
  options?: { headers: Record<string, string> },
) => WebSocket;

export function openLocationSocket(url: string, token: string): WebSocket {
  const NativeWebSocket = WebSocket as unknown as NativeSocketConstructor;
  return new NativeWebSocket(url, undefined, { headers: { Authorization: `Bearer ${token}` } });
}
