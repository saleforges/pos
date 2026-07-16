import { api, setAccessToken } from '@/lib/api';
import { rolesApi } from '@/features/roles/api/rolesApi';

export interface Role {
  id: number;
  name: string;
  merchant?: string | null;
  branch?: string | null;
  branchScope?: string;
  isDefault?: boolean;
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

export interface RoleDefinition {
  id: number;
  name: string;
  description: string;
  permissions: string[];
  is_system: boolean;
}

export async function resolveUserPermissions(user: User): Promise<string[]> {
  if (user.permissions?.length) return user.permissions;
  try {
    const roles = await rolesApi.list();
    const userRoleNames = new Set(user.roles.map((r) => r.name));
    const perms = new Set<string>();
    for (const role of roles) {
      if (userRoleNames.has(role.name) && role.permissions) {
        role.permissions.forEach((p: string) => perms.add(p));
      }
    }
    return Array.from(perms);
  } catch {
    return [];
  }
}

export function hasRole(user: User | null, ...roleNames: string[]): boolean {
  if (!user) return false;
  return roleNames.some((name) => user.roles.some((r) => r.name === name));
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
