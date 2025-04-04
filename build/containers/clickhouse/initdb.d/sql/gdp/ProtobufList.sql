--
-- gdp.ProtobufList.sql
--
DROP TABLE IF EXISTS gdp.ProtobufList;

CREATE TABLE IF NOT EXISTS gdp.ProtobufList (
  TimestampNs DateTime64(9,'UTC') CODEC(DoubleDelta, LZ4),
  Hostname LowCardinality(String) CODEC(LZ4),
  Pop LowCardinality(String) CODEC(LZ4),
  Label LowCardinality(String) CODEC(LZ4),
  Tag LowCardinality(String) CODEC(LZ4),
  PollCounter UInt64 CODEC(DoubleDelta, LZ4),
  RecordCounter UInt64 CODEC(DoubleDelta, LZ4),
  Function LowCardinality(String) CODEC(LZ4),
  Variable LowCardinality(String) CODEC(LZ4),
  Type LowCardinality(String) CODEC(LZ4),
  Value Float64,
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(TimestampNs)
ORDER BY (TimestampNs, Hostname, Pop, Label, Tag, PollCounter, RecordCounter)
TTL toDateTime(TimestampNs) + INTERVAL 14 DAY;

-- Note that ORDER BY clause implicitly specifies a primary key

-- SHOW CREATE TABLE gdp.ProtobufList;
-- SELECT * FROM gdp.ProtobufList LIMIT 20;

-- https://clickhouse.com/docs/guides/developer/ttl
-- https://clickhouse.com/docs/sql-reference/statements/alter/ttl
-- https://clickhouse.com/docs/engines/table-engines/mergetree-family/mergetree#table_engine-mergetree-ttl
-- https://clickhouse.com/docs/sql-reference/functions/type-conversion-functions#todatetime

-- end

