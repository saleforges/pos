import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Building2, Store } from 'lucide-react';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { Logo } from '@/components/ui/Logo';
import { Alert } from '@/components/ui/Alert';

export default function BranchSelectPage() {
  const { user, isLoading, contexts, activeContext, selectContext } = useAuth();
  const navigate = useNavigate();
  const [error, setError] = useState('');
  const [submittingId, setSubmittingId] = useState<number | null>(null);

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
  }, [isLoading, user, contexts, navigate]);

  if (isLoading) return <div>Loading...</div>;

  const handleSelect = async (ctx: (typeof contexts)[number]) => {
    if (submittingId !== null) return;
    setError('');
    setSubmittingId(ctx.branch.id);
    try {
      await selectContext(ctx);
      navigate('/dashboard', { replace: true });
    } catch {
      setError('Failed to select branch. Please try again.');
      setSubmittingId(null);
    }
  };

  return (
    <div className="flex min-h-screen">
      {/* Brand panel */}
      <div className="hidden w-1/2 flex-col justify-between bg-primary p-12 lg:flex">
        <Logo />
        <div>
          <h2 className="font-display text-3xl font-bold text-white">
            Pick a branch,<br />get to work.
          </h2>
          <p className="mt-3 max-w-sm text-sm text-neutral-400">
            You have access to more than one branch. Select the store you want
            to manage right now.
          </p>
        </div>
        <p className="text-xs text-neutral-500">
          &copy; {new Date().getFullYear()} SaleForges. Open source POS.
        </p>
      </div>

      {/* Form panel */}
      <div className="flex w-full items-center justify-center bg-neutral-50 px-6 lg:w-1/2">
        <div className="w-full max-w-md">
          <div className="mb-8 lg:hidden">
            <Logo />
          </div>

          <h1 className="font-display text-2xl font-bold text-neutral-900">
            Select branch
          </h1>
          <p className="mt-1 text-sm text-neutral-500">
            Signed in as <span className="font-medium text-neutral-700">{user?.username}</span>
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
                          Active
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
