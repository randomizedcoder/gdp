--
-- gdp.ProtobufSingle_kafka.sql
--
DROP TABLE IF EXISTS gdp.ProtobufSingle_kafka;

CREATE TABLE IF NOT EXISTS gdp.ProtobufSingle_kafka (
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
ENGINE = Kafka SETTINGS
  kafka_broker_list = 'redpanda-0:9092',
  kafka_topic_list = 'ProtobufSingle',
  kafka_schema = 'prometheus_protolist.proto:PromRecordCounter',
  kafka_max_rows_per_message = 10000,
  kafka_num_consumers = 1,
  kafka_thread_per_consumer = 0,
  kafka_group_name = 'ProtobufSingle',
  kafka_skip_broken_messages = 1,
  kafka_handle_error_mode = 'stream',
  kafka_format = 'ProtobufSingle';

-- SHOW CREATE TABLE gdp.ProtobufSingle_kafka;
-- SELECT * FROM system.kafka_consumers FORMAT Vertical;
-- DETACH TABLE gdp.ProtobufSingle_kafka;
-- SELECT * FROM gdp.ProtobufSingle_kafka LIMIT 20;

-- https://clickhouse.com/docs/integrations/kafka/kafka-table-engine
-- https://clickhouse.com/docs/engines/table-engines/integrations/kafka#creating-a-table
-- kafka_format last! = https://github.com/ClickHouse/ClickHouse/issues/37895

-- end

