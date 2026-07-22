import { api } from '@/lib/api';
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

export const authApi = {
  login: async (username: string, password: string) => {
    // Backend sets access_token + refresh_token as HttpOnly cookies via Set-Cookie.
    // Browser handles cookie storage and sending automatically.
    const res = await api<{ access_token: string; refresh_token: string; expires_in: number }>(
      '/auth/login',
      {
        method: 'POST',
        body: JSON.stringify({ username, password }),
      },
    );

    // Fetch user info — the auth middleware extracts the token from the cookie
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
    // Backend clears cookies server-side
    try {
      await api('/auth/logout', { method: 'POST' });
    } catch { /* ignore — cookies cleared either way */ }
  },

  /** Attempt to restore session on page load.
   *  If the access token cookie is stale, the refresh token cookie is used
   *  transparently by the backend to issue new tokens. */
  me: async () => {
    try {
      return await api<User>('/auth/me');
    } catch {
      // Access token expired — try refreshing via the refresh_token cookie
      try {
        await api('/auth/refresh', { method: 'POST' });
      } catch {
        throw new Error('Not authenticated');
      }
      return await api<User>('/auth/me');
    }
  },
};

// Exported for backward compat — no longer stores tokens client-side
export function getStoredRefreshToken(): string | null {
  return null;
}

export function clearStoredRefreshToken() {
  // no-op — tokens live in HttpOnly cookies now
}
