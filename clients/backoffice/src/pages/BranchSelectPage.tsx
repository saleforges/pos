import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Building2, Store } from 'lucide-react';
import type { Locale } from '@pos/i18n';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { Logo } from '@/components/ui/Logo';
import { Alert } from '@/components/ui/Alert';
import { LanguageSwitcher } from '@/components/ui/LanguageSwitcher';

export default function BranchSelectPage() {
  const { user, isLoading, contexts, activeContext, selectContext } = useAuth();
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const [error, setError] = useState('');
  const [submittingId, setSubmittingId] = useState<number | null>(null);
  const [autoSelecting, setAutoSelecting] = useState(false);

  const groups = useMemo(() => {
    const map = new Map<number, { name: string; branches: typeof contexts }>();
    for (const ctx of contexts) {
      if (!map.has(ctx.merchant.id)) {
        map.set(ctx.merchant.id, { name: ctx.merchant.name, branches: [] });
      }
      map.get(ctx.merchant.id)!.branches.push(ctx);
    }
    return Array.from(map.values());
  }, [contexts]);

  useEffect(() => {
    if (isLoading) return;
    if (!user) {
      navigate('/login', { replace: true });
      return;
    }
    if (contexts.length === 0) {
      // platform-level user (superadmin) — no branch context needed
      navigate('/dashboard', { replace: true });
      return;
    }
    if (contexts.length === 1) {
      // single branch — auto-select it and go straight to the dashboard
      if (autoSelecting) return;
      setAutoSelecting(true);
      selectContext(contexts[0])
        .then(() => navigate('/dashboard', { replace: true }))
        .catch(() => {
          setAutoSelecting(false);
          setError(t('branchSelect.selectError'));
        });
    }
  }, [isLoading, user, contexts, navigate, selectContext, autoSelecting, t]);

  if (isLoading || autoSelecting) return <div>{t('branchSelect.loading')}</div>;

  const handleSelect = async (ctx: (typeof contexts)[number]) => {
    if (submittingId !== null) return;
    setError('');
    setSubmittingId(ctx.branch.id);
    try {
      await selectContext(ctx);
      navigate('/dashboard', { replace: true });
    } catch {
      setError(t('branchSelect.selectError'));
      setSubmittingId(null);
    }
  };

  const locale = (i18n.resolvedLanguage ?? i18n.language) as Locale;

  return (
    <div className="relative flex min-h-screen">
      {/* Brand panel */}
      <div className="hidden w-1/2 flex-col items-start justify-between bg-primary p-12 lg:flex">
        <Logo variant="white" />
        <div>
          <h2 className="font-display text-3xl font-bold text-white">
            {t('branchSelect.heroTitle1')}
            <br />
            {t('branchSelect.heroTitle2')}
          </h2>
          <p className="mt-3 max-w-sm text-sm text-neutral-400">
            {t('branchSelect.heroSubtitle')}
          </p>
        </div>
        <p className="text-xs text-neutral-500">
          {t('login.footer', { year: new Date().getFullYear() })}
        </p>
      </div>

      {/* Form panel */}
      <div className="flex w-full items-center justify-center bg-neutral-50 px-6 lg:w-1/2">
        <div className="absolute right-6 top-6">
          <LanguageSwitcher locale={locale} />
        </div>
        <div className="w-full max-w-md">
          <div className="mb-8 lg:hidden">
            <Logo />
          </div>

          <h1 className="font-display text-2xl font-bold text-neutral-900">
            {t('branchSelect.title')}
          </h1>
          <p className="mt-1 text-sm text-neutral-500">
            {t('branchSelect.signedInAs')}{' '}
            <span className="font-medium text-neutral-700">{user?.username}</span>
          </p>

          {error && <div className="mt-6"><Alert variant="danger">{error}</Alert></div>}

          <div className="mt-8 space-y-6">
            {groups.map((group) => (
              <div key={group.name}>
                <div className="mb-2 flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wide text-neutral-500">
                  <Store size={14} />
                  {group.name}
                </div>
                <div className="space-y-2">
                  {group.branches.map((ctx) => (
                    <button
                      key={ctx.branch.id}
                      onClick={() => handleSelect(ctx)}
                      disabled={submittingId !== null}
                      className={`flex w-full items-center gap-3 rounded-lg border bg-white px-4 py-3 text-left transition-colors ${
                        activeContext?.branch.id === ctx.branch.id
                          ? 'border-secondary ring-1 ring-secondary'
                          : 'border-neutral-200 hover:border-neutral-300 hover:bg-neutral-50'
                      } disabled:opacity-50`}
                    >
                      <Building2 size={18} className="shrink-0 text-neutral-400" />
                      <span className="font-medium text-neutral-900">{ctx.branch.name}</span>
                      {activeContext?.branch.id === ctx.branch.id && (
                        <span className="ml-auto rounded-full bg-secondary/10 px-2 py-0.5 text-xs font-medium text-secondary">
                          {t('branchSelect.active')}
                        </span>
                      )}
                    </button>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}