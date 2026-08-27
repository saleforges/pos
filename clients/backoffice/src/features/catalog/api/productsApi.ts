import { api, apiPaginated } from '@/lib/api';
import type { Product, ProductListResponse } from '../types';

export interface VariantInput {
  name: string;
  sku?: string;
  unitId?: number | null;
  price: number;
  trackInventory: boolean;
  imageUrl?: string;
}

export interface CreateProductRequest {
  categoryId: number;
  name: string;
  description?: string;
  imageUrl?: string;
  items: VariantInput[];
}

export interface UpdateProductRequest {
  categoryId?: number;
  name?: string;
  description?: string;
  imageUrl?: string;
  status?: string;
}

export interface BulkCreateResponse {
  product: Product;
  items: { id: number; name: string; price: number }[];
}

const merchantHeaders = (merchantId: number): Record<string, string> => ({
  'X-Merchant-Id': String(merchantId),
});

export const productsApi = {
  list: (
    merchantId: number,
    params: { search?: string; offset?: number; limit?: number } = {},
  ) => {
    const q = new URLSearchParams();
    if (params.search) q.set('search', params.search);
    q.set('offset', String(params.offset ?? 0));
    q.set('limit', String(params.limit ?? 20));
    return apiPaginated<ProductListResponse>(`/products?${q.toString()}`, {
      headers: merchantHeaders(merchantId),
    });
  },

  get: (merchantId: number, id: number) =>
    api<Product>(`/products/${id}`, { headers: merchantHeaders(merchantId) }),

  /** Creates the product together with its initial variants in one call
   *  (the plain POST /products endpoint cannot set SKU/unit on the default item). */
  create: (merchantId: number, data: CreateProductRequest) =>
    api<BulkCreateResponse>('/products/bulk', {
      method: 'POST',
      headers: merchantHeaders(merchantId),
      body: JSON.stringify(data),
    }),

  update: (merchantId: number, id: number, data: UpdateProductRequest) =>
    api<Product>(`/products/${id}`, {
      method: 'PATCH',
      headers: merchantHeaders(merchantId),
      body: JSON.stringify(data),
    }),

  remove: (merchantId: number, id: number) =>
    api<void>(`/products/${id}`, {
      method: 'DELETE',
      headers: merchantHeaders(merchantId),
    }),
};