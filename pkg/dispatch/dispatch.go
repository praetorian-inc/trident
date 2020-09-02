package dispatch

import (
	"context"
	"encoding/json"
	"log"

	"cloud.google.com/go/pubsub"

	"trident/pkg/event"
)

type Dispatcher struct {
	wc WorkerClient

	sub     *pubsub.Subscription
	resultc *pubsub.Topic
}

type Options struct {
	ProjectID      string
	SubscriptionID string
	ResultTopicID  string
}

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

func (d *Dispatcher) Listen(ctx context.Context) error {
	return d.sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var req event.AuthRequest
		err := json.Unmarshal(msg.Data, &req)
		if err != nil {
			log.Printf("error unmarshaling: %s", err)
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
