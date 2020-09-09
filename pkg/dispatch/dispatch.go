package dispatch

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"cloud.google.com/go/pubsub"

	"trident/pkg/event"
)

// Dispatcher creates a data pipeline which accepts tasks, sends them to a
// worker, and publishes the result. This pipeline can be visualized as:
//  PubSub Subscription --> WorkerClient --> PubSub Topic
type Dispatcher struct {
	wc WorkerClient

	sub     *pubsub.Subscription
	resultc *pubsub.Topic
}

// Options is used to configure a Dispatcher
type Options struct {

	// ProjectID is the Google Cloud Platform project ID
	ProjectID string

	// SubscriptionID is the Pub/Sub subscription used by the dispatcher to
    // listen for incoming tasks.
	SubscriptionID string

	// ResultTopicID is the Pub/Sub topic ID used by the dispatcher to publish
    // results..
	ResultTopicID string
}

// NewDispatcher creates a dispatcher based on the provided options and worker.
func NewDispatcher(ctx context.Context, opts Options, wc WorkerClient) (*Dispatcher, error) {
	client, err := pubsub.NewClient(ctx, opts.ProjectID)
	if err != nil {
		return nil, err
	}

	sub := client.Subscription(opts.SubscriptionID)
	sub.ReceiveSettings.Synchronous = true
	sub.ReceiveSettings.MaxOutstandingMessages = 10

	return &Dispatcher{
		wc:      wc,
		sub:     sub,
		resultc: client.Topic(opts.ResultTopicID),
	}, nil
}

// Listen listens for task messages on the Pub/Sub subscription. Tasks are sent
// to the worker and results are then published to the Pub/Sub topic.
func (d *Dispatcher) Listen(ctx context.Context) error {
	return d.sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var req event.AuthRequest
		err := json.Unmarshal(msg.Data, &req)
		if err != nil {
			log.Printf("error unmarshaling: %s", err)
			msg.Ack()
			return
		}

		ts := time.Now()
		if ts.After(req.NotAfter) {
			log.Printf("received an event after end time, dropping")
			msg.Ack()
			return
		}

		resp, err := d.wc.Submit(req)
		if err != nil {
			log.Printf("error from worker: %s", err)
			msg.Nack()
			return
		}
		msg.Ack()

		b, _ := json.Marshal(resp)
		d.resultc.Publish(ctx, &pubsub.Message{
			Data: b,
		})
	})
}
