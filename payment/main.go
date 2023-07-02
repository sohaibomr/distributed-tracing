package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
	"github.com/sohaibomr/distributed-tracing/apmkafkago"
	"go.elastic.co/apm/module/apmgorilla/v2"
)

func main() {
	fmt.Println("Starting server...")
	apmUrl := os.Getenv("ELASTIC_APM_SERVER_URL")
	fmt.Println(apmUrl)
	brokers := os.Getenv("KAFKA_BROKER_URL")
	kafkaTopic := os.Getenv("KAFKA_PAYMENT_TOPIC")
	writerK := &kafka.Writer{
		Addr:  kafka.TCP(brokers),
		Topic: kafkaTopic,
	}
	writer = apmkafkago.WrapWriter(writerK)
	go startOrderConsumer()
	r := mux.NewRouter()
	apmgorilla.Instrument(r)
	r.HandleFunc("/payment", payment).Methods("POST")
	log.Fatal(http.ListenAndServe(os.Getenv("PORT"), r))
}
