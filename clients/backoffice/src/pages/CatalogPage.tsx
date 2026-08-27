import { useState } from 'react';
import { Store } from 'lucide-react';
import { useCatalogMerchant } from '@/features/catalog/catalogMerchantContext';
import { CatalogMerchantProvider } from '@/features/catalog/CatalogMerchantProvider';
import CatalogProductsPage from '@/pages/CatalogProductsPage';
import CatalogCategoriesPage from '@/pages/CatalogCategoriesPage';

type Tab = 'products' | 'categories';

const tabs: { id: Tab; label: string }[] = [
  { id: 'products', label: 'Products' },
  { id: 'categories', label: 'Categories' },
];

const tabClass = (active: boolean) =>
  `-mb-px border-b-2 px-3 py-2 text-sm transition-colors ${
    active
      ? 'border-primary font-medium text-primary'
      : 'border-transparent text-neutral-500 hover:text-neutral-700'
  }`;

function MerchantSelector() {
  const { merchant, merchants, isLoading, selectMerchant } = useCatalogMerchant();

  if (isLoading) {
    return <div className="text-sm text-neutral-400">Loading merchant...</div>;
  }
  if (!merchant) return null;

  if (merchants.length <= 1) {
    return (
      <div className="flex items-center gap-1.5 text-sm text-neutral-600">
        <Store size={14} className="text-neutral-400" />
        <span className="font-medium text-neutral-900">{merchant.name}</span>
      </div>
    );
  }

  return (
    <label className="flex items-center gap-1.5 text-sm">
      <Store size={14} className="text-neutral-400 shrink-0" />
      <select
        value={merchant.id}
        onChange={(e) => selectMerchant(Number(e.target.value))}
        className="rounded-lg border border-neutral-200 bg-white px-2 py-1.5 text-sm text-neutral-700 outline-none focus:border-neutral-400"
      >
        {merchants.map((m) => (
          <option key={m.id} value={m.id}>
            {m.name}
          </option>
        ))}
      </select>
    </label>
  );
}

export default function CatalogPage() {
  const [tab, setTab] = useState<Tab>('products');

  return (
    <CatalogMerchantProvider>
      <div className="space-y-4">
        <div className="flex items-center justify-between gap-4">
          <nav className="flex grow gap-1 border-b border-neutral-200">
            {tabs.map((t) => (
              <button
                key={t.id}
                onClick={() => setTab(t.id)}
                className={tabClass(tab === t.id)}
              >
                {t.label}
              </button>
            ))}
          </nav>
          <MerchantSelector />
        </div>
        {tab === 'products' ? <CatalogProductsPage /> : <CatalogCategoriesPage />}
      </div>
    </CatalogMerchantProvider>
  );
}