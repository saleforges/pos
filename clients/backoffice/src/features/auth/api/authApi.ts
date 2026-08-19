import { api } from '@/lib/api';
import { rolesApi } from '@/features/roles/api/rolesApi';
import { branchesApi, type BranchResponse } from '@/features/branches/api/branchesApi';
import { merchantsApi } from '@/features/merchants/api/merchantsApi';

export interface MerchantRef {
  id: number;
  name: string;
}

export interface BranchRef {
  id: number;
  name: string;
}

export interface Role {
  id: number;
  name: string;
  merchant?: MerchantRef | null;
  branch?: BranchRef | null;
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

/** A merchant+branch combination the user can actively work in. */
export interface BranchContext {
  userRoleId: number;
  merchant: MerchantRef;
  branch: BranchRef;
}

/** Merchant-service response for /v1/me/assignments (requires X-Merchant-Id). */
export interface StaffAssignmentResponse {
  id: number;
  merchantId: number;
  branchId: number;
  userId: number;
  role: string;
  status: string;
  isDefault: boolean;
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

/** Build the list of merchant/branch contexts the user can select from.
 *  Branch-scoped roles expose their branch directly; merchant-wide roles
 *  (branchScope === 'all') expose every branch of that merchant.
 *  When /auth/me roles carry no merchant/branch info (staff accounts whose
 *  assignments live in the merchant service), fall back to the existing
 *  merchant-service endpoints: list all merchants, read /v1/me/assignments
 *  per merchant, and resolve branch names via /v1/branches. A user assigned
 *  as an "owner" is granted every branch of that merchant. */
export async function resolveBranchContexts(user: User): Promise<BranchContext[]> {
  const contexts: BranchContext[] = [];
  const seen = new Set<string>();

  for (const role of user.roles) {
    if (!role.merchant) continue;

    if (role.branch) {
      const key = `${role.merchant.id}:${role.branch.id}`;
      if (seen.has(key)) continue;
      seen.add(key);
      contexts.push({
        userRoleId: role.id,
        merchant: role.merchant,
        branch: role.branch,
      });
      continue;
    }

    if (role.branchScope === 'all') {
      try {
        const branches = await branchesApi.list(role.merchant.id);
        for (const b of branches) {
          const key = `${role.merchant.id}:${b.id}`;
          if (seen.has(key)) continue;
          seen.add(key);
          contexts.push({
            userRoleId: role.id,
            merchant: role.merchant,
            branch: { id: b.id, name: b.name },
          });
        }
      } catch {
        // branch fetch failed — skip; the user can still pick from their roles
      }
    }
  }

  if (contexts.length === 0 && user.roles.length > 0) {
    try {
      const roleIdByName = new Map(user.roles.map((r) => [r.name, r.id]));
      const merchants = await merchantsApi.list();
      for (const merchant of merchants) {
        const assignments = await authApi.myAssignments(merchant.id);
        if (assignments.length === 0) continue;
        let branches: BranchResponse[] = [];
        try {
          branches = await branchesApi.list(merchant.id);
        } catch {
          // branch list unavailable — fall back to assigned branch ids
        }
        const branchById = new Map(branches.map((b) => [b.id, b]));

        // An owner manages the whole merchant, so every branch is selectable.
        const ownerRoleId = roleIdByName.get('owner');
        const isOwner = assignments.some((a) => a.role === 'owner');
        const candidates: { branchId: number; name: string; roleId: number }[] =
          isOwner && ownerRoleId != null
            ? branches.map((b) => ({ branchId: b.id, name: b.name, roleId: ownerRoleId }))
            : [];

        if (candidates.length === 0) {
          for (const a of assignments) {
            const roleId = roleIdByName.get(a.role);
            if (roleId == null) continue;
            candidates.push({
              branchId: a.branchId,
              name: branchById.get(a.branchId)?.name ?? `Branch ${a.branchId}`,
              roleId,
            });
          }
        }

        for (const c of candidates) {
          const key = `${merchant.id}:${c.branchId}`;
          if (seen.has(key)) continue;
          seen.add(key);
          contexts.push({
            userRoleId: c.roleId,
            merchant: { id: merchant.id, name: merchant.name },
            branch: { id: c.branchId, name: c.name },
          });
        }
      }
    } catch {
      // merchant service unavailable — no contexts resolved
    }
  }

  return contexts;
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
    await api<{ access_token: string; refresh_token: string; expires_in: number }>(
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

  /** Switch the active role/context. The backend derives the branch from the
   *  selected role assignment and re-issues the access token with the new
   *  mid/bid claims. */
  switchContext: (userRoleId: number) =>
    api<{ accessToken: string; expiresIn: number }>('/auth/switch-context', {
      method: 'POST',
      body: JSON.stringify({ userRoleId }),
    }),

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

  /** The user's staff assignments for a merchant, from the merchant service.
   *  Requires the merchant context header. */
  myAssignments: (merchantId: number) =>
    api<StaffAssignmentResponse[]>(`/me/assignments`, {
      headers: { 'X-Merchant-Id': String(merchantId) } as Record<string, string>,
    }),
};

// Exported for backward compat — no longer stores tokens client-side
export function getStoredRefreshToken(): string | null {
  return null;
}

export function clearStoredRefreshToken() {
  // no-op — tokens live in HttpOnly cookies now
}
