const BASE_URL = import.meta.env.VITE_API_URL ?? '/api/v1';

export async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };

  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers,
    credentials: 'include',
  });

  if (!res.ok) {
    const text = await res.text();
    throw new ApiError(res.status, text || res.statusText);
  }

  if (res.status === 204) return undefined as T;

  const text = await res.text();
  const body = JSON.parse(text);
  return 'data' in body ? body.data : body;
}

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

export interface PageMeta {
  total: number;
  offset: number;
  limit: number;
  count: number;
}

export interface Paginated<T> {
  data: T[];
  pagination: PageMeta;
}

/** Like api(), but for list endpoints answering with a { message, data, pagination } envelope. */
export async function apiPaginated<T>(path: string, options: RequestInit = {}): Promise<Paginated<T>> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };

  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers,
    credentials: 'include',
  });

  if (!res.ok) {
    const text = await res.text();
    throw new ApiError(res.status, text || res.statusText);
  }

  const body = JSON.parse(await res.text());
  return { data: body.data ?? [], pagination: body.pagination };
}
