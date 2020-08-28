package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trident/pkg/db"
)

type mockDB struct{}

func (m *mockDB) InsertCampaign(c *db.Campaign) error {
	return nil
}

func (m *mockDB) SelectResults(q db.Query) ([]db.Result, error) {
	var results []db.Result

	dec := json.NewDecoder(strings.NewReader("[{\"id\":18,\"created_at\":\"2020-08-28T14:02:34.844333414Z\",\"updated_at\":\"2020-08-28T14:02:34.844333414Z\",\"deleted_at\":null,\"not_before\":\"2020-08-28T00:00:00Z\",\"not_after\":\"2020-08-29T00:00:00Z\",\"schedule_interval\":500000000,\"users\":[\"anthony.weems+lockout@praetorian.com\"],\"passwords\":[\"Password0\",\"Password1\",\"Password1!\"],\"provider\":\"okta\",\"provider_metadata\":{\"domain\":\"dev-634850\"},\"results\":null}]"))

	err := dec.Decode(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (m *mockDB) InsertResult(r *db.Result) error {

	return nil
}

type mockScheduler struct{}

func (m *mockScheduler) Schedule(c db.Campaign) error {
	return nil
}

func (m *mockScheduler) ProduceTasks() {
}

func (m *mockScheduler) ConsumeResults() error {
	return nil
}

func initServer() Server {
	return Server{
		DB:  &mockDB{},
		Sch: &mockScheduler{},
	}
}

func TestHealthzHandler(t *testing.T) {
	s := initServer()

	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.HealthzHandler)

	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestCampaignHandler(t *testing.T) {
	s := initServer()

	requestBody, err := json.Marshal(map[string]interface{}{
		"not_before":        "2020-08-28T00:00:00Z",
		"not_after":         "2020-08-29T00:00:00Z",
		"schedule_interval": 500000000,
		"users":             []string{"anthony.weems+lockout@praetorian.com"},
		"passwords":         []string{"Password0", "Password1", "Password1!"},
		"provider":          "okta",
		"provider_metadata": map[string]interface{}{
			"domain": "dev-634850",
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/campaign", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.CampaignHandler)

	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestResultsHandler(t *testing.T) {
	s := initServer()
	requestBody, err := json.Marshal(map[string]interface{}{
		"ReturnedFields": []string{"*"},
		"Filter": map[string]interface{}{
			"valid": true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/results", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.ResultsHandler)

	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
