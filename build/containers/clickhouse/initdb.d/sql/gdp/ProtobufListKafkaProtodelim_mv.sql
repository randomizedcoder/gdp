--
-- gdp.ProtobufListKafkaProtodelim_mv.sql
--
DROP VIEW IF EXISTS gdp.ProtobufListKafkaProtodelim_mv;

CREATE MATERIALIZED VIEW gdp.ProtobufListKafkaProtodelim_mv TO gdp.ProtobufListKafkaProtodelim
  AS SELECT *
  FROM gdp.ProtobufListKafkaProtodelim_kafka
  WHERE length(_error) == 0;

-- SHOW CREATE TABLE gdp.ProtobufListKafkaProtodelim_mv;

-- https://clickhouse.com/docs/sql-reference/statements/create/view#materialized-view

-- See also:
-- https://github.com/ClickHouse/ClickHouse/blob/master/tests/integration/test_storage_kafka/test_batch_fast.py#L2679
-- https://github.com/ClickHouse/ClickHouse/blob/master/tests/integration/test_storage_kafka/test_batch_slow.py

-- end

