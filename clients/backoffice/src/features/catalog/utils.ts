import type { ApiError } from '@/lib/api';

const idrFormatter = new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
});

export function formatIDR(amount: number): string {
  return idrFormatter.format(amount);
}

export function formatPriceRange(min: number, max: number): string {
  return min === max ? formatIDR(min) : `${formatIDR(min)} – ${formatIDR(max)}`;
}

/** Backend error bodies are JSON like {"code":"CAT001","message":"..."} while
 *  ApiError carries the raw text — surface the human-readable part. */
export function extractErrorMessage(err: unknown, fallback: string): string {
  if (err instanceof Error && 'status' in err) {
    const apiErr = err as ApiError;
    try {
      const parsed = JSON.parse(apiErr.message);
      if (parsed?.message) return String(parsed.message);
    } catch {
      // not JSON — fall through
    }
    if (apiErr.message) return apiErr.message;
  }
  return fallback;
}