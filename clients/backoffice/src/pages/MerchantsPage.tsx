import { useEffect, useState, useCallback } from 'react';
import { Search, Plus, Trash2, ChevronLeft, ChevronRight, Building2, MapPin, Phone, Mail, Hash, Clock, CreditCard, Globe, AlertTriangle } from 'lucide-react';
import { merchantsApi } from '@/features/merchants/api/merchantsApi';
import type { Merchant, CreateMerchantPayload } from '@/features/merchants/types';
import { Modal } from '@/components/ui/Modal';

const PAGE_SIZE = 10;

const DEFAULT_SETTINGS = {
  tax_rate: 11,
  currency: 'IDR',
  timezone: 'Asia/Jakarta',
  low_stock_threshold: 5,
};

function formatDate(dateStr: string) {
  const d = new Date(dateStr);
  return d.toLocaleDateString('en-ID', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' });
}

function DetailRow({ icon, label, value }: { icon: React.ReactNode; label: string; value: string }) {
  return (
    <div className="flex items-start gap-2 text-sm">
      <span className="mt-0.5 text-neutral-400">{icon}</span>
      <div>
        <span className="text-neutral-500">{label}:</span>{' '}
        <span className="text-neutral-900">{value || '-'}</span>
      </div>
    </div>
  );
}

export default function MerchantsPage() {
  const [merchants, setMerchants] = useState<Merchant[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);

  const [showCreate, setShowCreate] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const [detailMerchant, setDetailMerchant] = useState<Merchant | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<number | null>(null);

  const fetchMerchants = useCallback(async (pageNum: number) => {
    setIsLoading(true);
    try {
      const offset = (pageNum - 1) * PAGE_SIZE;
      const data = await merchantsApi.list({ offset, limit: PAGE_SIZE + 1 });
      setHasMore(data.length > PAGE_SIZE);
      setMerchants(data.slice(0, PAGE_SIZE));
    } catch {
      console.error('Failed to fetch merchants');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchMerchants(page);
  }, [page, fetchMerchants]);

  const filtered = search
    ? merchants.filter(
        (m) =>
          m.name.toLowerCase().includes(search.toLowerCase()) ||
          m.email.toLowerCase().includes(search.toLowerCase()),
      )
    : merchants;

  const handleCreate = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setSubmitting(true);
    const form = new FormData(e.currentTarget);
    const payload: CreateMerchantPayload = {
      name: form.get('name') as string,
      legal_name: form.get('legal_name') as string,
      address: form.get('address') as string,
      phone: form.get('phone') as string,
      email: form.get('email') as string,
      tax_id: form.get('tax_id') as string,
      settings: {
        ...DEFAULT_SETTINGS,
        tax_rate: Number(form.get('tax_rate')) || DEFAULT_SETTINGS.tax_rate,
        low_stock_threshold: Number(form.get('low_stock_threshold')) || DEFAULT_SETTINGS.low_stock_threshold,
      },
    };
    try {
      await merchantsApi.create(payload);
      setShowCreate(false);
      setPage(1);
      await fetchMerchants(1);
    } catch (err) {
      console.error('Failed to create merchant', err);
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await merchantsApi.delete(id);
      setDetailMerchant(null);
      setConfirmDelete(null);
      await fetchMerchants(page);
    } catch (err) {
      console.error('Failed to delete merchant', err);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="relative flex-1 max-w-xs">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-400" />
          <input
            type="text"
            placeholder="Search merchants..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full rounded-lg border border-neutral-200 bg-white py-2 pl-9 pr-3 text-sm outline-none focus:border-neutral-400"
          />
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 rounded-lg bg-primary px-3 py-2 text-sm text-white hover:bg-primary-hover"
        >
          <Plus size={16} />
          Add Merchant
        </button>
      </div>

      {isLoading ? (
        <div className="py-12 text-center text-sm text-neutral-400">Loading merchants...</div>
      ) : filtered.length === 0 ? (
        <div className="py-12 text-center text-sm text-neutral-400">
          {search ? 'No merchants match your search.' : 'No merchants yet. Create your first merchant.'}
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            {filtered.map((m) => (
              <div
                key={m.id}
                onClick={() => setDetailMerchant(m)}
                className="cursor-pointer rounded-lg border border-neutral-200 bg-white p-5 transition-colors hover:border-neutral-300"
              >
                <div className="flex items-start justify-between">
                  <div>
                    <h3 className="font-display text-sm font-semibold text-neutral-900">{m.name}</h3>
                    {m.legal_name && (
                      <p className="mt-0.5 text-xs text-neutral-400">{m.legal_name}</p>
                    )}
                  </div>
                  <span
                    className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                      m.status === 'active'
                        ? 'bg-secondary/10 text-secondary'
                        : 'bg-neutral-100 text-neutral-500'
                    }`}
                  >
                    {m.status}
                  </span>
                </div>
                <div className="mt-4 flex items-center gap-4 text-xs text-neutral-500">
                  <span>{m.email}</span>
                  <span>{m.phone}</span>
                </div>
              </div>
            ))}
          </div>

          <div className="flex items-center justify-between pt-2">
            <span className="text-xs text-neutral-400">
              Page {page}
            </span>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="flex items-center gap-1 rounded-md border border-neutral-200 px-3 py-1.5 text-sm text-neutral-600 hover:bg-neutral-50 disabled:opacity-40 disabled:cursor-not-allowed"
              >
                <ChevronLeft size={14} />
                Previous
              </button>
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={!hasMore}
                className="flex items-center gap-1 rounded-md border border-neutral-200 px-3 py-1.5 text-sm text-neutral-600 hover:bg-neutral-50 disabled:opacity-40 disabled:cursor-not-allowed"
              >
                Next
                <ChevronRight size={14} />
              </button>
            </div>
          </div>
        </>
      )}

      <Modal open={showCreate} onClose={() => setShowCreate(false)} title="Add Merchant">
        <form onSubmit={handleCreate} className="space-y-3">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-neutral-600 mb-1">Name *</label>
              <input name="name" required className="w-full rounded-md border border-neutral-200 px-3 py-1.5 text-sm outline-none focus:border-neutral-400" />
            </div>
            <div>
              <label className="block text-xs font-medium text-neutral-600 mb-1">Legal Name</label>
              <input name="legal_name" className="w-full rounded-md border border-neutral-200 px-3 py-1.5 text-sm outline-none focus:border-neutral-400" />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-neutral-600 mb-1">Email *</label>
              <input name="email" type="email" required className="w-full rounded-md border border-neutral-200 px-3 py-1.5 text-sm outline-none focus:border-neutral-400" />
            </div>
            <div>
              <label className="block text-xs font-medium text-neutral-600 mb-1">Phone</label>
              <input name="phone" className="w-full rounded-md border border-neutral-200 px-3 py-1.5 text-sm outline-none focus:border-neutral-400" />
            </div>
          </div>
          <div>
            <label className="block text-xs font-medium text-neutral-600 mb-1">Address</label>
            <textarea name="address" rows={2} className="w-full rounded-md border border-neutral-200 px-3 py-1.5 text-sm outline-none focus:border-neutral-400" />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-neutral-600 mb-1">Tax ID</label>
              <input name="tax_id" className="w-full rounded-md border border-neutral-200 px-3 py-1.5 text-sm outline-none focus:border-neutral-400" />
            </div>
            <div>
              <label className="block text-xs font-medium text-neutral-600 mb-1">Tax Rate (%)</label>
              <input name="tax_rate" type="number" defaultValue={11} className="w-full rounded-md border border-neutral-200 px-3 py-1.5 text-sm outline-none focus:border-neutral-400" />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-neutral-600 mb-1">Currency</label>
              <input name="currency" defaultValue="IDR" className="w-full rounded-md border border-neutral-200 px-3 py-1.5 text-sm outline-none focus:border-neutral-400" />
            </div>
            <div>
              <label className="block text-xs font-medium text-neutral-600 mb-1">Timezone</label>
              <input name="timezone" defaultValue="Asia/Jakarta" className="w-full rounded-md border border-neutral-200 px-3 py-1.5 text-sm outline-none focus:border-neutral-400" />
            </div>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <button type="button" onClick={() => setShowCreate(false)} className="rounded-md px-3 py-1.5 text-sm text-neutral-600 hover:bg-neutral-100">Cancel</button>
            <button type="submit" disabled={submitting} className="rounded-md bg-primary px-4 py-1.5 text-sm text-white hover:bg-primary-hover disabled:opacity-50">
              {submitting ? 'Creating...' : 'Create'}
            </button>
          </div>
        </form>
      </Modal>

      <Modal open={!!detailMerchant} onClose={() => { setDetailMerchant(null); setConfirmDelete(null); }} title={detailMerchant?.name ?? ''}>
        {detailMerchant && (
          <div className="space-y-4">
            {confirmDelete === detailMerchant.id ? (
              <div className="rounded-lg border border-red-200 bg-red-50 p-4">
                <div className="flex items-center gap-2 text-sm font-medium text-red-700 mb-2">
                  <AlertTriangle size={16} />
                  Delete merchant?
                </div>
                <p className="text-xs text-red-600 mb-3">
                  This will permanently delete <strong>{detailMerchant.name}</strong> and all associated data.
                </p>
                <div className="flex justify-end gap-2">
                  <button onClick={() => setConfirmDelete(null)} className="rounded-md px-3 py-1.5 text-sm text-neutral-600 hover:bg-white">Cancel</button>
                  <button onClick={() => handleDelete(detailMerchant.id)} className="rounded-md bg-red-600 px-3 py-1.5 text-sm text-white hover:bg-red-700">Delete</button>
                </div>
              </div>
            ) : (
              <div className="space-y-3">
                <div className="flex items-center gap-2">
                  <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                    detailMerchant.status === 'active' ? 'bg-secondary/10 text-secondary' : 'bg-neutral-100 text-neutral-500'
                  }`}>
                    {detailMerchant.status}
                  </span>
                </div>

                <div className="space-y-2">
                  <DetailRow icon={<Building2 size={14} />} label="Legal Name" value={detailMerchant.legal_name} />
                  <DetailRow icon={<Mail size={14} />} label="Email" value={detailMerchant.email} />
                  <DetailRow icon={<Phone size={14} />} label="Phone" value={detailMerchant.phone} />
                  <DetailRow icon={<MapPin size={14} />} label="Address" value={detailMerchant.address} />
                  <DetailRow icon={<Hash size={14} />} label="Tax ID" value={detailMerchant.tax_id} />
                </div>

                <div className="border-t border-neutral-100 pt-3">
                  <p className="text-xs font-semibold uppercase tracking-wide text-neutral-400 mb-2">Settings</p>
                  <div className="space-y-2">
                    <DetailRow icon={<CreditCard size={14} />} label="Tax Rate" value={`${detailMerchant.settings.tax_rate}%`} />
                    <DetailRow icon={<CreditCard size={14} />} label="Currency" value={detailMerchant.settings.currency} />
                    <DetailRow icon={<Globe size={14} />} label="Timezone" value={detailMerchant.settings.timezone} />
                    <DetailRow icon={<AlertTriangle size={14} />} label="Low Stock Alert" value={String(detailMerchant.settings.low_stock_threshold)} />
                  </div>
                </div>

                <div className="border-t border-neutral-100 pt-3">
                  <div className="space-y-2">
                    <DetailRow icon={<Clock size={14} />} label="Created" value={formatDate(detailMerchant.created_at)} />
                    <DetailRow icon={<Clock size={14} />} label="Updated" value={formatDate(detailMerchant.updated_at)} />
                  </div>
                </div>
              </div>
            )}

            {confirmDelete !== detailMerchant.id && (
              <div className="flex justify-end gap-2 pt-2 border-t border-neutral-100">
                <button
                  onClick={() => setConfirmDelete(detailMerchant.id)}
                  className="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm text-red-600 hover:bg-red-50"
                >
                  <Trash2 size={14} />
                  Delete
                </button>
              </div>
            )}
          </div>
        )}
      </Modal>
    </div>
  );
}
