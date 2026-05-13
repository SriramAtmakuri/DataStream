"""
Real-time stream processor: Kafka → ClickHouse

Implements tumbling windows manually in Python using kafka-python:
  - 1-minute window  → orders_per_minute table
  - 5-minute window  → revenue_by_region + top_products tables
  - Raw flush every 5s / 100 msgs → orders table
  - Late data tolerance: 10 seconds
  - Bad messages → dead_letter_queue table
"""

import json
import logging
import os
import time
from collections import defaultdict
from datetime import datetime, timezone

from kafka import KafkaConsumer, KafkaProducer
import clickhouse_connect

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger(__name__)

KAFKA_BROKERS   = os.getenv("KAFKA_BROKERS",     "localhost:9092")
TOPIC_ORDERS    = os.getenv("KAFKA_TOPIC_ORDERS", "orders")
TOPIC_DLQ       = os.getenv("KAFKA_TOPIC_DLQ",   "orders-dlq")
CH_HOST         = os.getenv("CLICKHOUSE_HOST",    "localhost")
CH_PORT         = int(os.getenv("CLICKHOUSE_PORT", "8123"))
CH_DB           = os.getenv("CLICKHOUSE_DB",      "ecommerce")
CH_USER         = os.getenv("CLICKHOUSE_USER",    "default")
CH_PASS         = os.getenv("CLICKHOUSE_PASSWORD", "")

LATE_TOLERANCE_S = 10
WINDOW_1M_S      = 60
WINDOW_5M_S      = 300
REQUIRED_FIELDS  = ("id", "product", "category", "quantity", "price", "total_amount", "region", "status")


# ─── Connections ──────────────────────────────────────────────────────────────

def connect_clickhouse():
    while True:
        try:
            client = clickhouse_connect.get_client(
                host=CH_HOST, port=CH_PORT, database=CH_DB,
                username=CH_USER, password=CH_PASS,
            )
            client.command("SELECT 1")
            log.info("Connected to ClickHouse at %s:%d", CH_HOST, CH_PORT)
            return client
        except Exception as e:
            log.warning("ClickHouse not ready: %s — retry in 5s", e)
            time.sleep(5)


def connect_kafka_consumer():
    while True:
        try:
            consumer = KafkaConsumer(
                TOPIC_ORDERS,
                bootstrap_servers=KAFKA_BROKERS.split(","),
                group_id="python-stream-processor",
                auto_offset_reset="earliest",
                enable_auto_commit=True,
                consumer_timeout_ms=1000,
                value_deserializer=lambda v: v.decode("utf-8", errors="replace"),
            )
            log.info("Kafka consumer connected to %s", KAFKA_BROKERS)
            return consumer
        except Exception as e:
            log.warning("Kafka not ready: %s — retry in 5s", e)
            time.sleep(5)


def connect_kafka_producer():
    while True:
        try:
            return KafkaProducer(
                bootstrap_servers=KAFKA_BROKERS.split(","),
                value_serializer=lambda v: v.encode("utf-8"),
            )
        except Exception as e:
            log.warning("Kafka producer not ready: %s — retry in 5s", e)
            time.sleep(5)


# ─── Tumbling Window ──────────────────────────────────────────────────────────

