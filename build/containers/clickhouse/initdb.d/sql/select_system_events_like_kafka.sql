--
-- See the counters for the kafka events
--

SELECT
    event,
    value,
    description
FROM system.events
WHERE event LIKE '%Kafka%';


-- 114. │ KafkaRebalanceAssignments                             │          1 │ Number of partition assignments (the final stage of consumer group rebalance)                                                                                                                                                                              │
-- 115. │ KafkaMessagesPolled                                   │         17 │ Number of Kafka messages polled from librdkafka to ClickHouse                                                                                                                                                                                              │
-- 116. │ KafkaMessagesRead                                     │          4 │ Number of Kafka messages already processed by ClickHouse                                                                                                                                                                                                   │
-- 117. │ KafkaMessagesFailed                                   │          3 │ Number of Kafka messages ClickHouse failed to parse                                                                                                                                                                                                        │
-- 118. │ KafkaRowsRead                                         │          3 │ Number of rows parsed from Kafka messages                                                                                                                                                                                                                  │
-- 119. │ KafkaBackgroundReads                                  │          1 │ Number of background reads populating materialized views from Kafka since server start

-- 477f3e94967e :) SELECT
--     event,
--     value,
--     description
-- FROM system.events
-- WHERE event LIKE '%Kafka%';

-- SELECT
--     event,
--     value,
--     description
-- FROM system.events
-- WHERE event LIKE '%Kafka%'

-- Query id: 8a8a547d-a541-4a82-8645-11e18bf20311

--    ┌─event─────────────────────┬─value─┬─description────────────────────────────────────────────────────────────────────────────┐
-- 1. │ KafkaRebalanceAssignments │     1 │ Number of partition assignments (the final stage of consumer group rebalance)          │
-- 2. │ KafkaMessagesPolled       │    17 │ Number of Kafka messages polled from librdkafka to ClickHouse                          │
-- 3. │ KafkaMessagesRead         │     6 │ Number of Kafka messages already processed by ClickHouse                               │
-- 4. │ KafkaMessagesFailed       │     5 │ Number of Kafka messages ClickHouse failed to parse                                    │
-- 5. │ KafkaRowsRead             │     5 │ Number of rows parsed from Kafka messages                                              │
-- 6. │ KafkaBackgroundReads      │     1 │ Number of background reads populating materialized views from Kafka since server start │
--    └───────────────────────────┴───────┴────────────────────────────────────────────────────────────────────────────────────────┘

-- 6 rows in set. Elapsed: 0.003 sec.