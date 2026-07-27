import { useEffect, useState } from 'react';
import { Building2 } from 'lucide-react';
import { Alert } from '@/components/ui/Alert';
import { platformAdminMerchantsApi } from '../api/merchantsApi';
import { CreateMerchantForm } from '../components/CreateMerchantForm';
import type { PlatformMerchant } from '../types';

export default function AdminMerchantsPage() {
  const [merchants, setMerchants] = useState<PlatformMerchant[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [success, setSuccess] = useState('');

  const fetchMerchants = () => {
    setIsLoading(true);
    platformAdminMerchantsApi
      .list()
      .then(setMerchants)
      .finally(() => setIsLoading(false));
  };

  useEffect(() => {
    fetchMerchants();
  }, []);

  const handleCreated = () => {
    setShowForm(false);
    setSuccess('Merchant created successfully.');
    fetchMerchants();
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="font-display text-2xl font-bold text-neutral-900">Merchants</h1>
        <button
          onClick={() => setShowForm(true)}
          className="rounded-md bg-secondary px-4 py-2 text-sm font-medium text-white hover:bg-secondary-hover"
        >
          Create Merchant
        </button>
      </div>

      {success && (
        <Alert variant="success">{success}</Alert>
      )}

      {showForm && (
        <CreateMerchantForm
          onSuccess={handleCreated}
          onCancel={() => setShowForm(false)}
        />
      )}

      {isLoading ? (
        <div className="py-8 text-center text-sm text-neutral-400">Loading merchants...</div>
      ) : (
        <div className="overflow-x-auto rounded-lg border border-neutral-200 bg-white">
          <table className="w-full border-collapse text-left text-sm">
            <thead>
              <tr className="border-b border-neutral-200 bg-neutral-50">
                <th className="px-4 py-3 font-medium text-neutral-500">Name</th>
                <th className="px-4 py-3 font-medium text-neutral-500">Owner</th>
                <th className="px-4 py-3 font-medium text-neutral-500">Branches</th>
                <th className="px-4 py-3 font-medium text-neutral-500">Inventory Scoping</th>
                <th className="px-4 py-3 font-medium text-neutral-500">Status</th>
                <th className="px-4 py-3 font-medium text-neutral-500">Created</th>
              </tr>
            </thead>
            <tbody>
              {merchants.map((m) => (
                <tr key={m.id} className="border-b border-neutral-100 last:border-0">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <Building2 size={16} className="text-neutral-400" />
                      <span className="font-medium text-neutral-900">{m.name}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-neutral-600">
                    {m.ownerUsername ? (
                      <span>
                        {m.ownerUsername}
                        <span className="ml-1 text-neutral-400">({m.ownerEmail})</span>
                      </span>
                    ) : (
                      <span className="text-neutral-400">—</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-neutral-900">{m.branchCount}</td>
                  <td className="px-4 py-3">
                    <span className="rounded bg-secondary/10 px-2 py-0.5 text-xs font-medium text-secondary-hover">
                      {m.inventoryScoping === 'shared' ? 'Shared' : 'Independent'}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
                        m.status === 'active'
                          ? 'bg-secondary/10 text-secondary-hover'
                          : 'bg-neutral-100 text-neutral-600'
                      }`}
                    >
                      {m.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-neutral-500">
                    {new Date(m.createdAt).toLocaleDateString()}
                  </td>
                </tr>
              ))}
              {merchants.length === 0 && (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-neutral-400">
                    No merchants found.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
