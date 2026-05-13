from dagster import ScheduleDefinition

from .jobs import daily_summary_job, dlq_report_job

# Run daily summary every hour
hourly_summary_schedule = ScheduleDefinition(
    job=daily_summary_job,
    cron_schedule="0 * * * *",  # top of every hour
    name="hourly_summary_schedule",
)

# Check DLQ every 15 minutes
dlq_check_schedule = ScheduleDefinition(
    job=dlq_report_job,
    cron_schedule="*/15 * * * *",
    name="dlq_check_schedule",
)
