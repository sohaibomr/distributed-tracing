// consumer to process payments of new order events
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/segmentio/kafka-go"
	"github.com/sohaibomr/distributed-tracing/apmkafkago"
	"github.com/sohaibomr/distributed-tracing/logger"
	"go.elastic.co/apm/module/apmzap/v2"
	"go.elastic.co/apm/v2"
)

const NewOrderEvent = "NEW_ORDER"
const PaymentProcessedEvent = "PAYMENT_PROCESSED"

func startOrderConsumer() {
	kafkaBrokers := os.Getenv("KAFKA_BROKER_URL")
	topic := os.Getenv("KAFKA_TOPIC")
	kReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaBrokers},
		Topic:   topic,
		GroupID: "payment-group",
	})
	defer kReader.Close()
	reader := apmkafkago.WrapReader(kReader)
	for {
		msg, tx, err := reader.ReadMessageTx(context.Background())
		if err != nil {
			panic(err)
		}
		ctx := apm.ContextWithTransaction(context.Background(), tx)
		switch string(msg.Key) {
		case NewOrderEvent:
			traceContextFields := apmzap.TraceContext(ctx)
			logger.Log.With(traceContextFields...).Info("Processing payment...")
			item := Order{}
			err := json.Unmarshal(msg.Value, &item)
			if err != nil {
				logger.Log.With(traceContextFields...).Error(fmt.Sprintf("Error decoding request body: %s", err.Error()))
				apm.CaptureError(ctx, err).Send()

			}
			trans := newTransaction(item.ID)
			err = trans.processTransaction(ctx)
			if err != nil {
				logger.Log.With(traceContextFields...).Error(fmt.Sprintf("Error processing transaction: %s", err.Error()))
				apm.CaptureError(ctx, err).Send()
			}
			writer.WriteMessages(ctx, kafka.Message{
				Key:   []byte(PaymentProcessedEvent),
				Value: []byte(fmt.Sprintf(`{"id": "%s", "status": "%s"}`, item.ID, PaymentProcessedEvent)),
			})
			tx.End()
		default:
			fmt.Println("Received unknown event")
		}
		fmt.Printf("received key: %s message %s", string(msg.Key), string(msg.Value))
	}
}
