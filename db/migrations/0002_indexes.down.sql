SET search_path = indexer, public;

DROP INDEX IF EXISTS idx_balances_address;
DROP INDEX IF EXISTS idx_balances_token;
DROP INDEX IF EXISTS idx_transfers_token_to;
DROP INDEX IF EXISTS idx_transfers_token_from;
DROP INDEX IF EXISTS idx_transfers_to;
DROP INDEX IF EXISTS idx_transfers_from;
DROP INDEX IF EXISTS idx_transfers_token;
DROP INDEX IF EXISTS idx_transfers_block_height;
DROP INDEX IF EXISTS idx_events_txhash;
DROP INDEX IF EXISTS idx_events_pkg_path;
DROP INDEX IF EXISTS idx_events_func;
DROP INDEX IF EXISTS idx_events_type;
DROP INDEX IF EXISTS idx_txs_success;
DROP INDEX IF EXISTS idx_txs_block_height;
DROP INDEX IF EXISTS idx_blocks_height;
DROP INDEX IF EXISTS idx_blocks_time;
