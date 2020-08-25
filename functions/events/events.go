package events

type AuthRequest struct {
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
	// Valid indicates the provided credential was valid
	Valid bool `json:"valid"`

	// Locked will be true iff the account is known to be locked
	Locked bool `json:"locked"`

	// RateLimit indicates the provider has detected a large number of requests
	RateLimited bool `json:"rate_limited"`

	// Additional metadata from the auth provider (e.g. information about MFA)
	Metadata map[string]interface{} `json:"metadata"`
}
