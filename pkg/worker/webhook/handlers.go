package webhook

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"trident/pkg/event"
	"trident/pkg/nozzle"
	"trident/pkg/parse"
	"trident/pkg/util"
)

type Server struct {
	ip string
}

func NewWebhookServer() (*Server, error) {
	externalIP, err := util.ExternalIP()
	if err != nil {
		log.Fatal(err)
	}
	return &Server{
		ip: externalIP,
	}, nil
}

func (s *Server) HealthzHandler(w http.ResponseWriter, r *http.Request) {}

func (s *Server) EventHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("retrieving results for query")
	var req event.AuthRequest

	err := parse.DecodeJSONBody(w, r, &req)
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

	noz, err := nozzle.Open(req.Provider, req.ProviderMetadata)
	if err != nil {
		log.Errorf("error opening nozzle: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ts := time.Now()
	res, err := noz.Login(req.Username, req.Password)
	if err != nil {
		log.Errorf("error logging in to %s: %s", req.Provider, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// fill in generic AuthResult values
	res.CampaignID = req.CampaignID
	res.Username = req.Username
	res.Password = req.Password
	res.Timestamp = ts
	res.IP = s.ip

	json.NewEncoder(w).Encode(&res)
}
