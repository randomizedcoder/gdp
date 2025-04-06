--
-- gdp.ProtobufListKafkaProtodelim_insert_rows.sql
--
INSERT INTO gdp.ProtobufListKafkaProtodelim (
    Timestamp_Ns,
    Hostname,
    Pop,
    Label,
    Tag,
    PollCounter,
    RecordCounter,
    Function,
    Variable,
    Type,
    Value
) VALUES
(
    toDateTime64('2025-04-05 10:00:00.0', 9, 'UTC'),
    'host1',
    'pop1',
    'label1',
    'tag1',
    100,
    1,
    'function1',
    'variable1',
    'type1',
    123.45
),
(
    toDateTime64('2025-04-05 11:00:00.0', 9, 'UTC'),
    'host2',
    'pop2',
    'label2',
    'tag2',
    200,
    2,
    'function2',
    'variable2',
    'type2',
    678.90
);

-- SELECT * FROM gdp.ProtobufListProtodelim LIMIT 10;
-- TRUNCATE TABLE gdp.ProtobufListProtodelim;

-- end

