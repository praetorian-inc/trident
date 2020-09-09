package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
)

// Verifier returns a function that is used to verify the cloudflare access token
func Verifier(authDomain string, policyAUD string) func(http.Handler) http.Handler {
	certsURL := fmt.Sprintf("%s/cdn-cgi/access/certs", authDomain)

	config := &oidc.Config{
		ClientID: policyAUD,
	}

	ctx := context.TODO()
	keySet := oidc.NewRemoteKeySet(ctx, certsURL)
	verifier := oidc.NewVerifier(authDomain, keySet, config)

	return func(next http.Handler) http.Handler {
		return VerifyToken(verifier)(next)
	}
}

// VerifyToken is a middleware to verify a CF Access token
func VerifyToken(verifier *oidc.IDTokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			headers := r.Header
			// Verify the request contains an access token signed by CloudFlare
			accessJWT := headers.Get("Cf-Access-Jwt-Assertion")
			if accessJWT == "" {
				w.WriteHeader(http.StatusUnauthorized)
				_, err := w.Write([]byte("No token on the request"))
				if err != nil {
					log.Printf("error writing to http response: %s", err)
				}
				return
			}

			// Verify the access token
			ctx := r.Context()
			_, err := verifier.Verify(ctx, accessJWT)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				_, err = w.Write([]byte(fmt.Sprintf("Invalid token: %s", err.Error())))
				if err != nil {
					log.Printf("error writing to http response: %s", err)
				}
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}
