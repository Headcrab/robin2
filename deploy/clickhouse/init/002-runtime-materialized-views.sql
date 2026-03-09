CREATE MATERIALIZED VIEW IF NOT EXISTS runtime.tag
(
    `TagName` String
)
ENGINE = ReplacingMergeTree
ORDER BY TagName
SETTINGS index_granularity = 8192
AS
SELECT DISTINCT TagName
FROM runtime.history
GROUP BY TagName;

CREATE MATERIALIZED VIEW IF NOT EXISTS runtime.max
(
    `name` String,
    `max` DateTime64(3, 'Asia/Almaty')
)
ENGINE = ReplacingMergeTree
ORDER BY name
SETTINGS index_granularity = 8192
AS
SELECT
    'max' AS name,
    max(h.DateTime) AS max
FROM runtime.history AS h
LIMIT 1;
