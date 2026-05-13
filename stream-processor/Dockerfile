FROM flink:1.18.0-scala_2.12-java11

# Install Python 3 and pip
RUN apt-get update -y && \
    apt-get install -y python3 python3-pip python3-dev && \
    rm -rf /var/lib/apt/lists/*

RUN ln -sf /usr/bin/python3 /usr/bin/python

RUN pip3 install --no-cache-dir \
    clickhouse-connect==0.6.23 \
    kafka-python==2.0.2

WORKDIR /app
COPY . .
