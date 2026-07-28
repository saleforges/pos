import { api } from '@/lib/api';

export interface Account {
  id: number;
  username: string;
  email: string;
  role: string;
  status: string;
  merchantId?: number | null;
  merchantName?: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface CreateAccountRequest {
  username: string;
  email: string;
  password: string;
  role: string;
  merchantId?: number;
}

/*
 * TEMPORARY MOCK — swap for real API calls once /v1/accounts is available on the backend.
 *
 * Mock stores accounts in localStorage keyed by 'intools_mock_accounts'.
 * Default seed includes a superadmin for testing.
 */
const STORAGE_KEY = 'intools_mock_accounts';

function getStored(): Account[] {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '[]');
  } catch {
    return [];
  }
}

function setStored(accounts: Account[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(accounts));
}

function seed() {
  const existing = getStored();
  if (existing.length === 0) {
    setStored([
      { id: 1, username: 'admin', email: 'admin@saleforges.com', role: 'superadmin', status: 'active', merchantId: null, merchantName: null, createdAt: '2026-01-01T00:00:00Z', updatedAt: '2026-01-01T00:00:00Z' },
      { id: 2, username: 'merchant_owner', email: 'owner@example.com', role: 'admin', status: 'active', merchantId: 1, merchantName: 'Acme Corp', createdAt: '2026-02-15T00:00:00Z', updatedAt: '2026-02-15T00:00:00Z' },
    ]);
  }
}

seed();

export const accountsApi = {
  list: async (): Promise<Account[]> => {
    try {
      return await api<Account[]>('/accounts');
    } catch {
      return getStored();
    }
  },

  create: async (data: CreateAccountRequest): Promise<Account> => {
    try {
      return await api<Account>('/accounts', {
        method: 'POST',
        body: JSON.stringify(data),
      });
    } catch {
      const accounts = getStored();
      const newAccount: Account = {
        id: accounts.length > 0 ? Math.max(...accounts.map((a) => a.id)) + 1 : 1,
        username: data.username,
        email: data.email,
        role: data.role,
        status: 'active',
        merchantId: data.merchantId ?? null,
        merchantName: null,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      accounts.push(newAccount);
      setStored(accounts);
      return newAccount;
    }
  },
};
