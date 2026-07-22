// App.tsx
import { RouterProvider } from 'react-router-dom';
import { AuthProvider } from '@/features/auth/context/AuthContext';
import { router } from '@/router';
// comment for trigger deploy
function App() {
  return (
    <AuthProvider>
      <RouterProvider router={router} />
    </AuthProvider>
  );
}

export default App;