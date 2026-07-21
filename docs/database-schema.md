# Database Schema

Semua service berbagi **satu database PostgreSQL** (`devdb`). Migrasi dikelola per-service dengan `golang-migrate`.

---

## 1. IAM Service (`internal/iam`)

### `users`

User akun — platform (superadmin) atau merchant (owner, staff).

| Column     | Type              | Constraints                |
|------------|-------------------|---------------------------|
| id         | `BIGSERIAL`       | `PRIMARY KEY`             |
| username   | `VARCHAR(255)`    | `NOT NULL UNIQUE`         |
| email      | `VARCHAR(255)`    | `NOT NULL UNIQUE`         |
| password   | `VARCHAR(255)`    | `NOT NULL` (argon2 hash)  |
| type       | `VARCHAR(20)`     | `NOT NULL DEFAULT 'merchant'` |
| status     | `VARCHAR(20)`     | `NOT NULL DEFAULT 'active'`  |
| created_at | `TIMESTAMPTZ`     | `NOT NULL DEFAULT NOW()`  |
| updated_at | `TIMESTAMPTZ`     | `NOT NULL DEFAULT NOW()`  |

```
UserType:
  platform — superadmin
  merchant — owner, admin, staff, dll

UserStatus:
  active, disabled
```

### `roles`

System permission roles — bukan staff position.

| Column      | Type              | Constraints                |
|-------------|-------------------|---------------------------|
| id          | `BIGSERIAL`       | `PRIMARY KEY`             |
| name        | `VARCHAR(255)`    | `NOT NULL UNIQUE`         |
| description | `TEXT`            | `NOT NULL DEFAULT ''`     |
| is_system   | `BOOLEAN`         | `NOT NULL DEFAULT FALSE`  |
| created_at  | `TIMESTAMPTZ`     | `NOT NULL DEFAULT NOW()`  |
| updated_at  | `TIMESTAMPTZ`     | `NOT NULL DEFAULT NOW()`  |

**Seed roles:** superadmin, owner, admin, supervisor, cashier, viewer

### `permissions`

System permissions (resource.action).

| Column     | Type              | Constraints                |
|------------|-------------------|---------------------------|
| name       | `VARCHAR(255)`    | `PRIMARY KEY`             |
| created_at | `TIMESTAMPTZ`     | `NOT NULL DEFAULT NOW()`  |

**Seed count:** 28 permissions (catalog.read, sales.create, user.manage, dll)

### `role_permissions`

| Column          | Type              | Constraints                                    |
|-----------------|-------------------|------------------------------------------------|
| role_id         | `BIGINT`          | `REFERENCES roles(id) ON DELETE CASCADE`       |
| permission_name | `VARCHAR(255)`    | `REFERENCES permissions(name) ON DELETE CASCADE` |
|                 |                   | `PRIMARY KEY (role_id, permission_name)`       |

### `user_roles`

| Column  | Type     | Constraints                                    |
|---------|----------|------------------------------------------------|
| user_id | `BIGINT` | `REFERENCES users(id) ON DELETE CASCADE`       |
| role_id | `BIGINT` | `REFERENCES roles(id) ON DELETE CASCADE`       |
|         |          | `PRIMARY KEY (user_id, role_id)`               |

### `refresh_tokens`

| Column     | Type              | Constraints                                    |
|------------|-------------------|------------------------------------------------|
| id         | `BIGSERIAL`       | `PRIMARY KEY`                                 |
| user_id    | `BIGINT`          | `NOT NULL REFERENCES users(id) ON DELETE CASCADE` |
| token_hash | `VARCHAR(255)`    | `NOT NULL`                                    |
| expires_at | `TIMESTAMPTZ`     | `NOT NULL`                                    |
| created_at | `TIMESTAMPTZ`     | `NOT NULL DEFAULT NOW()`                      |
| revoked_at | `TIMESTAMPTZ`     |                                               |

**Indexes:** `idx_refresh_tokens_token_hash`, `idx_refresh_tokens_user_id`

### `login_audits`

| Column     | Type              | Constraints                |
|------------|-------------------|---------------------------|
| id         | `BIGSERIAL`       | `PRIMARY KEY`             |
| user_id    | `BIGINT`          |                           |
| email      | `VARCHAR(255)`    | `NOT NULL DEFAULT ''`     |
| success    | `BOOLEAN`         | `NOT NULL`                |
| ip_address | `VARCHAR(45)`     | `NOT NULL DEFAULT ''`     |
| user_agent | `TEXT`            | `NOT NULL DEFAULT ''`     |
| reason     | `VARCHAR(255)`    | `NOT NULL DEFAULT ''`     |
| created_at | `TIMESTAMPTZ`     | `NOT NULL DEFAULT NOW()`  |

