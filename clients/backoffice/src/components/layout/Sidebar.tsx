import { NavLink } from 'react-router-dom';
import { LayoutDashboard, Package, Settings, ShoppingCart, Building2, Shield, UserCog } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { hasRole } from '@/features/auth/api/authApi';
import { Logo } from '@/components/ui/Logo';

export function Sidebar() {
  const { user } = useAuth();
  const { t } = useTranslation();
  const isSuperadmin = hasRole(user, 'superadmin');

  return (
    <aside className="flex h-screen w-64 flex-col bg-primary">
      <div className="flex h-16 items-center border-b border-white/10 px-6">
        <Logo variant="white" />
      </div>

      <nav className="flex-1 space-y-1 px-3 py-4">
        <NavLink
          to="/dashboard"
          className={({ isActive }) =>
            `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${isActive ? 'bg-secondary text-white' : 'text-neutral-300 hover:bg-white/5 hover:text-white'}`
          }
        >
          <LayoutDashboard size={18} />
          {t('pages.dashboard')}
        </NavLink>

        <NavLink
          to="/orders"
          className={({ isActive }) =>
            `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${isActive ? 'bg-secondary text-white' : 'text-neutral-300 hover:bg-white/5 hover:text-white'}`
          }
        >
          <ShoppingCart size={18} />
          {t('pages.orders')}
        </NavLink>

        <NavLink
          to="/products"
          className={({ isActive }) =>
            `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${isActive ? 'bg-secondary text-white' : 'text-neutral-300 hover:bg-white/5 hover:text-white'}`
          }
        >
          <Package size={18} />
          {t('pages.products')}
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
              {t('pages.merchants')}
            </NavLink>
            <NavLink
              to="/users"
              className={({ isActive }) =>
                `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${isActive ? 'bg-secondary text-white' : 'text-neutral-300 hover:bg-white/5 hover:text-white'}`
              }
            >
              <UserCog size={18} />
              {t('pages.users')}
            </NavLink>
            <NavLink
              to="/roles"
              className={({ isActive }) =>
                `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${isActive ? 'bg-secondary text-white' : 'text-neutral-300 hover:bg-white/5 hover:text-white'}`
              }
            >
              <Shield size={18} />
              {t('pages.roles')}
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
          {t('pages.settings')}
        </NavLink>
      </nav>

      <div className="border-t border-white/10 p-4 text-xs text-neutral-400">
        v0.1.0 — Backoffice
      </div>
    </aside>
  );
}
