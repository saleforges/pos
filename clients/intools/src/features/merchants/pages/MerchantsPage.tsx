import { useEffect, useState } from 'react';
import { Search, Plus } from 'lucide-react';
import { merchantsApi, type Merchant } from '../api/merchantsApi';
import { CreateMerchantForm } from '../components/CreateMerchantForm';
import { PageLoader } from '@/components/ui/PageLoader';

export default function MerchantsPage() {
  const [merchants, setMerchants] = useState<Merchant[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [showForm, setShowForm] = useState(false);

  const fetchMerchants = () => {
    setIsLoading(true);
    merchantsApi.list().then(setMerchants).finally(() => setIsLoading(false));
  };

  useEffect(() => { fetchMerchants() }, []);

  const filtered = merchants.filter(
    (m) => m.name.toLowerCase().includes(search.toLowerCase()),
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
              placeholder="Search merchants…"
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
            Create Merchant
          </button>
        </div>

        <div className="rounded-lg border border-neutral-200 bg-white">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-neutral-100 text-left text-xs text-neutral-500">
                  <th className="px-4 py-3 font-medium">Name</th>
                  <th className="px-4 py-3 font-medium">Owner</th>
                  <th className="px-4 py-3 font-medium">Branches</th>
                  <th className="px-4 py-3 font-medium">Inventory Scoping</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium">Created</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((merchant) => (
                  <tr key={merchant.id} className="border-b border-neutral-100 last:border-0 hover:bg-neutral-50">
                    <td className="px-4 py-3 font-medium text-neutral-900">{merchant.name}</td>
                    <td className="px-4 py-3 text-neutral-600">{merchant.ownerName ?? '—'}</td>
                    <td className="px-4 py-3 text-neutral-600">{merchant.branchCount}</td>
                    <td className="px-4 py-3">
                      <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                        merchant.inventoryScoping
                          ? 'bg-secondary/10 text-secondary'
                          : 'bg-neutral-100 text-neutral-500'
                      }`}>
                        {merchant.inventoryScoping ? 'Enabled' : 'Disabled'}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                        merchant.status === 'active' ? 'bg-secondary/10 text-secondary' : 'bg-neutral-100 text-neutral-500'
                      }`}>
                        {merchant.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-xs text-neutral-400">
                      {new Date(merchant.createdAt).toLocaleDateString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {showForm && (
        <CreateMerchantForm
          onClose={() => setShowForm(false)}
          onCreated={() => { setShowForm(false); fetchMerchants() }}
        />
      )}
    </>
  );
}
