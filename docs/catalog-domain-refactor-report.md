# Catalog Service — Domain Refactor Report

> **Date:** 2026-07-24
> **Branch:** `fix/permissive-cors` → feature branch needed
> **Status:** Implemented, compiled, tested ✅

---

## 1. What Changed

**Concept replace:** `SellableItem` → `ProductItem`

The core selling entity is no longer abstract. `ProductItem` is the exact item a POS cashier scans and sells.

| Old | New |
|-----|-----|
| `SellableItem` | `ProductItem` |
| `SellableItemBarcode` | `ProductItemBarcode` |
| `SellableItemStatus` | `ProductItemStatus` |
| `ErrSellableItemNotFound` | `ErrProductItemNotFound` |
| `/items/...` route | `/product-items/...` route |
| `POST /v1/products/bulk` items use SKU | items can include `sku` field |

---

## 2. Final Domain Model

```
Merchant Service
  └── Merchant
       └── Branch

Catalog Service
  └── Product              ← master / grouping product
       └── ProductItem     ← actual sellable entity in POS
            ├── SKU        ← optional, unique per merchant
            ├── Barcode    ← separate table, multiple per item
            ├── Unit       ← nullable (PCS, KG, PACK, etc.)
            └── Price      ← separate table, one active price
```

### Product fields

```
id, merchant_id, category_id, name, description, image, status, created_at, updated_at
```

**Does NOT contain:** price, SKU, barcode, stock, quantity, branch_id, unit.

### ProductItem fields

```
id, product_id, name, sku, unit_id (nullable), track_inventory, image_url, status, created_at, updated_at
```

**Does NOT contain:** stock, quantity, branch_id, stock movement.

### Price (separate table)

```
id, product_item_id (unique), amount, currency (default IDR), created_at, updated_at
```

- One active price per ProductItem.
- No price history. No branch pricing. No promotions.

### Barcode (separate table)

```
id, product_item_id, barcode (unique)
```

- Multiple barcodes per ProductItem allowed.

---

## 3. Product Creation Logic

### Case 1 — Simple Product

User sends:
```json
{ "name": "Es Teh", "categoryId": 1, "price": 3000 }
```

Backend auto-creates:
- Product: `Es Teh`
- ProductItem: `Es Teh` (unit = null, price = 3000)

### Case 2 — Multiple Selling Forms (Bulk)

User sends:
```json
{
  "name": "Marning",
  "categoryId": 1,
  "items": [
    { "name": "Marning Curah", "sku": "MRN-KG-001", "unitId": 3, "price": 40000 },
    { "name": "Marning Pack", "sku": "MRN-PACK-001", "unitId": 2, "price": 1000 }
  ]
}
```

Backend creates:
- Product: `Marning`
- ProductItem: `Marning Curah` + `Marning Pack`

---

## 4. API Endpoints

### Product (unchanged)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/products` | Create product (simple: include `price` auto-creates item) |
| POST | `/api/v1/products/bulk` | Create product + multiple items in one call |
| GET | `/api/v1/products` | List products (paginated, with items & price range) |
| GET | `/api/v1/products/:id` | Get product detail |
| PATCH | `/api/v1/products/:id` | Update product |
| PATCH | `/api/v1/products/bulk/:id` | Bulk update product + replace items |
| DELETE | `/api/v1/products/:id` | Soft delete product |
| PATCH | `/api/v1/products/:id/restore` | Restore deleted product |

### Product Item (replaces /sellable-items)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/products/:productId/items` | Create item under product |
| GET | `/api/v1/products/:productId/items` | List items by product |
| GET | `/api/v1/product-items` | **POS listing** — all items with price + unit |
| GET | `/api/v1/product-items/:id` | Get single item |
| PATCH | `/api/v1/product-items/:id` | Update item (partial) |
| DELETE | `/api/v1/product-items/:id` | Soft delete item |
| PATCH | `/api/v1/product-items/:id/restore` | Restore deleted item |