class TumblingWindow:
    def __init__(self, size_s: int):
        self.size_s = size_s
        self._buckets: dict[int, list] = defaultdict(list)

    def add(self, event_time_ms: int, event: dict):
        bucket_ts = (event_time_ms // 1000 // self.size_s) * self.size_s
        current_bucket = (int(time.time()) // self.size_s) * self.size_s
        # Drop events older than (1 window + late tolerance)
        if bucket_ts < current_bucket - self.size_s - LATE_TOLERANCE_S:
            return
        self._buckets[bucket_ts].append(event)

    def closed_buckets(self):
        """Return and remove buckets whose window has fully elapsed (+ late tolerance)."""
        cutoff = (int(time.time()) // self.size_s) * self.size_s - LATE_TOLERANCE_S
        results = []
        for ts in sorted(self._buckets):
            if ts < cutoff:
                results.append((ts, self._buckets.pop(ts)))
        return results


# ─── Aggregations ─────────────────────────────────────────────────────────────

def agg_orders_per_minute(events: list) -> tuple:
    total   = len(events)
    revenue = sum(float(e.get("total_amount", 0)) for e in events)
    done    = sum(1 for e in events if e.get("status") == "completed")
    failed  = sum(1 for e in events if e.get("status") == "failed")
    avg     = round(revenue / total, 2) if total else 0.0
    return total, round(revenue, 2), done, failed, avg


def agg_by_region(events: list) -> dict:
    acc = defaultdict(lambda: {"count": 0, "revenue": 0.0})
    for e in events:
        r = e.get("region", "Unknown")
        acc[r]["count"]   += 1
        acc[r]["revenue"] += float(e.get("total_amount", 0))
    return acc


def agg_by_product(events: list) -> dict:
    acc = defaultdict(lambda: {"qty": 0, "revenue": 0.0, "count": 0, "category": ""})
    for e in events:
        p = e.get("product", "Unknown")
        acc[p]["qty"]      += int(e.get("quantity", 1))
        acc[p]["revenue"]  += float(e.get("total_amount", 0))
        acc[p]["count"]    += 1
        acc[p]["category"]  = e.get("category", "")
    return acc


# ─── ClickHouse writers ───────────────────────────────────────────────────────

def write_raw_batch(ch, batch: list):
    ch.insert("orders", batch, column_names=[
        "id", "customer_id", "product", "category", "quantity",
        "price", "total_amount", "region", "status", "timestamp", "event_time",
    ])
    log.info("Flushed %d raw orders", len(batch))


def write_1m_window(ch, ts: int, events: list):
    if not events:
        return
    dt = datetime.fromtimestamp(ts, tz=timezone.utc)
    total, revenue, done, failed, avg = agg_orders_per_minute(events)
    ch.insert("orders_per_minute",
              [[dt, total, revenue, done, failed, avg]],
              column_names=["minute", "order_count", "total_revenue",
                            "completed_count", "failed_count", "avg_order_value"])
    log.info("1-min window %s: %d orders $%.2f", dt.strftime("%H:%M"), total, revenue)


def write_5m_window(ch, ts: int, events: list):
    if not events:
        return
    dt = datetime.fromtimestamp(ts, tz=timezone.utc)

    for region, d in agg_by_region(events).items():
        avg = round(d["revenue"] / d["count"], 2) if d["count"] else 0
        ch.insert("revenue_by_region",
                  [[dt, region, d["count"], round(d["revenue"], 2), avg]],
                  column_names=["window_start", "region", "order_count",
                                "total_revenue", "avg_order_value"])

    for product, d in agg_by_product(events).items():
        ch.insert("top_products",
                  [[dt, product, d["category"], d["qty"], round(d["revenue"], 2), d["count"]]],
                  column_names=["window_start", "product", "category",
                                "quantity_sold", "total_revenue", "order_count"])

    log.info("5-min window %s: %d orders", dt.strftime("%H:%M"), len(events))


def write_dlq(ch, raw: str, reason: str):
    try:
        ch.insert("dead_letter_queue",
                  [[raw[:2000], reason]],
                  column_names=["raw_message", "error_reason"])
    except Exception as e:
        log.error("DLQ write failed: %s", e)


# ─── Main loop ────────────────────────────────────────────────────────────────

def main():
    ch       = connect_clickhouse()
    consumer = connect_kafka_consumer()

    win_1m = TumblingWindow(WINDOW_1M_S)
    win_5m = TumblingWindow(WINDOW_5M_S)

    raw_batch: list   = []
    last_flush: float = time.time()

    log.info("Stream processor running — topic: %s", TOPIC_ORDERS)

    while True:
        try:
            records = consumer.poll(timeout_ms=1000)
        except Exception as e:
            log.error("Kafka poll error: %s", e)
            consumer = connect_kafka_consumer()
            continue

        for partition_records in records.values():
            for msg in partition_records:
                raw = msg.value

                # ── Parse ────────────────────────────────────────────────────
                try:
                    order = json.loads(raw)
                except json.JSONDecodeError as e:
                    write_dlq(ch, raw, f"json_parse: {e}")
                    continue

                # ── Validate ─────────────────────────────────────────────────
                if not all(k in order for k in REQUIRED_FIELDS):
                    missing = [k for k in REQUIRED_FIELDS if k not in order]
                    write_dlq(ch, raw, f"missing_fields: {missing}")
                    continue

                event_time_ms = int(order.get("event_time", time.time() * 1000))

                # ── Add to windows ───────────────────────────────────────────
                win_1m.add(event_time_ms, order)
                win_5m.add(event_time_ms, order)

                # ── Buffer raw insert ────────────────────────────────────────
                ts_dt = datetime.fromtimestamp(event_time_ms / 1000, tz=timezone.utc)
                raw_batch.append([
                    order["id"],
                    order.get("customer_id", ""),
                    order["product"],
                    order["category"],
                    int(order["quantity"]),
                    float(order["price"]),
                    float(order["total_amount"]),
                    order["region"],
                    order["status"],
                    ts_dt,
                    event_time_ms,
                ])

        # ── Periodic flush ────────────────────────────────────────────────────
        now = time.time()
        if raw_batch and (len(raw_batch) >= 100 or now - last_flush >= 5):
            try:
                write_raw_batch(ch, raw_batch)
                raw_batch = []
                last_flush = now
            except Exception as e:
                log.error("Raw batch write failed: %s", e)

        # ── Flush closed windows ──────────────────────────────────────────────
        for ts, events in win_1m.closed_buckets():
            try:
                write_1m_window(ch, ts, events)
            except Exception as e:
                log.error("1m window write failed: %s", e)

        for ts, events in win_5m.closed_buckets():
            try:
                write_5m_window(ch, ts, events)
            except Exception as e:
                log.error("5m window write failed: %s", e)


if __name__ == "__main__":
    main()
