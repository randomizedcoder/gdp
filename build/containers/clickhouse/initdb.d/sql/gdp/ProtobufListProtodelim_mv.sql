--
-- gdp.ProtobufListProtodelim_mv.sql
--
DROP VIEW IF EXISTS gdp.ProtobufListProtodelim_mv;

CREATE MATERIALIZED VIEW gdp.ProtobufListProtodelim_mv TO gdp.ProtobufListProtodelim
  AS SELECT *
  FROM gdp.ProtobufListProtodelim_kafka;

--  WHERE length(_error) == 0;

-- SHOW CREATE TABLE gdp.ProtobufListProtodelim_mv;

-- https://clickhouse.com/docs/sql-reference/statements/create/view#materialized-view

-- See also:
-- https://github.com/ClickHouse/ClickHouse/blob/master/tests/integration/test_storage_kafka/test_batch_fast.py#L2679
-- https://github.com/ClickHouse/ClickHouse/blob/master/tests/integration/test_storage_kafka/test_batch_slow.py

-- end

