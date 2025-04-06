--
-- gdp.ProtobufSingle_mv.sql
--
DROP VIEW IF EXISTS gdp.ProtobufSingle_mv;

CREATE MATERIALIZED VIEW gdp.ProtobufSingle_mv TO gdp.ProtobufSingle
  AS SELECT
    *,
--    toDateTime64(Timestamp_Ns, 9, 'UTC') AS Timestamp_Ns,
--    Hostname,
--    Pop,
--    Label,
--    Tag,
--    Poll_Counter,
--    Record_Counter,
--    Function,
--    Variable,
--    Type,
--    Value,
  FROM gdp.ProtobufSingle_kafka;

-- SHOW CREATE TABLE gdp.ProtobufSingle_mv;

-- https://clickhouse.com/docs/sql-reference/statements/create/view#materialized-view

-- See also:
-- https://github.com/ClickHouse/ClickHouse/blob/master/tests/integration/test_storage_kafka/test_batch_fast.py#L2679
-- https://github.com/ClickHouse/ClickHouse/blob/master/tests/integration/test_storage_kafka/test_batch_slow.py

-- end

