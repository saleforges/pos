import { Search, Plus, MoreHorizontal } from 'lucide-react';

const MERCHANTS = [
  { id: 1, name: 'Saleforge Main', owner: 'John Doe', email: 'john@saleforges.com', stores: 3, status: 'Active' },
  { id: 2, name: 'TechMart', owner: 'Jane Smith', email: 'jane@techmart.com', stores: 5, status: 'Active' },
  { id: 3, name: 'FoodHub', owner: 'Bob Johnson', email: 'bob@foodhub.com', stores: 2, status: 'Active' },
  { id: 4, name: 'GreenLeaf', owner: 'Alice Williams', email: 'alice@greenleaf.com', stores: 1, status: 'Inactive' },
];

export default function MerchantsPage() {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="relative flex-1 max-w-xs">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-400" />
          <input
            type="text"
            placeholder="Search merchants..."
            className="w-full rounded-lg border border-neutral-200 bg-white py-2 pl-9 pr-3 text-sm outline-none focus:border-neutral-400"
          />
        </div>
        <button className="flex items-center gap-2 rounded-lg bg-primary px-3 py-2 text-sm text-white hover:bg-primary-hover">
          <Plus size={16} />
          Add Merchant
        </button>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        {MERCHANTS.map(({ id, name, owner, email, stores, status }) => (
          <div key={id} className="rounded-lg border border-neutral-200 bg-white p-5">
            <div className="flex items-start justify-between">
              <div>
                <h3 className="font-display text-sm font-semibold text-neutral-900">{name}</h3>
                <p className="mt-0.5 text-xs text-neutral-500">{owner}</p>
              </div>
              <button className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600">
                <MoreHorizontal size={16} />
              </button>
            </div>
            <div className="mt-4 flex items-center gap-4 text-xs text-neutral-500">
              <span>{email}</span>
              <span>{stores} store{stores > 1 ? 's' : ''}</span>
              <span className={`ml-auto inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                status === 'Active' ? 'bg-secondary/10 text-secondary' : 'bg-neutral-100 text-neutral-500'
              }`}>
                {status}
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
