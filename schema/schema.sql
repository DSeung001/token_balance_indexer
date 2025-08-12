-- schema/schema.sql
CREATE TABLE IF NOT EXISTS hello (
  id SERIAL PRIMARY KEY,
  msg TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
);
