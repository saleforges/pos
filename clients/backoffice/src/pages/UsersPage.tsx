import { useEffect, useState } from 'react';
import { Search, Plus, X } from 'lucide-react';
import { usersApi, type UserResponse } from '@/features/users/api/usersApi';
import { rolesApi } from '@/features/roles/api/rolesApi';
import type { Role } from '@/features/roles/types';

export default function UsersPage() {
  const [users, setUsers] = useState<UserResponse[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [roles, setRoles] = useState<Role[]>([]);

  const [form, setForm] = useState({ username: '', email: '', password: '', selectedRoles: [] as string[] });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  const fetchUsers = () => {
    setIsLoading(true);
    usersApi.list().then(setUsers).finally(() => setIsLoading(false));
  };

  useEffect(() => {
    fetchUsers();
    rolesApi.list().then(setRoles).catch(() => {});
  }, []);

  const filtered = users.filter(
    (u) =>
      u.username.toLowerCase().includes(search.toLowerCase()) ||
      u.email.toLowerCase().includes(search.toLowerCase()),
  );

  const openModal = () => {
    setForm({ username: '', email: '', password: '', selectedRoles: [] });
    setErrors({});
    setShowModal(true);
  };

  const validate = () => {
    const e: Record<string, string> = {};
    if (!form.username.trim()) e.username = 'Username is required';
    if (!form.email.trim()) e.email = 'Email is required';
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) e.email = 'Invalid email format';
    if (!form.password) e.password = 'Password is required';
    else if (form.password.length < 8) e.password = 'Password must be at least 8 characters';
    else if (!/[a-z]/.test(form.password)) e.password = 'Password must contain a lowercase letter';
    else if (!/[A-Z]/.test(form.password)) e.password = 'Password must contain an uppercase letter';
    else if (!/[0-9]/.test(form.password)) e.password = 'Password must contain a digit';
    setErrors(e);
    return Object.keys(e).length === 0;
  };

  const handleSubmit = async () => {
    if (!validate()) return;
    setSubmitting(true);
    try {
      await usersApi.create({
        username: form.username.trim(),
        email: form.email.trim(),
        password: form.password,
        roles: form.selectedRoles.length > 0 ? form.selectedRoles : undefined,
      });
      setShowModal(false);
      fetchUsers();
    } catch {
      setErrors({ form: 'Failed to create user. It may already exist.' });
    } finally {
      setSubmitting(false);
    }
  };

  const toggleRole = (name: string) => {
    setForm((prev) => ({
      ...prev,
      selectedRoles: prev.selectedRoles.includes(name)
        ? prev.selectedRoles.filter((r) => r !== name)
        : [...prev.selectedRoles, name],
    }));
  };

  if (isLoading) return <div>Loading users...</div>;

  return (
    <>
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="relative flex-1 max-w-xs">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-400" />
            <input
              type="text"
              placeholder="Search users..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full rounded-lg border border-neutral-200 bg-white py-2 pl-9 pr-3 text-sm outline-none focus:border-neutral-400"
            />
          </div>
          <button
            onClick={openModal}
            className="flex items-center gap-2 rounded-lg bg-primary px-3 py-2 text-sm text-white hover:bg-primary-hover"
          >
            <Plus size={16} />
            Add User
          </button>
        </div>

        <div className="rounded-lg border border-neutral-200 bg-white">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-neutral-100 text-left text-xs text-neutral-500">
                  <th className="px-4 py-3 font-medium">ID</th>
                  <th className="px-4 py-3 font-medium">Username</th>
                  <th className="px-4 py-3 font-medium">Email</th>
                  <th className="px-4 py-3 font-medium">Role</th>
                  <th className="px-4 py-3 font-medium">Type</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium">Created At</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((user) => (
                  <tr key={user.id} className="border-b border-neutral-100 last:border-0 hover:bg-neutral-50">
                    <td className="px-4 py-3 font-mono text-xs text-neutral-500">{user.id}</td>
                    <td className="px-4 py-3 font-medium text-neutral-900">{user.username}</td>
                    <td className="px-4 py-3 text-neutral-600">{user.email}</td>
                    <td className="px-4 py-3">
                      {user.role ? (
                        <span className="inline-block rounded-full bg-neutral-100 px-2 py-0.5 text-xs font-medium text-neutral-600">
                          {user.role}
                        </span>
                      ) : (
                        <span className="text-xs text-neutral-400">—</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                        user.type === 'platform' ? 'bg-info/10 text-info-hover' : 'bg-neutral-100 text-neutral-600'
                      }`}>
                        {user.type}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                        user.status === 'active' ? 'bg-secondary/10 text-secondary' : 'bg-neutral-100 text-neutral-500'
                      }`}>
                        {user.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-xs text-neutral-400">{new Date(user.createdAt).toLocaleDateString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <div className="flex items-center justify-between mb-5">
              <h2 className="font-display text-lg font-semibold text-neutral-900">Create User</h2>
              <button onClick={() => setShowModal(false)} className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600">
                <X size={18} />
              </button>
            </div>

            {errors.form && (
              <div className="mb-4 rounded-lg bg-danger/10 px-3 py-2 text-sm text-danger">{errors.form}</div>
            )}

            <div className="space-y-4">
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Username</label>
                <input
                  type="text"
                  value={form.username}
                  onChange={(e) => setForm({ ...form, username: e.target.value })}
                  className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${
                    errors.username ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'
                  }`}
                />
                {errors.username && <p className="mt-1 text-xs text-danger">{errors.username}</p>}
              </div>

              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Email</label>
                <input
                  type="email"
                  value={form.email}
                  onChange={(e) => setForm({ ...form, email: e.target.value })}
                  className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${
                    errors.email ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'
                  }`}
                />
                {errors.email && <p className="mt-1 text-xs text-danger">{errors.email}</p>}
              </div>

              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Password</label>
                <input
                  type="password"
                  value={form.password}
                  onChange={(e) => setForm({ ...form, password: e.target.value })}
                  className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${
                    errors.password ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'
                  }`}
                />
                {errors.password && <p className="mt-1 text-xs text-danger">{errors.password}</p>}
                <p className="mt-1 text-xs text-neutral-400">Min 8 characters, with uppercase, lowercase, and digit.</p>
              </div>

              {roles.length > 0 && (
                <div>
                  <label className="mb-1 block text-xs font-medium text-neutral-600">Roles</label>
                  <div className="max-h-40 space-y-1 overflow-y-auto rounded-lg border border-neutral-200 p-2">
                    {roles.map((role) => (
                      <label key={role.id} className="flex items-center gap-2 rounded px-2 py-1 text-sm hover:bg-neutral-50 cursor-pointer">
                        <input
                          type="checkbox"
                          checked={form.selectedRoles.includes(role.name)}
                          onChange={() => toggleRole(role.name)}
                          className="rounded border-neutral-300 text-secondary focus:ring-secondary"
                        />
                        {role.name}
                      </label>
                    ))}
                  </div>
                </div>
              )}

              <div className="flex justify-end gap-3 pt-2">
                <button
                  onClick={() => setShowModal(false)}
                  className="rounded-lg border border-neutral-200 px-4 py-2 text-sm text-neutral-600 hover:bg-neutral-50"
                >
                  Cancel
                </button>
                <button
                  onClick={handleSubmit}
                  disabled={submitting}
                  className="rounded-lg bg-primary px-4 py-2 text-sm text-white hover:bg-primary-hover disabled:opacity-50"
                >
                  {submitting ? 'Creating...' : 'Create User'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
