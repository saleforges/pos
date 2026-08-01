import { api } from '@/lib/api';

export interface BranchResponse {
  id: number;
  merchantId: number;
  name: string;
  code: string;
  address: string;
  phone: string;
  status: string;
  operatingDays?: string[];
  operatingHours?: { open: string; close: string } | null;
  createdAt: string;
  updatedAt: string;
}

export interface CreateBranchRequest {
  name: string;
  code: string;
  address?: string;
  phone?: string;
}

const STORAGE_KEY = 'intools_mock_branches';

function getStored(): BranchResponse[] {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '[]');
  } catch {
    return [];
  }
}

function setStored(branches: BranchResponse[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(branches));
}

function seed() {
  const existing = getStored();
  if (existing.length === 0) {
    setStored([
      { id: 1, merchantId: 1, name: 'Main Branch', code: 'main', address: 'Jl. Sudirman No. 1', phone: '021-1234567', status: 'active', operatingDays: [], operatingHours: { open: '08:00', close: '21:00' }, createdAt: '2026-02-01T00:00:00Z', updatedAt: '2026-02-01T00:00:00Z' },
      { id: 2, merchantId: 2, name: 'HQ', code: 'hq', address: 'Jl. Thamrin No. 5', phone: '021-7654321', status: 'active', operatingDays: [], operatingHours: { open: '08:00', close: '21:00' }, createdAt: '2026-03-10T00:00:00Z', updatedAt: '2026-03-10T00:00:00Z' },
    ]);
  }
}

seed();

export const branchesApi = {
  list: async (merchantId: number): Promise<BranchResponse[]> => {
    try {
      return await api<BranchResponse[]>('/branches', {
        headers: { 'X-Merchant-Id': String(merchantId) } as Record<string, string>,
      });
    } catch {
      return getStored().filter((b) => b.merchantId === merchantId);
    }
  },

  create: async (merchantId: number, data: CreateBranchRequest): Promise<BranchResponse> => {
    try {
      return await api<BranchResponse>('/branches', {
        method: 'POST',
        headers: { 'X-Merchant-Id': String(merchantId) } as Record<string, string>,
        body: JSON.stringify(data),
      });
    } catch {
      const branches = getStored();
      const branch: BranchResponse = {
        id: branches.length > 0 ? Math.max(...branches.map((b) => b.id)) + 1 : 1,
        merchantId,
        name: data.name,
        code: data.code,
        address: data.address ?? '',
        phone: data.phone ?? '',
        status: 'active',
        operatingDays: [],
        operatingHours: { open: '08:00', close: '21:00' },
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      branches.push(branch);
      setStored(branches);
      return branch;
    }
  },
};
