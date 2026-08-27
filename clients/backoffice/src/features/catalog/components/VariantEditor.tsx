import { Plus, X } from 'lucide-react';
import type { Unit } from '../types';
import type { VariantRow } from './variantRow';
import { emptyVariantRow } from './variantRow';

interface VariantEditorProps {
  rows: VariantRow[];
  onChange: (rows: VariantRow[]) => void;
  units: Unit[];
}

export function VariantEditor({ rows, onChange, units }: VariantEditorProps) {
  const update = (key: string, patch: Partial<VariantRow>) =>
    onChange(rows.map((r) => (r.key === key ? { ...r, ...patch } : r)));

  const remove = (key: string) => onChange(rows.filter((r) => r.key !== key));

  return (
    <div className="space-y-2">
      <label className="block text-xs font-medium text-neutral-600">Variants</label>

      <div className="grid grid-cols-[1fr_1fr_1fr_1fr_auto_auto] items-center gap-2 px-1 text-xs font-medium text-neutral-400">
        <span>Name</span>
        <span>SKU</span>
        <span>Unit</span>
        <span>Price</span>
        <span className="text-center">Track stock</span>
        <span />
      </div>

      {rows.map((row) => (
        <div key={row.key} className="grid grid-cols-[1fr_1fr_1fr_1fr_auto_auto] items-center gap-2">
          <input
            type="text"
            placeholder="e.g. Large"
            value={row.name}
            onChange={(e) => update(row.key, { name: e.target.value })}
            className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400"
          />
          <input
            type="text"
            placeholder="SKU"
            value={row.sku}
            onChange={(e) => update(row.key, { sku: e.target.value })}
            className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400"
          />
          <select
            value={row.unitId}
            onChange={(e) => update(row.key, { unitId: e.target.value })}
            className="w-full rounded-lg border border-neutral-200 bg-white px-2 py-2 text-sm outline-none focus:border-neutral-400"
          >
            <option value="">—</option>
            {units.map((u) => (
              <option key={u.id} value={String(u.id)}>
                {u.name}
              </option>
            ))}
          </select>
          <input
            type="number"
            min="0"
            step="any"
            placeholder="0"
            value={row.price}
            onChange={(e) => update(row.key, { price: e.target.value })}
            className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400"
          />
          <div className="flex justify-center">
            <input
              type="checkbox"
              checked={row.trackInventory}
              onChange={(e) => update(row.key, { trackInventory: e.target.checked })}
              className="rounded border-neutral-300 text-secondary focus:ring-secondary"
            />
          </div>
          <button
            type="button"
            onClick={() => remove(row.key)}
            disabled={rows.length === 1}
            className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-danger disabled:cursor-not-allowed disabled:opacity-30"
            title={rows.length === 1 ? 'At least one variant is required' : 'Remove variant'}
          >
            <X size={16} />
          </button>
        </div>
      ))}

      {rows.length === 0 && (
        <p className="text-xs text-danger">Add at least one variant.</p>
      )}

      <button
        type="button"
        onClick={() => onChange([...rows, emptyVariantRow()])}
        className="flex items-center gap-1 rounded-lg border border-dashed border-neutral-300 px-3 py-1.5 text-xs text-neutral-500 hover:border-neutral-400 hover:text-neutral-700"
      >
        <Plus size={14} />
        Add variant
      </button>
    </div>
  );
}