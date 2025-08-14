CREATE SCHEMA IF NOT EXISTS indexer;
SET search_path = indexer, public;

CREATE DOMAIN u64 AS NUMERIC
    CHECK (VALUE >= 0 AND scale(VALUE) = 0);

CREATE TABLE IF NOT EXISTS blocks (
    hash TEXT PRIMARY KEY,
    height BIGINT UNIQUE NOT NULL,
    prev_hash TEXT,
    time TIMESTAMPTZ NOT NULL,
    total_txs BIGINT,
    num_txs BIGINT
);

CREATE TABLE IF NOT EXISTS transactions (
    hash TEXT PRIMARY KEY,
    block_height BIGINT NOT NULL REFERENCES blocks(height) ON DELETE CASCADE,
    index_in_block INT NOT NULL,
    success BOOLEAN,
    gas_wanted BIGINT,
    gas_used BIGINT,
    memo TEXT,
    gas_fee JSONB,
    messages_json JSONB,
    response_json JSONB,
    UNIQUE (block_height, index_in_block)
);

CREATE TABLE IF NOT EXISTS tx_events (
    id BIGSERIAL PRIMARY KEY,
    tx_hash TEXT NOT NULL REFERENCES transactions(hash) ON DELETE CASCADE,
    event_index INT NOT NULL,
    type TEXT NOT NULL,
    func TEXT,
    pkg_path TEXT,
    UNIQUE (tx_hash, event_index)
);

CREATE TABLE IF NOT EXISTS tx_event_attrs (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES tx_events(id) ON DELETE CASCADE,
    attr_index INT NOT NULL,
    key TEXT NOT NULL,
    value TEXT,
    UNIQUE (event_id, attr_index)
);

CREATE TABLE IF NOT EXISTS tokens (
    token_path TEXT PRIMARY KEY,
    symbol TEXT,
    decimals INT,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS transfers (
    id BIGSERIAL PRIMARY KEY,
    tx_hash TEXT NOT NULL REFERENCES transactions(hash) ON DELETE CASCADE,
    event_index INT NOT NULL,
    token_path TEXT NOT NULL REFERENCES tokens(token_path) ON DELETE RESTRICT,
    from_address TEXT,
    to_address TEXT,
    amount u64 NOT NULL,
    block_height BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tx_hash, event_index)
);

CREATE TABLE IF NOT EXISTS balances (
    address TEXT NOT NULL,
    token_path TEXT NOT NULL REFERENCES tokens(token_path) ON DELETE RESTRICT,
    amount u64 NOT NULL,
    last_tx_hash TEXT,
    last_block_h BIGINT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (address, token_path)
);

CREATE TABLE IF NOT EXISTS app_state (
    component TEXT PRIMARY KEY,   -- 'block_sync' | 'event_consumer'
    last_block_h BIGINT,
    last_tx_hash TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
