import { NavLink } from 'react-router-dom';
import { LayoutDashboard, Package, Settings, ShoppingCart, Building2, Shield, UserCog, Users } from 'lucide-react';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { hasRole } from '@/features/auth/api/authApi';
import { useAllPermissions } from '@/features/auth/hooks/usePermission';
import { Logo } from '@/components/ui/Logo';

export function Sidebar() {
  const { user } = useAuth();
  const isSuperadmin = hasRole(user, 'superadmin');
  const showPlatformAdmin = useAllPermissions(['merchant.manage', 'user.create']);

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
              to="/users"
              className={({ isActive }) =>
                `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${isActive ? 'bg-secondary text-white' : 'text-neutral-300 hover:bg-white/5 hover:text-white'}`
              }
            >
              <UserCog size={18} />
              Users
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

        {showPlatformAdmin && (
          <>
            <div className="border-t border-white/10 pt-4" />
            <div className="px-3 pb-1 pt-1 text-[10px] font-semibold uppercase tracking-widest text-neutral-500">
              Platform Admin
            </div>
            <NavLink
              to="/admin/accounts"
              className={({ isActive }) =>
                `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${isActive ? 'bg-secondary text-white' : 'text-neutral-300 hover:bg-white/5 hover:text-white'}`
              }
            >
              <Users size={18} />
              Accounts
            </NavLink>
            <NavLink
              to="/admin/merchants"
              className={({ isActive }) =>
                `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${isActive ? 'bg-secondary text-white' : 'text-neutral-300 hover:bg-white/5 hover:text-white'}`
              }
            >
              <Building2 size={18} />
              Merchants
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
