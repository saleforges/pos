import { CheckCircle2, Info, AlertTriangle, XCircle, type LucideIcon } from 'lucide-react';
import type { ReactNode } from 'react';

type AlertVariant = 'default' | 'success' | 'info' | 'warning' | 'danger';

const VARIANT_STYLES: Record<AlertVariant, { bg: string; border: string; text: string; icon: LucideIcon }> = {
  default: {
    bg: 'bg-neutral-100',
    border: 'border-neutral-200',
    text: 'text-neutral-700',
    icon: Info,
  },
  success: {
    bg: 'bg-secondary/5',
    border: 'border-secondary/20',
    text: 'text-secondary-hover',
    icon: CheckCircle2,
  },
  info: {
    bg: 'bg-info/5',
    border: 'border-info/20',
    text: 'text-info-hover',
    icon: Info,
  },
  warning: {
    bg: 'bg-tertiary/5',
    border: 'border-tertiary/20',
    text: 'text-tertiary-hover',
    icon: AlertTriangle,
  },
  danger: {
    bg: 'bg-danger/5',
    border: 'border-danger/20',
    text: 'text-danger-hover',
    icon: XCircle,
  },
};

export function Alert({ variant = 'default', children }: { variant?: AlertVariant; children: ReactNode }) {
  const { bg, border, text, icon: Icon } = VARIANT_STYLES[variant];

  return (
    <div className={`flex items-start gap-2 rounded-md border px-3 py-2 text-sm ${bg} ${border} ${text}`}>
      <Icon size={16} className="mt-0.5 shrink-0" />
      <span>{children}</span>
    </div>
  );
}
