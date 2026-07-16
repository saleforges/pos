import { useEffect, useState } from 'react';
import { rolesApi } from '@/features/roles/api/rolesApi';
import { groupPermissionsByResource } from '@/features/roles/utils';
import type { Role } from '@/features/roles/types';

export default function RolesPage() {
  const [roles, setRoles] = useState<Role[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    rolesApi
      .list()
      .then(setRoles)
      .finally(() => setIsLoading(false));
  }, []);

  if (isLoading) return <div>Loading roles...</div>;

  return (
    <div className="space-y-4">
      <h1 className="font-display text-2xl font-bold text-neutral-900">Roles</h1>

      <div className="grid gap-4">
        {roles.map((role) => {
          const grouped = groupPermissionsByResource(role.permissions);
          return (
            <div key={role.id} className="rounded-lg border border-neutral-200 bg-white p-5">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="font-display text-lg font-semibold text-neutral-900">
                    {role.name}
                  </h2>
                  <p className="text-sm text-neutral-500">{role.description}</p>
                </div>
                {role.is_system && (
                  <span className="rounded-full bg-neutral-100 px-2.5 py-1 text-xs font-medium text-neutral-600">
                    System role
                  </span>
                )}
              </div>

              <div className="mt-4 flex flex-wrap gap-4">
                {Object.entries(grouped).map(([resource, actions]) => (
                  <div key={resource} className="min-w-[140px]">
                    <div className="text-xs font-semibold uppercase tracking-wide text-neutral-400">
                      {resource}
                    </div>
                    <div className="mt-1 flex flex-wrap gap-1">
                      {actions.map((action) => (
                        <span
                          key={action}
                          className="rounded bg-secondary/10 px-1.5 py-0.5 text-xs font-medium text-secondary-hover"
                        >
                          {action}
                        </span>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
