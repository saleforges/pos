# Plan: Auto-increment Integer ID

## Masalah

Saat ini ID pake campuran:
- `VARCHAR(64)` — users, refresh_tokens, login_audits (hex dari `pkg/id`)
- `UUID` — roles (baru migrasi Step 1)
- `VARCHAR` — merchant, branches, staff, categories, products, variants (hex)

Debug susah (`id = 'f47ac10b...'`), index fragmented (UUID v4 random).

## Solusi

Semua PK pake `SERIAL` / `BIGSERIAL`. Gak ada `display_id` — cukup `name` atau `code` buat human-readable.

## Daftar Tabel & Kolom

### IAM Service

| Tabel | PK Baru | FK Berubah |
|-------|---------|------------|
| `users` | `id SERIAL` | — |
| `roles` | `id SERIAL` | — |
| `permissions` | `name VARCHAR PK` (tetep) | — |
| `role_permissions` | `(role_id, permission_name)` | `role_id INTEGER REFERENCES roles(id)` |
| `user_roles` | `(user_id, role_id)` | keduanya INTEGER |
| `refresh_tokens` | `id SERIAL` | `user_id INTEGER REFERENCES users(id)` |
| `login_audits` | `id SERIAL` | `user_id INTEGER` (nullable) |

### Merchant Service

| Tabel | PK Baru | FK Berubah |
|-------|---------|------------|
| `merchants` | `id SERIAL` | — |
| `branches` | `id SERIAL` | `merchant_id INTEGER REFERENCES merchants(id)` |
| `staff` | `id SERIAL` | `merchant_id INTEGER`, `branch_id INTEGER` |

### Catalog Service

| Tabel | PK Baru | FK Berubah |
|-------|---------|------------|
| `categories` | `id SERIAL` | `merchant_id INTEGER` |
| `products` | `id SERIAL` | `merchant_id INTEGER` |
| `variants` | `id SERIAL` | `product_id INTEGER` |

## Domain Go — ID string → int

Semua struct:
```go
// SEBELUM
ID string
ParentID *string

// SESUDAH
ID int64
ParentID *int64
```

Yang berubah:
- `domain.User.ID`
- `domain.Role.ID` (hapus DisplayID)
- `domain.Branch.ID`, `Branch.MerchantID`
- `domain.StaffMember.ID`, `StaffMember.MerchantID`, `StaffMember.BranchID`
- `domain.Category.ID`, `Category.MerchantID`, `Category.ParentID`
- `domain.Product.ID`, `Product.MerchantID`
- `domain.Variant.ID`, `Variant.ProductID`
- `TokenClaims.UserID`
- `StaffAssignment.MerchantID`
- `RefreshToken.ID`, `RefreshToken.UserID`
- `LoginAudit.ID`, `LoginAudit.UserID`
- `APIKey.ID`
- dll.

## Repository

### Postgres

Semua `$1` jadi `$1::int`. `GetByID` parameter jadi `id int64`.
```go
// SEBELUM
row.Scan(&user.ID, &user.Username, ...)  // ID string

// SESUDAH
row.Scan(&user.ID, &user.Username, ...)  // ID int64 — pgx.Scan int4 otomatis
```

### Memory

Map key `string` → `int64`:
```go
// SEBELUM
users map[string]*domain.User

// SESUDAH
users map[int64]*domain.User
seq   int64  // auto-counter
```

## pkg/id — Hapus

`pkg/id.Generate()` gak dipake lagi. Semua ID dari `SERIAL` DB.
`generateID()` di `usecase/auth.go` juga dihapus.

## API Response

```json
// SEBELUM
{ "id": "abc123def456", "name": "admin" }

// SESUDAH
{ "id": 1, "name": "admin" }
```

## Migration Strategy

**Drop semua tabel, migration dari awal** (sama kayak Step 1 & 2). Update file `000001_init.up.sql` di setiap service.

## Tahapan Pengerjaan

1. **IAM domain** — semua struct ID string → int64
2. **IAM port** — interface signatures
3. **IAM Postgres** — queries + scan
4. **IAM Memory** — map key int64 + seq
5. **IAM usecase** — hapus `generateID()`, adjust logic
6. **IAM handler** — adjust kalo perlu
7. **IAM seed** — `gen_random_uuid()` → DEFAULT
8. **Merchant domain + port + repo**
9. **Catalog domain + port + repo**
10. **Migrasi SQL** — update semua 3 service
11. **Build + test**
