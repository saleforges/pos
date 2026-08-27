import { LogOut, User, Store, ChevronDown } from 'lucide-react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import type { Locale } from '@pos/i18n';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { LanguageSwitcher } from '@/components/ui/LanguageSwitcher';

export function Header() {
  const { user, contexts, activeContext, logout } = useAuth();
  const { t, i18n } = useTranslation();
  const location = useLocation();
  const navigate = useNavigate();
  const currentPath = location.pathname.split('/')[1];
  const titles: Record<string, string> = {
    dashboard: t('pages.dashboard'),
    orders: t('pages.orders'),
    catalog: t('pages.catalog'),
    staff: t('pages.staff'),
    roles: t('pages.roles'),
    merchants: t('pages.merchants'),
    settings: t('pages.settings'),
  };
  const title = titles[currentPath] ?? t('pages.dashboard');

  const showContextSwitch = contexts.length > 1;
  const locale = (i18n.resolvedLanguage ?? i18n.language) as Locale;

  return (
    <header className="flex h-16 items-center justify-between border-b border-neutral-200 bg-white px-6">
      <div>
        <h1 className="font-display text-lg font-semibold text-neutral-900">
          {title}
        </h1>
      </div>

      <div className="flex items-center gap-4">
        {activeContext && (
          <div className="hidden items-center gap-2 text-sm text-neutral-600 md:flex">
            <Store size={15} className="text-neutral-400" />
            <span className="font-medium text-neutral-900">{activeContext.branch.name}</span>
            <span className="text-neutral-400">·</span>
            <span>{activeContext.merchant.name}</span>
          </div>
        )}
        <LanguageSwitcher locale={locale} />
        {showContextSwitch && (
          <button
            onClick={() => navigate('/select-branch')}
            className="flex items-center gap-1 rounded-md border border-neutral-200 px-3 py-1.5 text-sm text-neutral-600 transition-colors hover:bg-neutral-100 hover:text-neutral-900"
          >
            {t('common.switchBranch')}
            <ChevronDown size={14} />
          </button>
        )}
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
          {t('common.logout')}
        </button>
      </div>
    </header>
  );
}