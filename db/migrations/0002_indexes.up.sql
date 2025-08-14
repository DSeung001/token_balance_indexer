SET search_path = indexer, public;

CREATE INDEX IF NOT EXISTS idx_blocks_time ON blocks(time DESC);
CREATE INDEX IF NOT EXISTS idx_blocks_height ON blocks(height DESC);
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
