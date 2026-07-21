# Plan: `/me` Response Redesign

Unified `roles` table dengan `merchant_id`/`branch_id` scope di `user_roles`. Response `/me` pake flat `roles[]`, tiap entry bawa scope merchant & branch.

## Core Concept

`user_roles` jadi satu-satunya tabel assignment — gak perlu `StaffInfo`/`StaffAssignment` terpisah.

| `user_roles` merchant_id | `user_roles` branch_id | Arti |
|--------------------------|------------------------|------|
| `NULL` | `NULL` | System role (superadmin, admin) |
| `1` | `NULL` | Merchant-wide role (owner) |
| `1` | `1` | Branch-specific role (manager, cashier) |

## Roles

| `roles.name` | Type | Scope |
|-------------|------|-------|
| `superadmin` | system | bypass semua |
| `admin` | system | platform ops, custom permissions |
| `owner` | merchant-wide | full akses ke merchant sendiri |
| `manager` | branch | manage branch |
| `supervisor` | branch | supervisory |
| `cashier` | branch | kasir |
| `viewer` | branch | read-only |

## Response Shape

### Superadmin
```json
{
  "id": 1,
  "email": "sa@pos.com",
  "name": "Super Admin",
  "type": "platform",
  "status": "active",
  "roles": [
    { "id": 1, "name": "superadmin" }
  ]
}
```

### Owner + Staff (multi-role)
```json
{
  "id": 2,
  "email": "budi@pos.com",
  "name": "Budi",
  "type": "merchant",
  "status": "active",
  "roles": [
    { "id": 3, "name": "owner",
      "merchant": { "id": 1, "name": "Warung Makmur" },
      "is_default": true },
    { "id": 4, "name": "manager",
      "merchant": { "id": 1, "name": "Warung Makmur" },
      "branch": { "id": 2, "name": "Cabang B" } },
    { "id": 6, "name": "cashier",
      "merchant": { "id": 1, "name": "Warung Makmur" },
      "branch": { "id": 3, "name": "Cabang C" } }
  ]
}
```

### Staff (single role)
```json
{
  "id": 3,
  "email": "siti@pos.com",
  "name": "Siti",
  "type": "merchant",
  "status": "active",
  "roles": [
    { "id": 6, "name": "cashier",
      "merchant": { "id": 1, "name": "Warung Makmur" },
      "branch": { "id": 1, "name": "Cabang A" },
      "is_default": true }
  ]
}
```

## 5 Steps Implementasi

### Step 1: Database — Migration + Seed

**Files:**
- `internal/iam/adapter/repository/postgres/migrations/000001_init.up.sql`
- `internal/iam/adapter/repository/postgres/seed.go`

**Changes:**
- Tambah `merchant_id BIGINT`, `branch_id BIGINT` di `user_roles`
- Hapus PRIMARY KEY `(user_id, role_id)` → UNIQUE `(user_id, role_id, COALESCE(merchant_id,0), COALESCE(branch_id,0))`
- Drop DB, run ulang migrasi
- Seed ulang: roles jadi 7 (superadmin, admin, owner, manager, supervisor, cashier, viewer)
- Seed user: superadmin + owner (dengan scope merchant_id=1 via `user_roles`)
- Merchant `staff` table: **nol perubahan** (`role VARCHAR(20)` tetap)

### Step 2: Domain + Port

**Files:**
- `internal/iam/domain/user.go`
- `internal/iam/domain/staff.go`
- `internal/iam/port/token_signer.go`
- `internal/iam/port/repository/staff_repository.go`

**Changes:**
- `User.Roles []string` → `User.SystemRole *domain.Role`
- `User` tambah field `Roles []UserRoleAssignment`
- Struct baru `UserRoleAssignment` di `user.go` (id, name, merchant, branch, is_default)
- `StaffInfo` dan `StaffAssignment` di `staff.go` — hapus
- `TokenClaims.Roles []string` → `TokenClaims.RoleName string`
- Interface `StaffRepository.ListByUserID` return typenya berubah

### Step 3: Repository — Query + Data

**Files:**
- `internal/iam/adapter/repository/postgres/user_repository.go`
- `internal/iam/adapter/repository/postgres/staff_repository.go`

**Changes:**
- `scanUser()` — tambah `type` di SELECT (`u.type`) dan scan param
- `loadUserRoles()` — rewrite jadi:
  - Query (a): `SELECT r.id, r.name FROM user_roles ur JOIN roles r ON r.id = ur.role_id WHERE ur.user_id = $1 AND ur.merchant_id IS NULL`
  - Query (b): `SELECT s.merchant_id, m.name, s.branch_id, b.name, r.id, r.name, s.is_default FROM staff s JOIN roles r ON r.name = s.role JOIN merchants m ON m.id = s.merchant_id JOIN branches b ON b.id = s.branch_id WHERE s.user_id = $1 AND s.status = 'active'`
- `ListByUserID()` di `staff_repository.go` sesuaikan return type + query

### Step 4: Usecase + Handler

**Files:**
- `internal/iam/usecase/auth.go`
- `internal/iam/transport/http/handler/auth_handler.go`
- `internal/iam/transport/http/handler/types.go`

**Changes:**
- `collectPermissions()` — ambil permissions dari `RoleName` di claims (kalo system role), atau kosong (scoped roles gak perlu full permission list di token — bisa di-cache nanti)
- `Me()` handler — assemble `meResponse` dari `User.SystemRole` + `User.Roles`
- `types.go` — baru: `meResponse`, `roleResponse`, `merchantRef`, `branchRef`
- `userResponse` dihapus/disingkron

### Step 5: Middleware + Cleanup

**Files:**
- `internal/iam/transport/http/middleware/middleware.go`
- `internal/iam/domain/role.go` (defaultRoles)
- `internal/iam/transport/http/handler/user_handler.go` (toUserResponse)

**Changes:**
- `AuthMiddleware` — set `claims.RoleName` dari DB
- `RBACMiddleware.hasPermission()` — cek `claims.Permissions` (system role) atau cek via DB (scoped role)
- `domain/role.go` — `DefaultRoles` sesuaikan definisi: hapus yang gak dipake, update permission set
- `toUserResponse()` — ganti jadi `toMeResponse()` atau hapus kalo gak dipake lagi

## Files Not Touched

- Merchant service (staff table, domain, repo, handler — 0 changes)
- IAM routes, bootstrap, error types
