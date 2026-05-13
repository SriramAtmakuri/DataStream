# Real-Time E-Commerce Streaming Pipeline

A fully local, production-grade streaming data pipeline.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Docker Compose Network                               │
│                                                                             │
│  ┌──────────────┐    orders     ┌──────────────┐   orders-dlq              │
│  │  Go Service  │──────────────▶│    Kafka     │◀──────────────┐           │
│  │  (Gin API)   │               │  (bitnami)   │               │           │
│  │  :8080       │               └──────┬───────┘               │           │
│  │              │                      │ consume                │           │
│  │  - Generator │                      ▼                        │           │
│  │  - REST API  │               ┌──────────────┐               │           │
│  │  - /metrics  │               │  PyFlink     │───────────────┘           │
│  └──────┬───────┘               │  Processor   │  invalid records          │
│         │                       │              │                            │
│         │                       │  Watermarks  │                            │
│         │                       │  (10s late   │                            │
│         │                       │   tolerance) │                            │
│         │                       └──────┬───────┘                           │
│         │                              │ write                              │
│         │                              ▼                                    │
│         │                       ┌──────────────┐                           │
│         └──────────────────────▶│  ClickHouse  │                           │
│           query metrics API     │  :8123/:9000 │                           │
│                                 │              │                            │
│                                 │  orders      │                            │
│                                 │  orders_pm   │                            │
│                                 │  revenue_rgn │                            │
│                                 │  top_prods   │                            │
│                                 │  dlq         │                            │
│                                 └──────┬───────┘                           │
│                                        │                                    │
│              ┌─────────────────────────┼──────────────────┐                │
│              ▼                         ▼                   ▼                │
│       ┌────────────┐          ┌──────────────┐    ┌──────────────┐         │
│       │  Dagster   │          │   Grafana    │    │  React UI    │         │
│       │  :3000     │          │   :3001      │    │  :3002       │         │
│       │            │          │              │    │              │         │
│       │  Hourly    │          │  ClickHouse  │    │  5s polling  │         │
│       │  batch     │          │  + Prometheus│    │  Recharts    │         │
│       │  jobs      │          │  dashboards  │    │              │         │
│       └────────────┘          └──────────────┘    └──────────────┘         │
│                                        ▲                                    │
│                                ┌───────────────┐                           │
│                                │  Prometheus   │◀── Go /metrics            │
│                                │  :9090        │                            │
│                                └───────────────┘                           │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Services & Ports

| Service | URL | Description |
|---------|-----|-------------|
| Go API | http://localhost:8080 | REST API + order generator |
| Go Metrics | http://localhost:8080/metrics | Prometheus scrape endpoint |
| Kafka UI | http://localhost:8090 | Browse topics & messages |
| Flink UI | http://localhost:8081 | Job monitoring |
| Dagster | http://localhost:3000 | Pipeline orchestration |
| Grafana | http://localhost:3001 | Pre-built dashboards (admin/admin) |
| Prometheus | http://localhost:9090 | Metrics storage |
| React Dashboard | http://localhost:3002 | Custom frontend |
| ClickHouse HTTP | http://localhost:8123 | Direct SQL queries |

## Tech Stack (100% Open Source, 100% Free)

| Layer | Technology | License |
|-------|-----------|---------|
| Event Streaming | Apache Kafka (bitnami/kafka) | Apache 2.0 |
| Stream Processing | Apache Flink + PyFlink | Apache 2.0 |
| Analytics DB | ClickHouse | Apache 2.0 |
| API / Producer | Go + Gin | MIT |
| Orchestration | Dagster | Apache 2.0 |
| Monitoring | Prometheus | Apache 2.0 |
| Dashboards | Grafana OSS | AGPL-3.0 |
| Frontend | React + TypeScript + Recharts | MIT |
| Container Runtime | Docker Compose | Apache 2.0 |

## Prerequisites

- Docker Desktop ≥ 4.x with at least **8 GB RAM** allocated
- Docker Compose v2 (`docker compose` command)

## Quick Start

```bash
# 1. Clone / enter the project
cd realtime-ecommerce-pipeline

# 2. Start all services
docker compose up --build -d

# 3. Watch logs (optional)
docker compose logs -f go-producer flink-processor

# 4. Open the React dashboard
open http://localhost:3002
```

Data starts flowing immediately. The Flink processor writes aggregated windows to ClickHouse every **1 minute** (orders_per_minute) and every **5 minutes** (revenue_by_region, top_products). The React dashboard polls every 5 seconds.

## API Endpoints

### Order Generator
The Go service auto-generates ~5 orders/second (configurable via `ORDERS_PER_SECOND`).

### REST API

```
POST   /api/orders                    Create a single order manually
GET    /api/orders/stats              Generator status
GET    /api/metrics/orders-per-minute Orders per minute (last 60 min)
GET    /api/metrics/revenue-by-region Revenue grouped by region (last 1h)
GET    /api/metrics/top-products      Top 10 products by revenue (last 1h)
GET    /api/metrics/error-rate        Failed order rate (last 5 min)
GET    /health                        Health check
GET    /metrics                       Prometheus metrics
```

