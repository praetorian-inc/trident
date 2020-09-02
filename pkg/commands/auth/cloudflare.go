package auth

import (
	"net/url"

	"github.com/cloudflare/cloudflared/cmd/cloudflared/token"
	"github.com/cloudflare/cloudflared/logger"
)

type Authenticator interface {
	Auth() error
}

type ArgoAuthenticator struct {
	Token string
	URL   *url.URL
}

func (a *ArgoAuthenticator) Auth() error {
	logger, err := logger.New()
	token, err := token.FetchToken(a.URL, logger)
	a.Token = token
	return err
}
