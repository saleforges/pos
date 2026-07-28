import { NavLink } from 'react-router-dom';
import { Building2, Users, Settings } from 'lucide-react';
import { Logo } from '@/components/ui/Logo';

export function Sidebar() {
  return (
    <aside className="flex h-screen w-64 flex-col bg-primary">
      <div className="flex h-16 items-center border-b border-white/10 px-6">
        <Logo />
      </div>

      <nav className="flex-1 space-y-1 px-3 py-4">
        <NavLink
          to="/accounts"
          className={({ isActive }) =>
            `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${isActive ? 'bg-secondary text-white' : 'text-neutral-300 hover:bg-white/5 hover:text-white'}`
          }
        >
          <Users size={18} />
          Accounts
        </NavLink>

        <NavLink
          to="/merchants"
          className={({ isActive }) =>
            `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${isActive ? 'bg-secondary text-white' : 'text-neutral-300 hover:bg-white/5 hover:text-white'}`
          }
        >
          <Building2 size={18} />
          Merchants
        </NavLink>

        <div className="mt-6 border-t border-white/10 pt-4">
          <p className="px-3 pb-2 text-xs font-medium uppercase tracking-wider text-neutral-500">
            Platform
          </p>

          <NavLink
            to="/settings"
            className={({ isActive }) =>
              `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${isActive ? 'bg-secondary text-white' : 'text-neutral-300 hover:bg-white/5 hover:text-white'}`
            }
          >
            <Settings size={18} />
            Settings
          </NavLink>
        </div>
      </nav>

      <div className="border-t border-white/10 p-4 text-xs text-neutral-400">
        v0.1.0 — Intools
      </div>
    </aside>
  );
}
