import { createBrowserRouter, Navigate } from 'react-router-dom';
import { ProtectedRoute } from '@/features/auth/components/ProtectedRoute';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import LoginPage from '@/pages/LoginPage';
import DashboardPage from '@/pages/DashboardPage';
import OrdersPage from '@/pages/OrdersPage';
import ProductsPage from '@/pages/ProductsPage';
import StaffPage from '@/pages/StaffPage';
import MerchantsPage from '@/pages/MerchantsPage';
import RolesPage from '@/pages/RolesPage';

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
          { path: '/staff', element: <StaffPage /> },
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
          { path: '/settings', element: <div>Admin Settings</div> },
        ],
      },
    ],
  },
  { path: '*', element: <div>404 - Page Not Found</div> },
]);
