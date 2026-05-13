from dagster import define_asset_job

daily_summary_job = define_asset_job(
    name="daily_summary_job",
    selection=["daily_order_summary", "product_performance"],
)

dlq_report_job = define_asset_job(
    name="dlq_report_job",
    selection=["dlq_report"],
)
