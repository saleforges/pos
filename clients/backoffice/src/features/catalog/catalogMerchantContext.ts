import { createContext, useContext } from 'react';
import type { MerchantRef } from '@/features/auth/api/authApi';

export const CATALOG_MERCHANT_KEY = 'backoffice.catalog.merchantId';

export interface CatalogMerchantValue {
  /** The merchant the catalog tab should operate on. */
  merchant: MerchantRef | null;
  /** All merchants the logged-in user can access (for the switcher). */
  merchants: MerchantRef[];
  isLoading: boolean;
  selectMerchant: (id: number) => void;
}

export const CatalogMerchantContext = createContext<CatalogMerchantValue | undefined>(undefined);

export function useCatalogMerchant(): CatalogMerchantValue {
  const ctx = useContext(CatalogMerchantContext);
  if (!ctx) throw new Error('useCatalogMerchant must be used within <CatalogMerchantProvider>');
  return ctx;
}