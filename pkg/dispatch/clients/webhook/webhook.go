package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"trident/pkg/dispatch"
	"trident/pkg/event"
)

func init() {
	dispatch.Register("webhook", WebhookDriver{})
}

type WebhookDriver struct{}

func (WebhookDriver) New(opts map[string]string) (dispatch.WorkerClient, error) {
	url, ok := opts["url"]
	if !ok {
		return nil, fmt.Errorf("webhook client requires 'url' config parameter")
	}
	token, ok := opts["token"]
	if !ok {
		return nil, fmt.Errorf("webhook client requires 'token' config parameter")
	}
	header, ok := opts["header"]
	if !ok {
		header = "X-Access-Token"
	}
	return &WebhookClient{
		URL:    url,
		Header: header,
		Token:  token,
	}, nil
}

type WebhookClient struct {
	// URL is the HTTPS URL to a worker
	URL string

	// Header is the HTTP header used to set the access token
	// Header defaults to X-Access-Token
	Header string

	// Token is an authorization token used to communicate with the worker
	Token string
}

func (w *WebhookClient) Submit(r event.AuthRequest) (*event.AuthResponse, error) {
	data, _ := json.Marshal(r)
	req, err := http.NewRequest("POST", w.URL, bytes.NewBuffer(data))
	req.Header.Set(w.Header, w.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	var res event.AuthResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	return &res, err
}
