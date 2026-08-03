import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import type { Locale } from '@pos/i18n';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { Logo } from '@/components/ui/Logo';
import { ApiError } from '@/lib/api';
import { Alert } from '@/components/ui/Alert';
import { LanguageSwitcher } from '@/components/ui/LanguageSwitcher';

export default function LoginPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { login } = useAuth();
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsSubmitting(true);
    try {
      await login(username, password);
      navigate('/select-branch');
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        setError(t('login.invalidCredentials'));
      } else {
        setError(t('login.somethingWentWrong'));
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const locale = (i18n.resolvedLanguage ?? i18n.language) as Locale;

  return (
    <div className="relative flex min-h-screen">
      {/* Brand panel */}
      <div className="hidden w-1/2 flex-col justify-between bg-primary p-12 lg:flex">
        <Logo />
        <div>
          <h2 className="font-display text-3xl font-bold text-white">
            {t('login.heroTitle1')}
            <br />
            {t('login.heroTitle2')}
          </h2>
          <p className="mt-3 max-w-sm text-sm text-neutral-400">
            {t('login.heroSubtitle')}
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
        <div className="w-full max-w-sm">
          <div className="mb-8 lg:hidden">
            <Logo />
          </div>

          <h1 className="font-display text-2xl font-bold text-neutral-900">
            {t('login.welcomeBack')}
          </h1>
          <p className="mt-1 text-sm text-neutral-500">
            {t('login.subtitle')}
          </p>

          <form onSubmit={handleSubmit} className="mt-8 space-y-4">
            {error && <Alert variant="danger">{error}</Alert>}

            <div>
              <label className="mb-1.5 block text-sm font-medium text-neutral-700">
                {t('login.username')}
              </label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="w-full rounded-md border border-neutral-200 px-3 py-2 text-sm text-neutral-900 outline-none transition-colors focus:border-secondary focus:ring-1 focus:ring-secondary"
                required
                autoFocus
              />
            </div>

            <div>
              <label className="mb-1.5 block text-sm font-medium text-neutral-700">
                {t('login.password')}
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full rounded-md border border-neutral-200 px-3 py-2 text-sm text-neutral-900 outline-none transition-colors focus:border-secondary focus:ring-1 focus:ring-secondary"
                required
              />
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full rounded-md bg-primary py-2.5 text-sm font-semibold text-white transition-colors hover:bg-primary-hover disabled:opacity-50"
            >
              {isSubmitting ? t('login.loggingIn') : t('login.logIn')}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}