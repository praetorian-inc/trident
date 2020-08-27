// Package p contains a Pub/Sub Cloud Function.
package p

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/kelseyhightower/envconfig"

	"git.praetorianlabs.com/mars/trident/functions/events"
	"git.praetorianlabs.com/mars/trident/functions/nozzle"
	"git.praetorianlabs.com/mars/trident/functions/util"

	_ "git.praetorianlabs.com/mars/trident/functions/nozzle/okta"
	// TODO: rate limit our http client
	// "golang.org/x/time/rate"
)

// PubSubMessage is the payload of a Pub/Sub event. Please refer to the docs for
// additional information regarding Pub/Sub events.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

type specification struct {
	ProjectID string `envconfig:"PROJECT_ID"`
	TopicID   string `envconfig:"TOPIC_ID"`
}

var spec specification
var pub *pubsub.Topic
var externalIP string

func init() {
	err := envconfig.Process("func", &spec)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, spec.ProjectID)
	if err != nil {
		log.Fatal(err)
	}
	pub = client.Topic(spec.TopicID)

	externalIP, err = util.ExternalIP()
	if err != nil {
		log.Fatal(err)
	}
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

	ts := time.Now()
	res, err := noz.Login(req.Username, req.Password)
	if err != nil {
		return err
	}

	// fill in generic AuthResult values
	res.CampaignID = req.CampaignID
	res.Username = req.Username
	res.Password = req.Password
	res.Timestamp = ts
	res.IP = externalIP

	b, _ := json.Marshal(res)
	pr := pub.Publish(ctx, &pubsub.Message{
		Data: b,
	})
	// NOTE: this enables us to surface errors, but will block until the publish completes
	_, err = pr.Get(ctx)
	if err != nil {
		return err
	}

	return nil
}
