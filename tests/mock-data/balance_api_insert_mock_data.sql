-- API 테스트를 위한 목 데이터 삽입 스크립트

-- 1. 블록 데이터 삽입 (transfers 테이블의 외래키 제약조건을 위해)
INSERT INTO indexer.blocks (hash, height, last_block_hash, time, total_txs, num_txs) VALUES
                                                                                         ('block_hash_1000', 1000, 'block_hash_999', NOW() - INTERVAL '1 day', 1000, 3),
                                                                                         ('block_hash_1001', 1001, 'block_hash_1000', NOW() - INTERVAL '12 hours', 1001, 3),
                                                                                         ('block_hash_1002', 1002, 'block_hash_1001', NOW() - INTERVAL '6 hours', 1002, 2),
                                                                                         ('block_hash_1003', 1003, 'block_hash_1002', NOW() - INTERVAL '3 hours', 1003, 1),
                                                                                         ('block_hash_1004', 1004, 'block_hash_1003', NOW() - INTERVAL '2 hours', 1004, 1),
                                                                                         ('block_hash_1005', 1005, 'block_hash_1004', NOW() - INTERVAL '1 hour', 1005, 1),
                                                                                         ('block_hash_1006', 1006, 'block_hash_1005', NOW() - INTERVAL '30 minutes', 1006, 1)
    ON CONFLICT (hash) DO NOTHING;

-- 2. 트랜잭션 데이터 삽입 (transfers 테이블의 외래키 제약조건을 위해)
INSERT INTO indexer.transactions (hash, block_height, tx_index, success, gas_wanted, gas_used, gas_fee, memo) VALUES
                                                                                                                  ('tx_hash_001', 1000, 0, true, 100000, 50000, '{"amount": "1000", "denom": "ugnot"}', 'Mint WUGNOT'),
                                                                                                                  ('tx_hash_002', 1000, 1, true, 100000, 50000, '{"amount": "1000", "denom": "ugnot"}', 'Mint WUGNOT'),
                                                                                                                  ('tx_hash_003', 1000, 2, true, 100000, 50000, '{"amount": "1000", "denom": "ugnot"}', 'Mint WUGNOT'),
                                                                                                                  ('tx_hash_004', 1001, 0, true, 100000, 50000, '{"amount": "1000", "denom": "ugnot"}', 'Mint GNS'),
                                                                                                                  ('tx_hash_005', 1001, 1, true, 100000, 50000, '{"amount": "1000", "denom": "ugnot"}', 'Mint GNS'),
                                                                                                                  ('tx_hash_006', 1001, 2, true, 100000, 50000, '{"amount": "1000", "denom": "ugnot"}', 'Mint GNS'),
                                                                                                                  ('tx_hash_007', 1002, 0, true, 100000, 50000, '{"amount": "1000", "denom": "ugnot"}', 'Mint BAR'),
                                                                                                                  ('tx_hash_008', 1002, 1, true, 100000, 50000, '{"amount": "1000", "denom": "ugnot"}', 'Mint BAR'),
                                                                                                                  ('tx_hash_009', 1003, 0, true, 100000, 50000, '{"amount": "1000", "denom": "ugnot"}', 'Mint FOO'),
                                                                                                                  ('tx_hash_010', 1004, 0, true, 100000, 50000, '{"amount": "1000", "denom": "ugnot"}', 'Transfer WUGNOT'),
                                                                                                                  ('tx_hash_011', 1005, 0, true, 100000, 50000, '{"amount": "1000", "denom": "ugnot"}', 'Transfer GNS'),
                                                                                                                  ('tx_hash_012', 1006, 0, true, 100000, 50000, '{"amount": "1000", "denom": "ugnot"}', 'Transfer BAR')
    ON CONFLICT (hash) DO NOTHING;

-- 3. 토큰 데이터 삽입
INSERT INTO indexer.tokens (token_path, symbol, decimals) VALUES
                                                              ('gno.land/r/demo/wugnot', 'WUGNOT', 6),
                                                              ('gno.land/r/gnoswap/v1/gns', 'GNS', 6),
                                                              ('gno.land/r/gnoswap/v1/test_token/bar', 'BAR', 6),
                                                              ('gno.land/r/gnoswap/v1/test_token/foo', 'FOO', 6)
    ON CONFLICT (token_path) DO NOTHING;

