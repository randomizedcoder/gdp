--
-- gdp.ProtobufProtodelim_kafka.sql
--
DROP TABLE IF EXISTS gdp.ProtobufProtodelim_kafka;

CREATE TABLE IF NOT EXISTS gdp.ProtobufProtodelim_kafka (
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
ENGINE = Kafka SETTINGS
  kafka_broker_list = 'redpanda-0:9092',
  kafka_topic_list = 'ProtobufProtodelim',
  kafka_group_name = 'ProtobufProtodelim',
  kafka_schema = 'prometheus_protolist.proto:PromRecordCounter',
  kafka_handle_error_mode = 'stream',
  kafka_poll_max_batch_size = 1024,
  kafka_format = 'Protobuf';

-- SHOW CREATE TABLE gdp.ProtobufProtodelim_kafka;
-- SELECT * FROM system.kafka_consumers FORMAT Vertical;
-- SELECT * FROM system.kafka_consumers WHERE table LIKE '%ProtobufProtodelim_kafka%' FORMAT Vertical;
-- DETACH TABLE gdp.ProtobufProtodelim_kafka;
-- SELECT * FROM gdp.ProtobufProtodelim_kafka LIMIT 20;

-- kafka_handle_error_mode — How to handle errors for Kafka engine. Possible values: default
-- (the exception will be thrown if we fail to parse a message), stream (the exception message
-- and raw message will be saved in virtual columns _error and _raw_message).

-- https://clickhouse.com/docs/integrations/kafka/kafka-table-engine
-- https://clickhouse.com/docs/engines/table-engines/integrations/kafka#creating-a-table
-- https://github.com/confluentinc/librdkafka/blob/master/CONFIGURATION.md
-- kafka_format last! = https://github.com/ClickHouse/ClickHouse/issues/37895

-- end

