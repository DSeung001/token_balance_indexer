CREATE SCHEMA IF NOT EXISTS indexer;
SET
search_path = indexer, public;

CREATE DOMAIN u64 AS NUMERIC
    CHECK (VALUE >= 0 AND scale(VALUE) = 0);

CREATE TABLE IF NOT EXISTS blocks
(
    hash
    TEXT
    PRIMARY
    KEY,
    height
    BIGINT
    UNIQUE
    NOT
    NULL,
    last_block_hash
    TEXT,
    time
    TIMESTAMPTZ
    NOT
    NULL,
    total_txs
    BIGINT,
    num_txs
    BIGINT
);

CREATE TABLE IF NOT EXISTS transactions
(
    hash
    TEXT
    PRIMARY
    KEY,
    block_height
    BIGINT
    NOT
    NULL
    REFERENCES
    blocks
(
    height
) ON DELETE CASCADE,
    tx_index INT NOT NULL, -- GraphQL index ↔ tx_index mapping
    success BOOLEAN NOT NULL DEFAULT FALSE,
    gas_wanted BIGINT,
    gas_used BIGINT,
    gas_fee JSONB,
    memo TEXT,
    content_raw TEXT,
    messages_json JSONB,
    response_json JSONB,
    UNIQUE
(
    block_height,
    tx_index
)
    );

CREATE TABLE IF NOT EXISTS tx_events
(
    id
    BIGSERIAL
    PRIMARY
    KEY,
    tx_hash
    TEXT
    NOT
    NULL
    REFERENCES
    transactions
(
    hash
) ON DELETE CASCADE,
    event_index INT NOT NULL,
    type TEXT NOT NULL,
    func TEXT,
    pkg_path TEXT,
    UNIQUE
(
    tx_hash,
    event_index
)
    );

CREATE TABLE IF NOT EXISTS tx_event_attrs
(
    id
    BIGSERIAL
    PRIMARY
    KEY,
    event_id
    BIGINT
    NOT
    NULL
    REFERENCES
    tx_events
(
    id
) ON DELETE CASCADE,
    attr_index INT NOT NULL,
    key TEXT NOT NULL,
    value TEXT,
    UNIQUE
(
    event_id,
    attr_index
)
    );

CREATE TABLE IF NOT EXISTS tokens
(
    token_path
    TEXT
    PRIMARY
    KEY,
    symbol
    TEXT,
    decimals
    INT,
    created_at
    TIMESTAMPTZ
    DEFAULT
    now
(
)
    );

CREATE TABLE IF NOT EXISTS transfers
(
    id
    BIGSERIAL
    PRIMARY
    KEY,
    tx_hash
    TEXT
    NOT
    NULL
    REFERENCES
    transactions
(
    hash
) ON DELETE CASCADE,
    event_index INT NOT NULL,
    token_path TEXT NOT NULL REFERENCES tokens
(
    token_path
)
  ON DELETE RESTRICT,
    from_address TEXT,
    to_address TEXT,
    amount u64 NOT NULL,
    block_height BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now
(
),
    UNIQUE
(
    tx_hash,
    event_index
)
    );

CREATE TABLE IF NOT EXISTS balances
(
    address
    TEXT
    NOT
    NULL,
    token_path
    TEXT
    NOT
    NULL
    REFERENCES
    tokens
(
    token_path
) ON DELETE RESTRICT,
    amount u64 NOT NULL,
    last_tx_hash TEXT,
    last_block_h BIGINT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now
(
),
    PRIMARY KEY
(
    address,
    token_path
)
    );

CREATE TABLE IF NOT EXISTS app_state
(
    component
    TEXT
    PRIMARY
    KEY, -- 'block_sync' | 'event_consumer'
    last_block_h
    BIGINT,
    last_tx_hash
    TEXT,
    updated_at
    TIMESTAMPTZ
    NOT
    NULL
    DEFAULT
    now
(
)
    );

SET
search_path = indexer, public;

CREATE INDEX IF NOT EXISTS idx_blocks_time ON blocks(time DESC);
CREATE INDEX IF NOT EXISTS idx_blocks_height ON blocks(height DESC);
CREATE INDEX IF NOT EXISTS idx_txs_block_index ON transactions (block_height ASC, tx_index ASC);
CREATE INDEX IF NOT EXISTS idx_txs_block_height ON transactions(block_height DESC);
CREATE INDEX IF NOT EXISTS idx_txs_success ON transactions(success);
CREATE INDEX IF NOT EXISTS idx_events_type ON tx_events(type);
CREATE INDEX IF NOT EXISTS idx_events_func ON tx_events(func);
CREATE INDEX IF NOT EXISTS idx_events_pkg_path ON tx_events(pkg_path);
CREATE INDEX IF NOT EXISTS idx_events_txhash ON tx_events(tx_hash);

CREATE INDEX IF NOT EXISTS idx_transfers_block_height ON transfers(block_height DESC);
CREATE INDEX IF NOT EXISTS idx_transfers_token ON transfers(token_path);
CREATE INDEX IF NOT EXISTS idx_transfers_from ON transfers(from_address);
CREATE INDEX IF NOT EXISTS idx_transfers_to ON transfers(to_address);
CREATE INDEX IF NOT EXISTS idx_transfers_token_from ON transfers(token_path, from_address);
CREATE INDEX IF NOT EXISTS idx_transfers_token_to ON transfers(token_path, to_address);

CREATE INDEX IF NOT EXISTS idx_balances_token ON balances(token_path);
CREATE INDEX IF NOT EXISTS idx_balances_address ON balances(address);