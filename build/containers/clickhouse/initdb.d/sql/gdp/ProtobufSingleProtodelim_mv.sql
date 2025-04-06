--
-- gdp.ProtobufSingleProtodelim_mv.sql
--
DROP VIEW IF EXISTS gdp.ProtobufSingleProtodelim_mv;

CREATE MATERIALIZED VIEW gdp.ProtobufSingleProtodelim_mv TO gdp.ProtobufSingleProtodelim
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
  FROM gdp.ProtobufSingleProtodelim_kafka;

-- SHOW CREATE TABLE gdp.ProtobufSingleProtodelim_mv;

-- https://clickhouse.com/docs/sql-reference/statements/create/view#materialized-view

-- See also:
-- https://github.com/ClickHouse/ClickHouse/blob/master/tests/integration/test_storage_kafka/test_batch_fast.py#L2679
-- https://github.com/ClickHouse/ClickHouse/blob/master/tests/integration/test_storage_kafka/test_batch_slow.py

-- end

