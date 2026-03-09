# ClickHouse Deployment

This directory contains the ClickHouse image used by the development stack.

## Current schema

The live container currently contains:

- `runtime.history`
- `runtime.tag` materialized view
- `runtime.max` materialized view
- `truckscales.stat`

The matching DDL is stored in `init/*.sql` and is copied into `/docker-entrypoint-initdb.d/`.
These scripts run automatically on first start when the ClickHouse data directory is empty.

## Truck scales import

`script_download.sh` supports both `av_*.json.gz` and `rail_*.json.gz`.

The dev stack mounts three source directories into the container:

- `D:\work\docker\copy_to_clickhouse` -> `/var/lib/clickhouse/copyed`
- `D:\work\docker\copy_to_clickhouse_` -> `/var/lib/clickhouse/copyed_legacy`
- `D:\work\docker\copy_to_clickhouse_from_rail` -> `/var/lib/clickhouse/copyed_truckscales`

- `result_*.json.gz` and `hs_*.json.gz` are imported as `JSONEachRow` into `runtime.history`
- `av_*.json.gz` and `rail_*.json.gz` are converted by `transform_truckscales_json.py` and inserted into `truckscales.stat`

The rail files from `D:\work\docker\copy_to_clickhouse_from_rail` are gzipped JSON arrays with string values, so they require normalization before insertion.

## Safe optimization plan

1. Stabilize the image build.
   Keep `clickhouse/clickhouse-server` pinned to a specific version and upgrade intentionally.

2. Normalize query semantics.
   Align all range queries to the same interval convention, preferably `[from, to)`, and use an explicit timezone consistently.

3. Split correctness from speed.
   Keep the interpolated path for range series, but use a direct aggregate fast-path for `avg`, `sum`, `min`, and `max` when only a scalar result is needed.

4. Remove dead or misleading SQL.
   Keep all ClickHouse queries on `runtime.history` rather than mixed legacy names, and prune templates that are no longer called.

5. Revisit storage only after measuring.
   The current `ORDER BY (TagName, DateTime)` matches the hot filters well. Changes like `LowCardinality(TagName)`, TTL, different partitions, or projections should be decided only after profiling a larger dataset.
