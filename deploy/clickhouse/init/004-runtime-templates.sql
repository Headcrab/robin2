CREATE TABLE IF NOT EXISTS runtime.templates
(
    `ID` UUID DEFAULT generateUUIDv4(),
    `Name` String,
    `Body` String
)
ENGINE = ReplacingMergeTree
ORDER BY Name
SETTINGS index_granularity = 8192;
