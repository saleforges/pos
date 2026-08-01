import { api } from '@/lib/api';

export interface StaffMemberResponse {
  id: number;
  merchantId: number;
  branchId: number;
  userId: number;
  role: string;
  status: string;
  isDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface AssignStaffRequest {
  userId: number;
  branchId: number;
  role: string;
  isDefault?: boolean;
}

const STORAGE_KEY = 'intools_mock_staff';

function getStored(): StaffMemberResponse[] {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '[]');
  } catch {
    return [];
  }
}

function setStored(staff: StaffMemberResponse[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(staff));
}

function seed() {
  const existing = getStored();
  if (existing.length === 0) {
    setStored([
      { id: 1, merchantId: 1, branchId: 1, userId: 2, role: 'owner', status: 'active', isDefault: true, createdAt: '2026-02-15T00:00:00Z', updatedAt: '2026-02-15T00:00:00Z' },
      { id: 2, merchantId: 1, branchId: 1, userId: 3, role: 'cashier', status: 'active', isDefault: true, createdAt: '2026-03-01T00:00:00Z', updatedAt: '2026-03-01T00:00:00Z' },
    ]);
  }
}

seed();

export const staffApi = {
  list: async (merchantId: number): Promise<StaffMemberResponse[]> => {
    try {
      return await api<StaffMemberResponse[]>('/staff', {
        headers: { 'X-Merchant-Id': String(merchantId) } as Record<string, string>,
      });
    } catch {
      return getStored().filter((s) => s.merchantId === merchantId);
    }
  },

  assign: async (merchantId: number, data: AssignStaffRequest): Promise<StaffMemberResponse> => {
    try {
      return await api<StaffMemberResponse>('/staff', {
        method: 'POST',
        headers: { 'X-Merchant-Id': String(merchantId) } as Record<string, string>,
        body: JSON.stringify(data),
      });
    } catch {
      const staff = getStored();
      const member: StaffMemberResponse = {
        id: staff.length > 0 ? Math.max(...staff.map((s) => s.id)) + 1 : 1,
        merchantId,
        branchId: data.branchId,
        userId: data.userId,
        role: data.role,
        status: 'active',
        isDefault: data.isDefault ?? false,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      staff.push(member);
      setStored(staff);
      return member;
    }
  },
};
