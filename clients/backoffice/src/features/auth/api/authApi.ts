// features/auth/api/authApi.ts
const MOCK_DELAY = 500;
const MOCK_USER = { id: '1', name: 'Ilham', role: 'admin' };

export const authApi = {
  login: async (email: string, password: string) => {
    await new Promise((r) => setTimeout(r, MOCK_DELAY));
    if (email === 'admin@test.com' && password === 'password') {
      localStorage.setItem('mock_session', 'true');
      return { user: MOCK_USER };
    }
    throw new Error('Invalid credentials');
  },

  logout: async () => {
    await new Promise((r) => setTimeout(r, 200));
    localStorage.removeItem('mock_session');
  },

  me: async () => {
    await new Promise((r) => setTimeout(r, 300));
    const hasSession = localStorage.getItem('mock_session') === 'true';
    if (!hasSession) throw new Error('Not authenticated');
    return MOCK_USER;
  },
};