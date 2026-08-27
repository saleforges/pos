export type ProductStatus = 'active' | 'inactive' | 'archived';

export type ProductItemStatus = 'active' | 'inactive';

export interface Unit {
  id: number;
  code: string;
  name: string;
}

export interface Category {
  id: number;
  merchantId: number;
  name: string;
  parentId?: number | null;
  createdAt: string;
  updatedAt: string;
}

export interface Price {
  amount: number;
  currency: string;
}

/** Product item as returned by GET /products/:productId/items. */
export interface ProductItem {
  id: number;
  productId: number;
  name: string;
  sku?: string;
  unitId?: number | null;
  price: Price;
  trackInventory: boolean;
  imageUrl?: string;
  status: ProductItemStatus;
}

/** Variant as embedded in the products list response (unit resolved by name). */
export interface ProductListVariant {
  id: number;
  name: string;
  sku?: string;
  unit?: Unit | null;
  price: Price;
  trackInventory: boolean;
  status: ProductItemStatus;
}

export interface Product {
  id: number;
  categoryId?: number;
  name: string;
  description?: string;
  imageUrl?: string;
  status: ProductStatus;
  createdAt: string;
  updatedAt: string;
}

/** Enriched product row from the paginated list endpoint. */
export interface ProductListResponse extends Product {
  category: { id: number; name: string };
  priceRange: { min: number; max: number };
  items: ProductListVariant[];
}