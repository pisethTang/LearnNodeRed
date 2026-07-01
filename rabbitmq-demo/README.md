# RabbitMQ Demo — Go Producer / Consumer

A minimal Go producer/consumer example using [RabbitMQ](https://www.rabbitmq.com/), a traditional message broker.

## Why RabbitMQ?

RabbitMQ is a message broker built around **queues**. A producer sends messages to a queue, and a consumer receives them. Once a consumer acknowledges a message, it is removed from the queue.

This is different from Redpanda/Kafka, which store messages in a persistent **topic log** that can be replayed.

## Run RabbitMQ locally

```bash
cd rabbitmq-demo
docker compose up -d
```

This starts RabbitMQ with the management UI enabled:

- AMQP port: `localhost:5672`
- Management UI: http://localhost:15672  (guest / guest)

## Produce messages

```bash
cd rabbitmq-demo/producer
go run .
```

## Consume messages

```bash
cd rabbitmq-demo/consumer
go run .
```

## Stop RabbitMQ

```bash
cd rabbitmq-demo
docker compose down
```
