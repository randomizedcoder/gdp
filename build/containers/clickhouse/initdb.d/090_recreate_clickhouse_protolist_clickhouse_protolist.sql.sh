#!/bin/bash
# apparently /usr/bin/bash doesn't exist in the container

#
# This is the clickhouse database table creation script for xtcp2
#

set -e;

if [ "${EUID}" -ne 0 ]
then
	echo "Please run as root";
	exit 1;
fi

CLICKHOUSE_CLIENT="clickhouse-client";

DIR="/docker-entrypoint-initdb.d/";

SQL_FILE="${DIR}sql/clickhouse_protolist_clickhouse_protolist.sql";

CMD="${CLICKHOUSE_CLIENT} --time < ${SQL_FILE}";

echo "${CMD}";
eval "${CMD}";


d=$(date +date_%Y_%m_%d_%H_%M_%S);
du=$(date --utc +date_utc_%Y_%m_%d_%H_%M_%S);

echo "${d}" > "${DIR}out/date_clickhouse_protolist";
echo "${du}" > "${DIR}out/date_utc_clickhouse_protolist";

# end