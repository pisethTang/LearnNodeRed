package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP("localhost:19092"),
		Topic:                  "sensor-readings",
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	defer writer.Close()
	// send exactly 20 messages for now.
	for i := 0; i < 20; i++ {
		bin := fmt.Sprintf("BIN_0%d", (i%5)+1)
		value := fmt.Sprintf(
			`{"sensor_id":"%s","weight_kg":%.2f,"timestamp":"%s"}`,
			bin,
			float64(i)*3.5+2.0,
			time.Now().Format(time.RFC3339),
		)

		msg := kafka.Message{
			Key:   []byte(bin),
			Value: []byte(value),
		}

		if err := writer.WriteMessages(context.Background(), msg); err != nil {
			log.Fatalf("failed to write message: %v", err)
		}

		fmt.Printf("produced: %s\n", value)
		time.Sleep(500 * time.Millisecond)
	}
}
