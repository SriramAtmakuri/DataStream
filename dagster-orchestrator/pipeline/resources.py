import os
from dagster import ConfigurableResource
import clickhouse_connect


class ClickHouseResource(ConfigurableResource):
    host: str = "clickhouse"
    port: int = 8123
    database: str = "ecommerce"
    username: str = "default"
    password: str = ""

    def get_client(self):
        return clickhouse_connect.get_client(
            host=self.host,
            port=self.port,
            database=self.database,
            username=self.username,
            password=self.password,
        )


def clickhouse_resource_from_env() -> ClickHouseResource:
    return ClickHouseResource(
        host=os.getenv("CLICKHOUSE_HOST", "clickhouse"),
        port=int(os.getenv("CLICKHOUSE_PORT", "8123")),
        database=os.getenv("CLICKHOUSE_DB", "ecommerce"),
        username=os.getenv("CLICKHOUSE_USER", "default"),
        password=os.getenv("CLICKHOUSE_PASSWORD", ""),
    )
