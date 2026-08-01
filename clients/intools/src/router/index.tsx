import { createBrowserRouter, Navigate } from 'react-router-dom';
import { ProtectedRoute } from '@/features/auth/ProtectedRoute';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import LoginPage from '@/pages/LoginPage';
import AccountsPage from '@/features/accounts/pages/AccountsPage';
import AccountDetailPage from '@/features/accounts/pages/AccountDetailPage';
import MerchantsPage from '@/features/merchants/pages/MerchantsPage';
import MerchantDetailPage from '@/features/merchants/pages/MerchantDetailPage';

export const router = createBrowserRouter([
  { path: '/', element: <Navigate to="/login" replace /> },
  { path: '/login', element: <LoginPage /> },
  {
    element: <ProtectedRoute allowedRoles={['superadmin']} />,
    children: [
      {
        element: <DashboardLayout />,
        children: [
          { path: '/accounts', element: <AccountsPage /> },
          { path: '/accounts/:id', element: <AccountDetailPage /> },
          { path: '/merchants', element: <MerchantsPage /> },
          { path: '/merchants/:id', element: <MerchantDetailPage /> },
          { path: '/settings', element: <div className="p-4 text-neutral-500">Platform settings coming soon.</div> },
        ],
      },
    ],
  },
  { path: '*', element: <div className="flex h-screen items-center justify-center text-neutral-500">404 — Page Not Found</div> },
]);
