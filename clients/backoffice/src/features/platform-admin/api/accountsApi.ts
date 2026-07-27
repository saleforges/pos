import { api, ApiError } from '@/lib/api';
import type { PlatformAccount, CreateAccountPayload } from '../types';

/*
 * MOCK-FIRST: These endpoints (/admin/accounts) may not exist on the backend yet.
 * When the real endpoints are deployed, replace the catch blocks with direct
 * `return api<T>('/admin/accounts', ...)` calls.
 */

const MOCK_ACCOUNTS: PlatformAccount[] = [
  {
    id: 1,
    username: 'superadmin',
    email: 'superadmin@saleforges.com',
    roleName: 'superadmin',
    status: 'active',
    merchantId: null,
    merchantName: null,
    createdAt: '2026-01-01T00:00:00Z',
  },
  {
    id: 2,
    username: 'jane',
    email: 'jane@acme.com',
    roleName: 'owner',
    status: 'active',
    merchantId: 1,
    merchantName: 'Acme Corp',
    createdAt: '2026-01-15T00:00:00Z',
  },
  {
    id: 3,
    username: 'bob',
    email: 'bob@acme.com',
    roleName: 'staff',
    status: 'active',
    merchantId: 1,
    merchantName: 'Acme Corp',
    createdAt: '2026-02-01T00:00:00Z',
  },
  {
    id: 4,
    username: 'michael',
    email: 'michael@initech.com',
    roleName: 'owner',
    status: 'active',
    merchantId: 3,
    merchantName: 'Initech',
    createdAt: '2026-05-01T00:00:00Z',
  },
];

const MOCK_ASSIGNMENT_ROLES = ['owner', 'staff'];

let nextMockId = 5;

let mockAccounts = [...MOCK_ACCOUNTS];

export const platformAdminAccountsApi = {
  list: async (): Promise<PlatformAccount[]> => {
    try {
      return await api<PlatformAccount[]>('/admin/accounts');
    } catch {
      return mockAccounts;
    }
  },

  listOwners: async (): Promise<PlatformAccount[]> => {
    try {
      return await api<PlatformAccount[]>('/admin/accounts?role=owner');
    } catch {
      return mockAccounts.filter((a) => a.roleName === 'owner');
    }
  },

  getAssignmentRoles: async (): Promise<string[]> => {
    try {
      return await api<string[]>('/admin/accounts/roles');
    } catch {
      return MOCK_ASSIGNMENT_ROLES;
    }
  },

  create: async (data: CreateAccountPayload): Promise<PlatformAccount> => {
    try {
      return await api<PlatformAccount>('/admin/accounts', {
        method: 'POST',
        body: JSON.stringify(data),
      });
    } catch (err) {
      if (err instanceof ApiError) throw err;
      const exists = mockAccounts.find(
        (a) => a.username.toLowerCase() === data.username.toLowerCase(),
      );
      if (exists) {
        throw new ApiError(409, JSON.stringify({ error: 'Username already exists' }));
      }
      if (mockAccounts.some((a) => a.email.toLowerCase() === data.email.toLowerCase())) {
        throw new ApiError(409, JSON.stringify({ error: 'Email already exists' }));
      }
      const account: PlatformAccount = {
        id: nextMockId++,
        username: data.username,
        email: data.email,
        roleName: data.role,
        status: 'active',
        merchantId: data.merchantId ?? null,
        merchantName: null,
        createdAt: new Date().toISOString(),
      };
      mockAccounts = [...mockAccounts, account];
      return account;
    }
  },
};
