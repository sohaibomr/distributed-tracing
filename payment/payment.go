package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sohaibomr/distributed-tracing/logger"
	"go.elastic.co/apm/module/apmzap/v2"
)

type Order struct {
	ID       string `json:"order_id,omitempty"`
	Quantity int    `json:"quantity,omitempty"`
	Name     string `json:"name,omitempty"`
}

func payment(w http.ResponseWriter, r *http.Request) {
	traceContextFields := apmzap.TraceContext(r.Context())
	logger.Log.With(traceContextFields...).Info("Processing payment...")
	item := Order{}
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		logger.Log.With(traceContextFields...).Error(fmt.Sprintf("Error decoding request body: %s", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	trans := newTransaction(item.ID)
	err = trans.processTransaction(r.Context())
	if err != nil {
		logger.Log.With(traceContextFields...).Error(fmt.Sprintf("Error processing transaction: %s", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)

}
