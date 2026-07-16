import { LogOut, User } from 'lucide-react';
import { useLocation } from 'react-router-dom';
import { useAuth } from '@/features/auth/hooks/useAuth';

const PAGE_TITLES: Record<string, string> = {
  dashboard: 'Dashboard',
  orders: 'Orders',
  products: 'Products',
  staff: 'Staff',
  roles: 'Roles',
  merchants: 'Merchants',
  settings: 'Settings',
};

export function Header() {
  const { user, logout } = useAuth();
  const location = useLocation();
  const currentPath = location.pathname.split('/')[1];
  const title = PAGE_TITLES[currentPath] ?? 'Dashboard';

  return (
    <header className="flex h-16 items-center justify-between border-b border-neutral-200 bg-white px-6">
      <div>
        <h1 className="font-display text-lg font-semibold text-neutral-900">
          {title}
        </h1>
      </div>

      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2 text-sm text-neutral-600">
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-neutral-100">
            <User size={16} />
          </div>
          <span className="font-medium text-neutral-900">{user?.username}</span>
        </div>
        <button
          onClick={logout}
          className="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm text-neutral-500 transition-colors hover:bg-neutral-100 hover:text-neutral-900"
        >
          <LogOut size={16} />
          Log out
        </button>
      </div>
    </header>
  );
}