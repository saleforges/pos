# Catalog Service Refactor Plan

## 1. Current Architecture Problems

### Domain Model Issues
- **Product contains price/cost/tax_rate** — violates the catalog boundary. Catalog should not own pricing.
- **Product has SKU and barcode directly** — conflates product identity with sellable unit identity.
- **No SellableItem entity** — "Variant" exists but it's a thin copy of Product with price/cost. Doesn't reflect the UMKM concept of "one product, multiple selling forms."
- **No Unit entity** — unit is a free-text string. No master data.
- **Missing SellableItemBarcode** — barcode stored directly on product/variant, can't support multiple barcodes per item.
- **Category has slug, description, sort_order** — over-engineered for UMKM use.

### Structural Issues
- **No `*_service.go` interface files** — usecase interfaces are defined inline in `interfaces.go` instead of being in separate `*_service.go` files (IAM/Merchant convention).
- **Handlers in flat `handler/` package** — IAM and Merchant use per-entity subfolders (`transport/http/product/`, `transport/http/category/`).
- **No DTO/mapper separation** — request/response types and conversion functions live in `types.go` alongside handlers. IAM/Merchant split these into `dto.go` and `mapper.go`.
- **Custom pagination** — uses local `PaginatedResult`/`PaginationMeta` instead of shared `pkg/pagination`.
- **PUT instead of PATCH** — current catalog uses PUT for updates. IAM/Merchant use PATCH with pointer fields for partial updates.
- **snake_case JSON** — current catalog uses `snake_case` JSON tags. IAM/Merchant use `camelCase`.
- **No `common/errors.go`** — no shared HTTP error codes/domain error prefixes at the transport layer.
- **No tests at any layer** — domain, usecase, handler, repo all untested.

### Auth / Multi-Tenant Issues
- **Auth middleware is inlined** in `router.go` instead of being in a proper `middleware/` package.
- **No RBAC middleware** — no permission checking on catalog endpoints.
- **merchant_id passed explicitly** in usecase inputs instead of extracted from JWT claims via context.

## 2. Proposed Architecture

```
services/internal/catalog/
├── adapter/
│   ├── repository/
│   │   ├── memory/
│   │   │   ├── product_repository.go
│   │   │   ├── product_repository_test.go
│   │   │   ├── sellable_item_repository.go
│   │   │   ├── sellable_item_repository_test.go
│   │   │   ├── category_repository.go
│   │   │   ├── category_repository_test.go
│   │   │   ├── unit_repository.go
│   │   │   └── unit_repository_test.go
│   │   └── postgres/
│   │       ├── db.go
│   │       ├── seed.go
│   │       ├── product_repository.go
│   │       ├── sellable_item_repository.go
│   │       ├── category_repository.go
│   │       ├── unit_repository.go
│   │       └── migrations/
│   └── storage/
│       └── minio/
│           └── client.go  (keep as-is)
├── bootstrap/
│   └── bootstrap.go
├── domain/
│   ├── errors.go
│   ├── product.go
│   ├── sellable_item.go
│   ├── category.go
│   └── unit.go
├── port/
│   └── repository/
│       ├── product_repository.go
│       ├── sellable_item_repository.go
│       ├── category_repository.go
│       └── unit_repository.go
├── transport/
│   └── http/
│       ├── middleware/
│       │   ├── middleware.go
│       │   └── middleware_test.go
│       ├── product/
│       │   ├── dto.go
│       │   ├── handler.go
│       │   ├── handler_test.go
│       │   └── mapper.go
│       ├── sellable_item/
│       │   ├── dto.go
│       │   ├── handler.go
│       │   ├── handler_test.go
│       │   └── mapper.go
│       ├── category/
│       │   ├── dto.go
│       │   ├── handler.go
│       │   ├── handler_test.go
│       │   └── mapper.go
│       ├── unit/
│       │   ├── dto.go
│       │   ├── handler.go
│       │   ├── handler_test.go
│       │   └── mapper.go
│       └── router.go
└── usecase/
    ├── product.go
    ├── product_service.go
    ├── product_test.go
    ├── sellable_item.go
    ├── sellable_item_service.go
    ├── sellable_item_test.go
    ├── category.go
    ├── category_service.go
    ├── category_test.go
    ├── unit.go
    ├── unit_service.go
    ├── unit_test.go
    └── mocks_test.go
```

