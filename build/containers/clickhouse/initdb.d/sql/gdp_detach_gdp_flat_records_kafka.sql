--
-- Detach and query gdp.gdp_flat_records_kafka
--

DETACH TABLE gdp.gdp_flat_records_kafka;

SELECT
    *
FROM
    gdp.gdp_flat_records_kafka
SETTINGS
    stream_like_engine_allow_direct_select = 1;