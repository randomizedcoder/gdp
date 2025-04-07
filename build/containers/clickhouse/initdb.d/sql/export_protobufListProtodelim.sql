SELECT
  *
FROM gdp.ProtobufListProtodelim
LIMIT 2
INTO OUTFILE 'gdp.ProtobufListProtodelim.bin'
FORMAT ProtobufList
SETTINGS format_schema = '/var/lib/clickhouse/format_schemas/prometheus_protolist.proto:PromRecordCounter'