## 3. Domain Model Changes

### Product (new)
```go
type Product struct {
    ID          int64           `json:"id"`
    MerchantID  int64           `json:"merchantId"`
    CategoryID  int64           `json:"categoryId"`
    Name        string          `json:"name"`
    Description string          `json:"description,omitempty"`
    Status      ProductStatus   `json:"status"`
    CreatedAt   time.Time       `json:"createdAt"`
    UpdatedAt   time.Time       `json:"updatedAt"`
}
```
Removed: SKU, Barcode, Price, Cost, TaxRate, Unit, ImageURL.

### SellableItem (new — replaces Variant)
```go
type SellableItem struct {
    ID             int64             `json:"id"`
    ProductID      int64             `json:"productId"`
    Name           string            `json:"name"`
    UnitID         int64             `json:"unitId"`
    TrackInventory bool              `json:"trackInventory"`
    ImageURL       string            `json:"imageUrl,omitempty"`
    Status         SellableItemStatus `json:"status"`
    CreatedAt      time.Time         `json:"createdAt"`
    UpdatedAt      time.Time         `json:"updatedAt"`
}
```
No SKU, Price, Cost on SellableItem either.

### Unit (new — master data)
```go
type Unit struct {
    ID   int64  `json:"id"`
    Code string `json:"code"`
    Name string `json:"name"`
}
```

### SellableItemBarcode (new entity)
```go
type SellableItemBarcode struct {
    ID             int64  `json:"id"`
    SellableItemID int64  `json:"sellableItemId"`
    Barcode        string `json:"barcode"`
}
```

### Category (simplified)
```go
type Category struct {
    ID         int64     `json:"id"`
    MerchantID int64     `json:"merchantId"`
    Name       string    `json:"name"`
    ParentID   *int64    `json:"parentId,omitempty"`
    CreatedAt  time.Time `json:"createdAt"`
    UpdatedAt  time.Time `json:"updatedAt"`
}
```
Removed: slug, description, sort_order, status.

### Error Codes (consistent with IAM/Merchant)
```
CAT001: product not found
CAT002: category not found  
CAT003: sellable item not found
CAT004: unit not found
CAT005: sellable item barcode exists
CAT006: invalid product data
CAT007: invalid category data
CAT008: invalid sellable item data
CAT500: internal error
```

## 4. Database Changes

### Migrations (new files in `adapter/repository/postgres/migrations/`)

**`000001_create_units.up.sql`**
```sql
CREATE TABLE units (
    id   BIGSERIAL PRIMARY KEY,
    code VARCHAR(10)  NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL
);

INSERT INTO units (code, name) VALUES
    ('PCS', 'Piece'),
    ('PACK', 'Pack'),
    ('KG', 'Kilogram'),
    ('GRAM', 'Gram'),
    ('LITER', 'Liter'),
    ('ML', 'Milliliter'),
    ('BOX', 'Box'),
    ('METER', 'Meter');
```

**`000002_create_categories.up.sql`**
```sql
CREATE TABLE categories (
    id          BIGSERIAL    PRIMARY KEY,
    merchant_id BIGINT       NOT NULL REFERENCES merchants(id),
    name        VARCHAR(100) NOT NULL,
    parent_id   BIGINT       REFERENCES categories(id),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_categories_merchant ON categories(merchant_id);
```

**`000003_create_products.up.sql`**
```sql
CREATE TABLE products (
    id          BIGSERIAL    PRIMARY KEY,
    merchant_id BIGINT       NOT NULL REFERENCES merchants(id),
    category_id BIGINT       NOT NULL REFERENCES categories(id),
    name        VARCHAR(200) NOT NULL,
    description TEXT,
    status      VARCHAR(20)  NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_products_merchant ON products(merchant_id);
CREATE INDEX idx_products_category ON products(category_id);
```

