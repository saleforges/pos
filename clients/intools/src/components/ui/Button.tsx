import type { ReactNode, ButtonHTMLAttributes } from 'react';

type ButtonVariant = 'primary' | 'secondary' | 'danger' | 'ghost';

const VARIANT_STYLES: Record<ButtonVariant, string> = {
  primary: 'bg-primary text-white hover:bg-primary-hover disabled:opacity-50',
  secondary: 'bg-white text-neutral-700 border border-neutral-200 hover:bg-neutral-50 disabled:opacity-50',
  danger: 'bg-danger text-white hover:bg-danger-hover disabled:opacity-50',
  ghost: 'text-neutral-500 hover:bg-neutral-100 hover:text-neutral-900 disabled:opacity-50',
};

export function Button({
  variant = 'primary',
  children,
  className = '',
  ...props
}: {
  variant?: ButtonVariant;
  children: ReactNode;
  className?: string;
} & ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      className={`inline-flex items-center justify-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors ${VARIANT_STYLES[variant]} ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}
