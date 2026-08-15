import {
  createContext,
  use,
  useCallback,
  useEffect,
  useMemo,
  useState,
  type PropsWithChildren,
} from 'react';

import { loadSessionToken, saveSessionToken } from '@/src/api/storage';

export type AuthStatus = 'loading' | 'signed-in' | 'signed-out';

interface AuthContextValue {
  status: AuthStatus;
  token: string | null;
  signIn: (token: string) => Promise<void>;
  signOut: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

/**
 * Session provider for the whole app. Restores the persisted token on
 * startup so the splash screen can stay up until the auth state is known.
 */
export function AuthProvider({ children }: PropsWithChildren) {
  const [token, setToken] = useState<string | null>(null);
  const [hydrated, setHydrated] = useState(false);

  useEffect(() => {
    let cancelled = false;
    loadSessionToken()
      .then((stored) => {
        if (cancelled) {
          return;
        }
        setToken(stored);
        setHydrated(true);
      })
      .catch(() => {
        if (!cancelled) {
          setHydrated(true);
        }
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const signIn = useCallback(async (nextToken: string) => {
    await saveSessionToken(nextToken);
    setToken(nextToken);
  }, []);

  const signOut = useCallback(async () => {
    await saveSessionToken(null);
    setToken(null);
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      status: hydrated ? (token ? 'signed-in' : 'signed-out') : 'loading',
      token,
      signIn,
      signOut,
    }),
    [hydrated, token, signIn, signOut],
  );

  return <AuthContext value={value}>{children}</AuthContext>;
}

export function useAuth(): AuthContextValue {
  const value = use(AuthContext);
  if (value === null) {
    throw new Error('useAuth must be used inside an <AuthProvider />');
  }
  return value;
}
