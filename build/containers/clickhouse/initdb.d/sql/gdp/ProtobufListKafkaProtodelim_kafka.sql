--
-- gdp.ProtobufListKafkaProtodelim_kafka.sql
--
DROP TABLE IF EXISTS gdp.ProtobufListKafkaProtodelim_kafka;

CREATE TABLE IF NOT EXISTS gdp.ProtobufListKafkaProtodelim_kafka (
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
  kafka_topic_list = 'ProtobufListKafkaProtodelim',
  kafka_group_name = 'ProtobufListKafkaProtodelim',
  kafka_schema = 'prometheus_protolist.proto:PromRecordCounter',
  kafka_handle_error_mode = 'stream',
  kafka_poll_max_batch_size = 2048,
  kafka_format = 'ProtobufList';

-- SHOW CREATE TABLE gdp.ProtobufListKafkaProtodelim_kafka;
-- SELECT * FROM system.kafka_consumers FORMAT Vertical;
-- DETACH TABLE gdp.ProtobufListKafkaProtodelim_kafka;
-- SELECT * FROM gdp.ProtobufListKafkaProtodelim_kafka LIMIT 20;

-- kafka_handle_error_mode — How to handle errors for Kafka engine. Possible values: default
-- (the exception will be thrown if we fail to parse a message), stream (the exception message
-- and raw message will be saved in virtual columns _error and _raw_message).

-- https://clickhouse.com/docs/integrations/kafka/kafka-table-engine
-- https://clickhouse.com/docs/engines/table-engines/integrations/kafka#creating-a-table
-- https://github.com/confluentinc/librdkafka/blob/master/CONFIGURATION.md
-- kafka_format last! = https://github.com/ClickHouse/ClickHouse/issues/37895

-- end

