import { useEffect, useState } from 'react';
import { Alert } from '@/components/ui/Alert';
import { platformAdminAccountsApi } from '../api/accountsApi';
import { AccountsTable } from '../components/AccountsTable';
import { CreateAccountForm } from '../components/CreateAccountForm';
import type { PlatformAccount } from '../types';

export default function AdminAccountsPage() {
  const [accounts, setAccounts] = useState<PlatformAccount[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [success, setSuccess] = useState('');

  const fetchAccounts = () => {
    setIsLoading(true);
    platformAdminAccountsApi
      .list()
      .then(setAccounts)
      .finally(() => setIsLoading(false));
  };

  useEffect(() => {
    fetchAccounts();
  }, []);

  const handleCreated = () => {
    setShowForm(false);
    setSuccess('Account created successfully.');
    fetchAccounts();
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="font-display text-2xl font-bold text-neutral-900">Accounts</h1>
        <button
          onClick={() => setShowForm(true)}
          className="rounded-md bg-secondary px-4 py-2 text-sm font-medium text-white hover:bg-secondary-hover"
        >
          Create Account
        </button>
      </div>

      {success && (
        <Alert variant="success">{success}</Alert>
      )}

      {showForm && (
        <CreateAccountForm
          onSuccess={handleCreated}
          onCancel={() => setShowForm(false)}
        />
      )}

      {isLoading ? (
        <div className="py-8 text-center text-sm text-neutral-400">Loading accounts...</div>
      ) : (
        <AccountsTable accounts={accounts} />
      )}
    </div>
  );
}
