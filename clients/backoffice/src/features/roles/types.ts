export interface Role {
  id: number;
  name: string;
  description: string;
  permissions: string[];
  is_system: boolean;
}

export interface ParsedPermission {
  raw: string;
  resource: string;
  action: string;
}

export function parsePermission(permission: string): ParsedPermission {
  const [resource, action] = permission.split('.');
  return { raw: permission, resource, action };
}
