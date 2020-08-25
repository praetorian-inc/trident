// Package p contains a Pub/Sub Cloud Function.
package p

import (
	"context"
	"encoding/json"
	"log"

	"git.praetorianlabs.com/mars/trident/functions/events"
	"git.praetorianlabs.com/mars/trident/functions/nozzle"

	_ "git.praetorianlabs.com/mars/trident/functions/nozzle/okta"


	// TODO: rate limit our http client
	// "golang.org/x/time/rate"
)

// PubSubMessage is the payload of a Pub/Sub event. Please refer to the docs for
// additional information regarding Pub/Sub events.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// HelloPubSub consumes a Pub/Sub message.
func HelloPubSub(ctx context.Context, m PubSubMessage) error {
	var req events.AuthRequest
	err := json.Unmarshal(m.Data, &req)
	if err != nil {
		return err
	}

	noz, err := nozzle.Open(req.Provider, req.ProviderMetadata)
	if err != nil {
		return err
	}

	res, err := noz.Login(req.Username, req.Password)
	if err != nil {
		return err
	}

	// TODO send nozzle response to pub/sub topic
	log.Printf("nozzle response: %+v", res)

	return nil
}
