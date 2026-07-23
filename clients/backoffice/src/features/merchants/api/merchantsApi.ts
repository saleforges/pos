import { api } from '@/lib/api';
import type { Merchant, CreateMerchantPayload } from '../types';

export const merchantsApi = {
  list: (params?: { offset?: number; limit?: number }) => {
    const qs = new URLSearchParams();
    if (params?.offset !== undefined) qs.set('offset', String(params.offset));
    if (params?.limit !== undefined) qs.set('limit', String(params.limit));
    const query = qs.toString();
    return api<Merchant[]>(`/merchants${query ? `?${query}` : ''}`);
  },

  get: (id: number) => api<Merchant>(`/merchants/${id}`),

  create: (data: CreateMerchantPayload) =>
    api<Merchant>('/merchants', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<Merchant>) =>
    api<Merchant>(`/merchants/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    api<void>(`/merchants/${id}`, { method: 'DELETE' }),
};
