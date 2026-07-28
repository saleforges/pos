import { useState } from 'react';
import { X } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { accountsApi, type CreateAccountRequest, type Account } from '../api/accountsApi';

export function CreateAccountForm({
  onClose,
  onCreated,
}: {
  onClose: () => void;
  onCreated: (account: Account) => void;
}) {
  const [form, setForm] = useState<CreateAccountRequest>({
    username: '',
    email: '',
    password: '',
    role: 'admin',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  const validate = () => {
    const e: Record<string, string> = {};
    if (!form.username.trim()) e.username = 'Required';
    if (!form.email.trim()) e.email = 'Required';
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) e.email = 'Invalid email';
    if (!form.password.trim()) e.password = 'Required';
    else if (form.password.length < 8) e.password = 'At least 8 characters';
    setErrors(e);
    return Object.keys(e).length === 0;
  };

  const handleSubmit = async () => {
    if (!validate()) return;
    setSubmitting(true);
    try {
      const account = await accountsApi.create(form);
      onCreated(account);
    } catch {
      setErrors({ form: 'Failed to create account.' });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
        <div className="mb-5 flex items-center justify-between">
          <h2 className="font-display text-lg font-semibold text-neutral-900">Create Account</h2>
          <button onClick={onClose} className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600">
            <X size={18} />
          </button>
        </div>

        {errors.form && (
          <div className="mb-4 rounded-lg bg-danger/10 px-3 py-2 text-sm text-danger">{errors.form}</div>
        )}

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

          <div>
            <label className="mb-1 block text-xs font-medium text-neutral-600">Temporary Password</label>
            <input
              type="password"
              value={form.password}
              onChange={(e) => setForm({ ...form, password: e.target.value })}
              className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${errors.password ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`}
            />
            {errors.password && <p className="mt-1 text-xs text-danger">{errors.password}</p>}
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-neutral-600">Role</label>
            <select
              value={form.role}
              onChange={(e) => setForm({ ...form, role: e.target.value })}
              className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400"
            >
              <option value="superadmin">Superadmin</option>
              <option value="admin">Admin</option>
              <option value="support">Support</option>
            </select>
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <Button variant="secondary" onClick={onClose}>Cancel</Button>
            <Button onClick={handleSubmit} disabled={submitting}>
              {submitting ? 'Creating…' : 'Create Account'}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
