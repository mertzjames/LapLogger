import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
} from 'react';
import type { ReactNode } from 'react';
import type { User } from '../types';
import { getMe } from '../services/api';

interface AuthState {
  user: User | null;
  loading: boolean;
  login: () => void;
  logout: () => void;
}

const AuthContext = createContext<AuthState | null>(null);

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check for token in URL (OAuth callback redirect)
    const params = new URLSearchParams(window.location.search);
    const urlToken = params.get('token');
    if (urlToken) {
      sessionStorage.setItem('token', urlToken);
      // Clean URL without reloading
      window.history.replaceState({}, '', window.location.pathname);
    }

    // Fetch user if we have a token
    const token = sessionStorage.getItem('token');
    if (token) {
      getMe()
        .then(setUser)
        .catch(() => {
          sessionStorage.removeItem('token');
          setUser(null);
        })
        .finally(() => setLoading(false));
    } else {
      // No token — schedule loading=false for next microtask to avoid
      // synchronous setState inside the effect body.
      Promise.resolve().then(() => setLoading(false));
    }
  }, []);

  const login = useCallback(() => {
    window.location.href = '/api/auth/google';
  }, []);

  const logout = useCallback(() => {
    sessionStorage.removeItem('token');
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider value={{ user, loading, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}
