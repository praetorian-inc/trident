// Copyright 2020 Praetorian Security, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/go-redis/redis/v7"

	"github.com/praetorian-inc/trident/pkg/db"
)

const (
	// CacheKey is the Redis key used to store the task set.
	CacheKey  = "tasks"
	CacheKeyF = "campaign%d.tasks"
	CacheKeyR = "campaign*.tasks" // CacheKey format string for the redis Scan function
)

// Scheduler is an interface which wraps several scheduling functions together.
type Scheduler interface {
	Schedule(db.Campaign) error
	ProduceTasks()
	ConsumeResults() error
}

// PubSubScheduler implements the scheduler interface and produces/consumes to
// Google Cloud Pub/Sub.
type PubSubScheduler struct {
	db    *db.TridentDB
	cache *redis.Client
	pub   *pubsub.Topic
	sub   *pubsub.Subscription
}

// Options is used to configure a PubSubScheduler.
type Options struct {
	// Database is a pointer to the database struct.
	Database *db.TridentDB

	// ProjectID is the Google Cloud Platform project ID
	ProjectID string

	// TopicID is the Pub/Sub topic ID used by the producer to publish tasks.
	TopicID string

	// SubscriptionID is the Pub/Sub subscription used by the consumer to pull
	// task results.
	SubscriptionID string

	// RedisURI is the URI to the Redis instance (used for storing the task schedule)
	RedisURI string

	// RedisPassword is the Redis password
	RedisPassword string
}

// NewPubSubScheduler creates a PubSubScheduler given the provided Options.
// This call will attempt to ping the provided RedisURI and error if this
// connection fails.
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

/*func (s *PubSubScheduler) pushTask(task *db.Task) error {
	return s.cache.ZAdd(CacheKey, &redis.Z{
		Score:  float64(task.NotBefore.UnixNano()),
		Member: task,
	}).Err()
}*/

func (s *PubSubScheduler) pushCampaignTask(task *db.Task, campaignID uint) error {
	return s.cache.ZAdd(fmt.Sprintf(CacheKeyF, campaignID), &redis.Z{
		Score:  float64(task.NotBefore.UnixNano()),
		Member: task,
	}).Err()
}

/*func (s *PubSubScheduler) popTask(task *db.Task) error {
	z, err := s.cache.BZPopMin(5*time.Second, CacheKey).Result()
	if err != nil {
		return err
	}
	return task.UnmarshalBinary([]byte(z.Member.(string)))
}*/

// Since there are multiple per-campaign queues, popTask must
// scan through the current keys in the cache and peek each
// minimum member of the queues in order to select a candidate
func (s *PubSubScheduler) popTask(task *db.Task) error {
	// KEYS is apparently inadvisable (according to docs
	// it can be unsafe in a production environment, since it can
	// leak ~every~ key in the db, but i /think/ in our case that's
	// inconsequential), plus the alternative solution involving
	// SCAN seems overly complicated and less deterministic
	campaignKeys, err := s.cache.Keys(CacheKeyR).Result()
	if err != nil {
		return err
	}

	var minKey string
	var minNotBefore float64
	// track lowest NotBefore time across all campaigns
	for _, key := range campaignKeys {
		minTask, err := s.cache.ZRangeWithScores(key, 0, 0).Result()
		if err != nil {
			return err
		}
		// warning this code will possibly behave weird if you're trying
		// to run campaigns starting on the Unix epoch
		if minNotBefore == 0 || minNotBefore > minTask[0].Score {
			minNotBefore = minTask[0].Score
			minKey = key
		}
	}

	// per weems' suggestion, we may be able to decrease the block
	// time on this
	z, err := s.cache.BZPopMin(5*time.Second, minKey).Result()
	if err != nil {
		return err
	}
	if z.Score != minNotBefore {
		log.Printf("unexpected score retrieved by popTask: expected %f got %f", minNotBefore, z.Score)
	}
	return task.UnmarshalBinary([]byte(z.Member.(string)))
}

// Schedule accepts a campaign and computes all required tasks based on the
// provided NotBefore, NotAfter, and ScheduleInterval values. Tasks are
// scheduled by continuously adding the ScheduleInterval to a running timestamp
// (starting at the NotBefore time). Tasks which would be scheduled after the
// NotAfter time are discarded.
//
// Additionally, this scheduler prefers to schedule credential guesses for a
// single password at a time, allowing the maximum time to pass before guessing
// a given username again.
func (s *PubSubScheduler) Schedule(campaign db.Campaign) error {
	t := campaign.NotBefore
	for _, p := range campaign.Passwords {
		for _, u := range campaign.Users {
			err := s.pushCampaignTask(&db.Task{
				CampaignID:       campaign.ID,
				NotBefore:        t,
				NotAfter:         campaign.NotAfter,
				Username:         u,
				Password:         p,
				Provider:         campaign.Provider,
				ProviderMetadata: campaign.ProviderMetadata,
			}, campaign.ID)
			if err != nil {
				log.Printf("error in redis push task: %s", err)
			}
		}
		t = t.Add(campaign.ScheduleInterval)
		if t.After(campaign.NotAfter) {
			return nil
		}
	}
	return nil
}

// ProduceTasks will poll the task schedule and publish tasks to pub/sub when
// the top task is ready.
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

		if time.Until(task.NotBefore) > 5*time.Second {
			// our task was not ready, reschedule it
			err = s.pushCampaignTask(&task, task.CampaignID)
			if err != nil {
				log.Printf("error rescheduling task: %s", err)
			}
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

// ConsumeResults will stream results from pub/sub and store them in the
// database. Valid results are written directly to the database and invalid
// results are batched by the db.StreamingInsertResults function.
func (s *PubSubScheduler) ConsumeResults() error {
	ctx := context.Background()
	results := s.db.StreamingInsertResults()
	return s.sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var res db.Result
		err := json.Unmarshal(msg.Data, &res)
		if err != nil {
			log.Printf("error unmarshaling: %s", err)
			msg.Nack()
			return
		}

		if res.Valid {
			err = s.db.InsertResult(&res)
			if err != nil {
				log.Printf("error inserting result into db: %s", err)
				results <- &res
			}
		} else {
			results <- &res
		}

		// ACK only if everything else succeeded
		msg.Ack()
	})
}
