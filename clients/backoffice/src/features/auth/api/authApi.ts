import { api, setAccessToken } from '@/lib/api';

export interface User {
  id: string;
  username: string;
  email: string;
  roles: string[];
  type: string;
  status: string;
  createdAt: string;
  updatedAt: string;
}

interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

const REFRESH_KEY = 'backoffice_refresh_token';

export function getStoredRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_KEY);
}

export function clearStoredRefreshToken() {
  localStorage.removeItem(REFRESH_KEY);
}

export const authApi = {
  login: async (username: string, password: string) => {
    let res: AuthResponse;
    try {
      res = await api<AuthResponse>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ username, password }),
      });
    } catch (err) {
      throw err;
    }

    setAccessToken(res.access_token);
    localStorage.setItem(REFRESH_KEY, res.refresh_token);

    let user: User;
    try {
      user = await api<User>('/auth/me');
    } catch (err) {
      console.error('[authApi] /auth/me failed', err);
      throw err;
    }
    return { user };
  },

  logout: async () => {
    const refreshToken = getStoredRefreshToken();
    if (refreshToken) {
      try {
        await api('/auth/logout', {
          method: 'POST',
          body: JSON.stringify({ refresh_token: refreshToken }),
        });
      } catch { /* ignore */ }
    }
    setAccessToken(null);
    clearStoredRefreshToken();
  },

  me: async () => {
    const refreshToken = getStoredRefreshToken();
    if (!refreshToken) throw new Error('Not authenticated');

    const refreshed = await api<AuthResponse>('/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    setAccessToken(refreshed.access_token);
    localStorage.setItem(REFRESH_KEY, refreshed.refresh_token);

    const user = await api<User>('/auth/me');
    return user;
  },
};
