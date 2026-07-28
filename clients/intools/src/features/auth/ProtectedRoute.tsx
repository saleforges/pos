import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from './useAuth';
import { PageLoader } from '@/components/ui/PageLoader';

export function ProtectedRoute({ allowedRoles }: { allowedRoles?: string[] }) {
  const { user, isLoading } = useAuth();

  if (isLoading) return <PageLoader />;
  if (!user) return <Navigate to="/login" replace />;

  if (allowedRoles && !allowedRoles.some((r) => user.roles?.some((ur) => ur.name === r))) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
}
