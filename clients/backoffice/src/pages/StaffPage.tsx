import { useEffect, useState } from 'react';
import { Search, Plus, X, Store } from 'lucide-react';
import { staffApi, type StaffMemberResponse } from '@/features/staff/api/staffApi';
import { merchantsApi, type MerchantResponse } from '@/features/merchants/api/merchantsApi';
import { branchesApi, type BranchResponse } from '@/features/branches/api/branchesApi';
import { usersApi, type UserResponse } from '@/features/users/api/usersApi';

export default function StaffPage() {
  const [merchants, setMerchants] = useState<MerchantResponse[]>([]);
  const [selectedMerchantId, setSelectedMerchantId] = useState<number | null>(null);
  const [staff, setStaff] = useState<StaffMemberResponse[]>([]);
  const [users, setUsers] = useState<UserResponse[]>([]);
  const [branches, setBranches] = useState<BranchResponse[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [showModal, setShowModal] = useState(false);

  const [form, setForm] = useState({ userId: '', branchId: '', role: 'cashier', isDefault: false });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    merchantsApi.list().then(setMerchants).catch(() => {});
    usersApi.list().then(setUsers).catch(() => {});
  }, []);

  const fetchStaff = (merchantId: number) => {
    setIsLoading(true);
    Promise.all([
      staffApi.list(merchantId),
      branchesApi.list(merchantId),
    ])
      .then(([staffData, branchData]) => {
        setStaff(staffData);
        setBranches(branchData);
      })
      .catch(() => {})
      .finally(() => setIsLoading(false));
  };

  useEffect(() => {
    if (selectedMerchantId) fetchStaff(selectedMerchantId);
    else setIsLoading(false);
  }, [selectedMerchantId]);

  const userMap = Object.fromEntries(users.map((u) => [u.id, u]));
  const branchMap = Object.fromEntries(branches.map((b) => [b.id, b]));

  const filtered = staff.filter((s) => {
    const user = userMap[s.userId];
    const branch = branchMap[s.branchId];
    const q = search.toLowerCase();
    return (
      (user?.username?.toLowerCase() || '').includes(q) ||
      (branch?.name?.toLowerCase() || '').includes(q) ||
      s.role.toLowerCase().includes(q)
    );
  });

  const openModal = () => {
    const firstBranch = branches[0];
    setForm({ userId: '', branchId: firstBranch ? String(firstBranch.id) : '', role: 'cashier', isDefault: false });
    setErrors({});
    setShowModal(true);
  };

  const validate = () => {
    const e: Record<string, string> = {};
    if (!form.userId) e.userId = 'User is required';
    if (!form.branchId) e.branchId = 'Branch is required';
    if (!form.role) e.role = 'Role is required';
    setErrors(e);
    return Object.keys(e).length === 0;
  };

  const handleSubmit = async () => {
    if (!validate() || !selectedMerchantId) return;
    setSubmitting(true);
    try {
      await staffApi.assign(selectedMerchantId, {
        userId: Number(form.userId),
        branchId: Number(form.branchId),
        role: form.role,
        isDefault: form.isDefault,
      });
      setShowModal(false);
      fetchStaff(selectedMerchantId);
    } catch {
      setErrors({ form: 'Failed to assign staff. This user may already be assigned to this branch.' });
    } finally {
      setSubmitting(false);
    }
  };

  if (!selectedMerchantId) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="w-full max-w-md">
          <div className="mb-8 text-center">
            <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-xl bg-secondary/10">
              <Store size={28} className="text-secondary" />
            </div>
            <h2 className="font-display text-lg font-semibold text-neutral-900">Select a Merchant</h2>
            <p className="mt-1 text-sm text-neutral-500">Choose a merchant to view its staff.</p>
          </div>
          <div className="space-y-3">
            {merchants.map((m) => (
              <button
                key={m.id}
                onClick={() => setSelectedMerchantId(m.id)}
                className="w-full rounded-lg border border-neutral-200 bg-white p-4 text-left transition-colors hover:border-secondary hover:shadow-sm"
              >
                <span className="font-medium text-neutral-900">{m.name}</span>
                <p className="mt-0.5 text-sm text-neutral-500">{m.email}</p>
              </button>
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <>
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="relative flex-1 max-w-xs">
              <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-400" />
              <input
                type="text"
                placeholder="Search staff..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-full rounded-lg border border-neutral-200 bg-white py-2 pl-9 pr-3 text-sm outline-none focus:border-neutral-400"
              />
            </div>
          </div>
          <div className="flex items-center gap-3">
            <select
              value={selectedMerchantId}
              onChange={(e) => setSelectedMerchantId(Number(e.target.value))}
              className="rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400"
            >
              {merchants.map((m) => (
                <option key={m.id} value={m.id}>{m.name}</option>
              ))}
            </select>
            <button
              onClick={openModal}
              disabled={branches.length === 0}
              className="flex items-center gap-2 rounded-lg bg-primary px-3 py-2 text-sm text-white hover:bg-primary-hover disabled:opacity-50"
            >
              <Plus size={16} />
              Assign Staff
            </button>
          </div>
        </div>

        <div className="rounded-lg border border-neutral-200 bg-white">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-neutral-100 text-left text-xs text-neutral-500">
                  <th className="px-4 py-3 font-medium">ID</th>
                  <th className="px-4 py-3 font-medium">User</th>
                  <th className="px-4 py-3 font-medium">Branch</th>
                  <th className="px-4 py-3 font-medium">Role</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium">Default</th>
                  <th className="px-4 py-3 font-medium">Created At</th>
                </tr>
              </thead>
              <tbody>
                {isLoading ? (
                  <tr><td colSpan={7} className="px-4 py-8 text-center text-sm text-neutral-400">Loading...</td></tr>
                ) : filtered.length === 0 ? (
                  <tr><td colSpan={7} className="px-4 py-8 text-center text-sm text-neutral-400">No staff found</td></tr>
                ) : (
                  filtered.map((s) => {
                    const user = userMap[s.userId];
                    const branch = branchMap[s.branchId];
                    return (
                      <tr key={s.id} className="border-b border-neutral-100 last:border-0 hover:bg-neutral-50">
                        <td className="px-4 py-3 font-mono text-xs text-neutral-500">{s.id}</td>
                        <td className="px-4 py-3 font-medium text-neutral-900">
                          {user ? `${user.username} (${user.email})` : `User #${s.userId}`}
                        </td>
                        <td className="px-4 py-3 text-neutral-600">{branch?.name || `Branch #${s.branchId}`}</td>
                        <td className="px-4 py-3">
                          <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                            s.role === 'manager' ? 'bg-info/10 text-info-hover' : 'bg-neutral-100 text-neutral-600'
                          }`}>
                            {s.role}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                            s.status === 'active' ? 'bg-secondary/10 text-secondary' : 'bg-neutral-100 text-neutral-500'
                          }`}>
                            {s.status}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-neutral-500">{s.isDefault ? 'Yes' : 'No'}</td>
                        <td className="px-4 py-3 text-xs text-neutral-400">{new Date(s.createdAt).toLocaleDateString()}</td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <div className="flex items-center justify-between mb-5">
              <h2 className="font-display text-lg font-semibold text-neutral-900">Assign Staff</h2>
              <button onClick={() => setShowModal(false)} className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600">
                <X size={18} />
              </button>
            </div>

            {errors.form && (
              <div className="mb-4 rounded-lg bg-danger/10 px-3 py-2 text-sm text-danger">{errors.form}</div>
            )}

            <div className="space-y-4">
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">User *</label>
                <select
                  value={form.userId}
                  onChange={(e) => setForm({ ...form, userId: e.target.value })}
                  className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${
                    errors.userId ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'
                  }`}
                >
                  <option value="">— Select user —</option>
                  {users.map((u) => (
                    <option key={u.id} value={u.id}>{u.username} ({u.email})</option>
                  ))}
                </select>
                {errors.userId && <p className="mt-1 text-xs text-danger">{errors.userId}</p>}
              </div>

              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Branch *</label>
                <select
                  value={form.branchId}
                  onChange={(e) => setForm({ ...form, branchId: e.target.value })}
                  className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${
                    errors.branchId ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'
                  }`}
                >
                  <option value="">— Select branch —</option>
                  {branches.map((b) => (
                    <option key={b.id} value={b.id}>{b.name} ({b.code})</option>
                  ))}
                </select>
                {errors.branchId && <p className="mt-1 text-xs text-danger">{errors.branchId}</p>}
              </div>

              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Role *</label>
                <select
                  value={form.role}
                  onChange={(e) => setForm({ ...form, role: e.target.value })}
                  className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400"
                >
                  <option value="manager">Manager</option>
                  <option value="supervisor">Supervisor</option>
                  <option value="cashier">Cashier</option>
                  <option value="viewer">Viewer</option>
                </select>
              </div>

              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={form.isDefault}
                  onChange={(e) => setForm({ ...form, isDefault: e.target.checked })}
                  className="rounded border-neutral-300 text-secondary focus:ring-secondary"
                />
                <span className="text-sm text-neutral-600">Set as default branch</span>
              </label>

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
                  {submitting ? 'Assigning...' : 'Assign Staff'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
