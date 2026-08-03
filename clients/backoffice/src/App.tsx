// App.tsx
import { RouterProvider } from 'react-router-dom';
import { I18nextProvider } from 'react-i18next';
import { AuthProvider } from '@/features/auth/context/AuthContext';
import { router } from '@/router';
import i18n from '@/i18n';

function App() {
  return (
    <I18nextProvider i18n={i18n}>
      <AuthProvider>
        <RouterProvider router={router} />
      </AuthProvider>
    </I18nextProvider>
  );
}

export default App;