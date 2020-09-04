package auth

import (
	"net/http"
	"net/url"

	"github.com/cloudflare/cloudflared/cmd/cloudflared/token"
	"github.com/cloudflare/cloudflared/logger"
)

type Authenticator interface {
	Auth(*http.Request) error
}

type ArgoAuthenticator struct {
	URL *url.URL
}

func (a *ArgoAuthenticator) Auth(req *http.Request) error {
	logger, err := logger.New()
	token, err := token.FetchToken(a.URL, logger)
	if err != nil {
		return err
	}

	req.Header.Add("cf-access-token", token)

	return nil
}
