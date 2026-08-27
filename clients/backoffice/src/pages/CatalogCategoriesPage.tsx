import { useCallback, useEffect, useState } from 'react';
import { Pencil, Plus, Trash2 } from 'lucide-react';
import { Can } from '@/components/ui/Can';
import { Alert } from '@/components/ui/Alert';
import { categoriesApi } from '@/features/catalog/api/categoriesApi';
import { useCatalogMerchant } from '@/features/catalog/catalogMerchantContext';
import type { Category } from '@/features/catalog/types';
import { extractErrorMessage } from '@/features/catalog/utils';
import { CategoryFormModal } from '@/features/catalog/components/CategoryFormModal';

export default function CatalogCategoriesPage() {
  const { merchant, isLoading: merchantLoading } = useCatalogMerchant();
  const merchantId = merchant?.id;

  const [categories, setCategories] = useState<Category[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');

  const [showCreate, setShowCreate] = useState(false);
  const [editing, setEditing] = useState<Category | null>(null);

  const fetchCategories = useCallback(async () => {
    if (!merchantId) return;
    setIsLoading(true);
    try {
      setCategories(await categoriesApi.list(merchantId));
    } catch {
      setError('Failed to load categories.');
    } finally {
      setIsLoading(false);
    }
  }, [merchantId]);

  useEffect(() => {
    fetchCategories();
  }, [fetchCategories]);

  if (merchantLoading) {
    return <div className="text-neutral-400">Loading categories...</div>;
  }

  if (!merchantId) {
    return (
      <Alert variant="warning">
        No merchants available yet. Create a merchant before managing categories.
      </Alert>
    );
  }

  const handleDelete = async (category: Category) => {
    if (!window.confirm(`Delete category "${category.name}"?`)) return;
    try {
      await categoriesApi.remove(merchantId, category.id);
      fetchCategories();
    } catch (err) {
      setError(extractErrorMessage(err, 'Failed to delete category.'));
    }
  };

  const nameById = new Map(categories.map((c) => [c.id, c.name]));

  return (
    <>
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <p className="text-sm text-neutral-500">
            Group products into categories. Categories with children cannot nest further via the same parent.
          </p>
          <Can permission="catalog.create">
            <button
              onClick={() => setShowCreate(true)}
              className="flex shrink-0 items-center gap-2 rounded-lg bg-primary px-3 py-2 text-sm text-white hover:bg-primary-hover"
            >
              <Plus size={16} />
              Add Category
            </button>
          </Can>
        </div>

        {error && <Alert variant="danger">{error}</Alert>}

        <div className="rounded-lg border border-neutral-200 bg-white">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-neutral-100 text-left text-xs text-neutral-500">
                  <th className="px-4 py-3 font-medium">Name</th>
                  <th className="px-4 py-3 font-medium">Parent</th>
                  <th className="px-4 py-3 font-medium">Created</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody>
                {isLoading ? (
                  <tr>
                    <td colSpan={4} className="px-4 py-8 text-center text-neutral-400">
                      Loading categories...
                    </td>
                  </tr>
                ) : categories.length === 0 ? (
                  <tr>
                    <td colSpan={4} className="px-4 py-8 text-center text-neutral-400">
                      No categories yet. Add one before creating products.
                    </td>
                  </tr>
                ) : (
                  categories.map((c) => (
                    <tr key={c.id} className="border-b border-neutral-100 last:border-0 hover:bg-neutral-50">
                      <td className="px-4 py-3 font-medium text-neutral-900">{c.name}</td>
                      <td className="px-4 py-3 text-neutral-500">
                        {c.parentId ? nameById.get(c.parentId) ?? `#${c.parentId}` : '—'}
                      </td>
                      <td className="px-4 py-3 text-xs text-neutral-400">
                        {new Date(c.createdAt).toLocaleDateString()}
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center justify-end gap-1">
                          <Can permission="catalog.update">
                            <button
                              onClick={() => setEditing(c)}
                              className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600"
                              title="Edit category"
                            >
                              <Pencil size={15} />
                            </button>
                          </Can>
                          <Can permission="catalog.delete">
                            <button
                              onClick={() => handleDelete(c)}
                              className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-danger"
                              title="Delete category"
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
        </div>
      </div>

      {showCreate && (
        <CategoryFormModal
          merchantId={merchantId}
          categories={categories}
          onClose={() => setShowCreate(false)}
          onSaved={() => {
            setShowCreate(false);
            fetchCategories();
          }}
        />
      )}

      {editing && (
        <CategoryFormModal
          merchantId={merchantId}
          categories={categories}
          category={editing}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null);
            fetchCategories();
          }}
        />
      )}
    </>
  );
}