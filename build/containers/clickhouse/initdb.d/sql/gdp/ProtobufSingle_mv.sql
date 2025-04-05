--
-- gdp.ProtobufSingle_mv.sql
--
DROP VIEW IF EXISTS gdp.ProtobufSingle_mv;

CREATE MATERIALIZED VIEW gdp.ProtobufSingle_mv TO gdp.ProtobufSingle
  AS SELECT *
  FROM gdp.ProtobufSingle_kafka;

--  WHERE length(_error) == 0;

-- SHOW CREATE TABLE gdp.ProtobufSingle_mv;

-- https://clickhouse.com/docs/sql-reference/statements/create/view#materialized-view

-- See also:
-- https://github.com/ClickHouse/ClickHouse/blob/master/tests/integration/test_storage_kafka/test_batch_fast.py#L2679
-- https://github.com/ClickHouse/ClickHouse/blob/master/tests/integration/test_storage_kafka/test_batch_slow.py

-- end

