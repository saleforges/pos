CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    dirty   BOOLEAN NOT NULL
);

INSERT INTO schema_migrations (version, dirty) VALUES (1, false);
INSERT INTO schema_migrations (version, dirty) VALUES (2, false);
INSERT INTO schema_migrations (version, dirty) VALUES (3, false);
INSERT INTO schema_migrations (version, dirty) VALUES (4, false);
INSERT INTO schema_migrations (version, dirty) VALUES (5, false);
INSERT INTO schema_migrations (version, dirty) VALUES (6, false);
INSERT INTO schema_migrations (version, dirty) VALUES (7, false);

CREATE TABLE IF NOT EXISTS units (
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
    ('METER', 'Meter')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS categories (
	id          BIGSERIAL    PRIMARY KEY,
	merchant_id BIGINT       NOT NULL,
	name        VARCHAR(100) NOT NULL,
	parent_id   BIGINT       REFERENCES categories(id),
	created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
	updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
	deleted_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_categories_merchant ON categories(merchant_id);

CREATE TABLE IF NOT EXISTS products (
	id          BIGSERIAL    PRIMARY KEY,
	merchant_id BIGINT       NOT NULL,
	category_id BIGINT       NOT NULL REFERENCES categories(id),
	name        VARCHAR(200) NOT NULL,
	description TEXT,
	image_url   TEXT,
	status      VARCHAR(20)  NOT NULL DEFAULT 'active',
	created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
	updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
	deleted_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_products_merchant ON products(merchant_id);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);

-- Product item (replaces sellable_items)
CREATE TABLE IF NOT EXISTS product_items (
	id               BIGSERIAL    PRIMARY KEY,
	product_id       BIGINT       NOT NULL REFERENCES products(id) ON DELETE CASCADE,
	merchant_id      BIGINT       NOT NULL,
	name             VARCHAR(200) NOT NULL,
	sku              VARCHAR(100),
	unit_id          BIGINT       REFERENCES units(id),
	track_inventory  BOOLEAN      NOT NULL DEFAULT FALSE,
	image_url        TEXT,
	status           VARCHAR(20)  NOT NULL DEFAULT 'active',
	created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
	updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
	deleted_at       TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_product_items_product ON product_items(product_id);
CREATE INDEX IF NOT EXISTS idx_product_items_merchant ON product_items(merchant_id);
-- SKU uniqueness within merchant scope (nullable, so partial index)
CREATE UNIQUE INDEX IF NOT EXISTS idx_product_items_sku_merchant ON product_items(merchant_id, sku) WHERE sku IS NOT NULL;

-- Price (separate table, one active price per product item)
CREATE TABLE IF NOT EXISTS prices (
    id               BIGSERIAL    PRIMARY KEY,
    product_item_id  BIGINT       NOT NULL REFERENCES product_items(id) ON DELETE CASCADE UNIQUE,
    amount           NUMERIC(12,2) NOT NULL DEFAULT 0,
    currency         VARCHAR(3)   NOT NULL DEFAULT 'IDR',
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Product item barcodes (separate table for multiple barcodes per item)
CREATE TABLE IF NOT EXISTS product_item_barcodes (
    id               BIGSERIAL    PRIMARY KEY,
    product_item_id  BIGINT       NOT NULL REFERENCES product_items(id) ON DELETE CASCADE,
    barcode          VARCHAR(100) NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_product_item_barcodes_unique ON product_item_barcodes(barcode);
CREATE INDEX IF NOT EXISTS idx_product_item_barcodes_item ON product_item_barcodes(product_item_id);
