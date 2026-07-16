import { useAuth } from './useAuth';

export function usePermission(permission: string): boolean {
  const { user } = useAuth();
  if (!user) return false;
  if (user.roles.some((r) => r.name === 'superadmin')) return true;
  return user.permissions?.includes(permission) ?? false;
}

export function useAnyPermission(permissions: string[]): boolean {
  const { user } = useAuth();
  if (!user) return false;
  if (user.roles.some((r) => r.name === 'superadmin')) return true;
  return permissions.some((p) => user.permissions?.includes(p));
}

export function useAllPermissions(permissions: string[]): boolean {
  const { user } = useAuth();
  if (!user) return false;
  if (user.roles.some((r) => r.name === 'superadmin')) return true;
  return permissions.every((p) => user.permissions?.includes(p));
}
