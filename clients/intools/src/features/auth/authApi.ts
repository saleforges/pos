import { api } from '@/lib/api';

export interface Role {
  id: number;
  name: string;
}

export interface User {
  id: number;
  username: string;
  email: string;
  type: string;
  status: string;
  roles: Role[];
  permissions: string[];
  createdAt: string;
  updatedAt: string;
}

export function hasRole(user: User | null, ...roleNames: string[]): boolean {
  if (!user) return false;
  return roleNames.some((name) => user.roles.some((r) => r.name === name));
}

export const authApi = {
  login: async (username: string, password: string) => {
    await api<{ access_token: string; refresh_token: string; expires_in: number }>(
      '/auth/login',
      {
        method: 'POST',
        body: JSON.stringify({ username, password }),
      },
    );

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
    try {
      await api('/auth/logout', { method: 'POST' });
    } catch { /* ignore */ }
  },

  me: async () => {
    try {
      return await api<User>('/auth/me');
    } catch {
      try {
        await api('/auth/refresh', { method: 'POST' });
      } catch {
        throw new Error('Not authenticated');
      }
      return await api<User>('/auth/me');
    }
  },
};
