export interface PlatformMerchant {
  id: number;
  name: string;
  ownerId: number | null;
  ownerUsername: string | null;
  ownerEmail: string | null;
  branchCount: number;
  inventoryScoping: 'shared' | 'independent_per_branch';
  status: string;
  createdAt: string;
}

export interface PlatformAccount {
  id: number;
  username: string;
  email: string;
  roleName: string;
  status: string;
  merchantId: number | null;
  merchantName: string | null;
  createdAt: string;
}

export interface CreateMerchantPayload {
  name: string;
  ownerId?: number;
  newOwner?: {
    username: string;
    email: string;
    password: string;
  };
  inventoryScoping: 'shared' | 'independent_per_branch';
}

export interface CreateAccountPayload {
  username: string;
  email: string;
  password: string;
  role: string;
  merchantId?: number;
}
