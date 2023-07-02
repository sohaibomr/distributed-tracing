package main

import (
	"fmt"
	"net/http"

	"github.com/segmentio/kafka-go"
	"github.com/sohaibomr/distributed-tracing/logger"
	"go.elastic.co/apm/module/apmzap/v2"
)

func placeNewOrderEvent(w http.ResponseWriter, r *http.Request) {
	traceContextFields := apmzap.TraceContext(r.Context())
	logger.Log.With(traceContextFields...).Info("Publishing new order event in to the kafka...")
	orderNumber := randomInt()
	order := orders[orderNumber]

	err := order.processOrder(r.Context())
	if err != nil {
		logger.Log.With(traceContextFields...).Error(fmt.Sprintf("Error processing order: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
		return
	}
	rawOrder, err := order.marshal()
	if err != nil {
		logger.Log.With(traceContextFields...).Error(fmt.Sprintf("Error marshalling order: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
		return
	}
	kMsg := kafka.Message{
		Key:   []byte(NewOrderEvent),
		Value: rawOrder,
	}
	writer.WriteMessages(r.Context(), kMsg)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 - OK"))

}