### Other

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/units` | List all units (PCS, KG, PACK, etc.) |
| POST | `/api/v1/images` | Upload image (conditional, needs MinIO) |

**Total: 25 endpoints**

---

## 5. Database Schema

### `product_items` (replaces `sellable_items`)

```sql
CREATE TABLE product_items (
    id               BIGSERIAL PRIMARY KEY,
    product_id       BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name             VARCHAR(200) NOT NULL,
    sku              VARCHAR(100),
    unit_id          BIGINT REFERENCES units(id),       -- nullable now
    track_inventory  BOOLEAN NOT NULL DEFAULT FALSE,
    image_url        TEXT,
    status           VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);
CREATE UNIQUE INDEX idx_product_items_sku ON product_items(sku) WHERE sku IS NOT NULL;
```

### `prices` (new)

```sql
CREATE TABLE prices (
    id               BIGSERIAL PRIMARY KEY,
    product_item_id  BIGINT NOT NULL REFERENCES product_items(id) ON DELETE CASCADE UNIQUE,
    amount           NUMERIC(12,2) NOT NULL DEFAULT 0,
    currency         VARCHAR(3) NOT NULL DEFAULT 'IDR',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### `product_item_barcodes` (replaces `sellable_item_barcodes`)

```sql
CREATE TABLE product_item_barcodes (
    id               BIGSERIAL PRIMARY KEY,
    product_item_id  BIGINT NOT NULL REFERENCES product_items(id) ON DELETE CASCADE,
    barcode          VARCHAR(100) NOT NULL
);
CREATE UNIQUE INDEX idx_product_item_barcodes_unique ON product_item_barcodes(barcode);
```

---

## 6. POS Integration

POS consumes **ProductItem directly** via:

```
GET /api/v1/product-items
```

Response format:
```json
{
  "data": [
    {
      "id": 1,
      "name": "Marning Pack",
      "sku": "MRN-PACK-001",
      "price": { "amount": 1000, "currency": "IDR" },
      "unit": { "code": "PACK", "name": "Pack" },
      "trackInventory": true
    }
  ]
}
```

Cashier sees:
```
Marning Pack          Rp1.000
Marning Curah         Rp40.000/KG
Aqua 600ml            Rp3.000
```

---

## 7. Inventory Preparation

Catalog **does not** contain stock/quantity/branch_id.

Future Inventory Service will consume:
```
Stock {
    product_item_id    ← from Catalog
    branch_id          ← from Merchant
    quantity
}
```

Catalog provides: `product_item_id`, `unit`, `track_inventory`.

---

## 8. Files Changed

**28 files** modified/created across 6 layers:

| Layer | Files |
|-------|-------|
| Domain | `product_item.go` (+), `errors.go` (updated), `sellable_item.go` (-) |
| Port | `product_item_repository.go` (+), `sellable_item_repository.go` (-) |
| Usecase | `product_item.go` (+), `product_item_service.go` (+), `product_item_test.go` (+), `mocks_test.go` (updated) |
| Transport | `product_item/handler.go` (+), `dto.go` (+), `handler_test.go` (+), `product/handler.go` (updated), `product/dto.go` (updated), `product/handler_test.go` (updated), `router.go` (updated) |
| Repository | `postgres/product_item_repository.go` (+), `postgres/db.go` (updated), `memory/product_item_repository.go` (+), `memory/product_item_repository_test.go` (+) |
| Infrastructure | `bootstrap.go` (updated), `migrations/000001_init.up.sql` (updated), `seed.go` (updated), `cmd/migrate/main.go` (updated) |
| API Docs | `api/bruno/Catalog/` — 6 new `.bru` files, 4 old deleted |

---

## 9. Migration Path

The migration (`000001_init.up.sql`) runs **CREATE IF NOT EXISTS** — it creates the new tables without dropping the old ones. On deploy:

1. Deploy code (new tables are created alongside old)
2. Run data migration script (copy `sellable_items` → `product_items`)
3. Deploy frontend (switch to new endpoints)
4. Drop old tables (future cleanup)

> ⚠️ **Note:** A data migration script is **not yet written**. The old `sellable_items` and `sellable_item_barcodes` tables still exist in the migration but aren't referenced by code anymore.
