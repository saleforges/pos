import { createBrowserRouter, Navigate } from 'react-router-dom';
import { lazy, Suspense } from 'react';
import { ProtectedRoute } from '@/features/auth/components/ProtectedRoute';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import LoginPage from '@/pages/LoginPage';
import DashboardPage from '@/pages/DashboardPage';
import OrdersPage from '@/pages/OrdersPage';
import ProductsPage from '@/pages/ProductsPage';
import MerchantsPage from '@/pages/MerchantsPage';
import MerchantDetailPage from '@/pages/MerchantDetailPage';
import RolesPage from '@/pages/RolesPage';
import UsersPage from '@/pages/UsersPage';

const AdminMerchantsPage = lazy(() => import('@/features/platform-admin/pages/MerchantsPage'));
const AdminAccountsPage = lazy(() => import('@/features/platform-admin/pages/AccountsPage'));

export const router = createBrowserRouter([
  { path: '/', element: <Navigate to="/login" replace /> },
  { path: '/login', element: <LoginPage /> },
  {
    element: <ProtectedRoute />,
    children: [
      {
        element: <DashboardLayout />,
        children: [
          { path: '/dashboard', element: <DashboardPage /> },
          { path: '/orders', element: <OrdersPage /> },
          { path: '/products', element: <ProductsPage /> },
        ],
      },
    ],
  },
  {
    element: <ProtectedRoute allowedRoles={['superadmin', 'admin']} />,
    children: [
      {
        element: <DashboardLayout />,
        children: [
          { path: '/roles', element: <RolesPage /> },
          { path: '/merchants', element: <MerchantsPage /> },
          { path: '/merchants/:merchantId', element: <MerchantDetailPage /> },
          { path: '/settings', element: <div>Admin Settings</div> },
        ],
      },
    ],
  },
  {
    element: <ProtectedRoute allowedRoles={['superadmin']} />,
    children: [
      {
        element: <DashboardLayout />,
        children: [
          { path: '/users', element: <UsersPage /> },
          { path: '/admin/merchants', element: <Suspense fallback={<div>Loading...</div>}><AdminMerchantsPage /></Suspense> },
          { path: '/admin/accounts', element: <Suspense fallback={<div>Loading...</div>}><AdminAccountsPage /></Suspense> },
        ],
      },
    ],
  },
  { path: '*', element: <div>404 - Page Not Found</div> },
]);
