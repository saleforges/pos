import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Save, RotateCcw, Store } from 'lucide-react';
import { accountsApi, type UserResponse } from '../api/accountsApi';
import { merchantsApi, type Merchant } from '@/features/merchants/api/merchantsApi';
import { staffApi, type StaffMemberResponse } from '@/features/merchants/api/staffApi';
import { PageLoader } from '@/components/ui/PageLoader';
import { Button } from '@/components/ui/Button';

export default function AccountDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [user, setUser] = useState<UserResponse | null>(null);
  const [edit, setEdit] = useState({ username: '', email: '', status: '' });
  const [assignments, setAssignments] = useState<{ merchant: Merchant; staff: StaffMemberResponse }[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    if (!id) return;
    setIsLoading(true);
    const userId = Number(id);

    accountsApi.get(userId)
      .then(async (u) => {
        setUser(u);
        setEdit({ username: u.username, email: u.email, status: u.status });

        const merchants = await merchantsApi.list();
        const results = await Promise.allSettled(
          merchants.map((m) =>
            staffApi.list(m.id).then((staff) => ({ merchant: m, staff: staff.filter((s) => s.userId === userId) })),
          ),
        );
        const found: { merchant: Merchant; staff: StaffMemberResponse }[] = [];
        for (const r of results) {
          if (r.status === 'fulfilled' && r.value.staff.length > 0) {
            for (const s of r.value.staff) {
              found.push({ merchant: r.value.merchant, staff: s });
            }
          }
        }
        setAssignments(found);
      })
      .catch(() => setError('User not found'))
      .finally(() => setIsLoading(false));
  }, [id]);

  const hasChanges = user && (
    edit.username !== user.username ||
    edit.email !== user.email ||
    edit.status !== user.status
  );

  const handleSave = async () => {
    if (!user) return;
    setSaving(true);
    setError('');
    setSuccess('');
    try {
      const updated = await accountsApi.update(user.id, {
        username: edit.username !== user.username ? edit.username : undefined,
        email: edit.email !== user.email ? edit.email : undefined,
        status: edit.status !== user.status ? edit.status : undefined,
      });
      setUser(updated);
      setEdit({ username: updated.username, email: updated.email, status: updated.status });
      setSuccess('Account updated successfully');
    } catch {
      setError('Failed to update account');
    } finally {
      setSaving(false);
    }
  };

  const handleReset = () => {
    if (!user) return;
    setEdit({ username: user.username, email: user.email, status: user.status });
    setError('');
    setSuccess('');
  };

  if (isLoading) return <PageLoader />;
  if (!user) return (
    <div className="flex h-64 flex-col items-center justify-center gap-3 text-neutral-500">
      <p>Account not found</p>
      <Button variant="secondary" onClick={() => navigate('/accounts')}>Back to Accounts</Button>
    </div>
  );

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div className="flex items-center gap-3">
        <button onClick={() => navigate('/accounts')} className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600">
          <ArrowLeft size={20} />
        </button>
        <h1 className="font-display text-lg font-semibold text-neutral-900">Account Details</h1>
      </div>

      {error && (
        <div className="rounded-lg bg-danger/10 px-3 py-2 text-sm text-danger">{error}</div>
      )}
      {success && (
        <div className="rounded-lg bg-secondary/10 px-3 py-2 text-sm text-secondary">{success}</div>
      )}

      <div className="rounded-lg border border-neutral-200 bg-white">
        <div className="border-b border-neutral-100 px-5 py-3">
          <span className="text-sm font-medium text-neutral-500">Account Information</span>
        </div>
        <div className="divide-y divide-neutral-100 px-5 py-2">
          <div className="flex items-center justify-between py-3">
            <span className="text-sm text-neutral-500">ID</span>
            <span className="text-sm font-medium text-neutral-900">{user.id}</span>
          </div>
          <div className="flex items-center justify-between py-3">
            <span className="text-sm text-neutral-500">Username</span>
            <input
              type="text"
              value={edit.username}
              onChange={(e) => setEdit({ ...edit, username: e.target.value })}
              className="w-64 rounded-lg border border-neutral-200 bg-white px-3 py-1.5 text-sm text-neutral-900 outline-none focus:border-neutral-400"
            />
          </div>
          <div className="flex items-center justify-between py-3">
            <span className="text-sm text-neutral-500">Email</span>
            <input
              type="email"
              value={edit.email}
              onChange={(e) => setEdit({ ...edit, email: e.target.value })}
              className="w-64 rounded-lg border border-neutral-200 bg-white px-3 py-1.5 text-sm text-neutral-900 outline-none focus:border-neutral-400"
            />
          </div>
          <div className="flex items-center justify-between py-3">
            <span className="text-sm text-neutral-500">Role</span>
            <span className="inline-block rounded-full bg-secondary/10 px-2.5 py-0.5 text-xs font-medium text-secondary">{user.role}</span>
          </div>
          <div className="flex items-center justify-between py-3">
            <span className="text-sm text-neutral-500">Type</span>
            <span className="inline-block rounded-full bg-primary/10 px-2.5 py-0.5 text-xs font-medium text-primary">{user.type}</span>
          </div>
          <div className="flex items-center justify-between py-3">
            <span className="text-sm text-neutral-500">Status</span>
            <select
              value={edit.status}
              onChange={(e) => setEdit({ ...edit, status: e.target.value })}
              className="w-40 rounded-lg border border-neutral-200 bg-white px-3 py-1.5 text-sm text-neutral-900 outline-none focus:border-neutral-400"
            >
              <option value="active">Active</option>
              <option value="disabled">Disabled</option>
            </select>
          </div>
          <div className="flex items-center justify-between py-3">
            <span className="text-sm text-neutral-500">Created</span>
            <span className="text-sm text-neutral-600">{new Date(user.createdAt).toLocaleString()}</span>
          </div>
          <div className="flex items-center justify-between py-3">
            <span className="text-sm text-neutral-500">Updated</span>
            <span className="text-sm text-neutral-600">{new Date(user.updatedAt).toLocaleString()}</span>
          </div>
        </div>
      </div>

      {assignments.length > 0 && (
        <div className="rounded-lg border border-neutral-200 bg-white">
          <div className="border-b border-neutral-100 px-5 py-3">
            <span className="text-sm font-medium text-neutral-500">Merchant Assignments</span>
          </div>
          <div className="divide-y divide-neutral-100">
            {assignments.map((a) => (
              <div key={`${a.merchant.id}-${a.staff.branchId}`} className="flex items-center justify-between px-5 py-3">
                <div className="flex items-center gap-3">
                  <Store size={16} className="text-neutral-400" />
                  <div>
                    <span className="text-sm font-medium text-neutral-900">{a.merchant.name}</span>
                    <span className="ml-2 text-xs text-neutral-400">Branch #{a.staff.branchId}</span>
                  </div>
                </div>
                <span className="inline-block rounded-full bg-secondary/10 px-2.5 py-0.5 text-xs font-medium text-secondary">
                  {a.staff.role}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="flex justify-end gap-2">
        {hasChanges && (
          <Button variant="ghost" onClick={handleReset}>
            <RotateCcw size={14} /> Reset
          </Button>
        )}
        <Button variant="secondary" onClick={() => navigate('/accounts')}>Back to List</Button>
        <Button onClick={handleSave} disabled={!hasChanges || saving}>
          <Save size={14} /> {saving ? 'Saving…' : 'Save Changes'}
        </Button>
      </div>
    </div>
  );
}