**`000004_create_sellable_items.up.sql`**
```sql
CREATE TABLE sellable_items (
    id               BIGSERIAL    PRIMARY KEY,
    product_id       BIGINT       NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name             VARCHAR(200) NOT NULL,
    unit_id          BIGINT       NOT NULL REFERENCES units(id),
    track_inventory  BOOLEAN      NOT NULL DEFAULT TRUE,
    image_url        TEXT,
    status           VARCHAR(20)  NOT NULL DEFAULT 'active',
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_sellable_items_product ON sellable_items(product_id);
```

**`000005_create_sellable_item_barcodes.up.sql`**
```sql
CREATE TABLE sellable_item_barcodes (
    id               BIGSERIAL    PRIMARY KEY,
    sellable_item_id BIGINT       NOT NULL REFERENCES sellable_items(id) ON DELETE CASCADE,
    barcode          VARCHAR(100) NOT NULL
);
CREATE UNIQUE INDEX idx_barcodes_unique ON sellable_item_barcodes(barcode);
CREATE INDEX idx_barcodes_item ON sellable_item_barcodes(sellable_item_id);
```

**Migration file naming:**
Use golang-migrate format (`NNNNNN_<name>.up.sql` / `NNNNNN_<name>.down.sql`), matching IAM's convention.

### Data Migration
Existing products will need a seed/migration script to:
1. Create system units (PCS, PACK, KG, etc.)
2. Migrate existing categories (strip slug, description, sort_order, status)
3. Transform old `products` table into new `products` + `sellable_items`:
   - Each old product → new product (drop price/cost/tax/sku/barcode/unit/image_url)
   - Each old variant → new sellable_item  
   - Old product unit → reasonable default unit in new schema
4. Drop columns: price, cost, tax_rate, sku, barcode, unit, image_url from products

## 5. API Changes

### Routes (standardized with IAM/Merchant style)

| Method | Path | Handler | Notes |
|--------|------|---------|-------|
| GET | /health | inline | |
| GET | /metrics | inline | |
| POST | /api/v1/products | product.Create | merchant_id from JWT |
| GET | /api/v1/products | product.List | paginated via pagination.Params |
| GET | /api/v1/products/:id | product.Get | |
| PATCH | /api/v1/products/:id | product.Update | partial update via pointer fields |
| DELETE | /api/v1/products/:id | product.Delete | |
| POST | /api/v1/products/:productId/items | sellable_item.Create | |
| GET | /api/v1/products/:productId/items | sellable_item.List | |
| PATCH | /api/v1/items/:id | sellable_item.Update | |
| DELETE | /api/v1/items/:id | sellable_item.Delete | |
| POST | /api/v1/categories | category.Create | |
| GET | /api/v1/categories | category.List | |
| PATCH | /api/v1/categories/:id | category.Update | |
| GET | /api/v1/units | unit.List | |

### Key changes from current API
- `PUT` → `PATCH` for updates (pointer fields for partial updates)
- `snake_case` → `camelCase` in JSON responses
- `Variant` endpoints → `SellableItem` endpoints
- No `price`, `cost`, `tax_rate` in product requests/responses
- No `merchant_id` or `branch_id` in request bodies (read from JWT claims)
- Pagination uses shared `pagination.Params` / `httputil.WritePaginated`
- Response format follows IAM style: `{"data": ..., "message": "success"}`

### Response format
```json
// Success
{"data": {...}, "message": "success"}
// Paginated
{"data": [...], "pagination": {"total": 100, "offset": 0, "limit": 20, "returnCount": 20}}
// Error
{"code": "CAT001", "message": "product not found"}
```

## 6. Code Changes Summary

### New files needed

