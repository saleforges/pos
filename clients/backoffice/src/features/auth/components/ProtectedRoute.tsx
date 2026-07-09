// features/auth/components/ProtectedRoute.tsx
import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

export function ProtectedRoute({ allowedRoles }: { allowedRoles?: string[] }) {
  const { user, isLoading } = useAuth();

  if (isLoading) return <div>Loading...</div>; // or a proper spinner

  if (!user) return <Navigate to="/login" replace />;

  if (allowedRoles && !allowedRoles.includes(user.role)) {
    return <Navigate to="/unauthorized" replace />;
  }

  return <Outlet />;
}