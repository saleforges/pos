# Catalog Service Refactor — Summary for PO/PM

## What Changed

The Catalog Service was refactored to match the same architecture as IAM and Merchant services (hexagonal architecture). The old catalog had mixed concerns — it contained pricing, inventory fields, and complex product variants that didn't match how UMKM businesses actually work.

## New Domain Model

### Product
A business product (e.g. "Marning", "Aqua 600ml").

**Contains:** name, description, category, image, status
**Does NOT contain:** price, cost, tax, SKU, barcode, stock

### Sellable Item (replaces old "Variant")
What the cashier actually sells to customers. One product can have multiple sellable items.

**Example:**
```
Product: "Marning"
  ├── Sellable Item: "Marning Curah" (per KG) — Rp15.000
  └── Sellable Item: "Marning Pack" (per PACK) — Rp20.000
```

**Contains:** name, unit (KG/PCS/PACK), price, barcode, inventory tracking flag

### Unit
System master data for units of measurement: PCS, PACK, KG, GRAM, LITER, ML, BOX, METER

### Category
Merchant-scoped categories (Food, Snack, Drink). Supports hierarchy (parent category).

### Barcode
Separate entity — one sellable item can have multiple barcodes.

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/units` | List all units |
| POST | `/api/v1/categories` | Create category |
| GET | `/api/v1/categories` | List categories |
| GET | `/api/v1/categories/:id` | Get category |
| PATCH | `/api/v1/categories/:id` | Update category |
| DELETE | `/api/v1/categories/:id` | Delete category (soft) |
| PATCH | `/api/v1/categories/:id/restore` | Restore deleted category |
| POST | `/api/v1/products` | Create product only |
| POST | `/api/v1/products/bulk` | Create product + sellable items (one call) |
| GET | `/api/v1/products` | List products with items + prices + category |
| GET | `/api/v1/products/:id` | Get product detail |
| PATCH | `/api/v1/products/:id` | Update product |
| PATCH | `/api/v1/products/bulk/:id` | Update product + replace items (one call) |
| DELETE | `/api/v1/products/:id` | Delete product (soft) |
| PATCH | `/api/v1/products/:id/restore` | Restore deleted product |
| POST | `/api/v1/products/:productId/items` | Add sellable item to product |
| GET | `/api/v1/products/:productId/items` | List sellable items for a product |
| PATCH | `/api/v1/items/:id` | Update sellable item |
| DELETE | `/api/v1/items/:id` | Delete sellable item (soft) |
| PATCH | `/api/v1/items/:id/restore` | Restore deleted sellable item |
| POST | `/api/v1/images` | Upload product image (MinIO storage) |

**23 endpoints total.**

## What Was Removed from Catalog

- **Price** → moved to Sellable Item level (one product can have multiple prices per unit)
- **Cost, Tax Rate** → removed entirely (belongs to future Order/Pricing service)
- **SKU** → removed from Product (if needed, can be added to Sellable Item later)
- **Branch ID** → removed (Catalog doesn't know about branches — belongs to Merchant service)

## Key Features

- **Soft delete** — products, categories, and sellable items are soft-deleted with `deleted_at`. Can be restored anytime via restore endpoints. Important for UMKM who often make mistakes.
- **Bulk create/update** — one API call to create product + all its sellable items. Saves 10+ API calls when adding a product with multiple selling forms.
- **Merchant isolation** — all data scoped by `merchant_id` from JWT token, no need to send merchant_id in requests.
- **Auth** — JWT verification via shared JWKS endpoint from IAM service.
- **CamelCase JSON** — consistent with IAM and Merchant services.

## Product List Response

```json
{
  "data": [
    {
      "id": 1,
      "name": "Marning",
      "description": "Marning original",
      "imageUrl": "https://minio.../abc.jpg",
      "status": "active",
      "category": {
        "id": 2,
        "name": "Snack"
      },
      "priceRange": {
        "min": 15000,
        "max": 20000
      },
      "items": [
        {
          "id": 1,
          "name": "Marning Curah",
          "unit": { "id": 3, "code": "KG", "name": "Kilogram" },
          "price": 15000,
          "trackInventory": true,
          "status": "active"
        }
      ]
    }
  ],
  "pagination": {
    "total": 50,
    "offset": 0,
    "limit": 20,
    "count": 10
  }
}
```

## What's Still Open (for Discussion)

1. **Pricing** — currently on Sellable Item. If the product has only one sellable item with one price, the UI flow is fine. If pricing needs to be managed by a separate service later (Order/Pricing), it can be moved.
2. **Image upload** — working via MinIO, but needs to be integrated with the product create flow (currently separate: upload image → get URL → create product with URL).
3. **Inventory tracking** — `trackInventory` flag on SellableItem is ready but no actual inventory service exists yet. Future Inventory/Stock service can use this flag.
4. **Barcode scanning** — barcode table exists and API is ready, but no scan endpoint. Can be added when POS hardware integration is needed.
5. **Data migration** — old products table has different columns. Since there's no production data, dropping and recreating tables is the plan.

## Architecture

```
catalog-service/
  domain/           → Business entities (Product, SellableItem, etc.)
  port/             → Interfaces (repository contracts)
  usecase/          → Business logic, validation
  adapter/
    repository/     → Database implementations (Postgres + in-memory for testing)
  transport/http/   → REST API handlers
  bootstrap/        → Dependency injection
```

## Tests

- 29 test cases across all layers
- Repository tests (memory), Usecase tests (with mocks), Handler tests (with Echo test helpers)
- Full project: 45+ test suites, all passing
