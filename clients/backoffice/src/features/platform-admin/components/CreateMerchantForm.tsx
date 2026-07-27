import { useState, useEffect } from 'react';
import { Alert } from '@/components/ui/Alert';
import { ApiError } from '@/lib/api';
import { platformAdminMerchantsApi } from '../api/merchantsApi';
import { platformAdminAccountsApi } from '../api/accountsApi';
import type { PlatformAccount } from '../types';

interface CreateMerchantFormProps {
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

export function CreateMerchantForm({ onSuccess, onCancel }: CreateMerchantFormProps) {
  const [name, setName] = useState('');
  const [assignMode, setAssignMode] = useState<'existing' | 'new'>('existing');
  const [ownerId, setOwnerId] = useState<number | undefined>(undefined);
  const [newUsername, setNewUsername] = useState('');
  const [newEmail, setNewEmail] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [inventoryScoping, setInventoryScoping] = useState<'shared' | 'independent_per_branch'>('shared');
  const [owners, setOwners] = useState<PlatformAccount[]>([]);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    platformAdminAccountsApi.listOwners().then(setOwners);
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSubmitting(true);
    try {
      await platformAdminMerchantsApi.create({
        name,
        ownerId: assignMode === 'existing' ? ownerId : undefined,
        newOwner: assignMode === 'new'
          ? { username: newUsername, email: newEmail, password: newPassword }
          : undefined,
        inventoryScoping,
      });
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
          Create Merchant
        </h2>

        {error && (
          <div className="mb-4">
            <Alert variant="danger">{error}</Alert>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-1 block text-sm font-medium text-neutral-700">
              Merchant Name
            </label>
            <input
              type="text"
              required
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full rounded-md border border-neutral-200 px-3 py-2 text-sm text-neutral-900 outline-none focus:border-secondary focus:ring-1 focus:ring-secondary"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-neutral-700">
              Owner Assignment
            </label>
            <div className="mb-2 flex gap-2">
              <button
                type="button"
                onClick={() => setAssignMode('existing')}
                className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                  assignMode === 'existing'
                    ? 'bg-secondary text-white'
                    : 'border border-neutral-200 bg-white text-neutral-700 hover:bg-neutral-50'
                }`}
              >
                Existing Owner
              </button>
              <button
                type="button"
                onClick={() => setAssignMode('new')}
                className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                  assignMode === 'new'
                    ? 'bg-secondary text-white'
                    : 'border border-neutral-200 bg-white text-neutral-700 hover:bg-neutral-50'
                }`}
              >
                Create New Owner
              </button>
            </div>

            {assignMode === 'existing' ? (
              <select
                value={ownerId ?? ''}
                onChange={(e) => setOwnerId(e.target.value ? Number(e.target.value) : undefined)}
                className="w-full rounded-md border border-neutral-200 px-3 py-2 text-sm text-neutral-900 outline-none focus:border-secondary focus:ring-1 focus:ring-secondary"
              >
                <option value="">Select an owner...</option>
                {owners.map((o) => (
                  <option key={o.id} value={o.id}>
                    {o.username} ({o.email})
                  </option>
                ))}
              </select>
            ) : (
              <div className="space-y-3 rounded-md border border-neutral-200 bg-neutral-50 p-3">
                <input
                  type="text"
                  placeholder="Username"
                  required
                  value={newUsername}
                  onChange={(e) => setNewUsername(e.target.value)}
                  className="w-full rounded-md border border-neutral-200 px-3 py-2 text-sm text-neutral-900 outline-none focus:border-secondary focus:ring-1 focus:ring-secondary"
                />
                <input
                  type="email"
                  placeholder="Email"
                  required
                  value={newEmail}
                  onChange={(e) => setNewEmail(e.target.value)}
                  className="w-full rounded-md border border-neutral-200 px-3 py-2 text-sm text-neutral-900 outline-none focus:border-secondary focus:ring-1 focus:ring-secondary"
                />
                <input
                  type="password"
                  placeholder="Temporary password"
                  required
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  className="w-full rounded-md border border-neutral-200 px-3 py-2 text-sm text-neutral-900 outline-none focus:border-secondary focus:ring-1 focus:ring-secondary"
                />
              </div>
            )}
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-neutral-700">
              Inventory Scoping
            </label>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => setInventoryScoping('shared')}
                className={`flex-1 rounded-md px-3 py-2 text-sm font-medium transition-colors ${
                  inventoryScoping === 'shared'
                    ? 'bg-secondary text-white'
                    : 'border border-neutral-200 bg-white text-neutral-700 hover:bg-neutral-50'
                }`}
              >
                Shared
              </button>
              <button
                type="button"
                onClick={() => setInventoryScoping('independent_per_branch')}
                className={`flex-1 rounded-md px-3 py-2 text-sm font-medium transition-colors ${
                  inventoryScoping === 'independent_per_branch'
                    ? 'bg-secondary text-white'
                    : 'border border-neutral-200 bg-white text-neutral-700 hover:bg-neutral-50'
                }`}
              >
                Independent per Branch
              </button>
            </div>
          </div>

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
              {submitting ? 'Creating...' : 'Create Merchant'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
