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

package webhook

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/praetorian-inc/trident/pkg/event"
	"github.com/praetorian-inc/trident/pkg/nozzle"
	"github.com/praetorian-inc/trident/pkg/parse"
	"github.com/praetorian-inc/trident/pkg/util"
)

// Server implements an HTTP server handler for handling tasks.
type Server struct {
	ip string
}

// NewWebhookServer creates a new Server.
func NewWebhookServer() (*Server, error) {
	externalIP, err := util.ExternalIP()
	if err != nil {
		log.Fatal(err)
	}
	return &Server{
		ip: externalIP,
	}, nil
}

// HealthzHandler returns an HTTP 200 ok always.
func (s *Server) HealthzHandler(w http.ResponseWriter, r *http.Request) {}

// EventHandler accepts an AuthRequest, executes the task using the nozzle
// interface and returns the AuthResponse via JSON.
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

	err = json.NewEncoder(w).Encode(&res)
	if err != nil {
		log.Printf("error writing to http response: %s", err)
	}
}
