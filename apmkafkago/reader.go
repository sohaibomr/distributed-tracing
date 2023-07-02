package apmkafkago

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm/module/apmhttp/v2"
	"go.elastic.co/apm/v2"
)

const (
	transactionConsumerType = "kafka-consumer"
	spanReadMessageType     = "ReadMessage"
)

// Reader is a wrapper around kafka.Read
type Reader struct {
	R *kafka.Reader
}

// WrapReader returns a new Reader wraper around kafka.Reader
func WrapReader(r *kafka.Reader) *Reader {
	return &Reader{
		R: r,
	}
}

// ReadMessage reads a message from kafka
func (r *Reader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	msg, err := r.R.ReadMessage(ctx)
	if err != nil {
		return msg, err
	}
	// start span reding traceparent header from kafka message
	traceParent := r.getTraceparentHeader(&msg)
	traceContext, _ := apmhttp.ParseTraceparentHeader(traceParent)
	opts := apm.TransactionOptions{
		TraceContext: traceContext,
	}
	transaction := apm.DefaultTracer().StartTransactionOptions("Consume "+r.R.Config().Topic+" "+string(msg.Key), transactionConsumerType, opts)
	ctx = apm.ContextWithTransaction(ctx, transaction)
	span, _ := apm.StartSpan(ctx, "Consume "+string(msg.Key), spanReadMessageType)
	span.Context.SetLabel("topic", r.R.Config().Topic)
	span.Context.SetLabel("key", string(msg.Key))
	span.Context.SetLabel("consumer-group", string(r.R.Config().GroupID))
	defer span.End()
	defer transaction.End()
	// TODO: add panic recovery
	return msg, nil

}

// ReadMessageTx reads a message from kafka and returns apm transaction
// to be used in a transactional context
// call transaction.end at the end of processing kafka message
func (r *Reader) ReadMessageTx(ctx context.Context) (msg kafka.Message, tx *apm.Transaction, err error) {
	msg, err = r.R.ReadMessage(ctx)
	if err != nil {
		return msg, tx, err
	}
	// start span reding traceparent header from kafka message
	traceParent := r.getTraceparentHeader(&msg)
	traceContext, _ := apmhttp.ParseTraceparentHeader(traceParent)
	opts := apm.TransactionOptions{
		TraceContext: traceContext,
	}
	transaction := apm.DefaultTracer().StartTransactionOptions("Consume "+r.R.Config().Topic, transactionConsumerType, opts)
	ctx = apm.ContextWithTransaction(ctx, transaction)
	span, _ := apm.StartSpan(ctx, "Consume "+string(msg.Key), spanReadMessageType)
	span.Context.SetLabel("topic", r.R.Config().Topic)
	span.Context.SetLabel("key", string(msg.Key))
	span.Context.SetLabel("consumer-group", string(r.R.Config().GroupID))
	defer span.End()
	// TODO: add panic recovery
	return msg, transaction, nil

}

func (r *Reader) getTraceparentHeader(msg *kafka.Message) string {
	for _, header := range msg.Headers {
		if header.Key == apmhttp.W3CTraceparentHeader {
			return string(header.Value)
		}
	}
	return ""
}
