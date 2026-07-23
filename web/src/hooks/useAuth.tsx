import { createContext, useContext, useState, useEffect, useCallback } from 'react';
import type { User } from '../types';

interface AuthContextType {
  user: User | null;
  perms: string[];
  loading: boolean;
  login: (token: string, userData: User, permList: string[]) => void;
  logout: () => void;
  hasPermission: (module: string, action: string) => boolean;
}

const AuthContext = createContext<AuthContextType>({
  user: null,
  perms: [],
  loading: true,
  login: () => {},
  logout: () => {},
  hasPermission: () => false,
});

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [perms, setPerms] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('token');
    const userStr = localStorage.getItem('user');
    const permsStr = localStorage.getItem('perms');
    if (token && userStr && permsStr) {
      try {
        setUser(JSON.parse(userStr));
        setPerms(JSON.parse(permsStr));
      } catch {
        localStorage.clear();
      }
    }
    setLoading(false);
  }, []);

  const login = useCallback((token: string, userData: User, permList: string[]) => {
    localStorage.setItem('token', token);
    localStorage.setItem('user', JSON.stringify(userData));
    localStorage.setItem('perms', JSON.stringify(permList));
    setUser(userData);
    setPerms(permList);
    window.location.href = '/';
  }, []);

  const logout = useCallback(() => {
    localStorage.clear();
    setUser(null);
    setPerms([]);
    window.location.href = '/login';
  }, []);

  const hasPermission = useCallback((module: string, action: string) => {
    return perms.includes(`${module}:${action}`);
  }, [perms]);

  return (
    <AuthContext.Provider value={{ user, perms, loading, login, logout, hasPermission }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
