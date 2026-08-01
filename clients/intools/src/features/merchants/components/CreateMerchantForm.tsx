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
    legalName: '',
    address: '',
    phone: '',
    taxId: '',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  const validate = () => {
    const e: Record<string, string> = {};
    if (!form.name.trim()) e.name = 'Required';
    if (!form.email.trim()) e.email = 'Required';
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) e.email = 'Invalid email';
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
      <div className="w-full max-w-lg rounded-xl bg-white p-6 shadow-xl">
        <div className="mb-5 flex items-center justify-between">
          <h2 className="font-display text-lg font-semibold text-neutral-900">Create Merchant</h2>
          <button onClick={onClose} className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600">
            <X size={18} />
          </button>
        </div>

        {errors.form && (
          <div className="mb-4 rounded-lg bg-danger/10 px-3 py-2 text-sm text-danger">{errors.form}</div>
        )}

        <div className="grid grid-cols-2 gap-4">
          <div className="col-span-2">
            <label className="mb-1 block text-xs font-medium text-neutral-600">Merchant Name</label>
            <input
              type="text"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${errors.name ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`}
            />
            {errors.name && <p className="mt-1 text-xs text-danger">{errors.name}</p>}
          </div>

          <div className="col-span-2">
            <label className="mb-1 block text-xs font-medium text-neutral-600">Email</label>
            <input
              type="email"
              value={form.email}
              onChange={(e) => setForm({ ...form, email: e.target.value })}
              className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${errors.email ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`}
            />
            {errors.email && <p className="mt-1 text-xs text-danger">{errors.email}</p>}
          </div>

          <div className="col-span-2">
            <label className="mb-1 block text-xs font-medium text-neutral-600">Legal Name</label>
            <input
              type="text"
              value={form.legalName ?? ''}
              onChange={(e) => setForm({ ...form, legalName: e.target.value })}
              className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400"
            />
          </div>

          <div className="col-span-2">
            <label className="mb-1 block text-xs font-medium text-neutral-600">Address</label>
            <input
              type="text"
              value={form.address ?? ''}
              onChange={(e) => setForm({ ...form, address: e.target.value })}
              className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400"
            />
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-neutral-600">Phone</label>
            <input
              type="text"
              value={form.phone ?? ''}
              onChange={(e) => setForm({ ...form, phone: e.target.value })}
              className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400"
            />
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-neutral-600">Tax ID</label>
            <input
              type="text"
              value={form.taxId ?? ''}
              onChange={(e) => setForm({ ...form, taxId: e.target.value })}
              className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400"
            />
          </div>
        </div>

        <div className="mt-6 flex justify-end gap-2">
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button onClick={handleSubmit} disabled={submitting}>
            {submitting ? 'Creating…' : 'Create Merchant'}
          </Button>
        </div>
      </div>
    </div>
  );
}
