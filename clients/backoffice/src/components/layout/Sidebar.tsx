import { NavLink } from 'react-router-dom';
import { LayoutDashboard, Package, Users, Settings, ShoppingCart, Building2, Shield } from 'lucide-react';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { hasRole } from '@/features/auth/api/authApi';
import { Logo } from '@/components/ui/Logo';

export function Sidebar() {
  const { user } = useAuth();
  const isSuperadmin = hasRole(user, 'superadmin');

  return (
    <aside className="flex h-screen w-64 flex-col bg-primary">
      <div className="flex h-16 items-center border-b border-white/10 px-6">
        <Logo />
      </div>

      <nav className="flex-1 space-y-1 px-3 py-4">
        <NavLink
          to="/dashboard"
          className={({ isActive }) =>
            `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${isActive ? 'bg-secondary text-white' : 'text-neutral-300 hover:bg-white/5 hover:text-white'}`
          }
        >
          <LayoutDashboard size={18} />
          Dashboard
        </NavLink>

        <NavLink
          to="/orders"
          className={({ isActive }) =>
            `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${isActive ? 'bg-secondary text-white' : 'text-neutral-300 hover:bg-white/5 hover:text-white'}`
          }
        >
          <ShoppingCart size={18} />
          Orders
        </NavLink>

        <NavLink
          to="/products"
          className={({ isActive }) =>
            `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${isActive ? 'bg-secondary text-white' : 'text-neutral-300 hover:bg-white/5 hover:text-white'}`
          }
        >
          <Package size={18} />
          Products
        </NavLink>

        <NavLink
          to="/staff"
          className={({ isActive }) =>
            `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${isActive ? 'bg-secondary text-white' : 'text-neutral-300 hover:bg-white/5 hover:text-white'}`
          }
        >
          <Users size={18} />
          Staff
        </NavLink>

        {isSuperadmin && (
          <>
            <NavLink
              to="/merchants"
              className={({ isActive }) =>
                `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${isActive ? 'bg-secondary text-white' : 'text-neutral-300 hover:bg-white/5 hover:text-white'}`
              }
            >
              <Building2 size={18} />
              Merchants
            </NavLink>
            <NavLink
              to="/roles"
              className={({ isActive }) =>
                `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${isActive ? 'bg-secondary text-white' : 'text-neutral-300 hover:bg-white/5 hover:text-white'}`
              }
            >
              <Shield size={18} />
              Roles
            </NavLink>
          </>
        )}

        <NavLink
          to="/settings"
          className={({ isActive }) =>
            `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${isActive ? 'bg-secondary text-white' : 'text-neutral-300 hover:bg-white/5 hover:text-white'}`
          }
        >
          <Settings size={18} />
          Settings
        </NavLink>
      </nav>

      <div className="border-t border-white/10 p-4 text-xs text-neutral-400">
        v0.1.0 — Backoffice
      </div>
    </aside>
  );
}
