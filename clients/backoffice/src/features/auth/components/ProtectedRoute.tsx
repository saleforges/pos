import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { hasRole } from '../api/authApi';

export function ProtectedRoute({
  allowedRoles,
  requiredPermissions,
}: {
  allowedRoles?: string[];
  requiredPermissions?: string[];
}) {
  const { user, isLoading } = useAuth();

  if (isLoading) return <div>Loading...</div>;

  if (!user) return <Navigate to="/login" replace />;

  if (allowedRoles && !hasRole(user, ...allowedRoles)) {
    return <Navigate to="/unauthorized" replace />;
  }

  if (
    requiredPermissions &&
    !requiredPermissions.every((p) => user.permissions?.includes(p))
  ) {
    return <Navigate to="/unauthorized" replace />;
  }

  return <Outlet />;
}
