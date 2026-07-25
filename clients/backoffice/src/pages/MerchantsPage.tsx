import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { Search, Plus, X } from 'lucide-react';
import { merchantsApi, type MerchantResponse } from '@/features/merchants/api/merchantsApi';
import { branchesApi } from '@/features/branches/api/branchesApi';
import { staffApi } from '@/features/staff/api/staffApi';
import { usersApi, type UserResponse } from '@/features/users/api/usersApi';

export default function MerchantsPage() {
  const [merchants, setMerchants] = useState<MerchantResponse[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [users, setUsers] = useState<UserResponse[]>([]);

  const [form, setForm] = useState({
    name: '', legalName: '', address: '', phone: '', email: '', taxId: '', ownerId: '',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  const fetchMerchants = () => {
    setIsLoading(true);
    merchantsApi.list().then(setMerchants).finally(() => setIsLoading(false));
  };

  useEffect(() => {
    fetchMerchants();
    usersApi.list().then(setUsers).catch(() => {});
  }, []);

  const filtered = merchants.filter(
    (m) =>
      m.name.toLowerCase().includes(search.toLowerCase()) ||
      m.email.toLowerCase().includes(search.toLowerCase()),
  );

  const openModal = () => {
    setForm({ name: '', legalName: '', address: '', phone: '', email: '', taxId: '', ownerId: '' });
    setErrors({});
    setShowModal(true);
  };

  const validate = () => {
    const e: Record<string, string> = {};
    if (!form.name.trim()) e.name = 'Name is required';
    if (!form.email.trim()) e.email = 'Email is required';
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) e.email = 'Invalid email format';
    setErrors(e);
    return Object.keys(e).length === 0;
  };

  const handleSubmit = async () => {
    if (!validate()) return;
    setSubmitting(true);
    try {
      const merchant = await merchantsApi.create({
        name: form.name.trim(),
        legalName: form.legalName.trim() || undefined,
        address: form.address.trim() || undefined,
        phone: form.phone.trim() || undefined,
        email: form.email.trim(),
        taxId: form.taxId.trim() || undefined,
      });

      if (form.ownerId) {
        const branch = await branchesApi.create(merchant.id, {
          name: 'Main Branch',
          code: 'main',
        });
        await staffApi.assign(merchant.id, {
          userId: Number(form.ownerId),
          branchId: branch.id,
          role: 'manager',
          isDefault: true,
        });
      }

      setShowModal(false);
      fetchMerchants();
    } catch {
      setErrors({ form: 'Failed to create merchant. It may already exist.' });
    } finally {
      setSubmitting(false);
    }
  };

  if (isLoading) return <div>Loading merchants...</div>;

  return (
    <>
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="relative flex-1 max-w-xs">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-400" />
            <input
              type="text"
              placeholder="Search merchants..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full rounded-lg border border-neutral-200 bg-white py-2 pl-9 pr-3 text-sm outline-none focus:border-neutral-400"
            />
          </div>
          <button
            onClick={openModal}
            className="flex items-center gap-2 rounded-lg bg-primary px-3 py-2 text-sm text-white hover:bg-primary-hover"
          >
            <Plus size={16} />
            Add Merchant
          </button>
        </div>

        <div className="rounded-lg border border-neutral-200 bg-white">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-neutral-100 text-left text-xs text-neutral-500">
                  <th className="px-4 py-3 font-medium">ID</th>
                  <th className="px-4 py-3 font-medium">Name</th>
                  <th className="px-4 py-3 font-medium">Legal Name</th>
                  <th className="px-4 py-3 font-medium">Email</th>
                  <th className="px-4 py-3 font-medium">Phone</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium">Created At</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((merchant) => (
                  <tr key={merchant.id} className="border-b border-neutral-100 last:border-0 hover:bg-neutral-50">
                    <td className="px-4 py-3 font-mono text-xs text-neutral-500">{merchant.id}</td>
                    <td className="px-4 py-3 font-medium">
                      <Link to={`/merchants/${merchant.id}`} className="text-secondary hover:underline">
                        {merchant.name}
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-neutral-600">{merchant.legalName || '—'}</td>
                    <td className="px-4 py-3 text-neutral-600">{merchant.email}</td>
                    <td className="px-4 py-3 text-neutral-500">{merchant.phone || '—'}</td>
                    <td className="px-4 py-3">
                      <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                        merchant.status === 'active' ? 'bg-secondary/10 text-secondary' : 'bg-neutral-100 text-neutral-500'
                      }`}>
                        {merchant.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-xs text-neutral-400">{new Date(merchant.createdAt).toLocaleDateString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <div className="flex items-center justify-between mb-5">
              <h2 className="font-display text-lg font-semibold text-neutral-900">Create Merchant</h2>
              <button onClick={() => setShowModal(false)} className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600">
                <X size={18} />
              </button>
            </div>

            {errors.form && (
              <div className="mb-4 rounded-lg bg-danger/10 px-3 py-2 text-sm text-danger">{errors.form}</div>
            )}

            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="mb-1 block text-xs font-medium text-neutral-600">Name *</label>
                  <input
                    type="text"
                    value={form.name}
                    onChange={(e) => setForm({ ...form, name: e.target.value })}
                    className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${
                      errors.name ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'
                    }`}
                  />
                  {errors.name && <p className="mt-1 text-xs text-danger">{errors.name}</p>}
                </div>
                <div>
                  <label className="mb-1 block text-xs font-medium text-neutral-600">Legal Name</label>
                  <input
                    type="text"
                    value={form.legalName}
                    onChange={(e) => setForm({ ...form, legalName: e.target.value })}
                    className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400"
                  />
                </div>
              </div>

              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Email *</label>
                <input
                  type="email"
                  value={form.email}
                  onChange={(e) => setForm({ ...form, email: e.target.value })}
                  className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${
                    errors.email ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'
                  }`}
                />
                {errors.email && <p className="mt-1 text-xs text-danger">{errors.email}</p>}
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="mb-1 block text-xs font-medium text-neutral-600">Phone</label>
                  <input
                    type="text"
                    value={form.phone}
                    onChange={(e) => setForm({ ...form, phone: e.target.value })}
                    className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400"
                  />
                </div>
                <div>
                  <label className="mb-1 block text-xs font-medium text-neutral-600">Tax ID</label>
                  <input
                    type="text"
                    value={form.taxId}
                    onChange={(e) => setForm({ ...form, taxId: e.target.value })}
                    className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400"
                  />
                </div>
              </div>

              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Address</label>
                <input
                  type="text"
                  value={form.address}
                  onChange={(e) => setForm({ ...form, address: e.target.value })}
                  className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400"
                />
              </div>

              {users.length > 0 && (
                <div>
                  <label className="mb-1 block text-xs font-medium text-neutral-600">Owner</label>
                  <select
                    value={form.ownerId}
                    onChange={(e) => setForm({ ...form, ownerId: e.target.value })}
                    className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400"
                  >
                    <option value="">— Select owner —</option>
                    {users.map((u) => (
                      <option key={u.id} value={u.id}>{u.username} ({u.email})</option>
                    ))}
                  </select>
                  <p className="mt-1 text-xs text-neutral-400">Owner will be assigned as staff with default branch access.</p>
                </div>
              )}

              <div className="flex justify-end gap-3 pt-2">
                <button
                  onClick={() => setShowModal(false)}
                  className="rounded-lg border border-neutral-200 px-4 py-2 text-sm text-neutral-600 hover:bg-neutral-50"
                >
                  Cancel
                </button>
                <button
                  onClick={handleSubmit}
                  disabled={submitting}
                  className="rounded-lg bg-primary px-4 py-2 text-sm text-white hover:bg-primary-hover disabled:opacity-50"
                >
                  {submitting ? 'Creating...' : 'Create Merchant'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
