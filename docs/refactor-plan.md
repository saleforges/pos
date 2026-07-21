# Refactor Plan

## Tahap 1: Roles → UUID ID (Stripe-like)

### Masalah
`roles.name` dipake sebagai PRIMARY KEY. Kalau rename role, semua FK (role_permissions, user_roles) ikut patah.

### Pendekatan: Stripe-like ID
```
DB:     id UUID PK, display_id VARCHAR UNIQUE ("role_xxx"), name VARCHAR UNIQUE
API:    GET /v1/roles/:id  (support by display_id or UUID)
JWT:    "roles": ["admin"]              ← tetep nama (small, readable)
Internal: role_id di FK                ← UUID (integrity)
```

Format ID: `role_<random>` — mirip Stripe (`role_abc123`).

### Migration DB

Opsi: drop DB dev, migration ulang dari awal (gak perlu ALTER).

```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    display_id VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_name VARCHAR(255) NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_name)
);

CREATE TABLE user_roles (
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);
```

### Domain
- `Role` struct → tambah field `ID string`, `DisplayID string`
- `User.Roles []string` → tetep nama (`["admin", "cashier"]`)
- `DefaultRoles` map → kuncinya tetep name, seed pake UUID + display_id
- API response: `{ id: "role_abc123", name: "admin", ... }`

### Port (interface)
- Tambah `GetByID(ctx, id string)` — support by UUID or display_id
- `GetByName` tetep ada buat convenience
- Semua internal method pake `roleID string` (UUID)
- CRUD by `display_id` di API, by UUID di internal

### Postgres repo
- Semua query `WHERE name = $1` → `WHERE id = $1`
- Method baru `GetByName` buat legacy lookup

### Memory repo
- Map key `string(name)` → `string(uuid)`
- Tambah index `name → uuid` buat lookup

### HTTP handler / Router
- `:name` path param → `:id`
- TAPI: biarin `name` masih bisa dipake di JSON request body buat convenience

### Seed
- Generate UUID pas insert default roles

### JWT
- Tetep `Roles []string` isinya **nama role** (small, readable)
- Permission check via `permissions[]` yang pre-computed pas login
- Gak perlu trace role → permission tiap request

### Bruno
- Update URL `/:name` → `/:id`
- TAPI: kalo endpoint support both, gak perlu

---

## Tahap 2: Merchant ID → Header

### Masalah
`merchant_id` di-path (`/merchants/:merchantId/...`) bikin URL panjang, inconsistent casing (`:merchantId` vs `:merchantID`), ribet di client.

### Middleware baru: `MerchantMiddleware`
Baca header `X-Merchant-Id`, validasi exist, set di context:

```go
func MerchantMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            merchantID := c.Request().Header.Get("X-Merchant-Id")
            if merchantID == "" {
                return c.JSON(http.StatusBadRequest, map[string]string{
                    "error": "X-Merchant-Id header is required",
                })
            }
            c.Set("merchant_id", merchantID)
            return next(c)
        }
    }
}
```

### Router — Merchant service
```go
// BEFORE
auth.POST("/api/v1/merchants/:merchantId/branches", ...)
auth.GET("/api/v1/merchants/:merchantId/branches", ...)
auth.GET("/api/v1/merchants/:merchantId/staff", ...)

// AFTER
branchGroup := api.Group("/v1/branches", MerchantMiddleware())
branchGroup.POST("", ...)
branchGroup.GET("", ...)
staffGroup := api.Group("/v1/staff", MerchantMiddleware())
staffGroup.GET("", ...)
```

### Router — Catalog service
```go
// BEFORE
cat := api.Group("/merchants/:merchantID/categories")
prod := api.Group("/merchants/:merchantID/products")

// AFTER
cat := api.Group("/v1/categories", MerchantMiddleware())
prod := api.Group("/v1/products", MerchantMiddleware())
```

### Handler
```go
// BEFORE
merchantID := c.Param("merchantId")

// AFTER
merchantID := c.Get("merchant_id").(string)
```

### BranchContext middleware
Ganti baca `merchant_id` dari context (set by MerchantMiddleware) daripada dari path param.

