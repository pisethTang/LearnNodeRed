package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

func main() {
	conn, err := amqp091.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"sensor-readings", // queue name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		log.Fatalf("failed to declare queue: %v", err)
	}

	for i := 0; i < 20; i++ {
		bin := fmt.Sprintf("BIN_0%d", (i%5)+1)
		body := fmt.Sprintf(
			`{"sensor_id":"%s","weight_kg":%.2f,"timestamp":"%s"}`,
			bin,
			float64(i)*3.5+2.0,
			time.Now().Format(time.RFC3339),
		)

		err = ch.PublishWithContext(
			context.Background(),
			"",     // exchange
			q.Name, // routing key (queue name)
			false,  // mandatory
			false,  // immediate
			amqp091.Publishing{
				ContentType: "application/json",
				Body:        []byte(body),
			},
		)
		if err != nil {
			log.Fatalf("failed to publish message: %v", err)
		}

		fmt.Printf("published: %s\n", body)
		time.Sleep(500 * time.Millisecond)
	}
}
