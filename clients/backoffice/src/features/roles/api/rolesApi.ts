import { api } from '@/lib/api';
import type { Role } from '../types';

export const rolesApi = {
  list: () => api<Role[]>('/roles'),
  get: (id: number) => api<Role>(`/roles/${id}`),
  update: (id: number, data: Partial<Role>) =>
    api<Role>(`/roles/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
};
