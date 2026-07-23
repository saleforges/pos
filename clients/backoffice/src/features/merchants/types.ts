export interface MerchantSettings {
  tax_rate: number;
  currency: string;
  timezone: string;
  low_stock_threshold: number;
}

export interface Merchant {
  id: number;
  name: string;
  legal_name: string;
  address: string;
  phone: string;
  email: string;
  tax_id: string;
  status: string;
  settings: MerchantSettings;
  created_at: string;
  updated_at: string;
}

export interface CreateMerchantPayload {
  name: string;
  legal_name: string;
  address: string;
  phone: string;
  email: string;
  tax_id: string;
  settings: MerchantSettings;
}
