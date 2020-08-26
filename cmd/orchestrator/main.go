package main

import (
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
)

type specification struct {
	LogLocation        string `default:"/var/log/orchestrator"`
	LogPrefix          string `default:"orchestrator: "`
	AdminListenerPort  int    `default:"9999"`
	DBConnectionString string `required:"true"`
}

type Server struct {
	logger *log.Logger
	db     *db.TridentDB
}

func (s *Server) campaignCreate(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("creating campaign")
	var c db.Campaign

	err := parse.DecodeJSONBody(w, r, &c)
	if err != nil {
		s.logger.Error("there was a json parse error")
		s.logger.Println(err.Error())

		var mr *parse.MalformedRequest

		if errors.As(err, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	c.CreatedAt = time.Now()
	s.db.InsertCampaign(&c)
	fmt.Fprintf(w, "Campaign: %+v", c)
}

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

	db, err := db.New(spec.DBConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	s := &Server{
		logger: initLogger(true, spec.LogLocation+"/server.log"),
		db:     db,
	}

	adminAPIServer := http.NewServeMux()
	adminAPIServer.HandleFunc("/campaign", s.campaignCreate)

	go func() {
		s.logger.Printf("starting server on port %d", spec.AdminListenerPort)
		s.logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", spec.AdminListenerPort), adminAPIServer))
	}()

	<-finish
}
