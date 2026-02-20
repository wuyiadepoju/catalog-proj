-- Initial schema for catalog service
-- Products table
CREATE TABLE products (
    product_id STRING(36) NOT NULL,
    name STRING(255) NOT NULL,
    description STRING(MAX) NOT NULL,
    category STRING(100) NOT NULL,
    base_price_numerator INT64 NOT NULL,
    base_price_denominator INT64 NOT NULL,
    discount_id STRING(36),
    discount_amount NUMERIC,
    discount_start_date TIMESTAMP,
    discount_end_date TIMESTAMP,
    status STRING(20) NOT NULL,
    archived_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
) PRIMARY KEY (product_id);

-- Index for category queries
CREATE INDEX idx_products_category ON products(category, status);

-- Index for status queries
CREATE INDEX idx_products_status ON products(status);

-- Index for active products (common query pattern)
-- Note: Spanner doesn't support WHERE clauses in indexes, so we index on status and archived_at
-- Filtering for NULL archived_at should be done in queries
CREATE INDEX idx_products_active ON products(status, archived_at);

-- Outbox table for domain events (transactional outbox pattern)
CREATE TABLE outbox_events (
    event_id STRING(36) NOT NULL,
    event_type STRING(100) NOT NULL,
    aggregate_id STRING(36) NOT NULL,
    payload JSON NOT NULL,
    status STRING(20) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    processed_at TIMESTAMP,
) PRIMARY KEY (event_id);

-- Index for unprocessed events (filter by status = 'pending' in queries)
CREATE INDEX idx_outbox_status ON outbox_events(status, created_at);
