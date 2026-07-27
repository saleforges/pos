import type { PlatformAccount } from '../types';

export function AccountsTable({ accounts }: { accounts: PlatformAccount[] }) {
  return (
    <div className="overflow-x-auto rounded-lg border border-neutral-200 bg-white">
      <table className="w-full border-collapse text-left text-sm">
        <thead>
          <tr className="border-b border-neutral-200 bg-neutral-50">
            <th className="px-4 py-3 font-medium text-neutral-500">Username</th>
            <th className="px-4 py-3 font-medium text-neutral-500">Email</th>
            <th className="px-4 py-3 font-medium text-neutral-500">Role</th>
            <th className="px-4 py-3 font-medium text-neutral-500">Status</th>
            <th className="px-4 py-3 font-medium text-neutral-500">Merchant</th>
            <th className="px-4 py-3 font-medium text-neutral-500">Created</th>
          </tr>
        </thead>
        <tbody>
          {accounts.map((account) => (
            <tr key={account.id} className="border-b border-neutral-100 last:border-0">
              <td className="px-4 py-3 font-medium text-neutral-900">{account.username}</td>
              <td className="px-4 py-3 text-neutral-600">{account.email}</td>
              <td className="px-4 py-3">
                <span className="rounded bg-secondary/10 px-2 py-0.5 text-xs font-medium text-secondary-hover">
                  {account.roleName}
                </span>
              </td>
              <td className="px-4 py-3">
                <span
                  className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
                    account.status === 'active'
                      ? 'bg-secondary/10 text-secondary-hover'
                      : 'bg-neutral-100 text-neutral-600'
                  }`}
                >
                  {account.status}
                </span>
              </td>
              <td className="px-4 py-3 text-neutral-600">
                {account.merchantName ?? '—'}
              </td>
              <td className="px-4 py-3 text-neutral-500">
                {new Date(account.createdAt).toLocaleDateString()}
              </td>
            </tr>
          ))}
          {accounts.length === 0 && (
            <tr>
              <td colSpan={6} className="px-4 py-8 text-center text-neutral-400">
                No accounts found.
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}
