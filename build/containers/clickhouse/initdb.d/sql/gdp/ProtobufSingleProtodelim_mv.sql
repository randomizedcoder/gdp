--
-- gdp.ProtobufSingleProtodelim_mv.sql
--
DROP VIEW IF EXISTS gdp.ProtobufSingleProtodelim_mv;

CREATE MATERIALIZED VIEW gdp.ProtobufSingleProtodelim_mv TO gdp.ProtobufSingleProtodelim
  AS SELECT *
  FROM gdp.ProtobufSingleProtodelim_kafka
  WHERE length(_error) == 0;

-- SHOW CREATE TABLE gdp.ProtobufSingleProtodelim_mv;

-- https://clickhouse.com/docs/sql-reference/statements/create/view#materialized-view

-- See also:
-- https://github.com/ClickHouse/ClickHouse/blob/master/tests/integration/test_storage_kafka/test_batch_fast.py#L2679
-- https://github.com/ClickHouse/ClickHouse/blob/master/tests/integration/test_storage_kafka/test_batch_slow.py

-- end

