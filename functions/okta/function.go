// Package p contains a Pub/Sub Cloud Function.
package p

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	// TODO: rate limit our http client
	// "golang.org/x/time/rate"
)

// PubSubMessage is the payload of a Pub/Sub event. Please refer to the docs for
// additional information regarding Pub/Sub events.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

type AuthEvent struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Metadata map[string]interface{}
}

type AuthResponse struct {
	// Valid indicates the provided credential was valid
	Valid bool

	// Locked will be true iff the account is known to be locked
	Locked bool

	// RateLimit indicates the provider has detected a large number of requests
	RateLimit bool

	// Additional metadata from the auth provider (e.g. information about MFA)
	Metadata map[string]interface{}
}

type Nozzle interface {
	Login(username, password string) (*AuthResponse, error)
}

type OktaNozzle struct {
	// Domain is the Okta subdomain (e.g. praetorianlabs)
	Domain string
}

type OktaAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type OktaAuthResponse struct {
	Status   string                 `json:"status"`
	Factor   string                 `json:"factorResult"`
	Embedded map[string]interface{} `json:"_embedded"`
}

func NewOktaNozzle(domain string) Nozzle {
	return &OktaNozzle{
		Domain: domain,
	}
}

func (n *OktaNozzle) Login(username, password string) (*AuthResponse, error) {
	url := fmt.Sprintf("https://%s.okta.com/api/v1/authn", n.Domain)
	data, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	// TODO: should we support custom user agents?
	// req.Header.Set("User-Agent", n.UserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 429 {
		return &AuthResponse{
			RateLimit: true,
		}, nil
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unhandled status code from okta provider: %d", resp.StatusCode)
	}

	var res OktaAuthResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Valid:    res.Status == "SUCCESS" || res.Factor == "SUCCESS",
		Locked:   res.Status == "LOCKED_OUT",
		Metadata: res.Embedded,
	}, nil
}

// TODO: nozzle registry with initialization from req.Metadata
var nozzle = &OktaNozzle{Domain: "dev-634850"}

// HelloPubSub consumes a Pub/Sub message.
func HelloPubSub(ctx context.Context, m PubSubMessage) error {
	var req AuthEvent
	err := json.Unmarshal(m.Data, &req)
	if err != nil {
		return err
	}
	res, err := nozzle.Login(req.Username, req.Password)
	if err != nil {
		return err
	}

	// TODO send nozzle response to pub/sub topic
	log.Printf("nozzle response: %+v", res)

	return nil
}
