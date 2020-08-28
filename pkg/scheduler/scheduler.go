package scheduler

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/go-redis/redis/v7"

	"trident/pkg/db"
)

const (
	CacheKey = "tasks"
)

type Scheduler interface {
	Schedule(db.Campaign) error
	ProduceTasks()
	ConsumeResults() error
}

type PubSubScheduler struct {
	db    *db.TridentDB
	cache *redis.Client
	pub   *pubsub.Topic
	sub   *pubsub.Subscription
}

type Options struct {
	Database       *db.TridentDB
	ProjectID      string
	TopicID        string
	SubscriptionID string
	RedisURI       string
	RedisPassword  string
}

func NewPubSubScheduler(opts Options) (*PubSubScheduler, error) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, opts.ProjectID)
	if err != nil {
		return nil, err
	}

	sub := client.SubscriptionInProject(opts.SubscriptionID, opts.ProjectID)
	sub.ReceiveSettings.Synchronous = true
	sub.ReceiveSettings.MaxOutstandingMessages = 10

	cache := redis.NewClient(&redis.Options{
		Addr:       opts.RedisURI,
		Password:   opts.RedisPassword,
		MaxRetries: 10,
		DB:         0,
	})
	_, err = cache.Ping().Result()
	if err != nil {
		return nil, err
	}

	return &PubSubScheduler{
		db:    opts.Database,
		cache: cache,
		sub:   sub,
		pub:   client.Topic(opts.TopicID),
	}, nil
}

func (s *PubSubScheduler) pushTask(task *db.Task) error {
	// TODO: do we need per-campaign queues?
	return s.cache.ZAdd(CacheKey, &redis.Z{
		Score:  float64(task.NotBefore.UnixNano()),
		Member: task,
	}).Err()
}

func (s *PubSubScheduler) popTask(task *db.Task) error {
	z, err := s.cache.BZPopMin(5*time.Second, CacheKey).Result()
	if err != nil {
		return err
	}
	return task.UnmarshalBinary([]byte(z.Member.(string)))
}

func (s *PubSubScheduler) Schedule(campaign db.Campaign) error {
	t := campaign.NotBefore
	for _, p := range campaign.Passwords {
		for _, u := range campaign.Users {
			// TODO: how do we want to handle error in task insertion?
			err := s.pushTask(&db.Task{
				CampaignID: campaign.ID,
				NotBefore:  t,
				NotAfter:   campaign.NotAfter,
				Username:   u,
				Password:   p,
				Provider:   campaign.Provider,
				// TODO: figure out jsonb -> map[string]string
				ProviderMetadata: campaign.ProviderMetadata,
			})
			if err != nil {
				log.Printf("error in redis push task: %s", err)
			}
		}
		t = t.Add(campaign.ScheduleInterval)
		if t.After(campaign.NotAfter) {
			// TODO: should this silently drops tasks that are outside the testing window
			return nil
		}
	}
	return nil
}

// ProduceTasks will publish tasks to pub/sub when ready
func (s *PubSubScheduler) ProduceTasks() {
	ctx := context.Background()
	for {
		var task db.Task
		err := s.popTask(&task)
		if err == redis.Nil {
			continue
		}
		if err != nil {
			log.Printf("error in redis pop task: %s", err)
			continue
		}

		if task.NotBefore.Sub(time.Now()) > 5*time.Second {
			// our task was not ready, reschedule it
			s.pushTask(&task)
			time.Sleep(1 * time.Second)
		} else {
			// our task was ready, run it!
			b, _ := json.Marshal(task)
			s.pub.Publish(ctx, &pubsub.Message{
				Data: b,
			})
		}
	}
}

// ConsumeResults will stream results from pub/sub and store them in the database
func (s *PubSubScheduler) ConsumeResults() error {
	ctx := context.Background()
	return s.sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var res db.Result
		err := json.Unmarshal(msg.Data, &res)
		if err != nil {
			log.Printf("error unmarshaling: %s", err)
			msg.Nack()
			return
		}

		err = s.db.InsertResult(&res)
		if err != nil {
			log.Printf("error inserting record: %s", err)
			msg.Nack()
			return
		}

		// ACK only if everything else succeeded
		msg.Ack()
	})
}
