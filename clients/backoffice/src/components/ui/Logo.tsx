import { Icon, Logo as BrandLogo } from '@saleforges/ui';

export type LogoVariant = 'color' | 'mono' | 'white';

export function Logo({
  collapsed = false,
  variant = 'color',
  height = 32,
}: {
  collapsed?: boolean;
  variant?: LogoVariant;
  height?: number;
}) {
  if (collapsed) {
    return <Icon variant={variant} size={height} />;
  }
  return <BrandLogo variant={variant} height={height} />;
}
