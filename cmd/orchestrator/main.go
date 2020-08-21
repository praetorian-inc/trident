package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang/gddo/httputil/header"
	"github.com/kelseyhightower/envconfig"
	"github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
)

type SprayingTask struct {
	TargetURL         string
	CandidatePassword string
	UserList          []string
}

type SprayingCampaign struct {
	CreatedAt time.Time
	CreatedBy string
	TaskList  []SprayingTask
}

type specification struct {
	LogLocation       string `default:"/var/log/orchestrator"`
	LogPrefix         string `default:"orchestrator: "`
	AdminListenerPort int    `default:9999`
}

type Server struct {
	logger *log.Logger
}

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

// big ups @ajmedwards
func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			return &malformedRequest{status: http.StatusUnsupportedMediaType, msg: msg}
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return &malformedRequest{status: http.StatusRequestEntityTooLarge, msg: msg}

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		return &malformedRequest{status: http.StatusBadRequest, msg: msg}
	}

	return nil
}

func (s *Server) campaignCreate(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("creating campaign")
	var c SprayingCampaign
	err := decodeJSONBody(w, r, &c)

	if err != nil {
		s.logger.Error("there was a json parse error")
		s.logger.Println(err.Error())

		var mr *malformedRequest

		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	c.CreatedAt = time.Now()

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

	s := &Server{
		logger: initLogger(true, spec.LogLocation+"/server.log"),
	}

	adminAPIServer := http.NewServeMux()
	adminAPIServer.HandleFunc("/campaign", s.campaignCreate)

	go func() {
		s.logger.Printf("starting server on port %d", spec.AdminListenerPort)
		s.logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", spec.AdminListenerPort), adminAPIServer))
	}()

	<-finish
}
