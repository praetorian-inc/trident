package main

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"

	"trident/pkg/worker/webhook"

	// TODO: is there a way to make this automatic for all nozzles?
	_ "trident/pkg/nozzle/okta"
)

type specification struct {
	LogLevel    string `envconfig:"LOG_LEVEL" default:"INFO"`
	Port        int    `envconfig:"PORT"`
	AccessToken []byte `envconfig:"ACCESS_TOKEN"`
}

var spec specification

func init() {
	err := envconfig.Process("worker", &spec)
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

func TokenVerifier(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Access-Token")
		if subtle.ConstantTimeCompare([]byte(token), spec.AccessToken) == 0 {
			http.Error(w, http.StatusText(403), 403)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	finish := make(chan bool)

	s, err := webhook.NewWebhookServer()
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger) // TODO: hook up to our logger
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// Insert authenication middleware to verify access token on all requests
	r.Use(TokenVerifier)

	r.Get("/healthz", s.HealthzHandler)
	r.Post("/", s.EventHandler)

	go func() {
		log.Printf("starting server on port %d", spec.Port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", spec.Port), r))
	}()

	<-finish
}
