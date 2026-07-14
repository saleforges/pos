import { NavLink } from 'react-router-dom';
import { LayoutDashboard, Package, Users, Settings, ShoppingCart } from 'lucide-react';
import { Logo } from '@/components/ui/Logo';

const NAV_ITEMS = [
  { to: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/orders', label: 'Orders', icon: ShoppingCart },
  { to: '/products', label: 'Products', icon: Package },
  { to: '/staff', label: 'Staff', icon: Users },
  { to: '/settings', label: 'Settings', icon: Settings },
];

export function Sidebar() {
  return (
    <aside className="flex h-screen w-64 flex-col bg-primary">
      <div className="flex h-16 items-center border-b border-white/10 px-6">
        <Logo />
      </div>

      <nav className="flex-1 space-y-1 px-3 py-4">
        {NAV_ITEMS.map(({ to, label, icon: Icon }) => (
          <NavLink
            key={to}
            to={to}
            className={({ isActive }) =>
              `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${isActive
                ? 'bg-secondary text-white'
                : 'text-neutral-300 hover:bg-white/5 hover:text-white'
              }`
            }
          >
            <Icon size={18} />
            {label}
          </NavLink>
        ))}
      </nav>

      <div className="border-t border-white/10 p-4 text-xs text-neutral-400">
        v0.1.0 — Backoffice
      </div>
    </aside>
  );
}