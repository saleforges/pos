import { api, ApiError } from '@/lib/api';
import type { PlatformMerchant, CreateMerchantPayload } from '../types';

/*
 * MOCK-FIRST: These endpoints (/admin/merchants) may not exist on the backend yet.
 * When the real endpoints are deployed, replace the catch blocks (or the entire
 * function body) with direct `return api<T>('/admin/merchants', ...)` calls.
 */

const MOCK_MERCHANTS: PlatformMerchant[] = [
  {
    id: 1,
    name: 'Acme Corp',
    ownerId: 2,
    ownerUsername: 'jane',
    ownerEmail: 'jane@acme.com',
    branchCount: 3,
    inventoryScoping: 'shared',
    status: 'active',
    createdAt: '2026-01-15T00:00:00Z',
  },
  {
    id: 2,
    name: 'Globex Inc',
    ownerId: null,
    ownerUsername: null,
    ownerEmail: null,
    branchCount: 1,
    inventoryScoping: 'independent_per_branch',
    status: 'active',
    createdAt: '2026-03-20T00:00:00Z',
  },
  {
    id: 3,
    name: 'Initech',
    ownerId: 4,
    ownerUsername: 'michael',
    ownerEmail: 'michael@initech.com',
    branchCount: 2,
    inventoryScoping: 'shared',
    status: 'inactive',
    createdAt: '2026-05-01T00:00:00Z',
  },
];

let nextMockId = 4;

let mockMerchants = [...MOCK_MERCHANTS];

export const platformAdminMerchantsApi = {
  list: async (): Promise<PlatformMerchant[]> => {
    try {
      return await api<PlatformMerchant[]>('/admin/merchants');
    } catch {
      return mockMerchants;
    }
  },

  create: async (data: CreateMerchantPayload): Promise<PlatformMerchant> => {
    try {
      return await api<PlatformMerchant>('/admin/merchants', {
        method: 'POST',
        body: JSON.stringify(data),
      });
    } catch (err) {
      if (err instanceof ApiError) throw err;
      if (mockMerchants.some((m) => m.name.toLowerCase() === data.name.toLowerCase())) {
        throw new ApiError(409, JSON.stringify({ error: 'A merchant with this name already exists' }));
      }
      const merchant: PlatformMerchant = {
        id: nextMockId++,
        name: data.name,
        ownerId: data.ownerId ?? null,
        ownerUsername: data.newOwner?.username ?? null,
        ownerEmail: data.newOwner?.email ?? null,
        branchCount: 0,
        inventoryScoping: data.inventoryScoping,
        status: 'active',
        createdAt: new Date().toISOString(),
      };
      mockMerchants = [...mockMerchants, merchant];
      return merchant;
    }
  },
};
