-- 000001_init: schema order service senior
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE orders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  status TEXT NOT NULL CHECK (status IN ('pending','reserved','confirmed','cancelled','failed')),
  idempotency_key TEXT NOT NULL UNIQUE,
  total_cents INT NOT NULL CHECK (total_cents >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_orders_status_created ON orders(status, created_at);

CREATE TABLE order_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id TEXT NOT NULL,
  qty INT NOT NULL CHECK (qty > 0),
  price_cents INT NOT NULL CHECK (price_cents >= 0),
  UNIQUE(order_id, product_id)
);
CREATE INDEX idx_order_items_order ON order_items(order_id);

CREATE TABLE inventory (
  product_id TEXT PRIMARY KEY,
  stock INT NOT NULL CHECK (stock >= 0),
  version INT NOT NULL DEFAULT 1,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE payments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id UUID NOT NULL UNIQUE REFERENCES orders(id),
  status TEXT NOT NULL CHECK (status IN ('pending','succeeded','failed')),
  amount_cents INT NOT NULL CHECK (amount_cents >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE outbox (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  aggregate_id UUID NOT NULL,
  event_type TEXT NOT NULL,
  payload JSONB NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('pending','published','failed')) DEFAULT 'pending',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  published_at TIMESTAMPTZ
);
CREATE INDEX idx_outbox_status_created ON outbox(status, created_at);

CREATE TABLE idempotency_keys (
  key TEXT PRIMARY KEY,
  response_code INT NOT NULL,
  response_body JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL DEFAULT now() + INTERVAL '24 hours'
);
CREATE INDEX idx_idempotency_expires ON idempotency_keys(expires_at);

CREATE TABLE order_events (
  id BIGSERIAL PRIMARY KEY,
  order_id UUID NOT NULL REFERENCES orders(id),
  from_status TEXT,
  to_status TEXT NOT NULL,
  payload JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_order_events_order ON order_events(order_id, created_at);

-- seed inventory
INSERT INTO inventory(product_id, stock) VALUES ('PROD-001', 100), ('PROD-002', 50), ('PROD-003', 200) ON CONFLICT DO NOTHING;
