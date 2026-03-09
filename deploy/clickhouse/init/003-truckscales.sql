CREATE DATABASE IF NOT EXISTS truckscales
ENGINE = Atomic;

CREATE TABLE IF NOT EXISTS truckscales.stat
(
    `DateTime` DateTime('Asia/Almaty'),
    `InvNum` Nullable(String),
    `VagNum` UInt16,
    `Brutto` UInt32,
    `Tare` Nullable(UInt32),
    `Netto` Nullable(Int32),
    `Difference` Nullable(Int32),
    `Carrying` Nullable(UInt32),
    `Velocity` Nullable(Float32),
    `Cargotype` LowCardinality(Nullable(String))
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(DateTime)
ORDER BY (DateTime, VagNum, ifNull(InvNum, ''))
SETTINGS index_granularity = 8192;
