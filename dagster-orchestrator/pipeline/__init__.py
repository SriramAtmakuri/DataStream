from dagster import Definitions

from .assets import daily_order_summary, dlq_report, product_performance
from .jobs import daily_summary_job, dlq_report_job
from .resources import ClickHouseResource, clickhouse_resource_from_env
from .schedules import dlq_check_schedule, hourly_summary_schedule

defs = Definitions(
    assets=[daily_order_summary, dlq_report, product_performance],
    jobs=[daily_summary_job, dlq_report_job],
    schedules=[hourly_summary_schedule, dlq_check_schedule],
    resources={
        "clickhouse": clickhouse_resource_from_env(),
    },
)
