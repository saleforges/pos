import { api } from '@/lib/api';

export interface Merchant {
  id: number;
  name: string;
  ownerName?: string | null;
  ownerId?: number | null;
  branchCount: number;
  inventoryScoping: boolean;
  status: string;
  email: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateMerchantRequest {
  name: string;
  email: string;
  ownerUsername?: string;
  ownerEmail?: string;
  ownerPassword?: string;
  inventoryScoping?: boolean;
}

/*
 * TEMPORARY MOCK — swap for real API calls once /v1/merchants is available on the backend.
 *
 * Mock stores merchants in localStorage keyed by 'intools_mock_merchants'.
 */
const STORAGE_KEY = 'intools_mock_merchants';

function getStored(): Merchant[] {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '[]');
  } catch {
    return [];
  }
}

function setStored(merchants: Merchant[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(merchants));
}

function seed() {
  const existing = getStored();
  if (existing.length === 0) {
    setStored([
      { id: 1, name: 'Acme Corp', ownerName: 'owner', ownerId: 2, branchCount: 3, inventoryScoping: false, status: 'active', email: 'acme@example.com', createdAt: '2026-02-01T00:00:00Z', updatedAt: '2026-02-01T00:00:00Z' },
      { id: 2, name: 'Globex Inc', ownerName: null, ownerId: null, branchCount: 1, inventoryScoping: true, status: 'active', email: 'globex@example.com', createdAt: '2026-03-10T00:00:00Z', updatedAt: '2026-03-10T00:00:00Z' },
    ]);
  }
}

seed();

let nextId = (getStored().length > 0 ? Math.max(...getStored().map((m) => m.id)) : 0) + 1;

export const merchantsApi = {
  list: async (): Promise<Merchant[]> => {
    try {
      return await api<Merchant[]>('/merchants');
    } catch {
      return getStored();
    }
  },

  create: async (data: CreateMerchantRequest): Promise<Merchant> => {
    try {
      return await api<Merchant>('/merchants', {
        method: 'POST',
        body: JSON.stringify(data),
      });
    } catch {
      const merchants = getStored();
      const merchant: Merchant = {
        id: nextId++,
        name: data.name,
        ownerName: data.ownerUsername ?? null,
        ownerId: null,
        branchCount: 0,
        inventoryScoping: data.inventoryScoping ?? false,
        status: 'active',
        email: data.email,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      merchants.push(merchant);
      setStored(merchants);
      return merchant;
    }
  },
};