**Indexes:** `idx_login_audits_user_id`, `idx_login_audits_created_at`

---

## 2. Merchant Service (`internal/merchant`)

### `merchants`

| Column              | Type              | Constraints                |
|---------------------|-------------------|---------------------------|
| id                  | `BIGSERIAL`       | `PRIMARY KEY`             |
| name                | `VARCHAR(255)`    | `NOT NULL`                |
| legal_name          | `VARCHAR(255)`    | `NOT NULL DEFAULT ''`     |
| address             | `TEXT`            | `NOT NULL DEFAULT ''`     |
| phone               | `VARCHAR(50)`     | `NOT NULL DEFAULT ''`     |
| email               | `VARCHAR(255)`    | `NOT NULL`                |
| logo_url            | `TEXT`            | `NOT NULL DEFAULT ''`     |
| tax_id              | `VARCHAR(100)`    | `NOT NULL DEFAULT ''`     |
| status              | `VARCHAR(20)`     | `NOT NULL DEFAULT 'active'` |
| tax_rate            | `NUMERIC(5,2)`    | `NOT NULL DEFAULT 0`      |
| currency            | `VARCHAR(10)`     | `NOT NULL DEFAULT 'IDR'`  |
| timezone            | `VARCHAR(100)`    | `NOT NULL DEFAULT 'Asia/Jakarta'` |
| receipt_footer      | `TEXT`            | `NOT NULL DEFAULT ''`     |
| receipt_logo        | `TEXT`            | `NOT NULL DEFAULT ''`     |
| order_prefix        | `VARCHAR(50)`     | `NOT NULL DEFAULT ''`     |
| low_stock_threshold | `INT`             | `NOT NULL DEFAULT 10`     |
| created_at          | `TIMESTAMPTZ`     | `NOT NULL DEFAULT NOW()`  |
| updated_at          | `TIMESTAMPTZ`     | `NOT NULL DEFAULT NOW()`  |

### `branches`

| Column         | Type              | Constraints                                    |
|----------------|-------------------|------------------------------------------------|
| id             | `BIGSERIAL`       | `PRIMARY KEY`                                 |
| merchant_id    | `BIGINT`          | `NOT NULL REFERENCES merchants(id) ON DELETE CASCADE` |
| name           | `VARCHAR(255)`    | `NOT NULL`                                    |
| code           | `VARCHAR(100)`    | `NOT NULL`                                    |
| address        | `TEXT`            | `NOT NULL DEFAULT ''`                         |
| phone          | `VARCHAR(50)`     | `NOT NULL DEFAULT ''`                         |
| status         | `VARCHAR(20)`     | `NOT NULL DEFAULT 'active'`                   |
| operating_days | `TEXT[]`          | `NOT NULL DEFAULT '{}'`                       |
| open_time      | `VARCHAR(10)`     | `NOT NULL DEFAULT '08:00'`                    |
| close_time     | `VARCHAR(10)`     | `NOT NULL DEFAULT '21:00'`                    |
| created_at     | `TIMESTAMPTZ`     | `NOT NULL DEFAULT NOW()`                      |
| updated_at     | `TIMESTAMPTZ`     | `NOT NULL DEFAULT NOW()`                      |

**Unique:** `(merchant_id, code)`
**Index:** `idx_branches_merchant_id`

### `staff`

Staff assignment — menghubungkan user, merchant, dan branch.

| Column      | Type              | Constraints                                    |
|-------------|-------------------|------------------------------------------------|
| id          | `BIGSERIAL`       | `PRIMARY KEY`                                 |
| merchant_id | `BIGINT`          | `NOT NULL REFERENCES merchants(id) ON DELETE CASCADE` |
| branch_id   | `BIGINT`          | `NOT NULL REFERENCES branches(id) ON DELETE CASCADE` |
| user_id     | `BIGINT`          | `NOT NULL` (ref: iam.users.id)                |
| role        | `VARCHAR(20)`     | `NOT NULL DEFAULT 'cashier'`                  |
| status      | `VARCHAR(20)`     | `NOT NULL DEFAULT 'active'`                   |
| is_default  | `BOOLEAN`         | `NOT NULL DEFAULT FALSE`                      |
| created_at  | `TIMESTAMPTZ`     | `NOT NULL DEFAULT NOW()`                      |
| updated_at  | `TIMESTAMPTZ`     | `NOT NULL DEFAULT NOW()`                      |

**Unique:** `(user_id, branch_id)` — satu user hanya bisa 1 role per branch
**Indexes:** `idx_staff_merchant_id`, `idx_staff_branch_id`, `idx_staff_user_id`

