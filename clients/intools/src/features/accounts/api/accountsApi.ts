import { api } from '@/lib/api';

export interface UserResponse {
  id: number;
  username: string;
  email: string;
  role: string;
  type: string;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  roles?: string[];
}

export interface UpdateUserRequest {
  username?: string;
  email?: string;
  status?: string;
}

const STORAGE_KEY = 'intools_mock_users';

function getStored(): UserResponse[] {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '[]');
  } catch {
    return [];
  }
}

function setStored(users: UserResponse[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(users));
}

function seed() {
  const existing = getStored();
  if (existing.length === 0) {
    setStored([
      { id: 1, username: 'superadmin', email: 'admin@saleforges.com', role: 'superadmin', type: 'platform', status: 'active', createdAt: '2026-01-01T00:00:00Z', updatedAt: '2026-01-01T00:00:00Z' },
      { id: 2, username: 'owner', email: 'owner@example.com', role: 'owner', type: 'merchant', status: 'active', createdAt: '2026-02-15T00:00:00Z', updatedAt: '2026-02-15T00:00:00Z' },
      { id: 3, username: 'cashier1', email: 'cashier@example.com', role: 'cashier', type: 'merchant', status: 'active', createdAt: '2026-03-01T00:00:00Z', updatedAt: '2026-03-01T00:00:00Z' },
    ]);
  }
}

seed();

export const accountsApi = {
  list: async (): Promise<UserResponse[]> => {
    try {
      return await api<UserResponse[]>('/users');
    } catch {
      return getStored();
    }
  },

  get: async (id: number): Promise<UserResponse> => {
    try {
      return await api<UserResponse>(`/users/${id}`);
    } catch {
      const user = getStored().find((u) => u.id === id);
      if (!user) throw new Error('User not found');
      return user;
    }
  },

  update: async (id: number, data: UpdateUserRequest): Promise<UserResponse> => {
    try {
      return await api<UserResponse>(`/users/${id}`, {
        method: 'PATCH',
        body: JSON.stringify(data),
      });
    } catch {
      const users = getStored();
      const idx = users.findIndex((u) => u.id === id);
      if (idx === -1) throw new Error('User not found');
      users[idx] = { ...users[idx], ...data, updatedAt: new Date().toISOString() };
      setStored(users);
      return users[idx];
    }
  },

  create: async (data: CreateUserRequest): Promise<{ accessToken: string; expiresIn: number }> => {
    try {
      return await api<{ accessToken: string; expiresIn: number }>('/users', {
        method: 'POST',
        body: JSON.stringify(data),
      });
    } catch {
      const users = getStored();
      const newUser: UserResponse = {
        id: users.length > 0 ? Math.max(...users.map((u) => u.id)) + 1 : 1,
        username: data.username,
        email: data.email,
        role: data.roles?.[0] ?? 'viewer',
        type: 'merchant',
        status: 'active',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      users.push(newUser);
      setStored(users);
      return { accessToken: '', expiresIn: 3600 };
    }
  },
};
