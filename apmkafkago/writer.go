package apmkafkago

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm/module/apmhttp/v2"
	"go.elastic.co/apm/v2"
)

const (
	transactionProducerType = "kafka-producer"
	spanWriteMessageType    = "WriteMessage"
)

// Writer is a wrapper around kafka.Writer
type Writer struct {
	W *kafka.Writer
}

// WrapWriter returns a new Writer wraper around kafka.Writer
func WrapWriter(w *kafka.Writer) *Writer {
	return &Writer{
		W: w,
	}
}

// WriteMessage writes a message to kafka
func (w *Writer) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	tx := apm.TransactionFromContext(ctx)
	if tx == nil {
		tx = apm.DefaultTracer().StartTransaction("Produce "+w.W.Topic, transactionProducerType)
		defer tx.End()
	}
	ctx = apm.ContextWithTransaction(ctx, tx)
	for i := range msgs {
		span, _ := apm.StartSpan(ctx, "Produce "+string(msgs[i].Key), spanWriteMessageType)
		span.Context.SetLabel("topic", w.W.Topic)
		span.Context.SetLabel("key", string(msgs[i].Key))
		traceParent := apmhttp.FormatTraceparentHeader(span.TraceContext())
		fmt.Println("traceparent", traceParent)
		w.addTraceparentHeader(&msgs[i], traceParent)
		span.End()
	}
	err := w.W.WriteMessages(ctx, msgs...)
	if err != nil {
		apm.CaptureError(ctx, err).Send()
	}
	// TODO: add panic recovery
	return err
}

func (w *Writer) addTraceparentHeader(msg *kafka.Message, traceParent string) {
	msg.Headers = append(msg.Headers, kafka.Header{
		Key:   apmhttp.W3CTraceparentHeader,
		Value: []byte(traceParent),
	})
}
