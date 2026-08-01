import { useState } from 'react';
import { X, Building2, User as UserIcon } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { accountsApi, type CreateUserRequest } from '../api/accountsApi';
import { merchantsApi, type CreateMerchantRequest } from '@/features/merchants/api/merchantsApi';
import { branchesApi, type CreateBranchRequest } from '@/features/merchants/api/branchesApi';
import { staffApi } from '@/features/merchants/api/staffApi';

const AVAILABLE_ROLES = ['owner', 'manager', 'supervisor', 'cashier', 'viewer'];

export function CreateAccountForm({
  onClose,
  onCreated,
}: {
  onClose: () => void;
  onCreated: () => void;
}) {
  const [form, setForm] = useState<CreateUserRequest>({
    username: '',
    email: '',
    password: '',
  });
  const [role, setRole] = useState('manager');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [merchant, setMerchant] = useState<CreateMerchantRequest>({
    name: '',
    email: '',
    phone: '',
    address: '',
  });
  const [branch, setBranch] = useState<CreateBranchRequest>({
    name: 'Main Branch',
    code: 'main',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  const isOwner = role === 'owner';

  const validate = () => {
    const e: Record<string, string> = {};
    if (!form.username.trim()) e.username = 'Required';
    if (!form.email.trim()) e.email = 'Required';
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) e.email = 'Invalid email';
    if (!form.password.trim()) e.password = 'Required';
    else if (form.password.length < 8) e.password = 'At least 8 characters';
    if (!confirmPassword.trim()) e.confirmPassword = 'Re-enter password';
    else if (form.password !== confirmPassword) e.confirmPassword = 'Passwords do not match';
    if (!role) e.roles = 'Select a role';
    if (isOwner) {
      if (!merchant.name?.trim()) e.merchantName = 'Required';
      if (!merchant.email?.trim()) e.merchantEmail = 'Required';
      else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(merchant.email ?? '')) e.merchantEmail = 'Invalid email';
      if (!branch.name.trim()) e.branchName = 'Required';
      if (!branch.code.trim()) e.branchCode = 'Required';
    }
    setErrors(e);
    return Object.keys(e).length === 0;
  };

  const handleSubmit = async () => {
    if (!validate()) return;
    setSubmitting(true);
    setErrors({});
    try {
      await accountsApi.create({ ...form, roles: [role] });

      if (isOwner) {
        const users = await accountsApi.list();
        const newUser = users.find((u) => u.username === form.username);
        if (!newUser) throw new Error('Created user not found');

        const createdMerchant = await merchantsApi.create({
          ...merchant,
          email: merchant.email ?? form.email,
        });
        const createdBranch = await branchesApi.create(createdMerchant.id, branch);
        await staffApi.assign(createdMerchant.id, {
          userId: newUser.id,
          branchId: createdBranch.id,
          role: 'owner',
          isDefault: true,
        });
      }

      onCreated();
    } catch {
      setErrors({ form: 'Failed to create account.' });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-xl bg-white p-6 shadow-xl">
        <div className="mb-5 flex items-center justify-between">
          <h2 className="font-display text-lg font-semibold text-neutral-900">Create Account</h2>
          <button onClick={onClose} className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600">
            <X size={18} />
          </button>
        </div>

        {errors.form && (
          <div className="mb-4 rounded-lg bg-danger/10 px-3 py-2 text-sm text-danger">{errors.form}</div>
        )}

        <div className="mb-2 flex items-center gap-2">
          <UserIcon size={16} className="text-neutral-400" />
          <span className="text-sm font-medium text-neutral-700">Account Details</span>
        </div>

        <div className="space-y-4">
          <div>
            <label className="mb-1 block text-xs font-medium text-neutral-600">Username</label>
            <input
              type="text"
              value={form.username}
              onChange={(e) => setForm({ ...form, username: e.target.value })}
              className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${errors.username ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`}
            />
            {errors.username && <p className="mt-1 text-xs text-danger">{errors.username}</p>}
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

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="mb-1 block text-xs font-medium text-neutral-600">Password</label>
              <input
                type="password"
                value={form.password}
                onChange={(e) => setForm({ ...form, password: e.target.value })}
                className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${errors.password ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`}
              />
              {errors.password && <p className="mt-1 text-xs text-danger">{errors.password}</p>}
            </div>

            <div>
              <label className="mb-1 block text-xs font-medium text-neutral-600">Confirm Password</label>
              <input
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${errors.confirmPassword ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`}
              />
              {errors.confirmPassword && <p className="mt-1 text-xs text-danger">{errors.confirmPassword}</p>}
            </div>
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-neutral-600">Role</label>
            <select
              value={role}
              onChange={(e) => setRole(e.target.value)}
              className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${errors.roles ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`}
            >
              {AVAILABLE_ROLES.map((r) => (
                <option key={r} value={r}>{r}</option>
              ))}
            </select>
            {errors.roles && <p className="mt-1 text-xs text-danger">{errors.roles}</p>}
          </div>
        </div>

        {isOwner && (
          <>
            <div className="mb-2 mt-6 flex items-center gap-2">
              <Building2 size={16} className="text-neutral-400" />
              <span className="text-sm font-medium text-neutral-700">Merchant &amp; Branch</span>
            </div>
            <p className="mb-3 text-xs text-neutral-500">
              An owner role also creates a merchant, its first branch, and assigns the account as owner.
            </p>

            <div className="space-y-4">
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Merchant Name</label>
                <input
                  type="text"
                  value={merchant.name ?? ''}
                  onChange={(e) => setMerchant({ ...merchant, name: e.target.value })}
                  className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${errors.merchantName ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`}
                />
                {errors.merchantName && <p className="mt-1 text-xs text-danger">{errors.merchantName}</p>}
              </div>

              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Merchant Email</label>
                <input
                  type="email"
                  value={merchant.email ?? form.email}
                  onChange={(e) => setMerchant({ ...merchant, email: e.target.value })}
                  className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${errors.merchantEmail ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`}
                />
                {errors.merchantEmail && <p className="mt-1 text-xs text-danger">{errors.merchantEmail}</p>}
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="mb-1 block text-xs font-medium text-neutral-600">Merchant Phone</label>
                  <input
                    type="text"
                    value={merchant.phone ?? ''}
                    onChange={(e) => setMerchant({ ...merchant, phone: e.target.value })}
                    className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400"
                  />
                </div>
                <div>
                  <label className="mb-1 block text-xs font-medium text-neutral-600">Merchant Address</label>
                  <input
                    type="text"
                    value={merchant.address ?? ''}
                    onChange={(e) => setMerchant({ ...merchant, address: e.target.value })}
                    className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="mb-1 block text-xs font-medium text-neutral-600">Branch Name</label>
                  <input
                    type="text"
                    value={branch.name}
                    onChange={(e) => setBranch({ ...branch, name: e.target.value })}
                    className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${errors.branchName ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`}
                  />
                  {errors.branchName && <p className="mt-1 text-xs text-danger">{errors.branchName}</p>}
                </div>
                <div>
                  <label className="mb-1 block text-xs font-medium text-neutral-600">Branch Code</label>
                  <input
                    type="text"
                    value={branch.code}
                    onChange={(e) => setBranch({ ...branch, code: e.target.value })}
                    className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${errors.branchCode ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`}
                  />
                  {errors.branchCode && <p className="mt-1 text-xs text-danger">{errors.branchCode}</p>}
                </div>
              </div>
            </div>
          </>
        )}

        <div className="mt-6 flex justify-end gap-2">
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button onClick={handleSubmit} disabled={submitting}>
            {submitting ? 'Creating…' : 'Create Account'}
          </Button>
        </div>
      </div>
    </div>
  );
}
