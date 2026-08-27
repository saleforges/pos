import { useCallback, useEffect, useState } from 'react';
import { Package, Pencil, Plus, Search, Trash2 } from 'lucide-react';
import { Can } from '@/components/ui/Can';
import { Alert } from '@/components/ui/Alert';
import { productsApi } from '@/features/catalog/api/productsApi';
import { categoriesApi } from '@/features/catalog/api/categoriesApi';
import { unitsApi } from '@/features/catalog/api/unitsApi';
import { useCatalogMerchant } from '@/features/catalog/catalogMerchantContext';
import type { Category, ProductListResponse, Unit } from '@/features/catalog/types';
import { formatPriceRange } from '@/features/catalog/utils';
import { ProductFormModal } from '@/features/catalog/components/ProductFormModal';

const PAGE_SIZE = 20;

export default function CatalogProductsPage() {
  const { merchant, isLoading: merchantLoading } = useCatalogMerchant();
  const merchantId = merchant?.id;

  const [products, setProducts] = useState<ProductListResponse[]>([]);
  const [meta, setMeta] = useState<{ total: number; count: number; offset: number } | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');

  const [searchInput, setSearchInput] = useState('');
  const [search, setSearch] = useState('');
  const [offset, setOffset] = useState(0);

  const [categories, setCategories] = useState<Category[]>([]);
  const [units, setUnits] = useState<Unit[]>([]);

  const [showCreate, setShowCreate] = useState(false);
  const [editing, setEditing] = useState<ProductListResponse | null>(null);

  // Debounce the search box so typing doesn't fire a request per keystroke.
  useEffect(() => {
    const t = setTimeout(() => {
      setSearch(searchInput.trim());
      setOffset(0);
    }, 300);
    return () => clearTimeout(t);
  }, [searchInput]);

  useEffect(() => {
    if (!merchantId) return;
    categoriesApi.list(merchantId).then(setCategories).catch(() => {});
    unitsApi.list().then(setUnits).catch(() => {});
  }, [merchantId]);

  const fetchProducts = useCallback(async () => {
    if (!merchantId) return;
    setIsLoading(true);
    setError('');
    try {
      const res = await productsApi.list(merchantId, { search, offset, limit: PAGE_SIZE });
      setProducts(res.data);
      setMeta(res.pagination);
    } catch {
      setError('Failed to load products. Is the catalog service running?');
    } finally {
      setIsLoading(false);
    }
  }, [merchantId, search, offset]);

  useEffect(() => {
    fetchProducts();
  }, [fetchProducts]);

  if (merchantLoading) {
    return <div className="text-neutral-400">Loading products...</div>;
  }

  if (!merchantId) {
    return (
      <Alert variant="warning">
        No merchants available yet. Create a merchant before managing products.
      </Alert>
    );
  }

  const handleDelete = async (product: ProductListResponse) => {
    if (!window.confirm(`Delete product "${product.name}"? Its variants will also be removed.`)) {
      return;
    }
    try {
      await productsApi.remove(merchantId, product.id);
      fetchProducts();
    } catch {
      setError('Failed to delete product.');
    }
  };

  const total = meta?.total ?? 0;
  const shownFrom = total === 0 ? 0 : offset + 1;
  const shownTo = offset + (meta?.count ?? 0);

  return (
    <>
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="relative flex-1 max-w-xs">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-400" />
            <input
              type="text"
              placeholder="Search products..."
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              className="w-full rounded-lg border border-neutral-200 bg-white py-2 pl-9 pr-3 text-sm outline-none focus:border-neutral-400"
            />
          </div>
          <Can permission="catalog.create">
            <button
              onClick={() => setShowCreate(true)}
              className="flex items-center gap-2 rounded-lg bg-primary px-3 py-2 text-sm text-white hover:bg-primary-hover"
            >
              <Plus size={16} />
              Add Product
            </button>
          </Can>
        </div>

        {error && <Alert variant="danger">{error}</Alert>}

        <div className="rounded-lg border border-neutral-200 bg-white">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-neutral-100 text-left text-xs text-neutral-500">
                  <th className="px-4 py-3 font-medium">Product</th>
                  <th className="px-4 py-3 font-medium">Category</th>
                  <th className="px-4 py-3 font-medium">Price</th>
                  <th className="px-4 py-3 font-medium">Variants</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody>
                {isLoading ? (
                  <tr>
                    <td colSpan={6} className="px-4 py-8 text-center text-neutral-400">
                      Loading products...
                    </td>
                  </tr>
                ) : products.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="px-4 py-8 text-center text-neutral-400">
                      {search ? 'No products match your search.' : 'No products yet. Add your first product to get started.'}
                    </td>
                  </tr>
                ) : (
                  products.map((p) => (
                    <tr key={p.id} className="border-b border-neutral-100 last:border-0 hover:bg-neutral-50">
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-3">
                          {p.imageUrl ? (
                            <img
                              src={p.imageUrl}
                              alt=""
                              className="h-9 w-9 shrink-0 rounded-lg object-cover"
                            />
                          ) : (
                            <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-neutral-100 text-neutral-400">
                              <Package size={16} />
                            </div>
                          )}
                          <span className="font-medium text-neutral-900">{p.name}</span>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-neutral-500">{p.category?.name || '—'}</td>
                      <td className="px-4 py-3 text-neutral-600">
                        {formatPriceRange(p.priceRange.min, p.priceRange.max)}
                      </td>
                      <td className="px-4 py-3 text-neutral-600">{p.items.length}</td>
                      <td className="px-4 py-3">
                        <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                          p.status === 'active'
                            ? 'bg-secondary/10 text-secondary'
                            : 'bg-neutral-100 text-neutral-500'
                        }`}>
                          {p.status}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center justify-end gap-1">
                          <Can permission="catalog.update">
                            <button
                              onClick={() => setEditing(p)}
                              className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600"
                              title="Edit product"
                            >
                              <Pencil size={15} />
                            </button>
                          </Can>
                          <Can permission="catalog.delete">
                            <button
                              onClick={() => handleDelete(p)}
                              className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-danger"
                              title="Delete product"
                            >
                              <Trash2 size={15} />
                            </button>
                          </Can>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          {!isLoading && total > 0 && (
            <div className="flex items-center justify-between border-t border-neutral-100 px-4 py-3 text-xs text-neutral-500">
              <span>
                Showing {shownFrom}–{shownTo} of {total}
              </span>
              <div className="flex gap-2">
                <button
                  onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
                  disabled={offset === 0}
                  className="rounded-md border border-neutral-200 px-3 py-1.5 text-neutral-600 hover:bg-neutral-50 disabled:cursor-not-allowed disabled:opacity-40"
                >
                  Previous
                </button>
                <button
                  onClick={() => setOffset(offset + PAGE_SIZE)}
                  disabled={shownTo >= total}
                  className="rounded-md border border-neutral-200 px-3 py-1.5 text-neutral-600 hover:bg-neutral-50 disabled:cursor-not-allowed disabled:opacity-40"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </div>
      </div>

      {showCreate && (
        <ProductFormModal
          merchantId={merchantId}
          categories={categories}
          units={units}
          onClose={() => setShowCreate(false)}
          onSaved={() => {
            setShowCreate(false);
            fetchProducts();
          }}
        />
      )}

      {editing && (
        <ProductFormModal
          merchantId={merchantId}
          categories={categories}
          units={units}
          product={editing}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null);
            fetchProducts();
          }}
        />
      )}
    </>
  );
}