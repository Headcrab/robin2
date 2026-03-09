#!/bin/bash
set -uo pipefail

exec 1>>/var/log/import_script.log 2>&1

declare -A direct_patterns=(
    ["result_*.json.gz"]="runtime.history (TagName, DateTime, Value)"
    ["hs_*.json.gz"]="runtime.history (TagName, DateTime, Value)"
)

declare -A truckscales_patterns=(
    ["av_*.json.gz"]="truckscales.stat"
    ["rail_*.json.gz"]="truckscales.stat"
)

path="/var/lib/clickhouse/copyed/"
clickhouse_client=(clickhouse-client --password password123 -u admin)

import_json_each_row() {
    local file=$1
    local table=$2
    "${clickhouse_client[@]}" --query="INSERT INTO ${table} FROM INFILE '${file}' COMPRESSION 'gzip' FORMAT JSONEachRow"
}

import_truckscales_json() {
    local file=$1
    local table=$2
    /usr/local/bin/transform_truckscales_json.py "$file" | \
        "${clickhouse_client[@]}" --query="INSERT INTO ${table} FORMAT JSONEachRow"
}

process_pattern_group() {
    local importer=$1
    shift
    local -n patterns_ref=$1

    for pattern in "${!patterns_ref[@]}"; do
        mapfile -t files < <(find "$path" -maxdepth 1 -type f -name "$pattern")
        for file in "${files[@]}"; do
            if "$importer" "$file" "${patterns_ref[$pattern]}"; then
                rm "$file"
                echo "Imported and removed $file"
            else
                echo "Import failed for $file"
            fi
        done
    done
}

process_pattern_group import_json_each_row direct_patterns
process_pattern_group import_truckscales_json truckscales_patterns
