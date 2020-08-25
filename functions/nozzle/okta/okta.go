package okta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"git.praetorianlabs.com/mars/trident/functions/events"
	"git.praetorianlabs.com/mars/trident/functions/nozzle"
)

func init() {
	nozzle.Register("okta", OktaDriver{})
}

type OktaDriver struct{}

func (OktaDriver) New(opts map[string]string) (nozzle.Nozzle, error) {
	domain, ok := opts["domain"]
	if !ok {
		return nil, fmt.Errorf("okta nozzle requires 'domain' config parameter")
	}
	return &OktaNozzle{
		Domain: domain,
	}, nil
}

type OktaNozzle struct {
	// Domain is the Okta subdomain
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

func NewOktaNozzle(domain string) nozzle.Nozzle {
	return &OktaNozzle{
		Domain: domain,
	}
}

func (n *OktaNozzle) Login(username, password string) (*events.AuthResponse, error) {
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
		return &events.AuthResponse{
			RateLimited: true,
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

	return &events.AuthResponse{
		Valid:    res.Status == "SUCCESS" || res.Factor == "SUCCESS",
		Locked:   res.Status == "LOCKED_OUT",
		Metadata: res.Embedded,
	}, nil
}
