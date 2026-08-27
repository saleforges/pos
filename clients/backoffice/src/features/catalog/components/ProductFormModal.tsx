import { useState } from 'react';
import { X } from 'lucide-react';
import { productsApi } from '../api/productsApi';
import { productItemsApi } from '../api/productItemsApi';
import { extractErrorMessage } from '../utils';
import { VariantEditor } from './VariantEditor';
import { emptyVariantRow, type VariantRow } from './variantRow';
import type { Category, ProductListResponse, Unit } from '../types';

interface ProductFormModalProps {
  merchantId: number;
  categories: Category[];
  units: Unit[];
  /** Present in edit mode. */
  product?: ProductListResponse | null;
  onClose: () => void;
  onSaved: () => void;
}

function rowsFromProduct(product: ProductListResponse): VariantRow[] {
  if (product.items.length === 0) return [emptyVariantRow()];
  return product.items.map(
    (item) =>
      ({
        key: `item-${item.id}`,
        itemId: item.id,
        name: item.name,
        sku: item.sku ?? '',
        unitId: item.unit ? String(item.unit.id) : '',
        price: String(item.price.amount),
        trackInventory: item.trackInventory,
      }) satisfies VariantRow,
  );
}

export function ProductFormModal({ merchantId, categories, units, product, onClose, onSaved }: ProductFormModalProps) {
  const isEdit = Boolean(product);
  const [categoryId, setCategoryId] = useState(product?.categoryId ? String(product.categoryId) : '');
  const [name, setName] = useState(product?.name ?? '');
  const [description, setDescription] = useState(product?.description ?? '');
  const [imageUrl, setImageUrl] = useState(product?.imageUrl ?? '');
  const [status, setStatus] = useState<string>(product?.status ?? 'active');
  const [rows, setRows] = useState<VariantRow[]>(
    product ? rowsFromProduct(product) : [emptyVariantRow()],
  );
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  const validate = (): boolean => {
    const e: Record<string, string> = {};
    if (!categoryId) e.categoryId = 'Category is required';
    if (!name.trim()) e.name = 'Name is required';
    if (rows.length === 0) e.variants = 'Add at least one variant';
    for (const row of rows) {
      if (!row.name.trim()) {
        e.variants = 'Every variant needs a name';
        break;
      }
      if (row.price === '' || !Number.isFinite(Number(row.price)) || Number(row.price) < 0) {
        e.variants = 'Every variant needs a valid price';
        break;
      }
    }
    setErrors(e);
    return Object.keys(e).length === 0;
  };

  const toVariantPayload = (row: VariantRow) => ({
    name: row.name.trim(),
    sku: row.sku.trim() || undefined,
    unitId: row.unitId ? Number(row.unitId) : undefined,
    price: Number(row.price),
    trackInventory: row.trackInventory,
  });

  const handleSubmit = async () => {
    if (!validate()) return;
    setSubmitting(true);
    setErrors({});
    try {
      if (isEdit && product) {
        await productsApi.update(merchantId, product.id, {
          categoryId: Number(categoryId),
          name: name.trim(),
          description: description.trim(),
          imageUrl: imageUrl.trim(),
          status,
        });

        const keptIds = new Set<number>();
        for (const row of rows) {
          if (row.itemId) {
            keptIds.add(row.itemId);
            await productItemsApi.update(merchantId, row.itemId, toVariantPayload(row));
          } else {
            await productItemsApi.create(merchantId, product.id, toVariantPayload(row));
          }
        }
        for (const item of product.items) {
          if (!keptIds.has(item.id)) {
            await productItemsApi.remove(merchantId, item.id);
          }
        }
      } else {
        // Bulk endpoint so SKU/unit/track-inventory land on the initial variants
        // (plain POST /products would create a bare default item).
        await productsApi.create(merchantId, {
          categoryId: Number(categoryId),
          name: name.trim(),
          description: description.trim() || undefined,
          imageUrl: imageUrl.trim() || undefined,
          items: rows.map(toVariantPayload),
        });
      }
      onSaved();
    } catch (err) {
      setErrors({ form: extractErrorMessage(err, 'Failed to save product.') });
    } finally {
      setSubmitting(false);
    }
  };

  const inputCls = (hasError?: string) =>
    `w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${
      hasError ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'
    }`;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="max-h-[90vh] w-full max-w-2xl overflow-y-auto rounded-xl bg-white p-6 shadow-xl">
        <div className="mb-5 flex items-center justify-between">
          <h2 className="font-display text-lg font-semibold text-neutral-900">
            {isEdit ? 'Edit Product' : 'Add Product'}
          </h2>
          <button onClick={onClose} className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600">
            <X size={18} />
          </button>
        </div>

        {errors.form && (
          <div className="mb-4 rounded-lg bg-danger/10 px-3 py-2 text-sm text-danger">{errors.form}</div>
        )}

        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="mb-1 block text-xs font-medium text-neutral-600">Category</label>
              <select
                value={categoryId}
                onChange={(e) => setCategoryId(e.target.value)}
                className={inputCls(errors.categoryId)}
              >
                <option value="">Select category...</option>
                {categories.map((c) => (
                  <option key={c.id} value={String(c.id)}>
                    {c.name}
                  </option>
                ))}
              </select>
              {errors.categoryId && <p className="mt-1 text-xs text-danger">{errors.categoryId}</p>}
            </div>

            {isEdit && (
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Status</label>
                <select
                  value={status}
                  onChange={(e) => setStatus(e.target.value)}
                  className={inputCls()}
                >
                  <option value="active">Active</option>
                  <option value="inactive">Inactive</option>
                  <option value="archived">Archived</option>
                </select>
              </div>
            )}
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-neutral-600">Name</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Cappuccino"
              className={inputCls(errors.name)}
            />
            {errors.name && <p className="mt-1 text-xs text-danger">{errors.name}</p>}
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-neutral-600">Description (optional)</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={2}
              className={inputCls()}
            />
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-neutral-600">
              Image URL (optional — upload support comes later)
            </label>
            <input
              type="text"
              value={imageUrl}
              onChange={(e) => setImageUrl(e.target.value)}
              placeholder="https://..."
              className={inputCls()}
            />
          </div>

          <VariantEditor rows={rows} onChange={setRows} units={units} />
          {errors.variants && <p className="text-xs text-danger">{errors.variants}</p>}

          <div className="flex justify-end gap-3 pt-2">
            <button
              onClick={onClose}
              className="rounded-lg border border-neutral-200 px-4 py-2 text-sm text-neutral-600 hover:bg-neutral-50"
            >
              Cancel
            </button>
            <button
              onClick={handleSubmit}
              disabled={submitting}
              className="rounded-lg bg-primary px-4 py-2 text-sm text-white hover:bg-primary-hover disabled:opacity-50"
            >
              {submitting ? 'Saving...' : isEdit ? 'Save Changes' : 'Add Product'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}