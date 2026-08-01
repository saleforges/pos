import { api } from '@/lib/api';

export interface MerchantSettings {
  taxRate: number;
  currency: string;
  timezone: string;
  receiptFooter?: string;
  receiptLogo?: string;
  orderPrefix?: string;
  lowStockThreshold: number;
}

export interface Merchant {
  id: number;
  name: string;
  legalName: string;
  address: string;
  phone: string;
  email: string;
  logoUrl?: string;
  taxId?: string;
  status: string;
  settings: MerchantSettings;
  createdAt: string;
  updatedAt: string;
}

export interface CreateMerchantRequest {
  name: string;
  legalName?: string;
  address?: string;
  phone?: string;
  email: string;
  taxId?: string;
  settings?: Partial<MerchantSettings>;
}

export interface UpdateMerchantRequest {
  name?: string;
  legalName?: string;
  address?: string;
  phone?: string;
  email?: string;
  taxId?: string;
}

const STORAGE_KEY = 'intools_mock_merchants';

function getStored(): Merchant[] {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '[]');
  } catch {
    return [];
  }
}

function setStored(merchants: Merchant[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(merchants));
}

function seed() {
  const existing = getStored();
  if (existing.length === 0) {
    setStored([
      {
        id: 1, name: 'Acme Corp', legalName: 'PT Acme Corporation', address: 'Jl. Sudirman No. 1', phone: '021-1234567', email: 'acme@example.com', logoUrl: '', taxId: '1234567890', status: 'active',
        settings: { taxRate: 11, currency: 'IDR', timezone: 'Asia/Jakarta', receiptFooter: 'Terima kasih', receiptLogo: '', orderPrefix: 'ACM', lowStockThreshold: 10 },
        createdAt: '2026-02-01T00:00:00Z', updatedAt: '2026-02-01T00:00:00Z',
      },
      {
        id: 2, name: 'Globex Inc', legalName: 'PT Globex Indonesia', address: 'Jl. Thamrin No. 5', phone: '021-7654321', email: 'globex@example.com', logoUrl: '', taxId: '0987654321', status: 'active',
        settings: { taxRate: 11, currency: 'IDR', timezone: 'Asia/Jakarta', receiptFooter: '', receiptLogo: '', orderPrefix: 'GLX', lowStockThreshold: 5 },
        createdAt: '2026-03-10T00:00:00Z', updatedAt: '2026-03-10T00:00:00Z',
      },
    ]);
  }
}

seed();

export const merchantsApi = {
  list: async (): Promise<Merchant[]> => {
    try {
      return await api<Merchant[]>('/merchants');
    } catch {
      return getStored();
    }
  },

  get: async (id: number): Promise<Merchant> => {
    try {
      return await api<Merchant>(`/merchants/${id}`);
    } catch {
      const merchant = getStored().find((m) => m.id === id);
      if (!merchant) throw new Error('Merchant not found');
      return merchant;
    }
  },

  update: async (id: number, data: UpdateMerchantRequest): Promise<Merchant> => {
    try {
      return await api<Merchant>(`/merchants/${id}`, {
        method: 'PATCH',
        body: JSON.stringify(data),
      });
    } catch {
      const merchants = getStored();
      const idx = merchants.findIndex((m) => m.id === id);
      if (idx === -1) throw new Error('Merchant not found');
      merchants[idx] = { ...merchants[idx], ...data, updatedAt: new Date().toISOString() };
      setStored(merchants);
      return merchants[idx];
    }
  },

  create: async (data: CreateMerchantRequest): Promise<Merchant> => {
    try {
      return await api<Merchant>('/merchants', {
        method: 'POST',
        body: JSON.stringify(data),
      });
    } catch {
      const merchants = getStored();
      const merchant: Merchant = {
        id: merchants.length > 0 ? Math.max(...merchants.map((m) => m.id)) + 1 : 1,
        name: data.name,
        legalName: data.legalName ?? '',
        address: data.address ?? '',
        phone: data.phone ?? '',
        email: data.email,
        logoUrl: '',
        taxId: data.taxId ?? '',
        status: 'active',
        settings: {
          taxRate: data.settings?.taxRate ?? 11,
          currency: data.settings?.currency ?? 'IDR',
          timezone: data.settings?.timezone ?? 'Asia/Jakarta',
          receiptFooter: data.settings?.receiptFooter ?? '',
          receiptLogo: data.settings?.receiptLogo ?? '',
          orderPrefix: data.settings?.orderPrefix ?? '',
          lowStockThreshold: data.settings?.lowStockThreshold ?? 10,
        },
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      merchants.push(merchant);
      setStored(merchants);
      return merchant;
    }
  },
};
