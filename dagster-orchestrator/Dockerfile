FROM python:3.11-slim
WORKDIR /app
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
RUN mkdir -p /opt/dagster/dagster_home
COPY . .
# dagster.yaml must be at $DAGSTER_HOME so Dagster picks it up at startup
RUN cp dagster.yaml /opt/dagster/dagster_home/dagster.yaml
RUN pip install -e .
ENV DAGSTER_HOME=/opt/dagster/dagster_home
