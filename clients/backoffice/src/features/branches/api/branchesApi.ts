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

export const branchesApi = {
  list: (merchantId: number) =>
    api<BranchResponse[]>('/branches', {
      headers: { 'X-Merchant-Id': String(merchantId) } as Record<string, string>,
    }),
  create: (merchantId: number, data: CreateBranchRequest) =>
    api<BranchResponse>('/branches', {
      method: 'POST',
      headers: { 'X-Merchant-Id': String(merchantId) } as Record<string, string>,
      body: JSON.stringify(data),
    }),
};