| Package | File | Purpose |
|---------|------|---------|
| domain | product.go | Product entity |
| domain | sellable_item.go | SellableItem + SellableItemBarcode entities |
| domain | category.go | Simplified Category entity |
| domain | unit.go | Unit entity |
| domain | errors.go | Error codes with CAT prefix |
| port/repository | product_repository.go | ProductRepository interface (no price/cost/sku/barcode) |
| port/repository | sellable_item_repository.go | SellableItemRepository + BarcodeRepository |
| port/repository | category_repository.go | CategoryRepository (simplified) |
| port/repository | unit_repository.go | UnitRepository |
| usecase | product_service.go | ProductUsecase interface |
| usecase | product.go | productUsecase implementation |
| usecase | product_test.go | Product usecase tests |
| usecase | sellable_item_service.go | SellableItemUsecase interface |
| usecase | sellable_item.go | sellableItemUsecase implementation |
| usecase | sellable_item_test.go | SellableItem usecase tests |
| usecase | category_service.go | CategoryUsecase interface |
| usecase | category.go | categoryUsecase implementation |
| usecase | category_test.go | Category usecase tests |
| usecase | unit_service.go | UnitUsecase interface |
| usecase | unit.go | unitUsecase implementation |
| usecase | unit_test.go | Unit usecase tests |
| usecase | mocks_test.go | Shared mock repositories |
| transport/http/middleware | middleware.go | Auth + MerchantContext middleware |
| transport/http/middleware | middleware_test.go | Middleware tests |
| transport/http/product | dto.go | Product request/response types |
| transport/http/product | handler.go | Product HTTP handler |
| transport/http/product | handler_test.go | Product handler tests |
| transport/http/product | mapper.go | Request → Params conversion |
| transport/http/sellable_item | dto.go | SellableItem request/response types |
| transport/http/sellable_item | handler.go | SellableItem HTTP handler |
| transport/http/sellable_item | handler_test.go | SellableItem handler tests |
| transport/http/sellable_item | mapper.go | Request → Params conversion |
| transport/http/category | dto.go | Category request/response types |
| transport/http/category | handler.go | Category HTTP handler |
| transport/http/category | handler_test.go | Category handler tests |
| transport/http/category | mapper.go | Request → Params conversion |
| transport/http/unit | dto.go | Unit response types |
| transport/http/unit | handler.go | Unit HTTP handler |
| transport/http/unit | handler_test.go | Unit handler tests |
| transport/http/unit | mapper.go | Request → Params conversion |
| transport/http | router.go | Route wiring (auth → merchant middleware → handlers) |
| adapter/repository/memory | product_repository.go | In-memory ProductRepository |
| adapter/repository/memory | sellable_item_repository.go | In-memory SellableItemRepository |
| adapter/repository/memory | category_repository.go | In-memory CategoryRepository |
| adapter/repository/memory | unit_repository.go | In-memory UnitRepository |
| adapter/repository/memory | *test.go | In-memory repo tests |
| adapter/repository/postgres | product_repository.go | Postgres ProductRepository |
| adapter/repository/postgres | sellable_item_repository.go | Postgres SellableItemRepository |
| adapter/repository/postgres | category_repository.go | Postgres CategoryRepository (simplified) |
| adapter/repository/postgres | unit_repository.go | Postgres UnitRepository |
| adapter/repository/postgres | db.go | Keep existing + add RunMigrations |
| adapter/repository/postgres | seed.go | Seed default units, sample categories |
| adapter/repository/postgres/migrations/ | *.up.sql / *.down.sql | Database migrations |
| bootstrap | bootstrap.go | Rewired DI |
| — | image handler | Keep as-is (MinIO upload) |

### Modified files
- `bootstrap/bootstrap.go` — wire new repos, usecases, handlers
- `adapter/storage/minio/client.go` — keep unchanged

### Deleted files
- `usecase/interfaces.go` — replaced by per-entity `*_service.go` files
- `usecase/product.go` — replaced
- `usecase/category.go` — replaced
- `usecase/variant.go` — replaced
- `transport/http/handler/` — entire directory replaced by per-entity subfolders
- `transport/http/router.go` — replaced
- `port/repository/product_repository.go` — replaced (new interface)
- `port/repository/variant_repository.go` — replaced by sellable_item_repository
- `port/repository/category_repository.go` — replaced (new interface)
- `adapter/repository/memory/product_repository.go` — replaced
- `adapter/repository/memory/category_repository.go` — replaced
- `adapter/repository/memory/variant_repository.go` — replaced
- `adapter/repository/postgres/product_repository.go` — replaced (new schema)
- `adapter/repository/postgres/category_repository.go` — replaced
- `adapter/repository/postgres/variant_repository.go` — replaced

