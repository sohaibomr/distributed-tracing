package main

import (
	"context"
	"fmt"
	"os"

	"github.com/segmentio/kafka-go"
	"github.com/sohaibomr/distributed-tracing/apmkafkago"
)

func main() {
	kafkaBrokers := os.Getenv("KAFKA_BROKER_URL")
	topic := os.Getenv("KAFKA_TOPIC")
	kReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaBrokers},
		Topic:   topic,
		GroupID: "report-group",
	})
	defer kReader.Close()
	reader := apmkafkago.WrapReader(kReader)
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			panic(err)
		}
		fmt.Printf("received key: %s message %s", string(msg.Key), string(msg.Value))
	}
}
