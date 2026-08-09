import type { ReactNode } from 'react';
import localFont from 'next/font/local';
import './globals.css';

const inter = localFont({
  src: [
    { path: './fonts/Inter-Regular.ttf', weight: '400', style: 'normal' },
    { path: './fonts/Inter-Medium.ttf', weight: '500', style: 'normal' },
    { path: './fonts/Inter-SemiBold.ttf', weight: '600', style: 'normal' },
    { path: './fonts/Inter-Bold.ttf', weight: '700', style: 'normal' },
    { path: './fonts/Inter-ExtraBold.ttf', weight: '800', style: 'normal' },
  ],
  variable: '--font-inter',
  display: 'swap',
});

// The [locale] layout sets <html lang>; this root layout only provides font vars.
export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html className={inter.variable} suppressHydrationWarning>
      <body>{children}</body>
    </html>
  );
}
