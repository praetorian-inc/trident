package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"
)

type Metadata map[string]string

func (a Metadata) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *Metadata) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &a)
}

type Campaign struct {
	gorm.Model
	CreatedBy        string //take it or leave it really meh
	NotBefore        time.Time
	NotAfter         time.Time
	ScheduleInterval time.Duration
	Users            pq.StringArray `gorm:"type:varchar(255)[]"`
	Passwords        pq.StringArray `gorm:"type:varchar(255)[]"`
	Provider         string
	ProviderMetadata postgres.Jsonb
}

type Result struct {
	gorm.Model

	// TODO: will this link properly in gorm to the campaign object?
	// CampaignID is used to track the results of the task
	CampaignID uint `json:"campaign_id"`

	// IP is the originating IP of the credential guess
	IP string `json:"ip"`

	// Timestamp is the time that we made the request
	Timestamp time.Time `json:"timestamp"`

	// Username is the username at the identity provider
	Username string `json:"username"`

	// Password is the password to guess against the identity provider
	Password string `json:"password"`

	// Valid indicates the provided credential was valid
	Valid bool `json:"valid"`

	// Locked will be true iff the account is known to be locked
	Locked bool `json:"locked"`

	// MFA will be true iff the account requires MFA to log in
	MFA bool `json:"mfa"`

	// RateLimited indicates the provider has detected a large number of requests
	RateLimited bool `json:"rate_limited"`

	// TODO: does this need to be postgres.Jsonb?
	// Additional metadata from the auth provider (e.g. information about MFA)
	Metadata map[string]interface{} `json:"metadata"`
}

type Task struct {
	// NotBefore will prevent execution until this time
	NotBefore time.Time `json:"not_before"`

	// NotAfter will prevent execution after this time
	NotAfter time.Time `json:"not_after"`

	// Username is the username at the identity provider
	Username string `json:"username"`

	// Password is the password to guess against the identity provider
	Password string `json:"password"`

	// Provider is the name of identity provider, used to look up the right nozzle
	Provider string `json:"provider"`

	// ProviderMetadata is any required configuration data for the provider
	ProviderMetadata map[string]string `json:"metadata"`
}

func (t *Task) MarshalBinary() (data []byte, err error) {
	return json.Marshal(t)
}

func (t *Task) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &t)
}
