import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { ArrowLeft, Plus, X, Building2, Users, PenLine } from 'lucide-react';
import { merchantsApi, type MerchantResponse } from '@/features/merchants/api/merchantsApi';
import { branchesApi, type BranchResponse } from '@/features/branches/api/branchesApi';
import { staffApi, type StaffMemberResponse } from '@/features/staff/api/staffApi';
import { usersApi, type UserResponse } from '@/features/users/api/usersApi';

export default function MerchantDetailPage() {
  const { merchantId } = useParams<{ merchantId: string }>();
  const id = Number(merchantId);

  const [merchant, setMerchant] = useState<MerchantResponse | null>(null);
  const [branches, setBranches] = useState<BranchResponse[]>([]);
  const [staff, setStaff] = useState<StaffMemberResponse[]>([]);
  const [users, setUsers] = useState<UserResponse[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const [isEditing, setIsEditing] = useState(false);
  const [showBranchModal, setShowBranchModal] = useState(false);
  const [showStaffModal, setShowStaffModal] = useState(false);

  const [editForm, setEditForm] = useState({ name: '', legalName: '', email: '', phone: '', address: '', taxId: '', ownerId: '' });
  const [editErrors, setEditErrors] = useState<Record<string, string>>({});
  const [editSubmitting, setEditSubmitting] = useState(false);

  const [branchForm, setBranchForm] = useState({ name: '', code: '', address: '', phone: '' });
  const [branchErrors, setBranchErrors] = useState<Record<string, string>>({});
  const [branchSubmitting, setBranchSubmitting] = useState(false);

  const [staffForm, setStaffForm] = useState({ userId: '', branchId: '', role: 'cashier', isDefault: false });
  const [staffErrors, setStaffErrors] = useState<Record<string, string>>({});
  const [staffSubmitting, setStaffSubmitting] = useState(false);

  const fetchData = () => {
    setIsLoading(true);
    Promise.all([
      merchantsApi.get(id),
      branchesApi.list(id),
      staffApi.list(id),
      usersApi.list(),
    ])
      .then(([merchantData, branchData, staffData, userData]) => {
        setMerchant(merchantData);
        setBranches(branchData);
        setStaff(staffData);
        setUsers(userData);
      })
      .catch(() => {})
      .finally(() => setIsLoading(false));
  };

  useEffect(() => {
    fetchData();
  }, [id]);

  const userMap = Object.fromEntries(users.map((u) => [u.id, u]));
  const branchMap = Object.fromEntries(branches.map((b) => [b.id, b]));

  const ownerStaff = staff.length > 0 ? staff.reduce((a, b) => a.id < b.id ? a : b) : null;
  const ownerUser = ownerStaff ? userMap[ownerStaff.userId] : null;

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
      await branchesApi.create(id, {
        name: branchForm.name.trim(),
        code: branchForm.code.trim(),
        address: branchForm.address.trim() || undefined,
        phone: branchForm.phone.trim() || undefined,
      });
      setShowBranchModal(false);
      const branchData = await branchesApi.list(id);
      setBranches(branchData);
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
      await staffApi.assign(id, {
        userId: Number(staffForm.userId),
        branchId: Number(staffForm.branchId),
        role: staffForm.role,
        isDefault: staffForm.isDefault,
      });
      setShowStaffModal(false);
      const staffData = await staffApi.list(id);
      setStaff(staffData);
    } catch {
      setStaffErrors({ form: 'Failed to assign staff. This user may already be assigned to this branch.' });
    } finally {
      setStaffSubmitting(false);
    }
  };

  const startEditing = () => {
    if (!merchant) return;
    setEditForm({
      name: merchant.name,
      legalName: merchant.legalName || '',
      email: merchant.email,
      phone: merchant.phone || '',
      address: merchant.address || '',
      taxId: merchant.taxId || '',
      ownerId: ownerStaff ? String(ownerStaff.userId) : '',
    });
    setEditErrors({});
    setIsEditing(true);
  };

  const cancelEditing = () => {
    setIsEditing(false);
    setEditErrors({});
  };

  const validateEdit = () => {
    const e: Record<string, string> = {};
    if (!editForm.name.trim()) e.name = 'Name is required';
    if (!editForm.email.trim()) e.email = 'Email is required';
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(editForm.email)) e.email = 'Invalid email format';
    setEditErrors(e);
    return Object.keys(e).length === 0;
  };

  const handleSave = async () => {
    if (!validateEdit()) return;
    setEditSubmitting(true);
    try {
      const updated = await merchantsApi.update(id, {
        name: editForm.name.trim(),
        legalName: editForm.legalName.trim() || undefined,
        email: editForm.email.trim(),
        phone: editForm.phone.trim() || undefined,
        address: editForm.address.trim() || undefined,
        taxId: editForm.taxId.trim() || undefined,
      });
      setMerchant(updated);

      const newOwnerId = Number(editForm.ownerId);
      const oldOwnerId = ownerStaff?.userId;
      if (newOwnerId && newOwnerId !== oldOwnerId) {
        if (ownerStaff) {
          await staffApi.remove(id, ownerStaff.id);
        }
        const branchList = await branchesApi.list(id);
        const targetBranch = branchList[0];
        if (!targetBranch) {
          throw new Error('Add a branch first before assigning an owner.');
        }
        setBranches(branchList);
        await staffApi.assign(id, {
          userId: newOwnerId,
          branchId: targetBranch.id,
          role: 'manager',
          isDefault: true,
        });
      }

      setIsEditing(false);
      const staffData = await staffApi.list(id);
      setStaff(staffData);
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Unknown error';
      setEditErrors({ form: `Failed to update merchant: ${msg}` });
    } finally {
      setEditSubmitting(false);
    }
  };

  if (isLoading) return <div>Loading...</div>;
  if (!merchant) return <div>Merchant not found</div>;

  return (
    <>
      <div className="space-y-6">
        <Link to="/merchants" className="inline-flex items-center gap-1.5 text-sm text-neutral-500 hover:text-neutral-700">
          <ArrowLeft size={16} />
          Merchants
        </Link>

        <div className="rounded-lg border border-neutral-200 bg-white p-6">
          <div className="flex items-center justify-between min-h-[28px]">
            <h1 className="font-display text-xl font-bold text-neutral-900">{merchant.name}</h1>
            {isEditing ? (
              <div className="flex items-center gap-2">
                {editErrors.form && <span className="text-xs text-danger">{editErrors.form}</span>}
                <button onClick={cancelEditing} className="rounded-lg border border-neutral-200 px-3 py-1.5 text-xs text-neutral-600 hover:bg-neutral-50">Cancel</button>
                <button onClick={handleSave} disabled={editSubmitting} className="rounded-lg bg-primary px-3 py-1.5 text-xs text-white hover:bg-primary-hover disabled:opacity-50">
                  {editSubmitting ? 'Saving...' : 'Save'}
                </button>
              </div>
            ) : (
              <button onClick={startEditing} className="flex items-center gap-1.5 rounded-lg border border-neutral-200 px-3 py-1.5 text-xs text-neutral-600 hover:bg-neutral-50">
                <PenLine size={14} /> Edit
              </button>
            )}
          </div>
          <div className="mt-4 grid grid-cols-[auto_1fr_auto_1fr] gap-x-6 gap-y-3 text-sm items-center">
            {isEditing ? (
              <>
                <span className="text-neutral-500">Legal Name:</span>
                <input type="text" value={editForm.legalName} onChange={(e) => setEditForm({ ...editForm, legalName: e.target.value })}
                  className="w-full rounded border border-neutral-200 bg-white px-2 py-1 text-sm outline-none focus:border-neutral-400" />
                <span className="text-neutral-500">Name:</span>
                <input type="text" value={editForm.name} onChange={(e) => setEditForm({ ...editForm, name: e.target.value })}
                  className={`w-full rounded border bg-white px-2 py-1 text-sm outline-none ${editErrors.name ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`} />
                {editErrors.name && <span className="text-xs text-danger col-start-2">{editErrors.name}</span>}

                <span className="text-neutral-500">Email:</span>
                <input type="email" value={editForm.email} onChange={(e) => setEditForm({ ...editForm, email: e.target.value })}
                  className={`w-full rounded border bg-white px-2 py-1 text-sm outline-none ${editErrors.email ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`} />
                {editErrors.email && <span className="text-xs text-danger col-start-2">{editErrors.email}</span>}
                <span className="text-neutral-500">Phone:</span>
                <input type="text" value={editForm.phone} onChange={(e) => setEditForm({ ...editForm, phone: e.target.value })}
                  className="w-full rounded border border-neutral-200 bg-white px-2 py-1 text-sm outline-none focus:border-neutral-400" />

                <span className="text-neutral-500">Tax ID:</span>
                <input type="text" value={editForm.taxId} onChange={(e) => setEditForm({ ...editForm, taxId: e.target.value })}
                  className="w-full rounded border border-neutral-200 bg-white px-2 py-1 text-sm outline-none focus:border-neutral-400" />
                <span className="text-neutral-500">Address:</span>
                <input type="text" value={editForm.address} onChange={(e) => setEditForm({ ...editForm, address: e.target.value })}
                  className="w-full rounded border border-neutral-200 bg-white px-2 py-1 text-sm outline-none focus:border-neutral-400" />

                <span className="text-neutral-500">Owner:</span>
                <select value={editForm.ownerId} onChange={(e) => setEditForm({ ...editForm, ownerId: e.target.value })}
                  className="w-full rounded border border-neutral-200 bg-white px-2 py-1 text-sm outline-none focus:border-neutral-400">
                  <option value="">— No owner —</option>
                  {users.map((u) => <option key={u.id} value={u.id}>{u.username} ({u.email})</option>)}
                </select>
                <span className="text-neutral-500">Status:</span>
                <span className={`inline-block justify-self-start rounded-full px-2 py-0.5 text-xs font-medium ${merchant.status === 'active' ? 'bg-secondary/10 text-secondary' : 'bg-neutral-100 text-neutral-500'}`}>{merchant.status}</span>
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
              </>
            )}
          </div>
        </div>

        <div className="rounded-lg border border-neutral-200 bg-white">
          <div className="flex items-center justify-between border-b border-neutral-200 px-4 py-3">
            <div className="flex items-center gap-2">
              <Building2 size={16} className="text-neutral-500" />
              <h2 className="font-display text-sm font-semibold text-neutral-900">Branches</h2>
            </div>
            <button onClick={() => { setBranchForm({ name: '', code: '', address: '', phone: '' }); setBranchErrors({}); setShowBranchModal(true); }}
              className="flex items-center gap-1.5 rounded-lg bg-primary px-3 py-1.5 text-xs text-white hover:bg-primary-hover">
              <Plus size={14} /> Add Branch
            </button>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-neutral-100 text-left text-xs text-neutral-500">
                  <th className="px-4 py-3 font-medium">ID</th>
                  <th className="px-4 py-3 font-medium">Name</th>
                  <th className="px-4 py-3 font-medium">Code</th>
                  <th className="px-4 py-3 font-medium">Address</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium">Created At</th>
                </tr>
              </thead>
              <tbody>
                {branches.length === 0 ? (
                  <tr><td colSpan={6} className="px-4 py-8 text-center text-sm text-neutral-400">No branches yet</td></tr>
                ) : (
                  branches.map((b) => (
                    <tr key={b.id} className="border-b border-neutral-100 last:border-0 hover:bg-neutral-50">
                      <td className="px-4 py-3 font-mono text-xs text-neutral-500">{b.id}</td>
                      <td className="px-4 py-3 font-medium text-neutral-900">{b.name}</td>
                      <td className="px-4 py-3 font-mono text-xs text-neutral-500">{b.code}</td>
                      <td className="px-4 py-3 text-neutral-600">{b.address || '—'}</td>
                      <td className="px-4 py-3">
                        <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                          b.status === 'active' ? 'bg-secondary/10 text-secondary' : 'bg-neutral-100 text-neutral-500'
                        }`}>{b.status}</span>
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
          <div className="flex items-center justify-between border-b border-neutral-200 px-4 py-3">
            <div className="flex items-center gap-2">
              <Users size={16} className="text-neutral-500" />
              <h2 className="font-display text-sm font-semibold text-neutral-900">Staff</h2>
            </div>
            <button onClick={() => { const firstBranch = branches[0]; setStaffForm({ userId: '', branchId: firstBranch ? String(firstBranch.id) : '', role: 'cashier', isDefault: false }); setStaffErrors({}); setShowStaffModal(true); }}
              className="flex items-center gap-1.5 rounded-lg bg-primary px-3 py-1.5 text-xs text-white hover:bg-primary-hover">
              <Plus size={14} /> Assign Staff
            </button>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-neutral-100 text-left text-xs text-neutral-500">
                  <th className="px-4 py-3 font-medium">ID</th>
                  <th className="px-4 py-3 font-medium">User</th>
                  <th className="px-4 py-3 font-medium">Branch</th>
                  <th className="px-4 py-3 font-medium">Role</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium">Default</th>
                  <th className="px-4 py-3 font-medium">Created At</th>
                </tr>
              </thead>
              <tbody>
                {staff.length === 0 ? (
                  <tr><td colSpan={7} className="px-4 py-8 text-center text-sm text-neutral-400">No staff assigned</td></tr>
                ) : (
                  staff.map((s) => {
                    const user = userMap[s.userId];
                    const branch = branchMap[s.branchId];
                    return (
                      <tr key={s.id} className="border-b border-neutral-100 last:border-0 hover:bg-neutral-50">
                        <td className="px-4 py-3 font-mono text-xs text-neutral-500">{s.id}</td>
                        <td className="px-4 py-3 font-medium text-neutral-900">
                          {user ? `${user.username} (${user.email})` : `User #${s.userId}`}
                        </td>
                        <td className="px-4 py-3 text-neutral-600">{branch?.name || `Branch #${s.branchId}`}</td>
                        <td className="px-4 py-3">
                          <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                            s.role === 'manager' ? 'bg-info/10 text-info-hover' : 'bg-neutral-100 text-neutral-600'
                          }`}>{s.role}</span>
                        </td>
                        <td className="px-4 py-3">
                          <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                            s.status === 'active' ? 'bg-secondary/10 text-secondary' : 'bg-neutral-100 text-neutral-500'
                          }`}>{s.status}</span>
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
      </div>

      {showBranchModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <div className="flex items-center justify-between mb-5">
              <h2 className="font-display text-lg font-semibold text-neutral-900">Add Branch</h2>
              <button onClick={() => setShowBranchModal(false)} className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600">
                <X size={18} />
              </button>
            </div>
            {branchErrors.form && <div className="mb-4 rounded-lg bg-danger/10 px-3 py-2 text-sm text-danger">{branchErrors.form}</div>}
            <div className="space-y-4">
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Name *</label>
                <input type="text" value={branchForm.name}
                  onChange={(e) => setBranchForm({ ...branchForm, name: e.target.value })}
                  className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${branchErrors.name ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`} />
                {branchErrors.name && <p className="mt-1 text-xs text-danger">{branchErrors.name}</p>}
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Code *</label>
                <input type="text" value={branchForm.code}
                  onChange={(e) => setBranchForm({ ...branchForm, code: e.target.value })}
                  className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${branchErrors.code ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`} />
                {branchErrors.code && <p className="mt-1 text-xs text-danger">{branchErrors.code}</p>}
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Address</label>
                <input type="text" value={branchForm.address}
                  onChange={(e) => setBranchForm({ ...branchForm, address: e.target.value })}
                  className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400" />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Phone</label>
                <input type="text" value={branchForm.phone}
                  onChange={(e) => setBranchForm({ ...branchForm, phone: e.target.value })}
                  className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400" />
              </div>
              <div className="flex justify-end gap-3 pt-2">
                <button onClick={() => setShowBranchModal(false)}
                  className="rounded-lg border border-neutral-200 px-4 py-2 text-sm text-neutral-600 hover:bg-neutral-50">Cancel</button>
                <button onClick={handleCreateBranch} disabled={branchSubmitting}
                  className="rounded-lg bg-primary px-4 py-2 text-sm text-white hover:bg-primary-hover disabled:opacity-50">
                  {branchSubmitting ? 'Creating...' : 'Create Branch'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {showStaffModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <div className="flex items-center justify-between mb-5">
              <h2 className="font-display text-lg font-semibold text-neutral-900">Assign Staff</h2>
              <button onClick={() => setShowStaffModal(false)} className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600">
                <X size={18} />
              </button>
            </div>
            {staffErrors.form && <div className="mb-4 rounded-lg bg-danger/10 px-3 py-2 text-sm text-danger">{staffErrors.form}</div>}
            <div className="space-y-4">
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">User *</label>
                <select value={staffForm.userId}
                  onChange={(e) => setStaffForm({ ...staffForm, userId: e.target.value })}
                  className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${staffErrors.userId ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`}>
                  <option value="">— Select user —</option>
                  {users.map((u) => <option key={u.id} value={u.id}>{u.username} ({u.email})</option>)}
                </select>
                {staffErrors.userId && <p className="mt-1 text-xs text-danger">{staffErrors.userId}</p>}
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Branch *</label>
                <select value={staffForm.branchId}
                  onChange={(e) => setStaffForm({ ...staffForm, branchId: e.target.value })}
                  className={`w-full rounded-lg border bg-white px-3 py-2 text-sm outline-none ${staffErrors.branchId ? 'border-danger' : 'border-neutral-200 focus:border-neutral-400'}`}>
                  <option value="">— Select branch —</option>
                  {branches.map((b) => <option key={b.id} value={b.id}>{b.name} ({b.code})</option>)}
                </select>
                {staffErrors.branchId && <p className="mt-1 text-xs text-danger">{staffErrors.branchId}</p>}
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Role *</label>
                <select value={staffForm.role}
                  onChange={(e) => setStaffForm({ ...staffForm, role: e.target.value })}
                  className="w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none focus:border-neutral-400">
                  <option value="manager">Manager</option>
                  <option value="supervisor">Supervisor</option>
                  <option value="cashier">Cashier</option>
                  <option value="viewer">Viewer</option>
                </select>
              </div>
              <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" checked={staffForm.isDefault}
                  onChange={(e) => setStaffForm({ ...staffForm, isDefault: e.target.checked })}
                  className="rounded border-neutral-300 text-secondary focus:ring-secondary" />
                <span className="text-sm text-neutral-600">Set as default branch</span>
              </label>
              <div className="flex justify-end gap-3 pt-2">
                <button onClick={() => setShowStaffModal(false)}
                  className="rounded-lg border border-neutral-200 px-4 py-2 text-sm text-neutral-600 hover:bg-neutral-50">Cancel</button>
                <button onClick={handleAssignStaff} disabled={staffSubmitting}
                  className="rounded-lg bg-primary px-4 py-2 text-sm text-white hover:bg-primary-hover disabled:opacity-50">
                  {staffSubmitting ? 'Assigning...' : 'Assign Staff'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
