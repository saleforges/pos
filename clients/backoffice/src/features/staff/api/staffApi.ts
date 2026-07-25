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
  branchId: number;
  userId: number;
  role: string;
  isDefault?: boolean;
}

export const staffApi = {
  list: (merchantId: number) =>
    api<StaffMemberResponse[]>('/staff', {
      headers: { 'X-Merchant-Id': String(merchantId) } as Record<string, string>,
    }),
  assign: (merchantId: number, data: AssignStaffRequest) =>
    api<StaffMemberResponse>('/staff', {
      method: 'POST',
      headers: { 'X-Merchant-Id': String(merchantId) } as Record<string, string>,
      body: JSON.stringify(data),
    }),
  remove: (merchantId: number, staffId: number) =>
    api<void>(`/staff/${staffId}`, {
      method: 'DELETE',
      headers: { 'X-Merchant-Id': String(merchantId) } as Record<string, string>,
    }),
};
