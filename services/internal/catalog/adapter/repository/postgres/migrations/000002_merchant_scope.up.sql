-- Migration v7: Add merchant scope to product_items and fix SKU uniqueness
-- Applied automatically for existing databases where product_items already exists
-- but the merchant_id column is missing.

INSERT INTO schema_migrations (version, dirty) VALUES (7, false)
ON CONFLICT (version) DO NOTHING;

-- Add merchant_id column (nullable initially for backfill)
ALTER TABLE product_items ADD COLUMN IF NOT EXISTS merchant_id BIGINT;

-- Backfill merchant_id from products table
UPDATE product_items
SET merchant_id = (
    SELECT merchant_id FROM products WHERE products.id = product_items.product_id
)
WHERE merchant_id IS NULL;

-- Make merchant_id NOT NULL after backfill
ALTER TABLE product_items ALTER COLUMN merchant_id SET NOT NULL;

-- Drop old global SKU unique index
DROP INDEX IF EXISTS idx_product_items_sku;

-- Create new merchant-scoped unique index
CREATE UNIQUE INDEX IF NOT EXISTS idx_product_items_sku_merchant ON product_items(merchant_id, sku) WHERE sku IS NOT NULL;

-- Add merchant index for product_items
CREATE INDEX IF NOT EXISTS idx_product_items_merchant ON product_items(merchant_id);
