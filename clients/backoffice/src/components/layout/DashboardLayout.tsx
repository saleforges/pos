import { Navigate, Outlet } from 'react-router-dom';
import { Sidebar } from './Sidebar';
import { Header } from './Header';
import { useAuth } from '@/features/auth/hooks/useAuth';

export function DashboardLayout() {
  const { contexts, activeContext } = useAuth();

  // If the user has selectable contexts but hasn't picked one, ask first.
  if (contexts.length > 1 && !activeContext) {
    return <Navigate to="/select-branch" replace />;
  }

  return (
    <div className="flex h-screen bg-neutral-50">
      <Sidebar />
      <div className="flex flex-1 flex-col overflow-hidden">
        <Header />
        <main className="flex-1 overflow-y-auto p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}