import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '@/features/auth/useAuth';
import { Logo } from '@/components/ui/Logo';
import { ApiError } from '@/lib/api';
import { Alert } from '@/components/ui/Alert';

export default function LoginPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsSubmitting(true);
    try {
      await login(username, password);
      navigate('/accounts');
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        setError('Invalid username or password');
      } else if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('Something went wrong. Please try again.');
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="flex min-h-screen">
      <div className="hidden w-1/2 flex-col justify-between bg-primary p-12 lg:flex">
        <Logo />
        <div>
          <h2 className="font-display text-3xl font-bold text-white">
            Internal tools,<br />built for the team.
          </h2>
          <p className="mt-3 max-w-sm text-sm text-neutral-400">
            Manage accounts, merchants, and platform settings — one place
            for SaleForges staff.
          </p>
        </div>
        <p className="text-xs text-neutral-500">
          &copy; {new Date().getFullYear()} SaleForges. Open source POS.
        </p>
      </div>

      <div className="flex w-full items-center justify-center bg-neutral-50 px-6 lg:w-1/2">
        <div className="w-full max-w-sm">
          <div className="mb-8 lg:hidden">
            <Logo />
          </div>

          <h1 className="font-display text-2xl font-bold text-neutral-900">
            Internal Tools
          </h1>
          <p className="mt-1 text-sm text-neutral-500">
            Staff authentication required
          </p>

          <form onSubmit={handleSubmit} className="mt-8 space-y-4">
            {error && <Alert variant="danger">{error}</Alert>}

            <div>
              <label className="mb-1.5 block text-sm font-medium text-neutral-700">
                Username
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
                Password
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
              {isSubmitting ? 'Logging in...' : 'Log in'}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
