#!/bin/bash
#
# check_protos.bash
#

echo "running check_protos.bash";

compare_files() {
	local file1="$1";
	local file2="$2";

	if [ $# -ne 2 ]; then
		echo "Usage: compare_files <file1> <file2>";
		return 1;
	fi

	if [ ! -f "${file1}" ]; then
		echo "Error: File '${file1}' not found.";
		return 1;
	fi

	if [ ! -f "${file2}" ]; then
		echo "Destination file '${file2}' does not exist. Copying from '${file1}'."
		sudo chown -R das:users ./build/containers/clickhouse/format_schemas/;
		cp "${file1}" "${file2}";
		return 0;
	fi

	echo "compare_files ${file1} ${file2}";

	diff_result=$(diff "${file1}" "${file2}"; echo $?);

	if [ "${diff_result}" -eq 0 ]; then
		echo "diff_result is zero: ${file1}:${file2}";
	else
		sudo chown -R das:users ./build/containers/clickhouse/format_schemas/;
		cp "${file1}" "${file2}";
		echo "Files differ. Copied ${file1}:${file2}";
	fi
}

file1="./proto/prometheus/v1/prometheus.proto";
file2="./build/containers/clickhouse/format_schemas/prometheus.proto";
compare_files "${file1}" "${file2}";

file1="./proto/prometheus/v1/prometheus_protolist.proto";
file2="./build/containers/clickhouse/format_schemas/prometheus_protolist.proto";
compare_files "${file1}" "${file2}";

file1="./proto/clickhouse_protolist/v1/clickhouse_protolist.proto";
file2="./build/containers/clickhouse/format_schemas/clickhouse_protolist.proto";
compare_files "${file1}" "${file2}";

# end