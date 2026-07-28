import { useEffect, useState } from 'react';
import { Search, Plus } from 'lucide-react';
import { accountsApi, type Account } from '../api/accountsApi';
import { CreateAccountForm } from '../components/CreateAccountForm';
import { PageLoader } from '@/components/ui/PageLoader';

export default function AccountsPage() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [showForm, setShowForm] = useState(false);

  const fetchAccounts = () => {
    setIsLoading(true);
    accountsApi.list().then(setAccounts).finally(() => setIsLoading(false));
  };

  useEffect(() => { fetchAccounts() }, []);

  const filtered = accounts.filter(
    (a) =>
      a.username.toLowerCase().includes(search.toLowerCase()) ||
      a.email.toLowerCase().includes(search.toLowerCase()),
  );

  if (isLoading) return <PageLoader />;

  return (
    <>
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="relative max-w-xs flex-1">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-400" />
            <input
              type="text"
              placeholder="Search accounts…"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full rounded-lg border border-neutral-200 bg-white py-2 pl-9 pr-3 text-sm outline-none focus:border-neutral-400"
            />
          </div>
          <button
            onClick={() => setShowForm(true)}
            className="flex items-center gap-2 rounded-lg bg-primary px-3 py-2 text-sm text-white hover:bg-primary-hover"
          >
            <Plus size={16} />
            Create Account
          </button>
        </div>

        <div className="rounded-lg border border-neutral-200 bg-white">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-neutral-100 text-left text-xs text-neutral-500">
                  <th className="px-4 py-3 font-medium">Username</th>
                  <th className="px-4 py-3 font-medium">Email</th>
                  <th className="px-4 py-3 font-medium">Role</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium">Merchant</th>
                  <th className="px-4 py-3 font-medium">Created</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((account) => (
                  <tr key={account.id} className="border-b border-neutral-100 last:border-0 hover:bg-neutral-50">
                    <td className="px-4 py-3 font-medium text-neutral-900">{account.username}</td>
                    <td className="px-4 py-3 text-neutral-600">{account.email}</td>
                    <td className="px-4 py-3">
                      <span className="inline-block rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
                        {account.role}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                        account.status === 'active' ? 'bg-secondary/10 text-secondary' : 'bg-neutral-100 text-neutral-500'
                      }`}>
                        {account.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-neutral-500">{account.merchantName ?? '—'}</td>
                    <td className="px-4 py-3 text-xs text-neutral-400">
                      {new Date(account.createdAt).toLocaleDateString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {showForm && (
        <CreateAccountForm
          onClose={() => setShowForm(false)}
          onCreated={() => { setShowForm(false); fetchAccounts() }}
        />
      )}
    </>
  );
}