### Bruno
- Hapus `:merchantId` dari URL
- Tambah header `X-Merchant-Id: {{merchant_id}}` di setiap request

### Caddy
- Update path rewrite rules kalo perlu

---

## Tahap 3: Platform vs Merchant Role

### Arsitektur Role Saat Ini

```
UserType = "platform"                    → superadmin, gak terikat merchant
UserType = "merchant"                    → owner/admin/cashier/etc, terikat ke merchant

JWT claims:
{
  "user_type": "platform" | "merchant",
  "roles": ["superadmin"],               ← campur aduk platform + merchant
  "merchant_id": "m1",                   ← kosong kalo platform
  "merchant_role": "",                   ← kosong kalo platform
  "permissions": ["*"]                   ← pre-computed pas login
}
```

### Masalah
1. **Nyampur** — `roles` di JWT gabungin platform role (superadmin) sama merchant role (owner/admin). Susah bedain mana akses global mana yang terbatas ke merchant.
2. **MerchantID di JWT** — Waktu user login di IAM, `merchant_id` langsung di-set dari claims. Tapi seharusnya merchant context itu dinamis, bisa ganti-ganti sesuai request.
3. **Staff info gak ada di JWT** — User bisa kerja di banyak merchant dan cabang, tapi JWT cuma bawa satu `merchant_id`.

### Usulan Desain

```
PLATFORM USER (superadmin):
├── Bisa akses semua merchant
├── X-Merchant-Id: optional (kalo dikasih, operasi terbatas ke merchant itu)
└── JWT:
    {
      "user_type": "platform",
      "roles": ["superadmin"],           ← platform roles
      "permissions": ["*"],
      "merchant_id": null
    }

MERCHANT USER (owner/admin/cashier):
├── Wajib pake X-Merchant-Id
├── Hanya bisa operasi di merchant itu
└── JWT:
    {
      "user_type": "merchant",
      "roles": ["owner"],               ← merchant roles
      "permissions": ["catalog.read", "user.create", ...],
      "merchant_id": "m1",             ← default merchant (optional)
      "staff": [                        ← assignments dari Merchant Service
        { "merchant_id": "m1", "branch_id": "b1", "role": "admin" },
        { "merchant_id": "m2", "branch_id": "b3", "role": "cashier" }
      ]
    }
```

### Flow Request

```
Request:
  GET /v1/branches
  X-Merchant-Id: m1
  Authorization: Bearer <JWT>

Middleware:
  1. AuthMiddleware → verify JWT, set claims (user_id, roles, permissions)
  2. MerchantMiddleware → baca X-Merchant-Id header
  
     ┌─ Kalo platform (user_type=platform):
     │    Header optional. Kalo dikasih, filter operasi ke merchant itu.
     │    Kalo gak dikasih, operasi global.
     │
     └─ Kalo merchant (user_type=merchant):
          Header WAJIB. Operasi cuma di merchant itu.

  3. BranchContextMiddleware → baca merchant dari header, load staff assignments
```

### Yang berubah

#### Domain
- `TokenClaims.MerchantID` → optional (`"merchant_id,omitempty"`)
- `TokenClaims.MerchantRole` → ganti jadi `Staff []StaffAssignment`
- `StaffAssignment` struct baru: `{ MerchantID, BranchID, Role }`

#### IAM Auth handler (pas login)
- Login sukses → panggil Merchant Service buat fetch staff assignments
- Masukin staff assignments ke JWT claims

#### Merchant Auth middleware
- Baca `X-Merchant-Id` dari header
- Validasi: kalo user_type=merchant, pastiin header cocok sama staff assignments
- Set `merchant_id` di context

#### MerchantMiddleware (baru)
- Baca header, bedain platform vs merchant
- Platform: header optional
- Merchant: header wajib, validasi cocok

#### Router
- Bedain route yang butuh merchant context vs yang gak

---

## Priority
1. **Tahap 1 (Roles ID)** — data integrity
2. **Tahap 2 (Merchant Header)** — API design
3. **Tahap 3 (Platform vs Merchant)** — authorization architecture
