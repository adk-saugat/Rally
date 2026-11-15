import { useCallback, useEffect, useRef, useState } from 'react';
import { AppState } from 'react-native';
import { API_URL } from '../config';
import { openLocationSocket } from '../services/locationSocket';
import { useSession } from '../session/SessionProvider';
import { parseMemberLocation, websocketURL, type MemberLocation } from '../utils/locations';
import type { Position } from './useDeviceLocation';

export default function useLiveLocations(
  groupID: string | undefined,
  position: Position | null,
  canShare: boolean,
) {
  const { session, expire } = useSession();
  const [sharing, setSharing] = useState(false);
  const [status, setStatus] = useState('Not sharing');
  const [points, setPoints] = useState<Record<string, MemberLocation>>({});
  const connection = useRef<WebSocket | null>(null);
  const requestedGroup = useRef<string | null>(null);
  const latest = useRef(position);
  latest.current = position;

  const stop = useCallback(() => {
    // Stop synchronously so no pending GPS callback can send another point.
    requestedGroup.current = null;
    const socket = connection.current;
    connection.current = null;
    socket?.close(1000, 'sharing stopped');
    setSharing(false);
    setStatus('Not sharing');
    setPoints({});
  }, []);

  useEffect(() => {
    stop();
  }, [groupID, stop]);
  useEffect(() => {
    if (!canShare) stop();
  }, [canShare, stop]);
  useEffect(() => {
    const subscription = AppState.addEventListener('change', (state) => {
      if (state !== 'active') stop();
    });
    return () => subscription.remove();
  }, [stop]);

  useEffect(() => {
    if (!sharing || !canShare || !session || !groupID || requestedGroup.current !== groupID) return;
    let disposed = false;
    let attempts = 0;
    let reconnect: ReturnType<typeof setTimeout> | undefined;
    let handshake: ReturnType<typeof setTimeout> | undefined;
    let sentAt = 0;
    let previousTimestamp = 0;

    function publish() {
      const current = latest.current;
      const socket = connection.current;
      if (
        !current ||
        disposed ||
        socket?.readyState !== WebSocket.OPEN ||
        current.timestamp <= previousTimestamp ||
        Date.now() - current.timestamp > 30000 ||
        Date.now() - sentAt < 4500
      )
        return;
      socket.send(
        JSON.stringify({
          latitude: current.latitude,
          longitude: current.longitude,
          timestamp_unix_ms: current.timestamp,
        }),
      );
      previousTimestamp = current.timestamp;
      sentAt = Date.now();
    }

    function connect() {
      if (disposed) return;
      setStatus(attempts ? 'Reconnecting…' : 'Connecting…');
      // React Native supports native handshake headers. No credential in the URL.
      const socket = openLocationSocket(websocketURL(API_URL, groupID!), session!.token);
      connection.current = socket;
      handshake = setTimeout(() => {
        if (socket.readyState === WebSocket.CONNECTING) socket.close();
      }, 12000);
      socket.onopen = () => {
        clearTimeout(handshake);
        if (disposed || connection.current !== socket) {
          socket.close();
          return;
        }
        setStatus('Sharing while Rally is open');
        sentAt = 0;
        previousTimestamp = 0;
        publish();
      };
      socket.onmessage = (event) => {
        if (disposed || connection.current !== socket) return;
        try {
          const point = parseMemberLocation(JSON.parse(String(event.data)), groupID!);
          if (!point || point.user_id === session!.userID) return;
          setPoints((current) => {
            if (current[point.user_id]?.timestamp_unix_ms >= point.timestamp_unix_ms)
              return current;
            return { ...current, [point.user_id]: point };
          });
        } catch {
          /* Ignore malformed frames without dropping a healthy stream. */
        }
      };
      socket.onerror = () => {
        if (!disposed) setStatus('Connection interrupted');
      };
      socket.onclose = (event) => {
        clearTimeout(handshake);
        if (disposed || connection.current !== socket) return;
        connection.current = null;
        if (Date.now() >= session!.expiresAt) {
          expire();
          return;
        }
        if (event.code === 1008) {
          setSharing(false);
          setStatus('Sharing access denied. Reopen your group or log in again.');
          return;
        }
        attempts++;
        if (attempts > 6) {
          setSharing(false);
          setStatus('Could not reconnect. Tap Start sharing to retry.');
          return;
        }
        setStatus('Reconnecting…');
        reconnect = setTimeout(
          connect,
          Math.min(1000 * 2 ** (attempts - 1), 15000) + Math.random() * 500,
        );
      };
    }
    connect();
    // Send only the newest fix; never queue an offline backlog.
    const interval = setInterval(publish, 1000);
    return () => {
      disposed = true;
      clearInterval(interval);
      clearTimeout(reconnect);
      clearTimeout(handshake);
      const socket = connection.current;
      connection.current = null;
      socket?.close(1000, 'stream closed');
    };
  }, [sharing, canShare, session, groupID, expire]);

  function start() {
    if (!canShare || !groupID || !position) return;
    requestedGroup.current = groupID;
    setSharing(true);
  }
  return { sharing, status, points, start, stop };
}
