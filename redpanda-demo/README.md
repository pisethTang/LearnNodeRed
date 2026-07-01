# Redpanda Demo — Go Producer / Consumer

A minimal Go producer/consumer example using [Redpanda](https://redpanda.com/), a Kafka-compatible streaming platform.

## Why Redpanda?

Redpanda is a drop-in replacement for Apache Kafka. It uses the same protocol, so any Kafka client works. It is simpler to run locally because it is a single binary and does not need ZooKeeper.

## Run Redpanda locally

```bash
cd redpanda-demo
docker compose up -d
```

This starts Redpanda on `localhost:19092`.

## Produce messages

In one terminal:

```bash
cd redpanda-demo/producer
go run .
```

## Consume messages

In another terminal:

```bash
cd redpanda-demo/consumer
go run .
```

The consumer joins a consumer group and reads from the `sensor-readings` topic.

## Stop Redpanda

```bash
cd redpanda-demo
docker compose down
```
