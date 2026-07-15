import { useState } from 'react';
import { Users, Package, ShoppingCart, DollarSign, Store, ChevronDown } from 'lucide-react';

interface Branch {
  id: number;
  name: string;
  address: string;
}

const BRANCHES: Branch[] = [
  { id: 1, name: 'Main Branch', address: '123 Main St, Downtown' },
  { id: 2, name: 'Downtown', address: '456 Oak Ave, Downtown' },
  { id: 3, name: 'Uptown', address: '789 Pine Rd, Uptown' },
];

const STATS = [
  { label: 'Total Revenue', value: '$12,340', icon: DollarSign, change: '+12.5%', up: true },
  { label: 'Active Orders', value: '24', icon: ShoppingCart, change: '+3.2%', up: true },
  { label: 'Products', value: '156', icon: Package, change: '+8.1%', up: true },
  { label: 'Staff', value: '18', icon: Users, change: '-2.4%', up: false },
];

const RECENT_ORDERS = [
  { id: '#1234', customer: 'John Doe', status: 'Completed', total: '$45.00', date: '2 min ago' },
  { id: '#1233', customer: 'Jane Smith', status: 'Processing', total: '$32.50', date: '15 min ago' },
  { id: '#1232', customer: 'Bob Johnson', status: 'Pending', total: '$78.20', date: '1 hr ago' },
  { id: '#1231', customer: 'Alice Williams', status: 'Completed', total: '$21.00', date: '2 hr ago' },
  { id: '#1230', customer: 'Charlie Brown', status: 'Cancelled', total: '$55.00', date: '3 hr ago' },
];

const STATUS_STYLES: Record<string, string> = {
  Completed: 'bg-secondary/10 text-secondary',
  Processing: 'bg-info/10 text-info-hover',
  Pending: 'bg-tertiary/10 text-tertiary-hover',
  Cancelled: 'bg-danger/10 text-danger',
};

export default function DashboardPage() {
  const [selectedBranch, setSelectedBranch] = useState<Branch | null>(null);
  const [showBranchPicker, setShowBranchPicker] = useState(false);

  if (!selectedBranch) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="w-full max-w-md">
          <div className="mb-8 text-center">
            <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-xl bg-secondary/10">
              <Store size={28} className="text-secondary" />
            </div>
            <h2 className="font-display text-lg font-semibold text-neutral-900">Select a Branch</h2>
            <p className="mt-1 text-sm text-neutral-500">Choose a branch to view its dashboard.</p>
          </div>
          <div className="space-y-3">
            {BRANCHES.map((branch) => (
              <button
                key={branch.id}
                onClick={() => setSelectedBranch(branch)}
                className="w-full rounded-lg border border-neutral-200 bg-white p-4 text-left transition-colors hover:border-secondary hover:shadow-sm"
              >
                <span className="font-medium text-neutral-900">{branch.name}</span>
                <p className="mt-0.5 text-sm text-neutral-500">{branch.address}</p>
              </button>
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="relative">
        <button
          onClick={() => setShowBranchPicker(!showBranchPicker)}
          className="flex items-center gap-2 rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-600 hover:bg-neutral-50"
        >
          <Store size={16} className="text-secondary" />
          <span className="font-medium text-neutral-900">{selectedBranch.name}</span>
          <ChevronDown size={14} />
        </button>
        {showBranchPicker && (
          <>
            <div className="fixed inset-0 z-10" onClick={() => setShowBranchPicker(false)} />
            <div className="absolute left-0 top-full z-20 mt-1 w-64 rounded-lg border border-neutral-200 bg-white shadow-lg">
              {BRANCHES.filter((b) => b.id !== selectedBranch.id).map((branch) => (
                <button
                  key={branch.id}
                  onClick={() => { setSelectedBranch(branch); setShowBranchPicker(false); }}
                  className="w-full px-3 py-2.5 text-left text-sm hover:bg-neutral-50 first:rounded-t-lg last:rounded-b-lg"
                >
                  <span className="font-medium text-neutral-900">{branch.name}</span>
                  <p className="text-xs text-neutral-400">{branch.address}</p>
                </button>
              ))}
            </div>
          </>
        )}
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {STATS.map(({ label, value, icon: Icon, change, up }) => (
          <div key={label} className="rounded-lg border border-neutral-200 bg-white p-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-neutral-500">{label}</span>
              <div className="flex h-9 w-9 items-center justify-center rounded-md bg-neutral-100">
                <Icon size={18} className="text-neutral-600" />
              </div>
            </div>
            <p className="mt-2 text-2xl font-semibold text-neutral-900">{value}</p>
            <span className={`text-xs ${up ? 'text-secondary' : 'text-danger'}`}>{change}</span>
          </div>
        ))}
      </div>

      <div className="rounded-lg border border-neutral-200 bg-white">
        <div className="border-b border-neutral-200 px-4 py-3">
          <h2 className="font-display text-sm font-semibold text-neutral-900">Recent Orders</h2>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-neutral-100 text-left text-xs text-neutral-500">
                <th className="px-4 py-3 font-medium">Order</th>
                <th className="px-4 py-3 font-medium">Customer</th>
                <th className="px-4 py-3 font-medium">Status</th>
                <th className="px-4 py-3 font-medium">Total</th>
                <th className="px-4 py-3 font-medium">Date</th>
              </tr>
            </thead>
            <tbody>
              {RECENT_ORDERS.map(({ id, customer, status, total, date }) => (
                <tr key={id} className="border-b border-neutral-100 last:border-0">
                  <td className="px-4 py-3 font-medium text-neutral-900">{id}</td>
                  <td className="px-4 py-3 text-neutral-600">{customer}</td>
                  <td className="px-4 py-3">
                    <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_STYLES[status]}`}>
                      {status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-neutral-600">{total}</td>
                  <td className="px-4 py-3 text-neutral-400">{date}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
