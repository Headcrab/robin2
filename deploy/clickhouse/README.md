# ClickHouse Deployment

This directory contains the ClickHouse image used by the development stack.

## Current schema

The live container currently contains:

- `runtime.history`
- `runtime.tag` materialized view
- `runtime.max` materialized view

The matching DDL is stored in `init/*.sql` and is copied into `/docker-entrypoint-initdb.d/`.
These scripts run automatically on first start when the ClickHouse data directory is empty.

## Safe optimization plan

1. Stabilize the image build.
   Pin `clickhouse/clickhouse-server` to a specific version instead of `:head`.

2. Normalize query semantics.
   Align all range queries to the same interval convention, preferably `[from, to)`, and use an explicit timezone consistently.

3. Split correctness from speed.
   Keep the current interpolated queries for compatibility, but add separate fast-path aggregate queries for `avg`, `sum`, `min`, and `max` that read directly from `runtime.history` without `WITH FILL`.

4. Remove dead or misleading SQL.
   Review and delete unused query templates such as `get_tag_from_to_group2`, and keep all ClickHouse queries on `runtime.history` rather than mixed legacy names.

5. Fix the incomplete ingest path.
   `script_download.sh` references `truckscales.stat`, but no matching schema exists in the running container or repository. Either add its DDL or drop that import pattern.

6. Revisit storage only after measuring.
   The current `ORDER BY (TagName, DateTime)` matches the hot filters well. Changes like `LowCardinality(TagName)`, TTL, different partitions, or projections should be decided only after profiling a larger dataset.
