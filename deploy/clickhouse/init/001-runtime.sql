CREATE DATABASE IF NOT EXISTS runtime
ENGINE = Atomic;

CREATE TABLE IF NOT EXISTS runtime.history
(
    `DateTime` DateTime64(3, 'Asia/Almaty'),
    `TagName` String,
    `Value` Float64
)
ENGINE = ReplacingMergeTree
PARTITION BY toYYYYMM(DateTime)
ORDER BY (TagName, DateTime)
SETTINGS index_granularity = 8192;
