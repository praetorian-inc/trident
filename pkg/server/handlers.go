package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"trident/pkg/db"
	"trident/pkg/parse"
	"trident/pkg/scheduler"
)

type Server struct {
	DB  db.Datastore
	Sch scheduler.Scheduler
}

func (s *Server) HealthzHandler(w http.ResponseWriter, r *http.Request) {}

//TODO: figure out what to do about the fact this still works if the scheduler is nil
func (s *Server) CampaignHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("creating campaign")
	var c db.Campaign

	err := parse.DecodeJSONBody(w, r, &c)
	if err != nil {
		log.Errorf("error parsing json: %s", err)

		var mr *parse.MalformedRequest

		if errors.As(err, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	err = s.DB.InsertCampaign(&c)
	if err != nil {
		log.WithFields(log.Fields{
			"campaign": c,
		}).Errorf("error inserting campaign: %s", err)
		return
	}

	go s.Sch.Schedule(c)

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&c)
}

func (s *Server) ResultsHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("retrieving results for query")
	var q db.Query

	err := parse.DecodeJSONBody(w, r, &q)
	if err != nil {
		log.Infof("error parsing json: %s", err)

		var mr *parse.MalformedRequest

		if errors.As(err, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			log.Errorf("there was something else we don't know: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	results, err := s.DB.SelectResults(q)
	if err != nil {
		message := fmt.Sprintf("there was an error collecting results from the database: %s", err)
		log.Error(message)
		http.Error(w, message, http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(&results)
}
