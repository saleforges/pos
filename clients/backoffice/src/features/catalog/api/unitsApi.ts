import { api } from '@/lib/api';
import type { Unit } from '../types';

export const unitsApi = {
  /** Units are global — no merchant context header needed. */
  list: async () => {
    const res = await api<Unit[] | null>('/units');
    return res ?? [];
  },
};