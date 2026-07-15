import { Search, Filter } from 'lucide-react';

const ORDERS = [
  { id: '#1234', customer: 'John Doe', status: 'Completed', items: 3, total: '$45.00', date: '2026-07-15' },
  { id: '#1233', customer: 'Jane Smith', status: 'Processing', items: 1, total: '$32.50', date: '2026-07-15' },
  { id: '#1232', customer: 'Bob Johnson', status: 'Pending', items: 5, total: '$78.20', date: '2026-07-14' },
  { id: '#1231', customer: 'Alice Williams', status: 'Completed', items: 2, total: '$21.00', date: '2026-07-14' },
  { id: '#1230', customer: 'Charlie Brown', status: 'Cancelled', items: 4, total: '$55.00', date: '2026-07-13' },
  { id: '#1229', customer: 'Diana Prince', status: 'Completed', items: 3, total: '$67.30', date: '2026-07-13' },
  { id: '#1228', customer: 'Edward Norton', status: 'Processing', items: 2, total: '$22.40', date: '2026-07-12' },
  { id: '#1227', customer: 'Fiona Apple', status: 'Pending', items: 1, total: '$12.00', date: '2026-07-12' },
];

const STATUS_STYLES: Record<string, string> = {
  Completed: 'bg-secondary/10 text-secondary',
  Processing: 'bg-info/10 text-info-hover',
  Pending: 'bg-tertiary/10 text-tertiary-hover',
  Cancelled: 'bg-danger/10 text-danger',
};

export default function OrdersPage() {
  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3">
        <div className="relative flex-1">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-400" />
          <input
            type="text"
            placeholder="Search orders..."
            className="w-full rounded-lg border border-neutral-200 bg-white py-2 pl-9 pr-3 text-sm outline-none focus:border-neutral-400"
          />
        </div>
        <button className="flex items-center gap-2 rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-600 hover:bg-neutral-50">
          <Filter size={16} />
          Filter
        </button>
      </div>

      <div className="rounded-lg border border-neutral-200 bg-white">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-neutral-100 text-left text-xs text-neutral-500">
                <th className="px-4 py-3 font-medium">Order</th>
                <th className="px-4 py-3 font-medium">Customer</th>
                <th className="px-4 py-3 font-medium">Status</th>
                <th className="px-4 py-3 font-medium">Items</th>
                <th className="px-4 py-3 font-medium">Total</th>
                <th className="px-4 py-3 font-medium">Date</th>
              </tr>
            </thead>
            <tbody>
              {ORDERS.map(({ id, customer, status, items, total, date }) => (
                <tr key={id} className="border-b border-neutral-100 last:border-0 hover:bg-neutral-50">
                  <td className="px-4 py-3 font-medium text-neutral-900">{id}</td>
                  <td className="px-4 py-3 text-neutral-600">{customer}</td>
                  <td className="px-4 py-3">
                    <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_STYLES[status]}`}>
                      {status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-neutral-600">{items}</td>
                  <td className="px-4 py-3 text-neutral-600">{total}</td>
                  <td className="px-4 py-3 text-neutral-400">{date}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="flex items-center justify-between border-t border-neutral-200 px-4 py-3">
          <span className="text-xs text-neutral-500">Showing 8 of 24 orders</span>
          <div className="flex gap-1">
            <button className="rounded-md border border-neutral-200 px-2.5 py-1 text-xs text-neutral-600 hover:bg-neutral-50">Previous</button>
            <button className="rounded-md bg-primary px-2.5 py-1 text-xs text-white">1</button>
            <button className="rounded-md border border-neutral-200 px-2.5 py-1 text-xs text-neutral-600 hover:bg-neutral-50">2</button>
            <button className="rounded-md border border-neutral-200 px-2.5 py-1 text-xs text-neutral-600 hover:bg-neutral-50">3</button>
            <button className="rounded-md border border-neutral-200 px-2.5 py-1 text-xs text-neutral-600 hover:bg-neutral-50">Next</button>
          </div>
        </div>
      </div>
    </div>
  );
}