## 7. Migration Steps

### Phase 1 — Foundation (new entities, no existing data touch)
1. Create new domain entities: Product (clean), SellableItem, Unit, Category (simplified), SellableItemBarcode, errors
2. Create port/repository interfaces for all 4 entities
3. Write migrations: units, categories, products, sellable_items, sellable_item_barcodes
4. Implement postgres repositories for all 4 entities
5. Implement memory repositories for all 4 entities (+ tests)
6. Add seed data (default units)

### Phase 2 — Usecase layer
7. Create `*_service.go` interface files for all 4 usecases
8. Implement usecases for all 4 entities (+ tests)
9. Create `usecase/mocks_test.go` with mock repositories

### Phase 3 — HTTP transport
10. Create `transport/http/middleware/` with auth + merchant context middleware
11. Create per-entity handler directories with dto, handler, mapper, tests
12. Rewrite `transport/http/router.go`
13. Wire auth middleware using `pkg/jwks.Verifier` (same pattern as Merchant)

### Phase 4 — Bootstrap wiring
14. Update `bootstrap/bootstrap.go` to wire new layers
15. Remove old handler/usecase imports

### Phase 5 — Cleanup
16. Delete old files (old product/category/variant handlers, usecases, repos)
17. Run full test suite
18. Data migration for existing records (if production data exists)

## 8. Risk Areas

### Data Migration Risk
- **Breaking change**: Existing products have price/cost/tax_rate. New schema drops these columns. Must ensure downstream services (Order, Payment) have already been migrated to read pricing from their own domain.
- **Strategy**: Deploy new schema alongside old code temporarily (dual-write if needed). Migrate data before cutting over.

### API Contract Changes
- **PUT → PATCH** changes the request format for updates. Frontend must be updated simultaneously.
- **camelCase JSON** breaks existing frontend code that reads snake_case fields. Frontend migration needed.
- **Endpoint path changes**: `/products/:productID/variants` → `/products/:productId/items`. Frontend must update.

### Service Boundary Impact
- **Pricing**: Currently on Product. After refactor, no catalog entity holds price. The Order Service must own pricing. If Order isn't ready, this creates a gap.
- **Mitigation**: Keep price as deprecated fields in the first phase, add warning logs, remove in Phase 2 after Order service is ready.

### Image Upload
- Image upload via MinIO currently works. The new SellableItem has `image_url` but the upload handler references products. Need to update or keep the existing handler.

### Testing Debt
- Current Catalog has zero tests. New code must hit high coverage before merging.
- Mock repositories in `usecase/mocks_test.go` follow Merchant's pattern (map-based, error injection).

## 9. Testing Strategy

| Layer | Tests | Pattern |
|-------|-------|---------|
| Domain | Product validation, SellableItem validation | Table-driven tests |
| Usecase | Create product (success, duplicate, missing category), merchant isolation, create sellable item, list with pagination | Mock repos from mocks_test.go |
| Repository | CRUD for all 4 entities (memory only — faster, no DB dependency) | In-memory implementation |
| HTTP Handler | Create (201, 400, 409, 500), Get (200, 404), List (200), Update (200, 400, 404), Delete (200, 404) | Echo test helpers + mock usecase |
| Middleware | Auth (valid token, invalid token, missing), Merchant context (from JWT, from header) | Echo test helpers |

## 10. Execution Order

1. Domain entities + errors
2. Port interfaces
3. Memory repositories + tests
4. Postgres repositories (with migrations)
5. Usecase interfaces + implementations + mocks + tests
6. HTTP middleware
7. HTTP handlers + DTOs + mappers + tests
8. Router
9. Bootstrap
10. Delete old files
11. Run full test suite
12. Data migration

This plan ensures every layer is testable independently, and the new code can coexist with the old during transition.