### Create an order manually

```bash
curl -X POST http://localhost:8080/api/orders \
  -H 'Content-Type: application/json' \
  -d '{
    "product": "Laptop Pro",
    "category": "Electronics",
    "quantity": 2,
    "price": 999.99,
    "region": "Europe",
    "status": "completed"
  }'
```

## Key Design Decisions

### Late Data & Out-of-Order Events
- The Go generator intentionally **backdates 10% of events** by up to 30 seconds to simulate real-world out-of-order delivery.
- PyFlink uses a **BoundedOutOfOrderness watermark strategy with 10s tolerance** — events arriving up to 10 seconds late are still included in the correct window.

### Dead Letter Queue
- ~5% of generated orders are deliberately malformed and routed directly to `orders-dlq` Kafka topic.
- The Flink processor also sends any unparseable JSON to the DLQ.
- DLQ messages are stored in ClickHouse `dead_letter_queue` table for analysis.
- Dagster runs a DLQ report job every 15 minutes.

### Fault Tolerance & Retries
- Flink checkpointing enabled every **30 seconds** (filesystem backend).
- Kafka producer in Go uses `MaxAttempts: 3` with automatic retry.
- Flink processor container restarts on failure (`restart: on-failure`).

### Prometheus Metrics (Go API)
| Metric | Type | Description |
|--------|------|-------------|
| `orders_published_total` | Counter | Successfully published orders |
| `orders_failed_total` | Counter | Failed publishes |
| `orders_dlq_total` | Counter | Orders sent to DLQ |
| `orders_revenue_total` | Counter | Cumulative revenue ($) |
| `kafka_publish_duration_seconds` | Histogram | Kafka publish latency |
| `http_request_duration_seconds` | Histogram | API request latency |

## ClickHouse Queries

```sql
-- Last 5 minutes orders per minute
SELECT minute, order_count, failed_count
FROM ecommerce.orders_per_minute
WHERE minute >= now() - INTERVAL 5 MINUTE
ORDER BY minute;

-- Revenue by region (last hour)
SELECT region, round(sum(total_revenue), 2) AS revenue
FROM ecommerce.revenue_by_region
WHERE window_start >= now() - INTERVAL 1 HOUR
GROUP BY region ORDER BY revenue DESC;

-- Top products (last hour)
SELECT product, sum(quantity_sold) AS qty, round(sum(total_revenue), 2) AS revenue
FROM ecommerce.top_products
WHERE window_start >= now() - INTERVAL 1 HOUR
GROUP BY product ORDER BY revenue DESC LIMIT 10;

-- Failed order rate (last 5 min)
SELECT countIf(status='failed') * 100.0 / count() AS error_pct
FROM ecommerce.orders
WHERE timestamp >= now() - INTERVAL 5 MINUTE;

-- DLQ last 24h
SELECT toStartOfHour(failed_at) AS hr, count() AS failures
FROM ecommerce.dead_letter_queue
WHERE failed_at >= now() - INTERVAL 24 HOUR
GROUP BY hr ORDER BY hr;
```

## Dagster Pipelines

Open http://localhost:3000 and navigate to **Assets** to run:

| Asset | Schedule | Description |
|-------|----------|-------------|
| `daily_order_summary` | Every hour | Writes daily KPI snapshot to ClickHouse |
| `product_performance` | Every hour | 24h product revenue aggregation |
| `dlq_report` | Every 15 min | Reports on dead-letter queue failures |

## Grafana Dashboards

Open http://localhost:3001 (admin / admin) → **E-Commerce** folder.

The pre-provisioned dashboard includes:
- Orders per minute time-series
- Revenue by region bar chart
- Top 10 products table
- Error rate stat panel
- Kafka publish rate (Prometheus)
- Cumulative revenue (Prometheus)

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `ORDERS_PER_SECOND` | `5` | Order generation rate |
| `KAFKA_BROKERS` | `kafka:9092` | Kafka bootstrap servers |
| `KAFKA_TOPIC_ORDERS` | `orders` | Main orders topic |
| `KAFKA_TOPIC_DLQ` | `orders-dlq` | Dead letter queue topic |
| `CLICKHOUSE_HOST` | `clickhouse` | ClickHouse hostname |

Copy `.env.example` to `.env` to override for local development outside Docker.

## Stopping / Cleanup

```bash
# Stop all services
docker compose down

# Stop and remove all data volumes
docker compose down -v
```

## Troubleshooting

**Flink processor keeps restarting:**
```bash
docker compose logs flink-processor
# Usually means Kafka or ClickHouse isn't ready yet — it will retry
```

**No data in React dashboard:**
- Check ClickHouse: http://localhost:8123/play (run `SELECT count() FROM ecommerce.orders`)
- Flink windows emit every 1–5 minutes; wait a moment after startup

**ClickHouse tables empty:**
```bash
docker compose logs flink-processor | grep -i "error\|exception"
```

**Port conflict:**
Edit `docker-compose.yml` and change the host port (left side of `:`).