> `staff.role` adalah string (`VARCHAR`) — menyimpan nama posisi seperti `cashier`, `supervisor`, `manager`, `viewer`. Beda konsep dari `roles` di IAM yang merupakan system permission. (Diskusi: rencana migrate ke `role_id BIGINT REFERENCES roles(id)` atau buat tabel `staff_roles` tersendiri.)

---

## 3. Catalog Service (`internal/catalog`)

Catalog belum punya migration SQL. Tabel dibuat manual atau via ORM. Berikut domain struct yang merepresentasikan skema yang diharapkan:

### `categories`

| Column      | Type              | Constraints                |
|-------------|-------------------|---------------------------|
| id          | `BIGSERIAL`       | `PRIMARY KEY`             |
| merchant_id | `BIGINT`          | `NOT NULL`                |
| name        | `VARCHAR(255)`    | `NOT NULL`                |
| slug        | `VARCHAR(255)`    | `NOT NULL`                |
| description | `TEXT`            |                           |
| parent_id   | `BIGINT`          | (self-ref, nullable)      |
| sort_order  | `INT`             | `DEFAULT 0`               |
| status      | `VARCHAR(20)`     | `NOT NULL DEFAULT 'active'` |
| created_at  | `TIMESTAMPTZ`     | `NOT NULL`                |
| updated_at  | `TIMESTAMPTZ`     | `NOT NULL`                |

### `products`

| Column      | Type              | Constraints                |
|-------------|-------------------|---------------------------|
| id          | `BIGSERIAL`       | `PRIMARY KEY`             |
| merchant_id | `BIGINT`          | `NOT NULL`                |
| category_id | `BIGINT`          | `NOT NULL`                |
| name        | `VARCHAR(255)`    | `NOT NULL`                |
| sku         | `VARCHAR(100)`    | `NOT NULL`                |
| barcode     | `VARCHAR(100)`    |                           |
| description | `TEXT`            |                           |
| price       | `NUMERIC(12,2)`   | `NOT NULL`                |
| cost        | `NUMERIC(12,2)`   |                           |
| tax_rate    | `NUMERIC(5,2)`    | `NOT NULL DEFAULT 0`      |
| unit        | `VARCHAR(50)`     | `NOT NULL`                |
| image_url   | `TEXT`            |                           |
| status      | `VARCHAR(20)`     | `NOT NULL DEFAULT 'active'` |
| created_at  | `TIMESTAMPTZ`     | `NOT NULL`                |
| updated_at  | `TIMESTAMPTZ`     | `NOT NULL`                |

### `variants`

| Column     | Type              | Constraints                |
|------------|-------------------|---------------------------|
| id         | `BIGSERIAL`       | `PRIMARY KEY`             |
| product_id | `BIGINT`          | `NOT NULL`                |
| name       | `VARCHAR(255)`    | `NOT NULL`                |
| sku        | `VARCHAR(100)`    | `NOT NULL`                |
| barcode    | `VARCHAR(100)`    |                           |
| price      | `NUMERIC(12,2)`   |                           |
| cost       | `NUMERIC(12,2)`   |                           |
| image_url  | `TEXT`            |                           |
| sort_order | `INT`             | `DEFAULT 0`               |
| created_at | `TIMESTAMPTZ`     | `NOT NULL`                |
| updated_at | `TIMESTAMPTZ`     | `NOT NULL`                |

---

## Entity Relationship

```
┌───────────┐     ┌──────────────────────┐     ┌──────────────┐
│  users    │1───┼│     user_roles       │1───┼│    roles     │
│  (IAM)    │     │  user_id, role_id    │     │  (IAM)       │
└───────────┘     └──────────────────────┘     └──────┬───────┘
       │ 1                                            │ 1
       │                                              │
       │ *                                  ┌─────────┴──────────┐
       │                                  │  role_permissions   │
       │                                  │  role_id, perm_name │
       │                                  └─────────────────────┘
       │
       │ * (via staff.user_id)
┌──────┴────────┐     ┌────────────────┐
│    staff      │┼───1│   branches     │
│  (Merchant)   │     │  (Merchant)    │
└──────┬────────┘     └───────┬────────┘
       │ *                    │ *
       │                      │
┌──────┴────────┐             │
│  merchants   │┼─────────────┘
│  (Merchant)  │
└──────────────┘
```

---

## Catatan Penting

1. **Shared database:** IAM, Merchant, dan Catalog semua pake database `devdb` yang sama. Tabel dipisah secara logika per-service, bukan per-database.
2. **`staff.user_id`** refer ke `users.id` (IAM) tapi tanpa FK constraint (karena beda migration file). Hanya integer reference.
3. **Catalog belum punya migration SQL.** Tabel perlu dibuat manual atau migration file perlu ditambahkan.
4. **`staff.role`** saat ini string. Rencana migrasi ke `role_id` dengan FK ke `roles(id)` atau tabel `staff_roles` sendiri.
