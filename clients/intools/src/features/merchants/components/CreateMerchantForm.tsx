import { useState } from 'react';
import { X } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { merchantsApi, type CreateMerchantRequest, type Merchant } from '../api/merchantsApi';

export function CreateMerchantForm({
  onClose,
  onCreated,
}: {
  onClose: () => void;
  onCreated: (merchant: Merchant) => void;
}) {
  const [form, setForm] = useState<CreateMerchantRequest>({
    name: '',
    email: '',
    inventoryScoping: false,
  });
  const [createOwner, setCreateOwner] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  const validate = () => {
    const e: Record<string, string> = {};
    if (!form.name.trim()) e.name = 'Required';
    if (!form.email.trim()) e.email = 'Required';
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) e.email = 'Invalid email';
    if (createOwner && !form.ownerUsername?.trim()) e.ownerUsername = 'Required';
    if (createOwner && !form.ownerPassword?.trim()) e.ownerPassword = 'Required';
    setErrors(e);
    return Object.keys(e).length === 0;
  };

  const handleSubmit = async () => {
    if (!validate()) return;
    setSubmitting(true);
    try {
      const merchant = await merchantsApi.create(form);
      onCreated(merchant);
    } catch {
      setErrors({ form: 'Failed to create merchant.' });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
        <div className="mb-5 flex items-center justify-between">
          <h2 className="font-display text-lg font-semibold text-neutral-900">Create Merchant</h2>
          <button onClick={onClose} className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600">
            <X size={18} />
          </button>
        </div>

        {errors.form && (
          <div className="mb-4 rounded-lg bg-danger/10 px-3 py-2 text-sm text-danger">{errors.form}</div>
        )}

        <div className="space-y-4">
          <div>
            <label className="mb-1 block text-xs font-medium text-neutral-600">Merchant Name</label>
            <input
              type="text"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${errors.name ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`}
            />
            {errors.name && <p className="mt-1 text-xs text-danger">{errors.name}</p>}
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-neutral-600">Email</label>
            <input
              type="email"
              value={form.email}
              onChange={(e) => setForm({ ...form, email: e.target.value })}
              className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${errors.email ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`}
            />
            {errors.email && <p className="mt-1 text-xs text-danger">{errors.email}</p>}
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="inventoryScoping"
              checked={form.inventoryScoping}
              onChange={(e) => setForm({ ...form, inventoryScoping: e.target.checked })}
              className="rounded border-neutral-300"
            />
            <label htmlFor="inventoryScoping" className="text-sm text-neutral-700">
              Enable inventory scoping
            </label>
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="createOwner"
              checked={createOwner}
              onChange={(e) => setCreateOwner(e.target.checked)}
              className="rounded border-neutral-300"
            />
            <label htmlFor="createOwner" className="text-sm text-neutral-700">
              Create owner account
            </label>
          </div>

          {createOwner && (
            <div className="space-y-3 rounded-lg border border-neutral-200 bg-neutral-50 p-3">
              <p className="text-xs font-medium text-neutral-600">Owner Account Details</p>
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Username</label>
                <input
                  type="text"
                  value={form.ownerUsername ?? ''}
                  onChange={(e) => setForm({ ...form, ownerUsername: e.target.value })}
                  className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${errors.ownerUsername ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`}
                />
                {errors.ownerUsername && <p className="mt-1 text-xs text-danger">{errors.ownerUsername}</p>}
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Password</label>
                <input
                  type="password"
                  value={form.ownerPassword ?? ''}
                  onChange={(e) => setForm({ ...form, ownerPassword: e.target.value })}
                  className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${errors.ownerPassword ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`}
                />
                {errors.ownerPassword && <p className="mt-1 text-xs text-danger">{errors.ownerPassword}</p>}
              </div>
            </div>
          )}

          <div className="flex justify-end gap-2 pt-2">
            <Button variant="secondary" onClick={onClose}>Cancel</Button>
            <Button onClick={handleSubmit} disabled={submitting}>
              {submitting ? 'Creating…' : 'Create Merchant'}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
