// Copyright 2020 Praetorian Security, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package o365

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/time/rate"

	"github.com/praetorian-inc/trident/pkg/event"
	"github.com/praetorian-inc/trident/pkg/nozzle"
)

const (
	// FrozenUserAgent is a static user agent that we use for all requests. This
	// value is based on the UA client hint work within browsers.
	// Additional details: https://bugs.chromium.org/p/chromium/issues/detail?id=955620
	FrozenUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64)" +
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3764.0 Safari/537.36"
)

// Driver implements the nozzle.Driver interface.
type Driver struct{}

func init() {
	nozzle.Register("o365", Driver{})
}

func (Driver) New(opts map[string]string) (nozzle.Nozzle, error) {
	domain, ok := opts["domain"]
	if !ok {
		return nil, fmt.Errorf("o365 nozzle require 'domain' config parameter")
	}

	rl := rate.NewLimiter(rate.Every(300*time.Millisecond), 1)

	return &Nozzle{
		Domain:      domain,
		UserAgent:   FrozenUserAgent,
		RateLimiter: rl,
	}, nil
}

type Nozzle struct {
	// Domain is the O365 domain
	Domain string

	// UserAgent will override the Go-http-client user-agent in requests
	UserAgent string

	// RateLimiter controls how frequently we send requests to O365
	RateLimiter *rate.Limiter
}

var (
	oauth2AuthURL  = "https://%s/common/oauth2/authorize"
	oauth2TokenURL = "https://%s/common/oauth2/token"
)

func (n *Nozzle) Login(username, password string) (*event.AuthResponse, error) {
	ctx := context.Background()
	err := n.RateLimiter.Wait(ctx)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
