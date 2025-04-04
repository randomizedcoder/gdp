--
-- gdp.ProtobufSingleKafkaProtodelim_mv.sql
--
DROP VIEW IF EXISTS gdp.ProtobufSingleKafkaProtodelim_mv;

CREATE MATERIALIZED VIEW gdp.ProtobufSingleKafkaProtodelim_mv TO gdp.ProtobufSingleKafkaProtodelim
  AS SELECT *
  FROM gdp.ProtobufSingleKafkaProtodelim_kafka
  WHERE length(_error) == 0;

-- SHOW CREATE TABLE gdp.ProtobufSingleKafkaProtodelim_mv;

-- https://clickhouse.com/docs/sql-reference/statements/create/view#materialized-view

-- See also:
-- https://github.com/ClickHouse/ClickHouse/blob/master/tests/integration/test_storage_kafka/test_batch_fast.py#L2679
-- https://github.com/ClickHouse/ClickHouse/blob/master/tests/integration/test_storage_kafka/test_batch_slow.py

-- end

