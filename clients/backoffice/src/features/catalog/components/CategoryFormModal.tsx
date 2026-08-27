import { useState } from 'react';
import { X } from 'lucide-react';
import { ApiError } from '@/lib/api';
import { categoriesApi } from '../api/categoriesApi';
import { extractErrorMessage } from '../utils';
import type { Category } from '../types';

interface CategoryFormModalProps {
  merchantId: number;
  categories: Category[];
  /** Present in edit mode. */
  category?: Category | null;
  onClose: () => void;
  onSaved: () => void;
}

export function CategoryFormModal({ merchantId, categories, category, onClose, onSaved }: CategoryFormModalProps) {
  const isEdit = Boolean(category);
  const [name, setName] = useState(category?.name ?? '');
  const [parentId, setParentId] = useState(category?.parentId ? String(category.parentId) : '');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const parentOptions = categories.filter((c) => c.id !== category?.id);

  const handleSubmit = async () => {
    if (!name.trim()) {
      setError('Name is required');
      return;
    }
    setSubmitting(true);
    setError('');
    try {
      if (isEdit && category) {
        await categoriesApi.update(merchantId, category.id, {
          name: name.trim(),
          parentId: parentId ? Number(parentId) : undefined,
        });
      } else {
        await categoriesApi.create(merchantId, {
          name: name.trim(),
          parentId: parentId ? Number(parentId) : undefined,
        });
      }
      onSaved();
    } catch (err) {
      setError(extractErrorMessage(err, 'Failed to save category.'));
      if (err instanceof ApiError && err.status === 403) {
        setError('You do not have permission to manage categories.');
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
        <div className="mb-5 flex items-center justify-between">
          <h2 className="font-display text-lg font-semibold text-neutral-900">
            {isEdit ? 'Edit Category' : 'Add Category'}
          </h2>
          <button onClick={onClose} className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600">
            <X size={18} />
          </button>
        </div>

        {error && (
          <div className="mb-4 rounded-lg bg-danger/10 px-3 py-2 text-sm text-danger">{error}</div>
        )}

        <div className="space-y-4">
          <div>
            <label className="mb-1 block text-xs font-medium text-neutral-600">Name</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Coffee"
              className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${
                error && !name.trim() ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'
              }`}
            />
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-neutral-600">Parent category (optional)</label>
            <select
              value={parentId}
              onChange={(e) => setParentId(e.target.value)}
              className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400"
            >
              <option value="">— None —</option>
              {parentOptions.map((c) => (
                <option key={c.id} value={String(c.id)}>
                  {c.name}
                </option>
              ))}
            </select>
          </div>

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
              {submitting ? 'Saving...' : isEdit ? 'Save Changes' : 'Add Category'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}