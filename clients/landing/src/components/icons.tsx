type IconProps = { className?: string };

const s = (className?: string) => ({
  viewBox: '0 0 24 24',
  fill: 'none',
  stroke: 'currentColor',
  strokeWidth: 1.8,
  strokeLinecap: 'round' as const,
  strokeLinejoin: 'round' as const,
  className,
});

export const IconPOS = ({ className }: IconProps) => (
  <svg {...s(className)}>
    <rect x="3" y="4" width="18" height="12" rx="2" />
    <path d="M3 20h18M8 16v4M16 16v4" />
  </svg>
);

export const IconBox = ({ className }: IconProps) => (
  <svg {...s(className)}>
    <path d="M21 8l-9-5-9 5 9 5 9-5zM3 8v8l9 5 9-5V8" />
    <path d="M12 13v8" />
  </svg>
);

export const IconCard = ({ className }: IconProps) => (
  <svg {...s(className)}>
    <rect x="2" y="5" width="20" height="14" rx="2" />
    <path d="M2 10h20M6 15h4" />
  </svg>
);

export const IconUsers = ({ className }: IconProps) => (
  <svg {...s(className)}>
    <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
    <circle cx="9" cy="7" r="4" />
    <path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" />
  </svg>
);

export const IconFlow = ({ className }: IconProps) => (
  <svg {...s(className)}>
    <circle cx="6" cy="6" r="2.4" />
    <circle cx="18" cy="18" r="2.4" />
    <path d="M8.4 6H16a2 2 0 0 1 2 2v7.5M6 8.4V16a2 2 0 0 0 2 2h7.5" />
  </svg>
);

export const IconChart = ({ className }: IconProps) => (
  <svg {...s(className)}>
    <path d="M4 20V10M10 20V4M16 20v-7M22 20H2" />
  </svg>
);

export const IconCheck = ({ className }: IconProps) => (
  <svg
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth={2.4}
    strokeLinecap="round"
    strokeLinejoin="round"
    className={className}
  >
    <path d="M20 6L9 17l-5-5" />
  </svg>
);

export const IconX = ({ className }: IconProps) => (
  <svg viewBox="0 0 24 24" fill="currentColor" className={className}>
    <path d="M18.9 2H22l-7 8 8.2 12h-6.4l-5-6.6L5.5 22H2.4l7.5-8.6L2 2h6.6l4.5 6L18.9 2z" />
  </svg>
);

export const IconLinkedIn = ({ className }: IconProps) => (
  <svg viewBox="0 0 24 24" fill="currentColor" className={className}>
    <path d="M6.94 5a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM3.3 8.5h3.3V22H3.3V8.5zM9.4 8.5h3.16v1.84h.05c.44-.83 1.5-1.7 3.1-1.7 3.32 0 3.93 2.18 3.93 5.02V22h-3.3v-6.1c0-1.45-.03-3.32-2.02-3.32-2.03 0-2.34 1.58-2.34 3.21V22H9.4V8.5z" />
  </svg>
);

export const IconInstagram = ({ className }: IconProps) => (
  <svg
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth={1.8}
    className={className}
  >
    <rect x="3" y="3" width="18" height="18" rx="5" />
    <circle cx="12" cy="12" r="4" />
    <circle cx="17.5" cy="6.5" r="1" fill="currentColor" stroke="none" />
  </svg>
);
