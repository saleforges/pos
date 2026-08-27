import { api } from '@/lib/api';
import type { Category } from '../types';

export interface SaveCategoryRequest {
  name: string;
  parentId?: number | null;
}

const merchantHeaders = (merchantId: number): Record<string, string> => ({
  'X-Merchant-Id': String(merchantId),
});

export const categoriesApi = {
  list: async (merchantId: number) => {
    const res = await api<Category[] | null>('/categories', { headers: merchantHeaders(merchantId) });
    return res ?? [];
  },

  create: (merchantId: number, data: SaveCategoryRequest) =>
    api<Category>('/categories', {
      method: 'POST',
      headers: merchantHeaders(merchantId),
      body: JSON.stringify(data),
    }),

  update: (merchantId: number, id: number, data: SaveCategoryRequest) =>
    api<Category>(`/categories/${id}`, {
      method: 'PATCH',
      headers: merchantHeaders(merchantId),
      body: JSON.stringify(data),
    }),

  remove: (merchantId: number, id: number) =>
    api<void>(`/categories/${id}`, {
      method: 'DELETE',
      headers: merchantHeaders(merchantId),
    }),
};