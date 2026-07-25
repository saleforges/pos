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

export const usersApi = {
  list: () => api<UserResponse[]>('/users'),
  create: (data: CreateUserRequest) =>
    api<{ accessToken: string; expiresIn: number }>('/users', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
};
