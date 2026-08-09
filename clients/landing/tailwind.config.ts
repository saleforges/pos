import type { Config } from 'tailwindcss';

const config: Config = {
  content: [
    './src/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          blue: '#1D5FF3',
          bluedark: '#123F9E',
          green: '#0BB489',
          navy: '#192539',
          navy2: '#0F1622',
        },
        ink: '#101828',
        muted: '#667085',
        muted2: '#98A2B3',
        line: '#EAECF0',
        line2: '#F2F4F7',
        soft: '#F8FAFC',
      },
      fontFamily: {
        sans: ['var(--font-inter)', 'system-ui', 'Arial', 'sans-serif'],
      },
      maxWidth: {
        container: '1160px',
      },
      boxShadow: {
        card: '0 12px 32px -16px rgba(16,24,40,.2)',
        frame: '0 24px 64px -24px rgba(16,24,40,.24)',
        btn: '0 6px 20px rgba(29,95,243,.28)',
      },
      keyframes: {
        reveal: {
          '0%': { opacity: '0', transform: 'translateY(16px)' },
          '100%': { opacity: '1', transform: 'none' },
        },
      },
      animation: {
        reveal: 'reveal .6s ease forwards',
      },
    },
  },
  plugins: [],
};

export default config;
