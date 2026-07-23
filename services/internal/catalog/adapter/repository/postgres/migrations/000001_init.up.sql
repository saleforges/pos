CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    dirty   BOOLEAN NOT NULL
);

INSERT INTO schema_migrations (version, dirty) VALUES (1, false);
INSERT INTO schema_migrations (version, dirty) VALUES (2, false);
INSERT INTO schema_migrations (version, dirty) VALUES (3, false);
INSERT INTO schema_migrations (version, dirty) VALUES (4, false);
INSERT INTO schema_migrations (version, dirty) VALUES (5, false);

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

CREATE TABLE IF NOT EXISTS sellable_items (
	id               BIGSERIAL    PRIMARY KEY,
	product_id       BIGINT       NOT NULL REFERENCES products(id) ON DELETE CASCADE,
	name             VARCHAR(200) NOT NULL,
	unit_id          BIGINT       NOT NULL REFERENCES units(id),
	price            NUMERIC(12,2) NOT NULL DEFAULT 0,
	track_inventory  BOOLEAN      NOT NULL DEFAULT TRUE,
	image_url        TEXT,
	status           VARCHAR(20)  NOT NULL DEFAULT 'active',
	created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
	updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
	deleted_at       TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_sellable_items_product ON sellable_items(product_id);

CREATE TABLE IF NOT EXISTS sellable_item_barcodes (
    id               BIGSERIAL    PRIMARY KEY,
    sellable_item_id BIGINT       NOT NULL REFERENCES sellable_items(id) ON DELETE CASCADE,
    barcode          VARCHAR(100) NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_barcodes_unique ON sellable_item_barcodes(barcode);
CREATE INDEX IF NOT EXISTS idx_barcodes_item ON sellable_item_barcodes(sellable_item_id);
