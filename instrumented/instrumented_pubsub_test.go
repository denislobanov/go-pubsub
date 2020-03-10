package instrumented_test

import (
	"context"
	"testing"
	"time"

	pubsub "github.com/denislobanov/go-pubsub"
	"github.com/denislobanov/go-pubsub/instrumented"
	"github.com/denislobanov/go-pubsub/mockqueue"
	"github.com/prometheus/client_golang/prometheus"
)

func TestInstrumentation(t *testing.T) {
	q := mockqueue.NewMockQueue()

	var source = q
	var sink = q

	instrumentedSink := instrumented.NewMessageSink(sink, prometheus.CounterOpts{
		Help: "help_sink",
		Name: "test_sink",
	}, "test_topic")

	go func() {

		for i := 0; i < 10; i++ {
			err := instrumentedSink.PutMessage(pubsub.SimpleProducerMessage([]byte("test")))
			if err != nil {
				t.Fatalf("error publishing message: [%+v]", err)
			}
		}
	}()
	instrumentedSource := instrumented.NewMessageSource(source, prometheus.CounterOpts{
		Help: "help_source",
		Name: "test_source",
	}, "test_topic")
	consumed := make(chan pubsub.ConsumerMessage)
	count := 0
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	handler := func(m pubsub.ConsumerMessage) error {
		consumed <- m
		count++
		if count == 9 {
			cancel()
		}
		return nil
	}
	errH := func(m pubsub.ConsumerMessage, e error) error {
		panic(e)
	}
	go func() {
		err := instrumentedSource.ConsumeMessages(ctx, handler, errH)
		if err != nil {
			panic(err)
		}
		close(consumed)
	}()
	for {
		_, ok := <-consumed
		if !ok {
			break
		}
	}
}

func TestInstrumentationSameCollector(t *testing.T) {
	q := mockqueue.NewMockQueue()
	var sink = q
	var source = q

	instrumented.NewMessageSink(sink, prometheus.CounterOpts{
		Help: "help_sink",
		Name: "test_sink",
	}, "test_topic")
	instrumented.NewMessageSink(sink, prometheus.CounterOpts{
		Help: "help_sink",
		Name: "test_sink",
	}, "test_topic1")

	instrumented.NewMessageSource(source, prometheus.CounterOpts{
		Help: "help_source",
		Name: "test_source",
	}, "test_topic")
	instrumented.NewMessageSource(source, prometheus.CounterOpts{
		Help: "help_source",
		Name: "test_source",
	}, "test_topic1")
}
