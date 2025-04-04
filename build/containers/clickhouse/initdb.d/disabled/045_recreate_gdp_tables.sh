#!/bin/bash
#
# 050_recreate_gdp_tables.sh
#
# This is the clickhouse database table creation script for gdp
#

set -e

if [ "${EUID}" -ne 0 ]; then
    echo "Please run as root"
    exit 1
fi

CLICKHOUSE_CLIENT="clickhouse-client";
SQL_DIR="/docker-entrypoint-initdb.d/sql/gdp/";
OUT_SQL_DIR="/docker-entrypoint-initdb.d/";

execute_sql() {
    local sql_file="$1"
    local cmd="${CLICKHOUSE_CLIENT} --time < ${sql_file}"
    echo "Executing: ${cmd}"
    eval "${cmd}"
}

# Find all base table .sql files (no underscores)
find "${SQL_DIR}" -maxdepth 1 -name "*.sql" ! -name "*_*" -print0 | while IFS= read -r -d $'\0' base_sql_file; do

    # Extract the base table name
    table_name=$(basename "${base_sql_file}" .sql)

    if ! [[ "${table_name}" =~ ^(ProtobufSingle|ProtobufList) ]]; then
        echo "Skipping table: ${table_name}"
        continue
    fi

    echo "Processing table: ${table_name}"

    if [[ -f "${SQL_DIR}${table_name}.sql" ]]; then
        execute_sql "${SQL_DIR}${table_name}.sql"
    fi

    if [[ -f "${SQL_DIR}${table_name}_kafka.sql" ]]; then
        execute_sql "${SQL_DIR}${table_name}_kafka.sql"
    fi

    if [[ -f "${SQL_DIR}${table_name}_mv.sql" ]]; then
        execute_sql "${SQL_DIR}${table_name}_mv.sql"
    fi
done

d=$(date +date_%Y_%m_%d_%H_%M_%S)
du=$(date --utc +date_utc_%Y_%m_%d_%H_%M_%S)

echo "${d}" > "${OUT_SQL_DIR}out/date_gdp_create_tables"
echo "${du}" > "${OUT_SQL_DIR}out/date_utc_gdp_create_tables"

echo "GDP table creation complete"

# end
