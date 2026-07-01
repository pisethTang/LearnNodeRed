# Apache Kafka Demo — Go Producer / Consumer

A minimal Go producer/consumer example using [Apache Kafka](https://kafka.apache.org/) running in KRaft mode (no ZooKeeper required).

## Why Kafka?

Apache Kafka is the original distributed event-streaming platform. Redpanda is API-compatible with Kafka, so the same client code works on both. This demo uses Kafka directly so you can say you have used the real thing.

## Run Kafka locally

```bash
cd kafka-demo
docker compose up -d
```

This starts Kafka on `localhost:9092`.

## Produce messages

```bash
cd kafka-demo/producer
go run .
```

## Consume messages

```bash
cd kafka-demo/consumer
go run .
```

## Stop Kafka

```bash
cd kafka-demo
docker compose down
```
