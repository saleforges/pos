// router/index.tsx
import { createBrowserRouter, Navigate } from 'react-router-dom';
import { ProtectedRoute } from '@/features/auth/components/ProtectedRoute';
import LoginPage from '@/pages/LoginPage';
// import DashboardPage from '@/pages/DashboardPage';

export const router = createBrowserRouter([
  { path: '/', element: <Navigate to="/login" replace /> },
  { path: '/login', element: <LoginPage /> },
  {
    element: <ProtectedRoute />,
    children: [
      { path: '/dashboard', element: <h1>Dashboard</h1> },
    ],
  },
  {
    element: <ProtectedRoute allowedRoles={['admin']} />,
    children: [
      { path: '/settings', element: <div>Admin Settings</div> },
    ],
  },
]);