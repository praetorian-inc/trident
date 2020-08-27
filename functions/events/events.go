package events

import (
	"time"
)

type AuthRequest struct {
	// CampaignID is used to track the results of the task
	CampaignID uint `json:"campaign_id"`

	// Username is the username at the identity provider
	Username string `json:"username"`

	// Password is the password to guess against the identity provider
	Password string `json:"password"`

	// Provider is the name of identity provider, used to look up the right nozzle
	Provider string `json:"provider"`

	// ProviderMetadata is any required configuration data for the provider
	ProviderMetadata map[string]string `json:"metadata"`
}

type AuthResponse struct {
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

	// MFA will be true iff the account is known to require MFA to log in
	MFA bool `json:"mfa"`

	// RateLimited indicates the provider has detected a large number of requests
	RateLimited bool `json:"rate_limited"`

	// Additional metadata from the auth provider (e.g. information about MFA)
	Metadata map[string]interface{} `json:"metadata"`
}
