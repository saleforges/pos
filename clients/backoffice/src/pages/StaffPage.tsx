import { Search, Plus, MoreHorizontal } from 'lucide-react';

const STAFF = [
  { name: 'John Doe', email: 'john@saleforges.com', role: 'Cashier', branch: 'Main Branch', status: 'Active' },
  { name: 'Jane Smith', email: 'jane@saleforges.com', role: 'Manager', branch: 'Main Branch', status: 'Active' },
  { name: 'Bob Johnson', email: 'bob@saleforges.com', role: 'Cashier', branch: 'Downtown', status: 'Active' },
  { name: 'Alice Williams', email: 'alice@saleforges.com', role: 'Supervisor', branch: 'Main Branch', status: 'Active' },
  { name: 'Charlie Brown', email: 'charlie@saleforges.com', role: 'Cashier', branch: 'Downtown', status: 'Inactive' },
  { name: 'Diana Prince', email: 'diana@saleforges.com', role: 'Manager', branch: 'Uptown', status: 'Active' },
];

export default function StaffPage() {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="relative flex-1 max-w-xs">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-400" />
          <input
            type="text"
            placeholder="Search staff..."
            className="w-full rounded-lg border border-neutral-200 bg-white py-2 pl-9 pr-3 text-sm outline-none focus:border-neutral-400"
          />
        </div>
        <button className="flex items-center gap-2 rounded-lg bg-primary px-3 py-2 text-sm text-white hover:bg-primary-hover">
          <Plus size={16} />
          Add Staff
        </button>
      </div>

      <div className="rounded-lg border border-neutral-200 bg-white">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-neutral-100 text-left text-xs text-neutral-500">
                <th className="px-4 py-3 font-medium">Name</th>
                <th className="px-4 py-3 font-medium">Email</th>
                <th className="px-4 py-3 font-medium">Role</th>
                <th className="px-4 py-3 font-medium">Branch</th>
                <th className="px-4 py-3 font-medium">Status</th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody>
              {STAFF.map(({ name, email, role, branch, status }) => (
                <tr key={email} className="border-b border-neutral-100 last:border-0 hover:bg-neutral-50">
                  <td className="px-4 py-3 font-medium text-neutral-900">{name}</td>
                  <td className="px-4 py-3 text-neutral-600">{email}</td>
                  <td className="px-4 py-3">
                    <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                      role === 'Manager' ? 'bg-info/10 text-info-hover' : 'bg-neutral-100 text-neutral-600'
                    }`}>
                      {role}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-neutral-500">{branch}</td>
                  <td className="px-4 py-3">
                    <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                      status === 'Active' ? 'bg-secondary/10 text-secondary' : 'bg-neutral-100 text-neutral-500'
                    }`}>
                      {status}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <button className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600">
                      <MoreHorizontal size={16} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
