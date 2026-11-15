import { useCallback, useEffect, useRef, useState } from 'react';
import { AppState, Linking } from 'react-native';
import * as Location from 'expo-location';

export type Position = { latitude: number; longitude: number; timestamp: number };
export type LocationAccess = 'loading' | 'explain' | 'granted' | 'denied' | 'blocked';

export default function useDeviceLocation() {
  const [access, setAccess] = useState<LocationAccess>('loading');
  const [position, setPosition] = useState<Position | null>(null);
  const [active, setActive] = useState(AppState.currentState === 'active');
  const [message, setMessage] = useState('');
  const [requesting, setRequesting] = useState(false);
  const mounted = useRef(true);
  const lock = useRef(false);
  const watcher = useRef<Location.LocationSubscription | null>(null);

  const refresh = useCallback(async () => {
    try {
      const permission = await Location.getForegroundPermissionsAsync();
      if (!mounted.current) return;
      setAccess(
        permission.granted
          ? 'granted'
          : permission.status === 'undetermined'
            ? 'explain'
            : permission.canAskAgain
              ? 'denied'
              : 'blocked',
      );
    } catch {
      if (mounted.current) {
        setAccess('denied');
        setMessage('Location is unavailable. You can still use the map.');
      }
    }
  }, []);

  useEffect(() => {
    mounted.current = true;
    void refresh();
    const subscription = AppState.addEventListener('change', (state) => {
      const foreground = state === 'active';
      if (!foreground) {
        watcher.current?.remove();
        watcher.current = null;
        setPosition(null);
      } else setAccess('loading');
      setActive(foreground);
      if (foreground) void refresh();
    });
    return () => {
      mounted.current = false;
      watcher.current?.remove();
      subscription.remove();
    };
  }, [refresh]);

  useEffect(() => {
    if (access !== 'granted' || !active) return;
    let cancelled = false;
    let localWatcher: Location.LocationSubscription | null = null;
    setMessage('Finding your location…');
    (async () => {
      try {
        if (!(await Location.hasServicesEnabledAsync())) {
          if (!cancelled) setMessage('Device location services are off. Enable them in Settings.');
          return;
        }
        if (cancelled) return;
        localWatcher = await Location.watchPositionAsync(
          { accuracy: Location.Accuracy.Balanced, timeInterval: 5000, distanceInterval: 10 },
          (location) => {
            if (cancelled) return;
            setPosition({
              latitude: location.coords.latitude,
              longitude: location.coords.longitude,
              timestamp: location.timestamp,
            });
            setMessage('');
          },
          () => {
            if (!cancelled) setMessage('Could not get a GPS fix. Check device location settings.');
          },
        );
        if (cancelled) localWatcher.remove();
        else watcher.current = localWatcher;
      } catch {
        if (!cancelled)
          setMessage('Could not get your location. Check permission and device location settings.');
      }
    })();
    return () => {
      cancelled = true;
      localWatcher?.remove();
      if (watcher.current === localWatcher) watcher.current = null;
    };
  }, [access, active]);

  async function requestAccess() {
    if (lock.current) return;
    lock.current = true;
    setRequesting(true);
    setMessage('');
    try {
      if (access === 'blocked') {
        await Linking.openSettings();
        return;
      }
      const permission = await Location.requestForegroundPermissionsAsync();
      if (mounted.current)
        setAccess(permission.granted ? 'granted' : permission.canAskAgain ? 'denied' : 'blocked');
    } catch {
      if (mounted.current) setMessage('Could not request location access. Try again in Settings.');
    } finally {
      lock.current = false;
      if (mounted.current) setRequesting(false);
    }
  }

  return { access, position, active, message, requesting, requestAccess };
}
