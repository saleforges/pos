import { Store, Globe, Bell, Shield } from 'lucide-react';

const SETTINGS_SECTIONS = [
  {
    title: 'Store Profile',
    description: 'Manage your store name, address, and contact information.',
    icon: Store,
  },
  {
    title: 'Preferences',
    description: 'Configure currency, timezone, and regional settings.',
    icon: Globe,
  },
  {
    title: 'Notifications',
    description: 'Set up email and push notification preferences.',
    icon: Bell,
  },
  {
    title: 'Security',
    description: 'Manage passwords, session timeouts, and access controls.',
    icon: Shield,
  },
];

export default function SettingsPage() {
  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
      {SETTINGS_SECTIONS.map(({ title, description, icon: Icon }) => (
        <div
          key={title}
          className="cursor-pointer rounded-lg border border-neutral-200 bg-white p-5 transition-colors hover:border-neutral-300"
        >
          <div className="flex items-start gap-4">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-neutral-100">
              <Icon size={20} className="text-neutral-600" />
            </div>
            <div>
              <h3 className="font-display text-sm font-semibold text-neutral-900">{title}</h3>
              <p className="mt-1 text-sm text-neutral-500">{description}</p>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
