package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"

	"trident/pkg/db"
	"trident/pkg/parse"
	"trident/pkg/scheduler"
)

type specification struct {
	LogLocation        string `envconfig:"LOG_LOCATION" default:"/var/log/orchestrator"`
	LogPrefix          string `envconfig:"LOG_PREFIX" default:"orchestrator: "`
	AdminListenerPort  int    `envconfig:"ADMIN_LISTENING_PORT" default:"9999"`
	DBConnectionString string `envconfig:"DB_CONNECTION_STRING" required:"true"`

	ProjectID      string `envconfig:"PROJECT_ID" required:"true"`
	TopicID        string `envconfig:"TOPIC_ID" required:"true"`
	SubscriptionID string `envconfig:"SUBSCRIPTION_ID" required:"true"`

	RedisURI      string `envconfig:"REDIS_URI" required:"true"`
	RedisPassword string `envconfig:"REDIS_PASSWORD"`
}

type Server struct {
	logger *log.Logger
	db     *db.TridentDB
	sch    *scheduler.Scheduler
}

func (s *Server) CampaignHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("creating campaign")
	var c db.Campaign

	err := parse.DecodeJSONBody(w, r, &c)
	if err != nil {
		s.logger.Errorf("error parsing json: %s", err)

		var mr *parse.MalformedRequest

		if errors.As(err, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	err = s.db.InsertCampaign(&c)
	if err != nil {
		s.logger.WithFields(log.Fields{
			"campaign": c,
		}).Errorf("error inserting campaign: %s", err)
		return
	}

	go s.sch.Schedule(c)

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&c)
}

func (s *Server) ResultsHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("retrieving results for query")
	var q db.Query

	err := parse.DecodeJSONBody(w, r, &q)
	if err != nil {
		s.logger.Errorf("error parsing json: %s", err)

		var mr *parse.MalformedRequest

		if errors.As(err, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	results, err := s.db.SelectResults(q)
	if err != nil {
		message := fmt.Sprintf("there was an error collecting results from the database: %s", err)
		http.Error(w, message, http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(&results)
}

func (s *Server) HealthzHandler(w http.ResponseWriter, r *http.Request) {}

func initLogger(debug bool, logPath string) *log.Logger {
	logger := log.New()

	var logLevel = log.InfoLevel
	if debug {
		logLevel = log.DebugLevel
	}

	rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
		Filename:   logPath,
		MaxSize:    50, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
		Level:      logLevel,
		Formatter: &log.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		},
	})

	if err != nil {
		log.Fatalf("Failed to initialize file rotate hook: %v", err)
	}

	logger.SetLevel(logLevel)
	logger.SetOutput(colorable.NewColorableStdout())
	logger.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	})

	logger.AddHook(rotateFileHook)
	return logger
}

func main() {
	finish := make(chan bool)

	var spec specification

	err := envconfig.Process("orchestrator", &spec)
	if err != nil {
		log.Fatal(err)
	}

	logger := initLogger(true, spec.LogLocation+"/server.log")

	db, err := db.New(spec.DBConnectionString)
	if err != nil {
		logger.WithFields(log.Fields{
			"connectionstring": spec.DBConnectionString,
		}).Fatal(err)
	}

	sch, err := scheduler.NewScheduler(scheduler.Options{
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

	s := &Server{
		logger: logger,
		db:     db,
		sch:    sch,
	}

	s.logger.WithFields(log.Fields{
		"spec": spec,
	}).Debug("server components successfully created")

	adminAPIServer := http.NewServeMux()
	adminAPIServer.HandleFunc("/healthz", s.HealthzHandler)
	adminAPIServer.HandleFunc("/campaign", s.CampaignHandler)
	adminAPIServer.HandleFunc("/results", s.ResultsHandler)

	go func() {
		s.logger.Printf("starting server on port %d", spec.AdminListenerPort)
		s.logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", spec.AdminListenerPort), adminAPIServer))
	}()

	go func() {
		s.logger.Printf("starting scheduler task production to %s", spec.TopicID)
		sch.ProduceTasks()
	}()

	go func() {
		s.logger.Printf("starting scheduler result consumption from %s", spec.SubscriptionID)
		sch.ConsumeResults()
	}()

	<-finish
}
