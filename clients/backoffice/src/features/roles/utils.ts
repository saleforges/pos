import { parsePermission } from './types';

export function groupPermissionsByResource(permissions: string[]) {
  const groups: Record<string, string[]> = {};
  for (const perm of permissions) {
    const { resource, action } = parsePermission(perm);
    groups[resource] ??= [];
    groups[resource].push(action);
  }
  return groups;
}
