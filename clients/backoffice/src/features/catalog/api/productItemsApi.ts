import { api } from '@/lib/api';
import type { ProductItem } from '../types';
import type { VariantInput } from './productsApi';

export interface UpdateVariantRequest {
  name?: string;
  sku?: string;
  unitId?: number | null;
  price?: number;
  trackInventory?: boolean;
  imageUrl?: string;
}

const merchantHeaders = (merchantId: number): Record<string, string> => ({
  'X-Merchant-Id': String(merchantId),
});

export const productItemsApi = {
  listByProduct: async (merchantId: number, productId: number) => {
    const res = await api<ProductItem[] | null>(`/products/${productId}/items`, {
      headers: merchantHeaders(merchantId),
    });
    return res ?? [];
  },

  create: (merchantId: number, productId: number, data: VariantInput) =>
    api<ProductItem>(`/products/${productId}/items`, {
      method: 'POST',
      headers: merchantHeaders(merchantId),
      body: JSON.stringify(data),
    }),

  update: (merchantId: number, id: number, data: UpdateVariantRequest) =>
    api<ProductItem>(`/product-items/${id}`, {
      method: 'PATCH',
      headers: merchantHeaders(merchantId),
      body: JSON.stringify(data),
    }),

  remove: (merchantId: number, id: number) =>
    api<void>(`/product-items/${id}`, {
      method: 'DELETE',
      headers: merchantHeaders(merchantId),
    }),
};