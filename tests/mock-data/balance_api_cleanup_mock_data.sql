-- API 테스트 데이터 정리 스크립트

-- 1. 전송 내역 데이터 삭제 (가장 최근에 추가된 데이터부터)
DELETE FROM indexer.transfers WHERE tx_hash IN (
                                                'tx_hash_001', 'tx_hash_002', 'tx_hash_003', 'tx_hash_004', 'tx_hash_005', 'tx_hash_006',
                                                'tx_hash_007', 'tx_hash_008', 'tx_hash_009', 'tx_hash_010', 'tx_hash_011', 'tx_hash_012'
    );

-- 2. 잔액 데이터 삭제
DELETE FROM indexer.balances WHERE (address, token_path) IN (
                                                             ('g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5', 'gno.land/r/demo/wugnot'),
                                                             ('g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae', 'gno.land/r/demo/wugnot'),
                                                             ('g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d', 'gno.land/r/demo/wugnot'),
                                                             ('g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5', 'gno.land/r/gnoswap/v1/gns'),
                                                             ('g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae', 'gno.land/r/gnoswap/v1/gns'),
                                                             ('g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d', 'gno.land/r/gnoswap/v1/gns'),
                                                             ('g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5', 'gno.land/r/gnoswap/v1/test_token/bar'),
                                                             ('g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae', 'gno.land/r/gnoswap/v1/test_token/bar'),
                                                             ('g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d', 'gno.land/r/gnoswap/v1/test_token/foo')
    );

-- 3. 토큰 데이터 삭제
DELETE FROM indexer.tokens WHERE token_path IN (
                                                'gno.land/r/demo/wugnot',
                                                'gno.land/r/gnoswap/v1/gns',
                                                'gno.land/r/gnoswap/v1/test_token/bar',
                                                'gno.land/r/gnoswap/v1/test_token/foo'
    );

-- 4. 트랜잭션 데이터 삭제
DELETE FROM indexer.transactions WHERE hash IN (
                                                'tx_hash_001', 'tx_hash_002', 'tx_hash_003', 'tx_hash_004', 'tx_hash_005', 'tx_hash_006',
                                                'tx_hash_007', 'tx_hash_008', 'tx_hash_009', 'tx_hash_010', 'tx_hash_011', 'tx_hash_012'
    );

-- 5. 블록 데이터 삭제
DELETE FROM indexer.blocks WHERE hash IN (
                                          'block_hash_1000', 'block_hash_1001', 'block_hash_1002', 'block_hash_1003',
                                          'block_hash_1004', 'block_hash_1005', 'block_hash_1006'
    );

-- 6. 정리 결과 확인
SELECT 'blocks' as table_name, COUNT(*) as count FROM indexer.blocks
UNION ALL
SELECT 'transactions' as table_name, COUNT(*) as count FROM indexer.transactions
UNION ALL
SELECT 'tokens' as table_name, COUNT(*) as count FROM indexer.tokens
UNION ALL
SELECT 'balances' as table_name, COUNT(*) as count FROM indexer.balances
UNION ALL
SELECT 'transfers' as table_name, COUNT(*) as count FROM indexer.transfers;

-- 7. 정리 완료 메시지
SELECT 'Test data cleanup completed successfully' as message;