package okta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"trident/pkg/event"
	"trident/pkg/nozzle"
)

type OktaDriver struct{}

func init() {
	nozzle.Register("okta", OktaDriver{})
}

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

func (n *OktaNozzle) Login(username, password string) (*event.AuthResponse, error) {
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

	switch resp.StatusCode {
	case 200:
		var res OktaAuthResponse
		err = json.NewDecoder(resp.Body).Decode(&res)
		if err != nil {
			return nil, err
		}

		return &event.AuthResponse{
			Valid:    res.Status != "LOCKED_OUT",
			MFA:      res.Status == "MFA_REQUIRED",
			Locked:   res.Status == "LOCKED_OUT",
			Metadata: res.Embedded,
		}, nil
	case 401:
		return &event.AuthResponse{
			Valid: false,
		}, nil
	case 429:
		return &event.AuthResponse{
			RateLimited: true,
		}, nil
	}

	return nil, fmt.Errorf("unhandled status code from okta provider: %d", resp.StatusCode)

}
