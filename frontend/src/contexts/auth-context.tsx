'use client';

import { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';

interface User {
  id: string;
  email: string;
  name: string;
  role: string;
  tenantId: string;
  tenantSlug: string;
}

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (accessToken: string, refreshToken: string, tenantSlug: string) => void;
  logout: () => void;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();

  const parseToken = useCallback((token: string, tenantSlug: string): User | null => {
    try {
      const payload = token.split('.')[1];
      const decoded = JSON.parse(atob(payload));
      return {
        id: decoded.user_id,
        email: decoded.email,
        name: decoded.name || decoded.email.split('@')[0],
        role: decoded.role,
        tenantId: decoded.tenant_id,
        tenantSlug: tenantSlug,
      };
    } catch {
      return null;
    }
  }, []);

  useEffect(() => {
    const initializeAuth = async () => {
      const token = localStorage.getItem('accessToken');
      const tenantSlug = localStorage.getItem('tenantSlug') || '';
      if (token) {
        // まずトークンから基本情報をセット
        const userData = parseToken(token, tenantSlug);
        setUser(userData);

        // APIから最新のユーザー情報を取得
        try {
          const { getMe } = await import('@/api/public/user/user');
          const response = await getMe();
          if (response) {
            setUser({
              id: response.id ?? '',
              email: response.email ?? '',
              name: response.name ?? '',
              role: response.role ?? '',
              tenantId: response.tenant_id ?? '',
              tenantSlug: tenantSlug,
            });
          }
        } catch (error) {
          console.error('Failed to fetch user info:', error);
        }
      }
      setIsLoading(false);
    };

    initializeAuth();
  }, [parseToken]);

  const login = useCallback((accessToken: string, refreshToken: string, tenantSlug: string) => {
    localStorage.setItem('accessToken', accessToken);
    localStorage.setItem('refreshToken', refreshToken);
    localStorage.setItem('tenantSlug', tenantSlug);
    const userData = parseToken(accessToken, tenantSlug);
    setUser(userData);
  }, [parseToken]);

  const logout = useCallback(() => {
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('tenantSlug');
    setUser(null);
    router.push('/login');
  }, [router]);

  const refreshUser = useCallback(async () => {
    const { getMe } = await import('@/api/public/user/user');
    try {
      const response = await getMe();
      const tenantSlug = localStorage.getItem('tenantSlug') || '';
      if (response) {
        setUser({
          id: response.id ?? '',
          email: response.email ?? '',
          name: response.name ?? '',
          role: response.role ?? '',
          tenantId: response.tenant_id ?? '',
          tenantSlug: tenantSlug,
        });
      }
    } catch (error) {
      console.error('Failed to refresh user:', error);
    }
  }, []);

  return (
    <AuthContext.Provider
      value={{
        user,
        isLoading,
        isAuthenticated: !!user,
        login,
        logout,
        refreshUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
