import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { merchantsApi } from '@/features/merchants/api/merchantsApi';
import type { MerchantRef } from '@/features/auth/api/authApi';
import { CatalogMerchantContext, CATALOG_MERCHANT_KEY } from './catalogMerchantContext';
import type { CatalogMerchantValue } from './catalogMerchantContext';

function readStoredId(): number | null {
  const raw = localStorage.getItem(CATALOG_MERCHANT_KEY);
  const id = raw ? Number(raw) : NaN;
  return Number.isFinite(id) && id > 0 ? id : null;
}

/** Resolves which merchant the catalog operates on without depending on the
 *  branch-selection flow: prefer the active branch context when present,
 *  otherwise fall back to the user's own merchant list (persisted choice). */
export function CatalogMerchantProvider({ children }: { children: ReactNode }) {
  const { activeContext } = useAuth();
  const [merchants, setMerchants] = useState<MerchantRef[]>([]);
  const [storedId, setStoredId] = useState<number | null>(readStoredId);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    merchantsApi
      .list()
      .then((res) => {
        if (cancelled) return;
        setMerchants(res.map((m) => ({ id: m.id, name: m.name })));
      })
      .catch(() => {
        // merchant service unavailable — leave the list empty
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const selectMerchant = useCallback((id: number) => {
    setStoredId(id);
    localStorage.setItem(CATALOG_MERCHANT_KEY, String(id));
  }, []);

  const value = useMemo<CatalogMerchantValue>(() => {
    const contextMerchantId = activeContext?.merchant.id ?? null;
    const fallbackId =
      storedId != null && merchants.some((m) => m.id === storedId)
        ? storedId
        : merchants[0]?.id ?? null;
    const id = contextMerchantId ?? fallbackId;
    return {
      merchant: merchants.find((m) => m.id === id) ?? null,
      merchants,
      isLoading,
      selectMerchant,
    };
  }, [activeContext, merchants, storedId, isLoading, selectMerchant]);

  return (
    <CatalogMerchantContext.Provider value={value}>{children}</CatalogMerchantContext.Provider>
  );
}