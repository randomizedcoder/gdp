--
-- gdp.ProtobufListKafka.sql
--
DROP TABLE IF EXISTS gdp.ProtobufListKafka;

CREATE TABLE IF NOT EXISTS gdp.ProtobufListKafka (
  Timestamp_Ns DateTime64(9,'UTC') CODEC(DoubleDelta, LZ4),
  Hostname LowCardinality(String) CODEC(LZ4),
  Pop LowCardinality(String) CODEC(LZ4),
  Label LowCardinality(String) CODEC(LZ4),
  Tag LowCardinality(String) CODEC(LZ4),
  Poll_Counter UInt64 CODEC(DoubleDelta, LZ4),
  Record_Counter UInt64 CODEC(DoubleDelta, LZ4),
  Function LowCardinality(String) CODEC(LZ4),
  Variable LowCardinality(String) CODEC(LZ4),
  Type LowCardinality(String) CODEC(LZ4),
  Value Float64,
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(Timestamp_Ns)
ORDER BY (Timestamp_Ns, Hostname, Pop, Label, Tag, Poll_Counter, Record_Counter)
TTL toDateTime(Timestamp_Ns) + INTERVAL 14 DAY;

-- Note that ORDER BY clause implicitly specifies a primary key

-- SHOW CREATE TABLE gdp.ProtobufListKafka;
-- SELECT * FROM gdp.ProtobufListKafka LIMIT 20;

-- SELECT
--   count(DISTINCT Poll_Counter)
-- FROM gdp.ProtobufListKafka;

-- https://clickhouse.com/docs/guides/developer/ttl
-- https://clickhouse.com/docs/sql-reference/statements/alter/ttl
-- https://clickhouse.com/docs/engines/table-engines/mergetree-family/mergetree#table_engine-mergetree-ttl
-- https://clickhouse.com/docs/sql-reference/functions/type-conversion-functions#todatetime

-- end

