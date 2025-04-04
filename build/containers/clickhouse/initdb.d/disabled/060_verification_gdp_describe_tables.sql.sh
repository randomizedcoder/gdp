#!/bin/bash
# apparently /usr/bin/bash doesn't exist in the container

#
# This is the clickhouse database table creation script for gdp
#

set -e;

if [ "${EUID}" -ne 0 ]
then
	echo "Please run as root";
	exit 1;
fi

CLICKHOUSE_CLIENT="clickhouse-client";

DIR="/docker-entrypoint-initdb.d/";

DATE=$(cat ${DIR}out/date);

#SQL_FILE="${DIR}/sql/gdp_describe_tables.sql";

#d1="DESCRIBE TABLE gdp.gdp_flat_records_kafka TRUNCATE INTO OUTFILE ${DIR}gdp.gdp_flat_records_kafka;";
#d2="DESCRIBE TABLE gdp.gdp_flat_records TRUNCATE INTO OUTFILE ${DIR}gdp.gdp_flat_records;";

d1="'DESCRIBE TABLE gdp.Protobuf_Kafka'";
d2="'DESCRIBE TABLE gdp.Protobuf;'";

#d3="DESCRIBE TABLE gdp.gdp_flat_records_kafka TRUNCATE INTO OUTFILE ${DIR}gdp.gdp_flat_records_kafka.csv FORMAT CSV;";
#d4="DESCRIBE TABLE gdp.gdp_flat_records TRUNCATE INTO OUTFILE ${DIR}gdp.gdp_flat_records.csv FORMAT CSV;";

CMD="${CLICKHOUSE_CLIENT} --time --query ${d1} > ${DIR}out/gdp.Protobuf_Kafka_${DATE}";

echo "${CMD}";
eval "${CMD}";

CMD="${CLICKHOUSE_CLIENT} --time --query ${d2} > ${DIR}out/gdp.Protobuf_${DATE}";

echo "${CMD}";
eval "${CMD}";


d=$(date +date_%Y_%m_%d_%H_%M_%S);
du=$(date --utc +date_utc_%Y_%m_%d_%H_%M_%S);

echo "${d}" > "${DIR}out/date_describe";
echo "${du}" > "${DIR}out/date_utc_describe";

# end