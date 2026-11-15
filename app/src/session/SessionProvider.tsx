import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from 'react';
import { AppState } from 'react-native';
import * as SecureStore from 'expo-secure-store';
import { sessionFromToken, type Session } from '../utils/session';

const STORAGE_KEY = 'rally.session.v1';
type SessionContextValue = {
  session: Session | null;
  restoring: boolean;
  notice: string;
  signIn: (token: string) => Promise<void>;
  signOut: () => Promise<void>;
  expire: () => void;
};
const SessionContext = createContext<SessionContextValue | null>(null);

export function SessionProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<Session | null>(null);
  const [restoring, setRestoring] = useState(true);
  const [notice, setNotice] = useState('');

  useEffect(() => {
    let active = true;
    (async () => {
      try {
        const token = await SecureStore.getItemAsync(STORAGE_KEY);
        if (token) {
          const restored = sessionFromToken(token);
          if (active) setSession(restored);
        }
      } catch {
        await SecureStore.deleteItemAsync(STORAGE_KEY).catch(() => {});
        if (active) setNotice('Please log in again to continue.');
      } finally {
        if (active) setRestoring(false);
      }
    })();
    return () => {
      active = false;
    };
  }, []);

  const signIn = useCallback(async (token: string) => {
    const next = sessionFromToken(token);
    try {
      await SecureStore.setItemAsync(STORAGE_KEY, token, {
        keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
      });
    } catch {
      throw new Error('Could not save your login on this device. Please try again.');
    }
    setNotice('');
    setSession(next);
  }, []);

  const signOut = useCallback(async () => {
    // Do not report successful logout if the saved credential could not be removed.
    await SecureStore.deleteItemAsync(STORAGE_KEY);
    setSession(null);
    setNotice('');
  }, []);

  const expire = useCallback(() => {
    setSession(null);
    setNotice('Your session expired. Please log in again.');
    void SecureStore.deleteItemAsync(STORAGE_KEY).catch(() => {});
  }, []);

  useEffect(() => {
    if (!session) return;
    const check = () => {
      if (Date.now() >= session.expiresAt) expire();
    };
    const timeout = setTimeout(
      check,
      Math.min(Math.max(session.expiresAt - Date.now(), 0), 2_147_483_647),
    );
    const subscription = AppState.addEventListener('change', (state) => {
      if (state === 'active') check();
    });
    return () => {
      clearTimeout(timeout);
      subscription.remove();
    };
  }, [session, expire]);

  return (
    <SessionContext.Provider value={{ session, restoring, notice, signIn, signOut, expire }}>
      {children}
    </SessionContext.Provider>
  );
}

export function useSession() {
  const context = useContext(SessionContext);
  if (!context) throw new Error('SessionProvider is required.');
  return context;
}
