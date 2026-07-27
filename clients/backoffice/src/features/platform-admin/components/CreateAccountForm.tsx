import { useState, useEffect } from 'react';
import { Alert } from '@/components/ui/Alert';
import { ApiError } from '@/lib/api';
import { platformAdminAccountsApi } from '../api/accountsApi';
import { platformAdminMerchantsApi } from '../api/merchantsApi';
import type { PlatformMerchant } from '../types';

interface CreateAccountFormProps {
  onSuccess: () => void;
  onCancel: () => void;
}

function parseApiError(err: unknown): string {
  if (err instanceof ApiError) {
    try {
      const body = JSON.parse(err.message);
      return body.error ?? err.message;
    } catch {
      return err.message;
    }
  }
  return 'An unexpected error occurred.';
}

export function CreateAccountForm({ onSuccess, onCancel }: CreateAccountFormProps) {
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState('staff');
  const [merchantId, setMerchantId] = useState<number | undefined>(undefined);
  const [roles, setRoles] = useState<string[]>([]);
  const [merchants, setMerchants] = useState<PlatformMerchant[]>([]);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    platformAdminAccountsApi.getAssignmentRoles().then(setRoles);
    platformAdminMerchantsApi.list().then(setMerchants);
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSubmitting(true);
    try {
      await platformAdminAccountsApi.create({ username, email, password, role, merchantId });
      onSuccess();
    } catch (err) {
      setError(parseApiError(err));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="w-full max-w-lg rounded-lg border border-neutral-200 bg-white p-6 shadow-lg">
        <h2 className="mb-4 font-display text-lg font-semibold text-neutral-900">
          Create Account
        </h2>

        {error && (
          <div className="mb-4">
            <Alert variant="danger">{error}</Alert>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-1 block text-sm font-medium text-neutral-700">Username</label>
            <input
              type="text"
              required
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full rounded-md border border-neutral-200 px-3 py-2 text-sm text-neutral-900 outline-none focus:border-secondary focus:ring-1 focus:ring-secondary"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-neutral-700">Email</label>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full rounded-md border border-neutral-200 px-3 py-2 text-sm text-neutral-900 outline-none focus:border-secondary focus:ring-1 focus:ring-secondary"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-neutral-700">Password</label>
            <input
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full rounded-md border border-neutral-200 px-3 py-2 text-sm text-neutral-900 outline-none focus:border-secondary focus:ring-1 focus:ring-secondary"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-neutral-700">Role</label>
            <select
              value={role}
              onChange={(e) => setRole(e.target.value)}
              className="w-full rounded-md border border-neutral-200 px-3 py-2 text-sm text-neutral-900 outline-none focus:border-secondary focus:ring-1 focus:ring-secondary"
            >
              {roles.map((r) => (
                <option key={r} value={r}>
                  {r}
                </option>
              ))}
            </select>
          </div>

          {role !== 'superadmin' && (
            <div>
              <label className="mb-1 block text-sm font-medium text-neutral-700">
                Merchant <span className="text-neutral-400">(optional)</span>
              </label>
              <select
                value={merchantId ?? ''}
                onChange={(e) => setMerchantId(e.target.value ? Number(e.target.value) : undefined)}
                className="w-full rounded-md border border-neutral-200 px-3 py-2 text-sm text-neutral-900 outline-none focus:border-secondary focus:ring-1 focus:ring-secondary"
              >
                <option value="">No merchant</option>
                {merchants.map((m) => (
                  <option key={m.id} value={m.id}>
                    {m.name}
                  </option>
                ))}
              </select>
            </div>
          )}

          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onCancel}
              className="rounded-md border border-neutral-200 bg-white px-4 py-2 text-sm font-medium text-neutral-700 hover:bg-neutral-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="rounded-md bg-secondary px-4 py-2 text-sm font-medium text-white hover:bg-secondary-hover disabled:opacity-50"
            >
              {submitting ? 'Creating...' : 'Create Account'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
