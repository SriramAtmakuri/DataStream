from datetime import date, datetime, timezone

import pandas as pd
from dagster import AssetExecutionContext, asset

from .resources import ClickHouseResource


@asset(group_name="analytics", required_resource_keys={"clickhouse"})
def daily_order_summary(context: AssetExecutionContext) -> pd.DataFrame:
    """Compute and persist today's order summary to ClickHouse."""
    client = context.resources.clickhouse.get_client()

    df = client.query_df("""
        SELECT
            toDate(timestamp)              AS date,
            count()                        AS total_orders,
            sum(total_amount)              AS total_revenue,
            countIf(status = 'completed')  AS completed_orders,
            countIf(status = 'failed')     AS failed_orders,
            avg(total_amount)              AS avg_order_value
        FROM ecommerce.orders
        WHERE toDate(timestamp) = today()
        GROUP BY date
    """)

    if df.empty:
        context.log.warning("No orders found for today")
        return df

    # Enrich with top region and top product
    top_region_df = client.query_df("""
        SELECT region, sum(total_amount) AS rev
        FROM ecommerce.orders
        WHERE toDate(timestamp) = today()
        GROUP BY region ORDER BY rev DESC LIMIT 1
    """)
    top_product_df = client.query_df("""
        SELECT product, sum(total_amount) AS rev
        FROM ecommerce.orders
        WHERE toDate(timestamp) = today()
        GROUP BY product ORDER BY rev DESC LIMIT 1
    """)

    df["top_region"] = top_region_df["region"].iloc[0] if not top_region_df.empty else ""
    df["top_product"] = top_product_df["product"].iloc[0] if not top_product_df.empty else ""
    df["computed_at"] = datetime.now(tz=timezone.utc)

    client.insert_df("daily_summary", df[[
        "date", "total_orders", "total_revenue",
        "completed_orders", "failed_orders", "avg_order_value",
        "top_region", "top_product", "computed_at",
    ]])

    context.log.info(f"Daily summary written: {len(df)} rows")
    return df


@asset(group_name="analytics", required_resource_keys={"clickhouse"})
def dlq_report(context: AssetExecutionContext) -> pd.DataFrame:
    """Report on dead-letter queue entries from the last 24 hours."""
    client = context.resources.clickhouse.get_client()

    df = client.query_df("""
        SELECT
            toStartOfHour(failed_at) AS hour,
            count()                  AS dlq_count,
            any(error_reason)        AS sample_error
        FROM ecommerce.dead_letter_queue
        WHERE failed_at >= now() - INTERVAL 24 HOUR
        GROUP BY hour ORDER BY hour DESC
    """)

    context.log.info(f"DLQ report: {len(df)} hours, {df['dlq_count'].sum() if not df.empty else 0} total failures")
    return df


@asset(group_name="analytics", required_resource_keys={"clickhouse"})
def product_performance(context: AssetExecutionContext) -> pd.DataFrame:
    """Aggregate product performance over last 24 hours."""
    client = context.resources.clickhouse.get_client()

    df = client.query_df("""
        SELECT
            product,
            category,
            sum(quantity_sold)  AS total_qty,
            sum(total_revenue)  AS total_revenue,
            sum(order_count)    AS order_count
        FROM ecommerce.top_products
        WHERE window_start >= now() - INTERVAL 24 HOUR
        GROUP BY product, category
        ORDER BY total_revenue DESC
        LIMIT 20
    """)

    context.log.info(f"Product performance: top product = {df['product'].iloc[0] if not df.empty else 'N/A'}")
    return df
