package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"

	"trident/pkg/db"
	"trident/pkg/scheduler"
	"trident/pkg/server"
)

type specification struct {
	LogLevel           string `envconfig:"LOG_LEVEL" default:"INFO"`
	AdminListenerPort  int    `envconfig:"ADMIN_LISTENING_PORT" default:"9999"`
	DBConnectionString string `envconfig:"DB_CONNECTION_STRING" required:"true"`

	ProjectID      string `envconfig:"PROJECT_ID" required:"true"`
	TopicID        string `envconfig:"TOPIC_ID" required:"true"`
	SubscriptionID string `envconfig:"SUBSCRIPTION_ID" required:"true"`

	RedisURI      string `envconfig:"REDIS_URI" required:"true"`
	RedisPassword string `envconfig:"REDIS_PASSWORD"`
}

var spec specification

func init() {
	err := envconfig.Process("orchestrator", &spec)
	if err != nil {
		log.Fatal(err)
	}

	level, err := log.ParseLevel(spec.LogLevel)
	if err != nil {
		log.Fatal(err)
	}

	log.SetLevel(level)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	})
}

func main() {
	finish := make(chan bool)

	db, err := db.New(spec.DBConnectionString)
	if err != nil {
		log.WithFields(log.Fields{
			"connectionstring": spec.DBConnectionString,
		}).Fatal(err)
	}

	sch, err := scheduler.NewPubSubScheduler(scheduler.Options{
		Database:       db,
		ProjectID:      spec.ProjectID,
		TopicID:        spec.TopicID,
		SubscriptionID: spec.SubscriptionID,
		RedisURI:       spec.RedisURI,
		RedisPassword:  spec.RedisPassword,
	})
	if err != nil {
		log.Fatal(err)
	}

	s := &server.Server{
		DB:  db,
		Sch: sch,
	}

	log.WithFields(log.Fields{
		"spec": spec,
	}).Debug("server components successfully created")

	adminAPIServer := http.NewServeMux()
	adminAPIServer.HandleFunc("/healthz", s.HealthzHandler)
	adminAPIServer.HandleFunc("/campaign", s.CampaignHandler)
	adminAPIServer.HandleFunc("/results", s.ResultsHandler)

	go func() {
		log.Printf("starting server on port %d", spec.AdminListenerPort)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", spec.AdminListenerPort), adminAPIServer))
	}()

	go func() {
		log.Printf("starting scheduler task production to %s", spec.TopicID)
		sch.ProduceTasks()
	}()

	go func() {
		log.Printf("starting scheduler result consumption from %s", spec.SubscriptionID)
		sch.ConsumeResults()
	}()

	<-finish
}
