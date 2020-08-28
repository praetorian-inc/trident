package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"

	"trident/pkg/auth"
	"trident/pkg/db"
	"trident/pkg/scheduler"
	"trident/pkg/server"
)

type specification struct {
	LogLevel           string `envconfig:"LOG_LEVEL" default:"INFO"`
	AdminListenerPort  int    `envconfig:"ADMIN_LISTENING_PORT" default:"9999"`
	DBConnectionString string `envconfig:"DB_CONNECTION_STRING" required:"true"`

	AuthDomain string `envconfig:"CF_AUTH_DOMAIN"`
	PolicyAUD string `envconfig:"CF_AUDIENCE"`

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


	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP) // TODO: use Cf-Connecting-Ip
	r.Use(middleware.Logger) // TODO: hook up to our logger
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// Insert authenication middleware to verify JWTs on all requests
	r.Use(auth.Verifier(spec.AuthDomain, spec.PolicyAUD))

	r.Get("/healthz", s.HealthzHandler)
	r.Post("/campaign", s.CampaignHandler)
	r.Post("/results", s.ResultsHandler)

	go func() {
		log.Printf("starting server on port %d", spec.AdminListenerPort)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", spec.AdminListenerPort), r))
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
