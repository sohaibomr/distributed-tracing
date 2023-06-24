package main

import (
	"context"
	"fmt"
	"math/rand"

	"go.elastic.co/apm/v2"
)

type transaction struct {
	OrderID       string `json:"order_id"`
	TransactionId int    `json:"transaction_id"`
}

func newTransaction(orderId string) transaction {

	return transaction{
		OrderID:       orderId,
		TransactionId: rand.Intn(1000),
	}
}

func (t transaction) processTransaction(ctx context.Context) error {
	span, sCtx := apm.StartSpan(ctx, "processTransaction", "custom")
	span.Context.SetLabel("order_id", t.OrderID)
	span.Context.SetLabel("transaction_id", t.TransactionId)
	defer span.End()
	fmt.Printf("Processing transaction for order %s \n", t.OrderID)
	if t.OrderID == "" {
		err := fmt.Errorf("error processing transaction. Order ID is empty")
		apm.CaptureError(sCtx, err).Send()
		return err
	}
	return nil
}
