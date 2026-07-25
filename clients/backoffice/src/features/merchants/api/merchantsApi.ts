import { api } from '@/lib/api';

interface MerchantSettings {
  taxRate: number;
  currency: string;
  timezone: string;
  receiptFooter?: string;
  receiptLogo?: string;
  orderPrefix?: string;
  lowStockThreshold: number;
}

export interface MerchantResponse {
  id: number;
  name: string;
  legalName: string;
  address: string;
  phone: string;
  email: string;
  logoUrl?: string;
  taxId?: string;
  status: string;
  settings: MerchantSettings;
  createdAt: string;
  updatedAt: string;
}

export interface CreateMerchantRequest {
  name: string;
  legalName?: string;
  address?: string;
  phone?: string;
  email: string;
  taxId?: string;
}

export interface UpdateMerchantRequest {
  name?: string;
  legalName?: string;
  address?: string;
  phone?: string;
  email?: string;
  taxId?: string;
}

export const merchantsApi = {
  list: () => api<MerchantResponse[]>('/merchants'),
  get: (id: number) => api<MerchantResponse>(`/merchants/${id}`),
  create: (data: CreateMerchantRequest) =>
    api<MerchantResponse>('/merchants', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  update: (id: number, data: UpdateMerchantRequest) =>
    api<MerchantResponse>(`/merchants/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    }),
};
