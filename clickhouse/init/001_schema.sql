-- ─── Raw Orders ──────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS ecommerce.orders
(
    id            String,
    customer_id   String,
    product       String,
    category      String,
    quantity      UInt32,
    price         Float64,
    total_amount  Float64,
    region        LowCardinality(String),
    status        LowCardinality(String),
    timestamp     DateTime('UTC'),
    event_time    Int64,
    processed_at  DateTime('UTC') DEFAULT now()
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, id)
TTL timestamp + INTERVAL 30 DAY
SETTINGS index_granularity = 8192;

-- ─── Orders per Minute (written by Flink 1-min tumbling window) ───────────────
CREATE TABLE IF NOT EXISTS ecommerce.orders_per_minute
(
    minute           DateTime('UTC'),
    order_count      UInt64,
    total_revenue    Float64,
    completed_count  UInt64,
    failed_count     UInt64,
    avg_order_value  Float64
)
ENGINE = ReplacingMergeTree()
ORDER BY minute
TTL minute + INTERVAL 7 DAY;

-- ─── Revenue by Region (written by Flink 5-min tumbling window) ───────────────
CREATE TABLE IF NOT EXISTS ecommerce.revenue_by_region
(
    window_start     DateTime('UTC'),
    region           LowCardinality(String),
    order_count      UInt64,
    total_revenue    Float64,
    avg_order_value  Float64
)
ENGINE = ReplacingMergeTree()
ORDER BY (window_start, region)
TTL window_start + INTERVAL 7 DAY;

-- ─── Top Products (written by Flink 5-min tumbling window) ───────────────────
CREATE TABLE IF NOT EXISTS ecommerce.top_products
(
    window_start   DateTime('UTC'),
    product        String,
    category       LowCardinality(String),
    quantity_sold  UInt64,
    total_revenue  Float64,
    order_count    UInt64
)
ENGINE = ReplacingMergeTree()
ORDER BY (window_start, product)
TTL window_start + INTERVAL 7 DAY;

-- ─── Dead Letter Queue ────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS ecommerce.dead_letter_queue
(
    id              String DEFAULT generateUUIDv4(),
    raw_message     String,
    error_reason    String,
    failed_at       DateTime('UTC') DEFAULT now(),
    kafka_topic     String DEFAULT 'orders-dlq'
)
ENGINE = MergeTree()
ORDER BY failed_at
TTL failed_at + INTERVAL 14 DAY;

-- ─── Daily Summary (populated by Dagster) ────────────────────────────────────
CREATE TABLE IF NOT EXISTS ecommerce.daily_summary
(
    date             Date,
    total_orders     UInt64,
    total_revenue    Float64,
    completed_orders UInt64,
    failed_orders    UInt64,
    avg_order_value  Float64,
    top_region       String,
    top_product      String,
    computed_at      DateTime('UTC') DEFAULT now()
)
ENGINE = ReplacingMergeTree()
ORDER BY (date)
TTL date + INTERVAL 90 DAY;

-- ─── Materialized views for fast queries ─────────────────────────────────────
CREATE MATERIALIZED VIEW IF NOT EXISTS ecommerce.orders_last_hour_mv
ENGINE = SummingMergeTree()
ORDER BY (region, status)
POPULATE
AS
SELECT
    region,
    status,
    count()       AS order_count,
    sum(total_amount) AS total_revenue
FROM ecommerce.orders
WHERE timestamp >= now() - INTERVAL 1 HOUR
GROUP BY region, status;
