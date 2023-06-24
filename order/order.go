package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"

	"github.com/sohaibomr/distributed-tracing/logger"
	"go.elastic.co/apm/module/apmhttp/v2"
	"go.elastic.co/apm/module/apmzap/v2"
	"go.elastic.co/apm/v2"
	"golang.org/x/net/context/ctxhttp"
)

// get random integer in range 1 to 8
func randomInt() int {
	min := 1
	max := 8
	return rand.Intn(max-min+1) + min
}

type Order struct {
	ID       string `json:"order_id,omitempty"`
	Quantity int    `json:"quantity,omitempty"`
	Name     string `json:"name,omitempty"`
}

// create list of orders with random item names
var orders = []Order{
	{ID: randomOrderId(), Quantity: 1, Name: "Item 1"},
	{ID: randomOrderId(), Quantity: 2, Name: "Item 2"},
	{ID: randomOrderId(), Quantity: 3, Name: "Item 3"},
	{ID: randomOrderId(), Quantity: 4, Name: "Item 4"},
	{ID: randomOrderId(), Quantity: 5, Name: "Item 5"},
	{ID: randomOrderId(), Quantity: 6, Name: "Item 6"},
	{ID: randomOrderId(), Quantity: 7, Name: "Item 7"},
	{ID: randomOrderId(), Quantity: 8, Name: "Item 8"},
	{ID: randomOrderId(), Quantity: 9, Name: "Item 9"},
}

func randomOrderId() string {
	// generate random string
	return "123456789"
}

func placeNewOrder(w http.ResponseWriter, r *http.Request) {
	traceContextFields := apmzap.TraceContext(r.Context())
	logger.Log.With(traceContextFields...).Info("Placing new order...")
	// get req args
	errArg := r.URL.Query().Get("error")
	orderNumber := randomInt()
	order := orders[orderNumber]
	if errArg == "true" {
		err := order.processOrderWithError(r.Context())
		if err != nil {
			logger.Log.With(traceContextFields...).Error(fmt.Sprintf("Error processing order: %s", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
	err := order.processOrder(r.Context())
	if err != nil {
		logger.Log.With(traceContextFields...).Error(fmt.Sprintf("Error processing order: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
		return
	}
	err = order.processPayment(r.Context())
	if err != nil {
		logger.Log.With(traceContextFields...).Error(fmt.Sprintf("Error processing payment: %v", err))
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
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(rawOrder))
}

func (o Order) marshal() ([]byte, error) {
	return json.Marshal(o)
}
func (o Order) processOrder(ctx context.Context) error {
	fmt.Println("Processing order...")
	// create new apm span with attributes
	span, _ := apm.StartSpan(ctx, "processOrder", "custom")
	span.Context.SetLabel("order.id", o.ID)
	span.Context.SetLabel("order.quantity", o.Quantity)
	span.Context.SetLabel("order.name", o.Name)
	defer span.End()

	// simulate processing time
	fmt.Println("Order processed!")
	return nil
}

func (o Order) processOrderWithError(ctx context.Context) error {
	fmt.Println("Processing order...")
	// create new span with error
	span, _ := apm.StartSpan(ctx, "processOrder", "custom")
	span.Context.SetLabel("order.id", o.ID)
	span.Context.SetLabel("order.quantity", o.Quantity)
	span.Context.SetLabel("order.name", o.Name)
	span.End()
	o.ID = ""
	return o.processPayment(ctx)
}

func (o Order) processPayment(ctx context.Context) error {
	traceContextFields := apmzap.TraceContext(ctx)
	logger.Log.With(traceContextFields...).Info("Processing payment...")
	paymentSvcUrl := fmt.Sprintf("%s/payment", os.Getenv("PAYMENT_URL"))
	// wrap http with apmoprions

	client := apmhttp.WrapClient(&http.Client{})
	rawOrder, err := o.marshal()
	if err != nil {
		logger.Log.With(traceContextFields...).Error(fmt.Sprintf("Error marshalling order: %v", err))
		return err
	}
	// make post request with body
	resp, err := ctxhttp.Post(ctx, client, paymentSvcUrl, "application/json", bytes.NewBuffer(rawOrder))
	if err != nil || resp.StatusCode != http.StatusOK {
		apm.CaptureError(ctx, fmt.Errorf("error processing payment. response: %s %v", resp.Status, err)).Send()
		err = fmt.Errorf("error processing payment. response: %s %v", resp.Status, err)
		logger.Log.With(traceContextFields...).Error(err.Error())
	}
	defer resp.Body.Close()
	fmt.Println(resp.Status)
	fmt.Println("Payment processed!")
	return err
}
