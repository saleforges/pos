import { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { ArrowLeft, Plus, X, Building2, Users, PenLine, Save, RotateCcw } from 'lucide-react';
import { merchantsApi, type Merchant } from '../api/merchantsApi';
import { branchesApi, type BranchResponse } from '../api/branchesApi';
import { staffApi, type StaffMemberResponse } from '../api/staffApi';
import { accountsApi, type UserResponse } from '@/features/accounts/api/accountsApi';
import { PageLoader } from '@/components/ui/PageLoader';
import { Button } from '@/components/ui/Button';

export default function MerchantDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const merchantId = Number(id);

  const [merchant, setMerchant] = useState<Merchant | null>(null);
  const [branches, setBranches] = useState<BranchResponse[]>([]);
  const [staff, setStaff] = useState<StaffMemberResponse[]>([]);
  const [users, setUsers] = useState<UserResponse[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');

  const [isEditing, setIsEditing] = useState(false);
  const [edit, setEdit] = useState({ name: '', legalName: '', email: '', phone: '', address: '', taxId: '' });
  const [editError, setEditError] = useState('');
  const [saving, setSaving] = useState(false);

  const [showBranchModal, setShowBranchModal] = useState(false);
  const [branchForm, setBranchForm] = useState({ name: '', code: '', address: '', phone: '' });
  const [branchErrors, setBranchErrors] = useState<Record<string, string>>({});
  const [branchSubmitting, setBranchSubmitting] = useState(false);

  const [showStaffModal, setShowStaffModal] = useState(false);
  const [staffForm, setStaffForm] = useState({ userId: '', branchId: '', role: 'cashier', isDefault: false });
  const [staffErrors, setStaffErrors] = useState<Record<string, string>>({});
  const [staffSubmitting, setStaffSubmitting] = useState(false);

  const fetchData = () => {
    setIsLoading(true);
    Promise.all([
      merchantsApi.get(merchantId),
      branchesApi.list(merchantId),
      staffApi.list(merchantId),
      accountsApi.list(),
    ])
      .then(([m, b, s, u]) => {
        setMerchant(m);
        setBranches(b);
        setStaff(s);
        setUsers(u);
      })
      .catch(() => setError('Failed to load merchant'))
      .finally(() => setIsLoading(false));
  };

  useEffect(() => { fetchData() }, [merchantId]);

  const userMap = Object.fromEntries(users.map((u) => [u.id, u]));
  const branchMap = Object.fromEntries(branches.map((b) => [b.id, b]));
  const ownerStaff = staff.find((s) => s.role === 'owner') ?? staff[0];
  const ownerUser = ownerStaff ? userMap[ownerStaff.userId] : null;

  const startEditing = () => {
    if (!merchant) return;
    setEdit({ name: merchant.name, legalName: merchant.legalName, email: merchant.email, phone: merchant.phone, address: merchant.address, taxId: merchant.taxId ?? '' });
    setEditError('');
    setIsEditing(true);
  };

  const cancelEditing = () => {
    setIsEditing(false);
    setEditError('');
  };

  const validateEdit = () => {
    if (!edit.name.trim()) { setEditError('Name is required'); return false; }
    if (!edit.email.trim()) { setEditError('Email is required'); return false; }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(edit.email)) { setEditError('Invalid email format'); return false; }
    return true;
  };

  const handleSave = async () => {
    if (!merchant || !validateEdit()) return;
    setSaving(true);
    setEditError('');
    try {
      const updated = await merchantsApi.update(merchant.id, {
        name: edit.name.trim(),
        legalName: edit.legalName.trim() || undefined,
        email: edit.email.trim(),
        phone: edit.phone.trim() || undefined,
        address: edit.address.trim() || undefined,
        taxId: edit.taxId.trim() || undefined,
      });
      setMerchant(updated);
      setIsEditing(false);
    } catch {
      setEditError('Failed to update merchant');
    } finally {
      setSaving(false);
    }
  };

  const validateBranch = () => {
    const e: Record<string, string> = {};
    if (!branchForm.name.trim()) e.name = 'Name is required';
    if (!branchForm.code.trim()) e.code = 'Code is required';
    setBranchErrors(e);
    return Object.keys(e).length === 0;
  };

  const handleCreateBranch = async () => {
    if (!validateBranch()) return;
    setBranchSubmitting(true);
    try {
      await branchesApi.create(merchantId, {
        name: branchForm.name.trim(),
        code: branchForm.code.trim(),
        address: branchForm.address.trim() || undefined,
        phone: branchForm.phone.trim() || undefined,
      });
      setShowBranchModal(false);
      setBranches(await branchesApi.list(merchantId));
    } catch {
      setBranchErrors({ form: 'Failed to create branch. Code may already exist.' });
    } finally {
      setBranchSubmitting(false);
    }
  };

  const validateStaff = () => {
    const e: Record<string, string> = {};
    if (!staffForm.userId) e.userId = 'User is required';
    if (!staffForm.branchId) e.branchId = 'Branch is required';
    if (!staffForm.role) e.role = 'Role is required';
    setStaffErrors(e);
    return Object.keys(e).length === 0;
  };

  const handleAssignStaff = async () => {
    if (!validateStaff()) return;
    setStaffSubmitting(true);
    try {
      await staffApi.assign(merchantId, {
        userId: Number(staffForm.userId),
        branchId: Number(staffForm.branchId),
        role: staffForm.role,
        isDefault: staffForm.isDefault,
      });
      setShowStaffModal(false);
      setStaff(await staffApi.list(merchantId));
    } catch {
      setStaffErrors({ form: 'Failed to assign staff. This user may already be assigned to this branch.' });
    } finally {
      setStaffSubmitting(false);
    }
  };

  if (isLoading) return <PageLoader />;
  if (!merchant) return (
    <div className="flex h-64 flex-col items-center justify-center gap-3 text-neutral-500">
      <p>Merchant not found</p>
      <Button variant="secondary" onClick={() => navigate('/merchants')}>Back to Merchants</Button>
    </div>
  );

  return (
    <div className="mx-auto max-w-4xl space-y-6">
      <Link to="/merchants" className="inline-flex items-center gap-1.5 text-sm text-neutral-500 hover:text-neutral-700">
        <ArrowLeft size={16} />
        Merchants
      </Link>

      {error && <div className="rounded-lg bg-danger/10 px-3 py-2 text-sm text-danger">{error}</div>}

      <div className="rounded-lg border border-neutral-200 bg-white">
        <div className="flex items-center justify-between border-b border-neutral-100 px-5 py-3">
          <h1 className="font-display text-lg font-semibold text-neutral-900">{merchant.name}</h1>
          {isEditing ? (
            <div className="flex items-center gap-2">
              <Button variant="ghost" onClick={cancelEditing}><RotateCcw size={14} /> Cancel</Button>
              <Button onClick={handleSave} disabled={saving}><Save size={14} /> {saving ? 'Saving…' : 'Save'}</Button>
            </div>
          ) : (
            <Button variant="secondary" onClick={startEditing}><PenLine size={14} /> Edit</Button>
          )}
        </div>
        {editError && <div className="border-b border-neutral-100 px-5 py-2 text-sm text-danger">{editError}</div>}
        <div className="grid grid-cols-[auto_1fr_auto_1fr] gap-x-6 gap-y-3 px-5 py-4 text-sm">
          {isEditing ? (
            <>
              <span className="text-neutral-500">Legal Name:</span>
              <input type="text" value={edit.legalName} onChange={(e) => setEdit({ ...edit, legalName: e.target.value })} className="w-full rounded border border-neutral-200 bg-white px-2 py-1 text-sm outline-none focus:border-neutral-400" />
              <span className="text-neutral-500">Name:</span>
              <input type="text" value={edit.name} onChange={(e) => setEdit({ ...edit, name: e.target.value })} className="w-full rounded border border-neutral-200 bg-white px-2 py-1 text-sm outline-none focus:border-neutral-400" />

              <span className="text-neutral-500">Email:</span>
              <input type="email" value={edit.email} onChange={(e) => setEdit({ ...edit, email: e.target.value })} className="w-full rounded border border-neutral-200 bg-white px-2 py-1 text-sm outline-none focus:border-neutral-400" />
              <span className="text-neutral-500">Phone:</span>
              <input type="text" value={edit.phone} onChange={(e) => setEdit({ ...edit, phone: e.target.value })} className="w-full rounded border border-neutral-200 bg-white px-2 py-1 text-sm outline-none focus:border-neutral-400" />

              <span className="text-neutral-500">Tax ID:</span>
              <input type="text" value={edit.taxId} onChange={(e) => setEdit({ ...edit, taxId: e.target.value })} className="w-full rounded border border-neutral-200 bg-white px-2 py-1 text-sm outline-none focus:border-neutral-400" />
              <span className="text-neutral-500">Address:</span>
              <input type="text" value={edit.address} onChange={(e) => setEdit({ ...edit, address: e.target.value })} className="w-full rounded border border-neutral-200 bg-white px-2 py-1 text-sm outline-none focus:border-neutral-400" />
            </>
          ) : (
            <>
              <span className="text-neutral-500">Legal Name:</span>
              <span className="text-neutral-900">{merchant.legalName || '—'}</span>
              <span className="text-neutral-500">Email:</span>
              <span className="text-neutral-900">{merchant.email}</span>

              <span className="text-neutral-500">Phone:</span>
              <span className="text-neutral-900">{merchant.phone || '—'}</span>
              <span className="text-neutral-500">Tax ID:</span>
              <span className="text-neutral-900">{merchant.taxId || '—'}</span>

              <span className="text-neutral-500">Address:</span>
              <span className="text-neutral-900">{merchant.address || '—'}</span>
              <span className="text-neutral-500">Owner:</span>
              <span className="text-neutral-900">{ownerUser ? `${ownerUser.username} (${ownerUser.email})` : '—'}</span>

              <span className="text-neutral-500">Status:</span>
              <span className={`inline-block justify-self-start rounded-full px-2 py-0.5 text-xs font-medium ${merchant.status === 'active' ? 'bg-secondary/10 text-secondary' : 'bg-neutral-100 text-neutral-500'}`}>{merchant.status}</span>
              <span className="text-neutral-500">Currency:</span>
              <span className="text-neutral-900">{merchant.settings.currency}</span>

              <span className="text-neutral-500">Timezone:</span>
              <span className="text-neutral-900">{merchant.settings.timezone}</span>
              <span className="text-neutral-500">Tax Rate:</span>
              <span className="text-neutral-900">{merchant.settings.taxRate}%</span>
            </>
          )}
        </div>
      </div>

      <div className="rounded-lg border border-neutral-200 bg-white">
        <div className="flex items-center justify-between border-b border-neutral-100 px-4 py-3">
          <div className="flex items-center gap-2">
            <Building2 size={16} className="text-neutral-500" />
            <h2 className="text-sm font-semibold text-neutral-900">Branches</h2>
          </div>
          <Button onClick={() => { setBranchForm({ name: '', code: '', address: '', phone: '' }); setBranchErrors({}); setShowBranchModal(true); }}>
            <Plus size={14} /> Add Branch
          </Button>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-neutral-100 text-left text-xs text-neutral-500">
                <th className="px-4 py-3 font-medium">Name</th>
                <th className="px-4 py-3 font-medium">Code</th>
                <th className="px-4 py-3 font-medium">Address</th>
                <th className="px-4 py-3 font-medium">Status</th>
                <th className="px-4 py-3 font-medium">Created</th>
              </tr>
            </thead>
            <tbody>
              {branches.length === 0 ? (
                <tr><td colSpan={5} className="px-4 py-8 text-center text-sm text-neutral-400">No branches yet</td></tr>
              ) : (
                branches.map((b) => (
                  <tr key={b.id} className="border-b border-neutral-100 last:border-0 hover:bg-neutral-50">
                    <td className="px-4 py-3 font-medium text-neutral-900">{b.name}</td>
                    <td className="px-4 py-3 font-mono text-xs text-neutral-500">{b.code}</td>
                    <td className="px-4 py-3 text-neutral-600">{b.address || '—'}</td>
                    <td className="px-4 py-3">
                      <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${b.status === 'active' ? 'bg-secondary/10 text-secondary' : 'bg-neutral-100 text-neutral-500'}`}>{b.status}</span>
                    </td>
                    <td className="px-4 py-3 text-xs text-neutral-400">{new Date(b.createdAt).toLocaleDateString()}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      <div className="rounded-lg border border-neutral-200 bg-white">
        <div className="flex items-center justify-between border-b border-neutral-100 px-4 py-3">
          <div className="flex items-center gap-2">
            <Users size={16} className="text-neutral-500" />
            <h2 className="text-sm font-semibold text-neutral-900">Staff</h2>
          </div>
          <Button onClick={() => { const first = branches[0]; setStaffForm({ userId: '', branchId: first ? String(first.id) : '', role: 'cashier', isDefault: false }); setStaffErrors({}); setShowStaffModal(true); }}>
            <Plus size={14} /> Assign Staff
          </Button>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-neutral-100 text-left text-xs text-neutral-500">
                <th className="px-4 py-3 font-medium">User</th>
                <th className="px-4 py-3 font-medium">Branch</th>
                <th className="px-4 py-3 font-medium">Role</th>
                <th className="px-4 py-3 font-medium">Status</th>
                <th className="px-4 py-3 font-medium">Default</th>
                <th className="px-4 py-3 font-medium">Created</th>
              </tr>
            </thead>
            <tbody>
              {staff.length === 0 ? (
                <tr><td colSpan={6} className="px-4 py-8 text-center text-sm text-neutral-400">No staff assigned</td></tr>
              ) : (
                staff.map((s) => {
                  const user = userMap[s.userId];
                  const branch = branchMap[s.branchId];
                  return (
                    <tr key={s.id} className="border-b border-neutral-100 last:border-0 hover:bg-neutral-50">
                      <td className="px-4 py-3 font-medium text-neutral-900">
                        {user ? `${user.username} (${user.email})` : `User #${s.userId}`}
                      </td>
                      <td className="px-4 py-3 text-neutral-600">{branch?.name || `Branch #${s.branchId}`}</td>
                      <td className="px-4 py-3">
                        <span className="inline-block rounded-full bg-secondary/10 px-2 py-0.5 text-xs font-medium text-secondary">{s.role}</span>
                      </td>
                      <td className="px-4 py-3">
                        <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${s.status === 'active' ? 'bg-secondary/10 text-secondary' : 'bg-neutral-100 text-neutral-500'}`}>{s.status}</span>
                      </td>
                      <td className="px-4 py-3 text-neutral-500">{s.isDefault ? 'Yes' : 'No'}</td>
                      <td className="px-4 py-3 text-xs text-neutral-400">{new Date(s.createdAt).toLocaleDateString()}</td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>

      {showBranchModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <div className="mb-5 flex items-center justify-between">
              <h2 className="font-display text-lg font-semibold text-neutral-900">Add Branch</h2>
              <button onClick={() => setShowBranchModal(false)} className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600"><X size={18} /></button>
            </div>
            {branchErrors.form && <div className="mb-4 rounded-lg bg-danger/10 px-3 py-2 text-sm text-danger">{branchErrors.form}</div>}
            <div className="space-y-4">
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Name *</label>
                <input type="text" value={branchForm.name} onChange={(e) => setBranchForm({ ...branchForm, name: e.target.value })} className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${branchErrors.name ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`} />
                {branchErrors.name && <p className="mt-1 text-xs text-danger">{branchErrors.name}</p>}
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Code *</label>
                <input type="text" value={branchForm.code} onChange={(e) => setBranchForm({ ...branchForm, code: e.target.value })} className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${branchErrors.code ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`} />
                {branchErrors.code && <p className="mt-1 text-xs text-danger">{branchErrors.code}</p>}
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Address</label>
                <input type="text" value={branchForm.address} onChange={(e) => setBranchForm({ ...branchForm, address: e.target.value })} className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400" />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Phone</label>
                <input type="text" value={branchForm.phone} onChange={(e) => setBranchForm({ ...branchForm, phone: e.target.value })} className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400" />
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <Button variant="secondary" onClick={() => setShowBranchModal(false)}>Cancel</Button>
                <Button onClick={handleCreateBranch} disabled={branchSubmitting}>{branchSubmitting ? 'Creating…' : 'Create Branch'}</Button>
              </div>
            </div>
          </div>
        </div>
      )}

      {showStaffModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <div className="mb-5 flex items-center justify-between">
              <h2 className="font-display text-lg font-semibold text-neutral-900">Assign Staff</h2>
              <button onClick={() => setShowStaffModal(false)} className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600"><X size={18} /></button>
            </div>
            {staffErrors.form && <div className="mb-4 rounded-lg bg-danger/10 px-3 py-2 text-sm text-danger">{staffErrors.form}</div>}
            <div className="space-y-4">
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">User *</label>
                <select value={staffForm.userId} onChange={(e) => setStaffForm({ ...staffForm, userId: e.target.value })} className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${staffErrors.userId ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`}>
                  <option value="">— Select user —</option>
                  {users.map((u) => <option key={u.id} value={u.id}>{u.username} ({u.email})</option>)}
                </select>
                {staffErrors.userId && <p className="mt-1 text-xs text-danger">{staffErrors.userId}</p>}
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Branch *</label>
                <select value={staffForm.branchId} onChange={(e) => setStaffForm({ ...staffForm, branchId: e.target.value })} className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${staffErrors.branchId ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`}>
                  <option value="">— Select branch —</option>
                  {branches.map((b) => <option key={b.id} value={b.id}>{b.name} ({b.code})</option>)}
                </select>
                {staffErrors.branchId && <p className="mt-1 text-xs text-danger">{staffErrors.branchId}</p>}
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Role *</label>
                <select value={staffForm.role} onChange={(e) => setStaffForm({ ...staffForm, role: e.target.value })} className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400">
                  <option value="manager">Manager</option>
                  <option value="supervisor">Supervisor</option>
                  <option value="cashier">Cashier</option>
                  <option value="viewer">Viewer</option>
                </select>
              </div>
              <label className="flex cursor-pointer items-center gap-2">
                <input type="checkbox" checked={staffForm.isDefault} onChange={(e) => setStaffForm({ ...staffForm, isDefault: e.target.checked })} className="rounded border-neutral-300" />
                <span className="text-sm text-neutral-600">Set as default branch</span>
              </label>
              <div className="flex justify-end gap-2 pt-2">
                <Button variant="secondary" onClick={() => setShowStaffModal(false)}>Cancel</Button>
                <Button onClick={handleAssignStaff} disabled={staffSubmitting}>{staffSubmitting ? 'Assigning…' : 'Assign Staff'}</Button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
