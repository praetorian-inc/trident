package auth

import (
	"net/http"
)

// Authenticator types define the Auth method that will modify a provided
// http.Request to include the necessary information to authenticate that
// request
type Authenticator interface {
	Auth(*http.Request) error
}