-- 4. 잔액 데이터 삽입
INSERT INTO indexer.balances (address, token_path, amount, last_tx_hash, last_block_h) VALUES
-- WUGNOT 토큰 잔액
('g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5', 'gno.land/r/demo/wugnot', 500000, 'tx_hash_001', 1000),
('g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae', 'gno.land/r/demo/wugnot', 500000, 'tx_hash_002', 1000),
('g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d', 'gno.land/r/demo/wugnot', 1000000, 'tx_hash_003', 1000),

-- GNS 토큰 잔액
('g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5', 'gno.land/r/gnoswap/v1/gns', 1000000, 'tx_hash_004', 1001),
('g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae', 'gno.land/r/gnoswap/v1/gns', 2000000, 'tx_hash_005', 1001),
('g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d', 'gno.land/r/gnoswap/v1/gns', 1500000, 'tx_hash_006', 1001),

-- BAR 토큰 잔액
('g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5', 'gno.land/r/gnoswap/v1/test_token/bar', 300000, 'tx_hash_007', 1002),
('g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae', 'gno.land/r/gnoswap/v1/test_token/bar', 700000, 'tx_hash_008', 1002),

-- FOO 토큰 잔액
('g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d', 'gno.land/r/gnoswap/v1/test_token/foo', 2500000, 'tx_hash_009', 1003)
    ON CONFLICT (address, token_path) DO UPDATE SET
    amount = EXCLUDED.amount,
                                             last_tx_hash = EXCLUDED.last_tx_hash,
                                             last_block_h = EXCLUDED.last_block_h;

-- 5. 전송 내역 데이터 삽입
INSERT INTO indexer.transfers (tx_hash, event_index, token_path, from_address, to_address, amount, block_height, created_at) VALUES
-- WUGNOT 전송 내역
('tx_hash_001', 0, 'gno.land/r/demo/wugnot', '', 'g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5', 500000, 1000, NOW() - INTERVAL '1 day'),
('tx_hash_002', 0, 'gno.land/r/demo/wugnot', '', 'g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae', 500000, 1000, NOW() - INTERVAL '1 day'),
('tx_hash_003', 0, 'gno.land/r/demo/wugnot', '', 'g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d', 1000000, 1000, NOW() - INTERVAL '1 day'),

-- GNS 전송 내역
('tx_hash_004', 0, 'gno.land/r/gnoswap/v1/gns', '', 'g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5', 1000000, 1001, NOW() - INTERVAL '12 hours'),
('tx_hash_005', 0, 'gno.land/r/gnoswap/v1/gns', '', 'g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae', 2000000, 1001, NOW() - INTERVAL '12 hours'),
('tx_hash_006', 0, 'gno.land/r/gnoswap/v1/gns', '', 'g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d', 1500000, 1001, NOW() - INTERVAL '12 hours'),

-- BAR 토큰 전송 내역
('tx_hash_007', 0, 'gno.land/r/gnoswap/v1/test_token/bar', '', 'g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5', 300000, 1002, NOW() - INTERVAL '6 hours'),
('tx_hash_008', 0, 'gno.land/r/gnoswap/v1/test_token/bar', '', 'g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae', 700000, 1002, NOW() - INTERVAL '6 hours'),

-- FOO 토큰 전송 내역
('tx_hash_009', 0, 'gno.land/r/gnoswap/v1/test_token/foo', '', 'g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d', 2500000, 1003, NOW() - INTERVAL '3 hours'),

-- 실제 전송 이벤트들 (Mint가 아닌 Transfer)
('tx_hash_010', 0, 'gno.land/r/demo/wugnot', 'g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5', 'g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae', 100000, 1004, NOW() - INTERVAL '2 hours'),
('tx_hash_011', 0, 'gno.land/r/gnoswap/v1/gns', 'g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae', 'g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d', 500000, 1005, NOW() - INTERVAL '1 hour'),
('tx_hash_012', 0, 'gno.land/r/gnoswap/v1/test_token/bar', 'g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5', 'g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d', 150000, 1006, NOW() - INTERVAL '30 minutes');

-- 데이터 삽입 확인
SELECT 'blocks' as table_name, COUNT(*) as count FROM indexer.blocks
UNION ALL
SELECT 'transactions' as table_name, COUNT(*) as count FROM indexer.transactions
UNION ALL
SELECT 'tokens' as table_name, COUNT(*) as count FROM indexer.tokens
UNION ALL
SELECT 'balances' as table_name, COUNT(*) as count FROM indexer.balances
UNION ALL
SELECT 'transfers' as table_name, COUNT(*) as count FROM indexer.transfers;